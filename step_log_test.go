package terrakube_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

const (
	stepLogPath    = "/tfoutput/v1/organization/org-1/job/job-1/step/step-1"
	stepStreamPath = stepLogPath + "/stream"
)

func TestStepLogService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepLogPath, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Range"); got != "" {
			t.Errorf("Range = %q, want none", got)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = io.WriteString(w, "Initializing...\nDone\n")
	})

	log, err := newTestClient(t, srv).StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.Incomplete {
		t.Error("Incomplete = true, want false")
	}
	if got := readAllClose(t, log.Body); got != "Initializing...\nDone\n" {
		t.Errorf("body = %q", got)
	}
}

func TestStepLogService_Get_Range(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepLogPath, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Range"); got != "bytes=5-" {
			t.Errorf("Range = %q, want bytes=5-", got)
		}
		w.Header().Set("Content-Range", "bytes 5-9/10")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "56789")
	})

	log, err := newTestClient(t, srv).StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", &terrakube.StepLogOptions{Offset: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readAllClose(t, log.Body); got != "56789" {
		t.Errorf("body = %q", got)
	}
}

// A running step's live log is sent whole, without honouring Range.
func TestStepLogService_Get_RangeIgnored(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepLogPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "0123456789")
	})
	client := newTestClient(t, srv)

	log, err := client.StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", &terrakube.StepLogOptions{Offset: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readAllClose(t, log.Body); got != "456789" {
		t.Errorf("body = %q, want the log from the offset", got)
	}

	log, err = client.StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", &terrakube.StepLogOptions{Offset: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readAllClose(t, log.Body); got != "" {
		t.Errorf("body = %q, want empty past the end", got)
	}
}

func TestStepLogService_Get_RangeNotSatisfiable(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepLogPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Range", "bytes */10")
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
	})

	log, err := newTestClient(t, srv).StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", &terrakube.StepLogOptions{Offset: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readAllClose(t, log.Body); got != "" {
		t.Errorf("body = %q, want empty", got)
	}
}

func TestStepLogService_Get_Incomplete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepLogPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Terrakube-Output-Incomplete", "true")
		_, _ = io.WriteString(w, "partial")
	})

	log, err := newTestClient(t, srv).StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !log.Incomplete {
		t.Error("Incomplete = false, want true")
	}
	if got := readAllClose(t, log.Body); got != "partial" {
		t.Errorf("body = %q", got)
	}
}

func TestStepLogService_Get_Errors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /tfoutput/v1/organization/org-1/job/job-1/step/missing", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound) // the controller's answer for an unknown job or step
	})
	srv.HandleFunc("GET /tfoutput/v1/organization/org-1/job/job-1/step/denied", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	srv.HandleFunc("GET /tfoutput/v1/organization/org-1/job/job-1/step/busy", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	srv.HandleFunc("GET /tfoutput/v1/organization/org-1/job/job-1/step/old", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // Spring's answer for a route it does not have
		_, _ = io.WriteString(w, `{"timestamp":"2026-11-15T10:00:00Z","status":404,"error":"Not Found","path":"/x"}`)
	})
	client := newTestClient(t, srv)
	ctx := context.Background()

	_, err := client.StepLogs.Get(ctx, "org-1", "job-1", "missing", nil)
	if !terrakube.IsNotFound(err) || errors.Is(err, terrakube.ErrUnsupported) {
		t.Errorf("missing step: err = %v, want a 404 that is not ErrUnsupported", err)
	}
	_, err = client.StepLogs.Get(ctx, "org-1", "job-1", "denied", nil)
	if !terrakube.IsForbidden(err) {
		t.Errorf("denied: err = %v, want 403", err)
	}
	_, err = client.StepLogs.Get(ctx, "org-1", "job-1", "busy", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok || apiErr.StatusCode != http.StatusServiceUnavailable || apiErr.RetryAfter != 5*time.Second {
		t.Errorf("busy: err = %#v, want a 503 with RetryAfter 5s", err)
	}
	_, err = client.StepLogs.Get(ctx, "org-1", "job-1", "old", nil)
	if !errors.Is(err, terrakube.ErrUnsupported) {
		t.Errorf("old server: err = %v, want ErrUnsupported", err)
	}

	_, err = client.StepLogs.Get(ctx, "", "job-1", "step-1", nil)
	assertValidationError(t, err, "organizationID")
	_, err = client.StepLogs.Get(ctx, "org-1", "", "step-1", nil)
	assertValidationError(t, err, "jobID")
	_, err = client.StepLogs.Get(ctx, "org-1", "job-1", "", nil)
	assertValidationError(t, err, "stepID")
	_, err = client.StepLogs.Get(ctx, "org-1", "job-1", "step-1", &terrakube.StepLogOptions{Offset: -1})
	if ve, ok := errors.AsType[*terrakube.ValidationError](err); !ok || ve.Field != "Offset" {
		t.Errorf("negative offset: err = %v, want a *ValidationError for Offset", err)
	}
}

func TestAPIError_RetryAfterDate(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepLogPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	_, err := newTestClient(t, srv).StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok || apiErr.RetryAfter < 58*time.Minute || apiErr.RetryAfter > time.Hour {
		t.Errorf("err = %#v, want RetryAfter of about an hour", err)
	}
}

func TestAPIError_RetryAfterInvalid(t *testing.T) {
	t.Parallel()

	for _, v := range []string{"soon", "-5", "0", "99999999999999999999"} {
		srv := testutil.NewServer(t)
		srv.HandleFunc("GET "+stepLogPath, status(http.StatusServiceUnavailable, v))

		_, err := newTestClient(t, srv).StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", nil)
		if apiErr, ok := errors.AsType[*terrakube.APIError](err); !ok || apiErr.RetryAfter != 0 {
			t.Errorf("Retry-After %q: err = %#v, want RetryAfter 0", v, err)
		}
	}
}

func TestStepLogService_Get_ReadErrorWhileSkipping(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepLogPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = io.WriteString(w, "short") // the connection closes before the promised length
	})

	_, err := newTestClient(t, srv).StepLogs.Get(context.Background(), "org-1", "job-1", "step-1", &terrakube.StepLogOptions{Offset: 50})
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("err = %v, want io.ErrUnexpectedEOF", err)
	}
}

// sseServer answers the step log stream with the handlers in order, one per
// connection, and records the Last-Event-ID header of each request.
type sseServer struct {
	t        *testing.T
	handlers []http.HandlerFunc
	mu       sync.Mutex
	ids      []string // Last-Event-ID of each request, "<none>" when absent
}

func newSSEServer(t *testing.T, handlers ...http.HandlerFunc) (*sseServer, *terrakube.Client) {
	t.Helper()
	s := &sseServer{t: t, handlers: handlers}
	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stepStreamPath, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Errorf("Accept = %q", got)
		}
		s.mu.Lock()
		id := "<none>"
		if v, ok := r.Header["Last-Event-Id"]; ok {
			id = v[0]
		}
		s.ids = append(s.ids, id)
		n := len(s.ids)
		s.mu.Unlock()
		if n > len(s.handlers) {
			t.Errorf("unexpected connection %d", n)
			w.WriteHeader(http.StatusGone)
			return
		}
		s.handlers[n-1](w, r)
	})
	return s, newTestClient(t, srv)
}

func (s *sseServer) lastEventIDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.ids...)
}

// sse returns a handler that writes body as an event stream and returns.
func sse(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, body)
	}
}

func status(code int, retryAfter string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if retryAfter != "" {
			w.Header().Set("Retry-After", retryAfter)
		}
		w.WriteHeader(code)
	}
}

// fastFollow returns options that never wait and record each wait.
func fastFollow(waits *[]time.Duration) *terrakube.StepLogFollowOptions {
	return &terrakube.StepLogFollowOptions{
		Backoff: func(int) time.Duration { return time.Millisecond },
		Sleep: func(_ context.Context, d time.Duration) error {
			if waits != nil {
				*waits = append(*waits, d)
			}
			return nil
		},
	}
}

func collect(events *[]terrakube.StepLogEvent) func(terrakube.StepLogEvent) error {
	return func(e terrakube.StepLogEvent) error {
		*events = append(*events, e)
		return nil
	}
}

func dataOf(events []terrakube.StepLogEvent) string {
	var b strings.Builder
	for _, e := range events {
		if e.Gap != nil {
			b.WriteString("[gap]")
			continue
		}
		b.WriteString(e.ID + "=" + e.Data + ";")
	}
	return b.String()
}

func TestStepLogService_Follow_Events(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t, sse(""+
		": heartbeat\n\n"+
		"event: gap\n"+
		`data: {"reason":"history-limited","afterId":"0-0","firstAvailableId":"5-0","recovery":"step-log"}`+"\n\n"+
		"id:5-0\ndata:plain\n\n"+
		"id: 6-0\ndata: line one\ndata: line two\n\n"+
		": heartbeat\n\n"+
		"id: 7-0\r\ndata: crlf\r\n\r\n"+
		"id: 8-0\ndata:\n\n"+
		"id: 9-0\rdata: cr\r\r"+
		"id: 9-x\ndata: odd id\n\n"+
		"event: unknown\ndata: ignored\n\n"+
		"retry: 1000\n\n"+
		"event: end\ndata: {\"reason\":\"complete\"}\n\n"))

	var events []terrakube.StepLogEvent
	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), collect(&events))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) == 0 || events[0].Gap == nil {
		t.Fatalf("events = %+v, want a gap first", events)
	}
	want := terrakube.StepLogGap{Reason: "history-limited", AfterID: "0-0", FirstAvailableID: "5-0", Recovery: "step-log"}
	if *events[0].Gap != want {
		t.Errorf("gap = %+v, want %+v", *events[0].Gap, want)
	}
	if got := dataOf(events); got != "[gap]5-0=plain;6-0=line one\nline two;7-0=crlf;8-0=;9-0=cr;9-x=odd id;" {
		t.Errorf("events = %q", got)
	}
	if ids := srv.lastEventIDs(); len(ids) != 1 || ids[0] != "<none>" {
		t.Errorf("Last-Event-ID per request = %q, want one request without it", ids)
	}
}

func TestStepLogService_Follow_LastEventID(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t, sse("id: 10-1\ndata: x\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"))

	opts := fastFollow(nil)
	opts.LastEventID = "10-0"
	var events []terrakube.StepLogEvent
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", opts, collect(&events)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ids := srv.lastEventIDs(); len(ids) != 1 || ids[0] != "10-0" {
		t.Errorf("Last-Event-ID per request = %q", ids)
	}
	if got := dataOf(events); got != "10-1=x;" {
		t.Errorf("events = %q", got)
	}
}

func TestStepLogService_Follow_Rotate(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t,
		sse("id: 1-0\ndata: a\n\nid: 2-0\ndata: b\n\nevent: end\ndata: {\"reason\":\"rotate\"}\n\n"),
		// A replay of what the client already has is not delivered again.
		sse("id: 1-0\ndata: a\n\nid: 2-0\ndata: b\n\nid: 10-0\ndata: c\n\nevent: end\ndata: {\"reason\":\"overflow\"}\n\n"),
		sse("id: 9-5\ndata: stale\n\nid: 10-1\ndata: d\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"),
	)

	var waits []time.Duration
	var events []terrakube.StepLogEvent
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(&waits), collect(&events)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := dataOf(events); got != "1-0=a;2-0=b;10-0=c;10-1=d;" {
		t.Errorf("events = %q", got)
	}
	if ids := srv.lastEventIDs(); strings.Join(ids, ",") != "<none>,2-0,10-0" {
		t.Errorf("Last-Event-ID per request = %q", ids)
	}
	if len(waits) != 2 {
		t.Errorf("waits = %v, want one before each reconnect", waits)
	}
}

func TestStepLogService_Follow_DroppedConnection(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t,
		sse("id: 1-0\ndata: a\n\nid: 2-0\ndata: cut"), // closed in the middle of an event
		sse("id: 2-0\ndata: b\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"),
	)

	var events []terrakube.StepLogEvent
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), collect(&events)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := dataOf(events); got != "1-0=a;2-0=b;" {
		t.Errorf("events = %q, want the cut event dropped and read again", got)
	}
	if ids := srv.lastEventIDs(); strings.Join(ids, ",") != "<none>,1-0" {
		t.Errorf("Last-Event-ID per request = %q", ids)
	}
}

func TestStepLogService_Follow_RetryAfter(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t,
		status(http.StatusServiceUnavailable, "7"),
		status(http.StatusServiceUnavailable, "3600"), // capped
		status(http.StatusBadGateway, ""),             // no Retry-After: the backoff applies
		sse("id: 1-0\ndata: a\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"),
	)

	var waits []time.Duration
	var events []terrakube.StepLogEvent
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(&waits), collect(&events)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []time.Duration{7 * time.Second, time.Minute, time.Millisecond}
	if len(waits) != len(want) || waits[0] != want[0] || waits[1] != want[1] || waits[2] != want[2] {
		t.Errorf("waits = %v, want %v", waits, want)
	}
	if got := dataOf(events); got != "1-0=a;" {
		t.Errorf("events = %q", got)
	}
	if n := len(srv.lastEventIDs()); n != 4 {
		t.Errorf("requests = %d, want 4", n)
	}
}

func TestStepLogService_Follow_GivesUp(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t,
		status(http.StatusServiceUnavailable, "1"),
		sse(""), // accepted and closed without an event
		sse("event: end\ndata: {\"reason\":\"rotate\"}\n\n"), // ended before any output
	)

	opts := fastFollow(nil)
	opts.MaxFailures = 3
	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", opts, collect(new([]terrakube.StepLogEvent)))
	if err == nil || !strings.Contains(err.Error(), "3 failed connection attempts") {
		t.Fatalf("err = %v, want it to give up after 3 attempts", err)
	}
	if n := len(srv.lastEventIDs()); n != 3 {
		t.Errorf("requests = %d, want 3", n)
	}
}

func TestStepLogService_Follow_GivesUpWrapsCause(t *testing.T) {
	t.Parallel()

	_, client := newSSEServer(t,
		status(http.StatusServiceUnavailable, ""),
		status(http.StatusServiceUnavailable, ""),
	)

	opts := fastFollow(nil)
	opts.MaxFailures = 2
	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", opts, collect(new([]terrakube.StepLogEvent)))
	if apiErr, ok := errors.AsType[*terrakube.APIError](err); !ok || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("err = %v, want it to wrap the last 503", err)
	}
}

func TestStepLogService_Follow_EventsResetFailures(t *testing.T) {
	t.Parallel()

	// With MaxFailures 1 any connection that ends without output gives up,
	// but each of these delivers output (or a heartbeat) before it drops.
	_, client := newSSEServer(t,
		sse("id: 1-0\ndata: a\n\n"),
		sse("id: 2-0\ndata: b\n\n"),
		sse(": heartbeat\n\n"),
		sse("id: 3-0\ndata: c\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"),
	)

	opts := fastFollow(nil)
	opts.MaxFailures = 1
	var events []terrakube.StepLogEvent
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", opts, collect(&events)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := dataOf(events); got != "1-0=a;2-0=b;3-0=c;" {
		t.Errorf("events = %q", got)
	}
}

func TestStepLogService_Follow_FatalStatus(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t, status(http.StatusForbidden, ""))

	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), collect(new([]terrakube.StepLogEvent)))
	if !terrakube.IsForbidden(err) {
		t.Fatalf("err = %v, want 403", err)
	}
	if n := len(srv.lastEventIDs()); n != 1 {
		t.Errorf("requests = %d, want no retry", n)
	}
}

func TestStepLogService_Follow_FnError(t *testing.T) {
	t.Parallel()

	stop := errors.New("stop")
	srv, client := newSSEServer(t, sse("id: 1-0\ndata: a\n\nid: 2-0\ndata: b\n\n"))

	var calls int
	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), func(terrakube.StepLogEvent) error {
		calls++
		return stop
	})
	if !errors.Is(err, stop) {
		t.Fatalf("err = %v, want fn's error", err)
	}
	if calls != 1 || len(srv.lastEventIDs()) != 1 {
		t.Errorf("calls = %d, requests = %d, want 1 and 1", calls, len(srv.lastEventIDs()))
	}
}

func TestStepLogService_Follow_ContextCanceled(t *testing.T) {
	t.Parallel()

	_, client := newSSEServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "id: 1-0\ndata: a\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done() // the stream stays open until the client leaves
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := client.StepLogs.Follow(ctx, "org-1", "job-1", "step-1", fastFollow(nil), func(terrakube.StepLogEvent) error {
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

// The default sleep returns when the context ends.
func TestStepLogService_Follow_ContextCanceledWhileWaiting(t *testing.T) {
	t.Parallel()

	_, client := newSSEServer(t, status(http.StatusServiceUnavailable, "30"))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := client.StepLogs.Follow(ctx, "org-1", "job-1", "step-1", nil, collect(new([]terrakube.StepLogEvent)))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if time.Since(start) > 10*time.Second {
		t.Error("Follow waited out Retry-After instead of returning when the context ended")
	}
}

// The default sleep waits for the backoff before reconnecting.
func TestStepLogService_Follow_DefaultSleep(t *testing.T) {
	t.Parallel()

	_, client := newSSEServer(t,
		sse("id: 1-0\ndata: a\n\n"),
		sse("event: end\ndata: {\"reason\":\"complete\"}\n\n"),
	)

	opts := &terrakube.StepLogFollowOptions{Backoff: func(int) time.Duration { return 20 * time.Millisecond }}
	start := time.Now()
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", opts, collect(new([]terrakube.StepLogEvent))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d := time.Since(start); d < 20*time.Millisecond {
		t.Errorf("Follow took %v, want at least the 20ms backoff", d)
	}
}

func TestStepLogService_Follow_RedirectToForeignHost(t *testing.T) {
	t.Parallel()

	var foreignCalls atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { foreignCalls.Add(1) }))
	t.Cleanup(foreign.Close)

	srv, client := newSSEServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, foreign.URL+"/steal", http.StatusFound)
	})

	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), collect(new([]terrakube.StepLogEvent)))
	if !errors.Is(err, terrakube.ErrForeignHost) {
		t.Fatalf("err = %v, want ErrForeignHost", err)
	}
	if foreignCalls.Load() != 0 || len(srv.lastEventIDs()) != 1 {
		t.Errorf("foreign calls = %d, requests = %d, want 0 and 1", foreignCalls.Load(), len(srv.lastEventIDs()))
	}
}

func TestStepLogService_Follow_TransportError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.NotFoundHandler())
	endpoint := srv.URL
	srv.Close() // nothing listens there any more

	client := newTestClientFromURL(t, endpoint)
	opts := fastFollow(nil)
	opts.MaxFailures = 2
	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", opts, collect(new([]terrakube.StepLogEvent)))
	if err == nil || !strings.Contains(err.Error(), "2 failed connection attempts") {
		t.Fatalf("err = %v, want it to give up after 2 attempts", err)
	}
}

func TestStepLogService_Follow_GapFnError(t *testing.T) {
	t.Parallel()

	stop := errors.New("stop")
	_, client := newSSEServer(t, sse("event: gap\ndata: {\"reason\":\"trimmed\"}\n\n"))

	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), func(e terrakube.StepLogEvent) error {
		if e.Gap == nil || e.Gap.Reason != "trimmed" {
			t.Errorf("event = %+v, want a trimmed gap", e)
		}
		return stop
	})
	if !errors.Is(err, stop) {
		t.Fatalf("err = %v, want fn's error", err)
	}
}

func TestStepLogService_Follow_ContextDoneBeforeConnecting(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.StepLogs.Follow(ctx, "org-1", "job-1", "step-1", fastFollow(nil), collect(new([]terrakube.StepLogEvent)))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if n := len(srv.lastEventIDs()); n != 0 {
		t.Errorf("requests = %d, want 0", n)
	}
}

// A line ending in "\r\n" split across two reads is one line ending.
func TestStepLogService_Follow_CRLFAcrossReads(t *testing.T) {
	t.Parallel()

	_, client := newSSEServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "id: 1-0\r")
		w.(http.Flusher).Flush()
		time.Sleep(20 * time.Millisecond)
		_, _ = io.WriteString(w, "\ndata: a\r\n\r\nevent: end\r\ndata: {\"reason\":\"complete\"}\r\n\r\n")
	})

	var events []terrakube.StepLogEvent
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), collect(&events)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := dataOf(events); got != "1-0=a;" {
		t.Errorf("events = %q", got)
	}
}

func TestStepLogService_Follow_ReadError(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1000")
			_, _ = io.WriteString(w, "id: 1-0\ndata: a\n\n") // the connection closes before the promised length
		},
		sse("id: 2-0\ndata: b\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"),
	)

	var events []terrakube.StepLogEvent
	if err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), collect(&events)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := dataOf(events); got != "1-0=a;2-0=b;" {
		t.Errorf("events = %q", got)
	}
	if ids := srv.lastEventIDs(); strings.Join(ids, ",") != "<none>,1-0" {
		t.Errorf("Last-Event-ID per request = %q", ids)
	}
}

func TestStepLogService_Follow_LineTooLong(t *testing.T) {
	t.Parallel()

	srv, client := newSSEServer(t, sse("data: "+strings.Repeat("x", 8<<20)+"\n\n"))

	err := client.StepLogs.Follow(context.Background(), "org-1", "job-1", "step-1", fastFollow(nil), collect(new([]terrakube.StepLogEvent)))
	if err == nil || !strings.Contains(err.Error(), "longer than") {
		t.Fatalf("err = %v, want a line-too-long error", err)
	}
	if n := len(srv.lastEventIDs()); n != 1 {
		t.Errorf("requests = %d, want no retry", n)
	}
}

func TestStepLogService_Follow_Validation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, testutil.NewServer(t))
	ctx := context.Background()
	fn := collect(new([]terrakube.StepLogEvent))

	assertValidationError(t, client.StepLogs.Follow(ctx, "", "job-1", "step-1", nil, fn), "organizationID")
	assertValidationError(t, client.StepLogs.Follow(ctx, "org-1", "", "step-1", nil, fn), "jobID")
	assertValidationError(t, client.StepLogs.Follow(ctx, "org-1", "job-1", "", nil, fn), "stepID")
	err := client.StepLogs.Follow(ctx, "org-1", "job-1", "step-1", nil, nil)
	if ve, ok := errors.AsType[*terrakube.ValidationError](err); !ok || ve.Field != "fn" {
		t.Errorf("nil fn: err = %v, want a *ValidationError for fn", err)
	}
}

func TestDefaultStepLogBackoff(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		failures int
		min, max time.Duration
	}{
		{0, time.Second, 2 * time.Second},
		{1, 2 * time.Second, 3 * time.Second},
		{3, 8 * time.Second, 9 * time.Second},
		{10, 30 * time.Second, 31 * time.Second},
		{100, 30 * time.Second, 31 * time.Second},
	} {
		for range 20 {
			if d := terrakube.DefaultStepLogBackoff(tc.failures); d < tc.min || d >= tc.max {
				t.Errorf("backoff(%d) = %v, want in [%v, %v)", tc.failures, d, tc.min, tc.max)
			}
		}
	}
}
