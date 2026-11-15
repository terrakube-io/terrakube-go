package terrakube

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	// stepLogIncompleteHeader marks a log that holds only the output archived
	// while the step ran, because the step ended without saving its own.
	stepLogIncompleteHeader = "X-Terrakube-Output-Incomplete"
	// defaultStepLogMaxFailures is the default StepLogFollowOptions.MaxFailures.
	defaultStepLogMaxFailures = 5
	// maxStepLogRetryAfter caps the wait a server's Retry-After asks of Follow.
	maxStepLogRetryAfter = time.Minute
	// maxStepLogLine caps one line of the event stream.
	maxStepLogLine = 8 << 20
)

// StepLogService reads the logs of job steps from the /tfoutput/v1 endpoints.
// These are not JSON:API endpoints: Get returns a log as a stream of bytes,
// and Follow reads a running step's log as Server-Sent Events.
type StepLogService struct {
	client *Client
}

// StepLogOptions specifies optional parameters for StepLogService.Get.
type StepLogOptions struct {
	// Offset is the byte offset to read the log from; 0 reads the whole log.
	Offset int64
}

// StepLog is a step's log as StepLogService.Get returns it.
type StepLog struct {
	// Body is the log from the requested offset. The caller must close it.
	Body io.ReadCloser
	// Incomplete reports that the server marked the log as incomplete
	// (X-Terrakube-Output-Incomplete): it holds only the output archived
	// while the step ran, because the step ended without saving its own.
	Incomplete bool
}

// StepLogFollowOptions specifies optional parameters for StepLogService.Follow.
type StepLogFollowOptions struct {
	// LastEventID resumes the stream after this event ID, the ID of the last
	// event already received. Empty starts at the beginning of the log as the
	// server keeps it.
	LastEventID string
	// MaxFailures is how many connection attempts in a row may fail before
	// Follow gives up; 0 means 5. An attempt fails when it cannot connect, the
	// server answers 429 or 5xx, or the stream ends before it delivers any
	// output, gap or heartbeat.
	MaxFailures int
	// Backoff returns the wait before the next connection attempt, given the
	// number of failed attempts in a row (0 after a stream that delivered
	// output). A Retry-After from the server replaces it, capped at a minute.
	// nil waits about a second, doubled for each failure up to 30 seconds,
	// plus up to a second of jitter.
	Backoff func(failures int) time.Duration
	// Sleep waits for d, or until ctx is done and then returns ctx.Err().
	// nil uses a timer. Tests can replace it to not wait.
	Sleep func(ctx context.Context, d time.Duration) error
}

// StepLogEvent is one event of a followed step log: output, or a gap.
type StepLogEvent struct {
	// ID is the event's stream ID, such as "1700000000123-0"; empty for a gap.
	ID string
	// Data is the output the event carries, with lines joined by "\n";
	// empty for a gap.
	Data string
	// Gap is set for output the server could not replay; ID and Data are
	// then empty.
	Gap *StepLogGap
}

// StepLogGap describes output the stream skipped. Read the whole log with
// StepLogService.Get to recover it. A field the server did not send is empty.
type StepLogGap struct {
	// Reason is "trimmed" (the server no longer holds the output after
	// AfterID) or "history-limited" (a new viewer gets only the latest output).
	Reason string `json:"reason"`
	// AfterID is the ID the stream resumed after ("0-0" for a new viewer).
	AfterID string `json:"afterId"`
	// FirstAvailableID is the ID of the first event the stream still has.
	FirstAvailableID string `json:"firstAvailableId"`
	// Recovery names how to get the skipped output: "step-log" means Get.
	Recovery string `json:"recovery"`
}

// Get returns the log of a step. While the step runs, it is the output so
// far. With opts.Offset set it returns the log from that byte on, sending a
// Range header; when the server sends the whole log anyway (it does for a
// running step), Get skips to the offset itself. An offset at or past the end
// of the log returns an empty body.
//
// It returns a *ValidationError if an ID is empty or the offset is negative,
// and a *APIError on server errors: 404 for an unknown job or step, 403 when
// the token may not read the job, and 503 with RetryAfter set while an
// archived log is temporarily unavailable.
func (s *StepLogService) Get(ctx context.Context, orgID, jobID, stepID string, opts *StepLogOptions) (*StepLog, error) {
	p, err := stepLogPath(orgID, jobID, stepID)
	if err != nil {
		return nil, err
	}
	var offset int64
	if opts != nil {
		offset = opts.Offset
	}
	if offset < 0 {
		return nil, &ValidationError{Field: "Offset", Message: "must not be negative"}
	}
	var hdr http.Header
	if offset > 0 {
		hdr = http.Header{"Range": {"bytes=" + strconv.FormatInt(offset, 10) + "-"}}
	}

	resp, err := s.client.requestStream(ctx, http.MethodGet, p, hdr)
	if err != nil {
		if apiErr, ok := errors.AsType[*APIError](err); ok && offset > 0 && apiErr.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			return &StepLog{Body: http.NoBody}, nil
		}
		return nil, err
	}
	if offset > 0 && resp.StatusCode != http.StatusPartialContent {
		if _, err := io.CopyN(io.Discard, resp.Body, offset); err != nil && !errors.Is(err, io.EOF) {
			resp.Body.Close() //nolint:errcheck,gosec // the body is discarded
			return nil, fmt.Errorf("reading step log: %w", err)
		}
	}
	return &StepLog{
		Body:       resp.Body,
		Incomplete: len(resp.Header.Values(stepLogIncompleteHeader)) > 0,
	}, nil
}

// Follow reads a step's log as it is written and calls fn for each output
// event and each gap, in order. A gap means the stream skipped output; read
// the whole log with Get to recover it.
//
// Follow reconnects, resuming after the last output event it delivered, when
// the server ends the stream to rotate it or because the client fell behind,
// when the connection drops, and when the server answers 429 or 5xx, waiting
// for the Retry-After the server sent (capped at a minute) or for the backoff.
// Output the server sends again after a reconnect is not delivered twice.
// After opts.MaxFailures connection attempts in a row that fail, it gives up
// with an error that wraps the last failure.
//
// Follow returns nil when the server reports the job complete, ctx.Err() when
// ctx is done, and fn's error as is when fn returns one. It returns a
// *ValidationError if an ID is empty or fn is nil, and a *APIError without
// retrying on other server errors, such as 404 for an unknown job or step or
// 403 when the token may not read the job. On Terrakube 2.34.1 that 404 also
// matches ErrUnsupported (see there).
func (s *StepLogService) Follow(ctx context.Context, orgID, jobID, stepID string, opts *StepLogFollowOptions, fn func(StepLogEvent) error) error {
	p, err := stepLogPath(orgID, jobID, stepID)
	if err != nil {
		return err
	}
	if fn == nil {
		return &ValidationError{Field: "fn", Message: "must not be nil"}
	}
	var o StepLogFollowOptions
	if opts != nil {
		o = *opts
	}
	if o.MaxFailures <= 0 {
		o.MaxFailures = defaultStepLogMaxFailures
	}
	if o.Backoff == nil {
		o.Backoff = defaultStepLogBackoff
	}
	if o.Sleep == nil {
		o.Sleep = sleepContext
	}

	lastID := o.LastEventID
	failures := 0
	for {
		res, err := s.followOnce(ctx, p+"/stream", &lastID, fn)
		if err != nil {
			return err
		}
		if res.complete {
			return nil
		}
		if res.healthy {
			failures = 0
		} else {
			failures++
			if failures >= o.MaxFailures {
				return fmt.Errorf("terrakube: following step log: gave up after %d failed connection attempts: %w", failures, res.err)
			}
		}
		wait := o.Backoff(failures)
		if res.retryAfter > 0 {
			wait = min(res.retryAfter, maxStepLogRetryAfter)
		}
		if err := o.Sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// followResult is how one connection of Follow ended.
type followResult struct {
	complete   bool          // the server reported the job complete
	healthy    bool          // the stream delivered output, a gap or a heartbeat
	retryAfter time.Duration // the wait the server asked for
	err        error         // why the stream ended, when it was not complete
}

// followOnce reads the stream at p from one connection, resuming after
// *lastID, and keeps *lastID at the ID of the last output event delivered. It
// returns an error only when Follow must stop: fn failed, ctx is done, or the
// server refused the request in a way a retry does not fix.
func (s *StepLogService) followOnce(ctx context.Context, p string, lastID *string, fn func(StepLogEvent) error) (followResult, error) {
	hdr := http.Header{"Accept": {"text/event-stream"}}
	if *lastID != "" {
		hdr.Set("Last-Event-ID", *lastID)
	}
	resp, err := s.client.requestStream(ctx, http.MethodGet, p, hdr)
	if err != nil {
		if ctx.Err() != nil {
			return followResult{}, ctx.Err()
		}
		if apiErr, ok := errors.AsType[*APIError](err); ok {
			if apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500 {
				return followResult{retryAfter: apiErr.RetryAfter, err: err}, nil
			}
			return followResult{}, err
		}
		if errors.Is(err, ErrForeignHost) {
			return followResult{}, err
		}
		return followResult{err: err}, nil
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are inconsequential

	res := followResult{err: errors.New("stream ended without an end event")}
	var ev sseEvent
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(nil, maxStepLogLine)
	sc.Split(scanSSELines)
	for sc.Scan() {
		line := sc.Text()
		if line != "" {
			if strings.HasPrefix(line, ":") { // a comment: the server's heartbeat
				res.healthy = true
				continue
			}
			field, value, _ := strings.Cut(line, ":")
			value = strings.TrimPrefix(value, " ")
			switch field {
			case "event":
				ev.kind = value
			case "data":
				ev.data = append(ev.data, value)
			case "id":
				if !strings.ContainsRune(value, 0) {
					ev.id, ev.hasID = value, true
				}
			}
			continue
		}

		// A blank line ends the event; one without data is not dispatched.
		cur := ev
		ev = sseEvent{}
		if cur.data == nil {
			continue
		}
		text := strings.Join(cur.data, "\n")
		switch cur.kind {
		case "", "message":
			if cur.hasID && !streamIDAfter(cur.id, *lastID) {
				continue // already delivered before a reconnect
			}
			if cur.hasID {
				*lastID = cur.id
			}
			res.healthy = true
			if err := fn(StepLogEvent{ID: cur.id, Data: text}); err != nil {
				return res, err
			}
		case "gap":
			gap := &StepLogGap{}
			_ = json.Unmarshal([]byte(text), gap) // a malformed gap still tells the caller to recover
			res.healthy = true
			if err := fn(StepLogEvent{Gap: gap}); err != nil {
				return res, err
			}
		case "end":
			var end struct {
				Reason string `json:"reason"`
			}
			_ = json.Unmarshal([]byte(text), &end) // an unknown reason reconnects
			if end.Reason == "complete" {
				res.complete = true
				return res, nil
			}
			res.err = fmt.Errorf("server ended the stream (%q)", end.Reason)
			return res, nil
		}
	}
	if ctx.Err() != nil {
		return res, ctx.Err()
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return res, fmt.Errorf("terrakube: step log stream line longer than %d bytes", maxStepLogLine)
		}
		res.err = err
	}
	return res, nil
}

// sseEvent collects the fields of one event of an event stream.
type sseEvent struct {
	kind  string   // the event field; empty for a plain message
	id    string   // the id field
	hasID bool     // whether the event had an id field
	data  []string // the data fields, one per line
}

// scanSSELines is a bufio.SplitFunc for event stream lines, which end in
// "\r\n", "\n" or "\r". An unterminated last line is dropped: the event it
// belongs to is incomplete.
func scanSSELines(data []byte, atEOF bool) (int, []byte, error) {
	i := bytes.IndexAny(data, "\r\n")
	if i < 0 {
		return 0, nil, nil
	}
	if data[i] == '\r' {
		if i+1 == len(data) && !atEOF {
			return 0, nil, nil // "\r\n" may continue in the next read
		}
		if i+1 < len(data) && data[i+1] == '\n' {
			return i + 2, data[:i], nil
		}
	}
	return i + 1, data[:i], nil
}

// streamIDAfter reports whether the stream ID id ("1700000000123-0") comes
// after last. When either is not a stream ID it reports true, so an event is
// never dropped on a guess.
func streamIDAfter(id, last string) bool {
	a, okA := parseStreamID(id)
	b, okB := parseStreamID(last)
	if !okA || !okB {
		return true
	}
	return a[0] > b[0] || (a[0] == b[0] && a[1] > b[1])
}

// parseStreamID splits a Redis stream ID "ms-seq" (or "ms") into numbers.
func parseStreamID(s string) ([2]uint64, bool) {
	msPart, seqPart, hasSeq := strings.Cut(s, "-")
	ms, err := strconv.ParseUint(msPart, 10, 64)
	if err != nil {
		return [2]uint64{}, false
	}
	var seq uint64
	if hasSeq {
		if seq, err = strconv.ParseUint(seqPart, 10, 64); err != nil {
			return [2]uint64{}, false
		}
	}
	return [2]uint64{ms, seq}, true
}

// defaultStepLogBackoff waits a second, doubled for each failure in a row up
// to 30 seconds, plus up to a second of jitter so viewers do not reconnect at
// once.
func defaultStepLogBackoff(failures int) time.Duration {
	d := 30 * time.Second
	if failures < 5 {
		d = time.Second << failures
	}
	return d + rand.N(time.Second) //nolint:gosec // jitter needs no secure randomness
}

// sleepContext waits for d or until ctx is done.
func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// stepLogPath validates the IDs and returns the path of a step's log.
func stepLogPath(orgID, jobID, stepID string) (string, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return "", err
	}
	if err := validateID("jobID", jobID); err != nil {
		return "", err
	}
	if err := validateID("stepID", stepID); err != nil {
		return "", err
	}
	return path.Join("/tfoutput/v1/organization", orgID, "job", jobID, "step", stepID), nil
}
