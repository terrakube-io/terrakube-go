package terrakube

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// ErrStateNotFound is returned when the server has no state file to send. The
// local storage backend answers a missing file with 200 and an empty body.
var ErrStateNotFound = errors.New("terrakube: state not found")

// StateFormat selects which file of a state version Download returns.
type StateFormat string

const (
	// StateRaw is the Terraform state file itself ({name}.raw.json).
	StateRaw StateFormat = "raw"
	// StateJSON is the "terraform show -json" form of the state ({name}.json).
	StateJSON StateFormat = "json"
)

// StateService handles the workspace state endpoints under /tfstate/v1. These
// are not JSON:API endpoints: state files are returned as streams.
type StateService struct {
	client *Client
}

// Current returns the workspace's current raw Terraform state. The caller must
// close the returned reader.
// It returns a *ValidationError if orgID or wsID is empty, ErrStateNotFound if
// the server has no state, and a *APIError on server errors.
func (s *StateService) Current(ctx context.Context, orgID, wsID string) (io.ReadCloser, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", wsID); err != nil {
		return nil, err
	}

	p := path.Join("/tfstate/v1/organization", orgID, "workspace", wsID, "state/terraform.tfstate")
	resp, err := s.client.requestStream(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}
	return nonEmptyBody(resp)
}

// Download returns one file of the state version h: the raw Terraform state
// (StateRaw) or its "terraform show -json" form (StateJSON). The caller must
// close the returned reader.
//
// The file's location comes from h.Output, whose host is whatever URL the
// server or executor was configured with, often an in-cluster address. When
// h.Output is on the configured endpoint, Download requests it as it is, so a
// path prefix such as the /api of a single-domain install is kept. When it is
// not on the endpoint (another host, scheme or port, or outside the endpoint's
// path) but has the state path of the API
// (/tfstate/v1/organization/{org}/workspace/{ws}/state/{name}.json), Download
// requests that path from the configured endpoint. Otherwise it returns
// ErrForeignHost without sending anything: the token never leaves the endpoint.
//
// It returns a *ValidationError for a missing h.Output or an unknown format,
// ErrStateNotFound if the server has no such file, and a *APIError on server errors.
func (s *StateService) Download(ctx context.Context, h *History, format StateFormat) (io.ReadCloser, error) {
	name, err := StateName(h)
	if err != nil {
		return nil, err
	}
	var file string
	switch format {
	case StateRaw:
		file = name + ".raw.json"
	case StateJSON:
		file = name + ".json"
	default:
		return nil, &ValidationError{Field: "format", Message: `must be StateRaw or StateJSON, got "` + string(format) + `"`}
	}

	u, err := url.Parse(h.Output) // StateName has parsed it already
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	if dir, ok := apiStateDir(u.Path); ok && !sameOrigin(s.client.baseURL, u) {
		resp, err = s.client.requestStream(ctx, http.MethodGet, dir+file, nil)
	} else {
		target := url.URL{Scheme: u.Scheme, Host: u.Host, Path: path.Join(path.Dir(u.Path), file)}
		resp, err = s.client.requestStreamURL(ctx, http.MethodGet, target.String(), nil)
	}
	if err != nil {
		return nil, err
	}
	return nonEmptyBody(resp)
}

// apiStateDir returns the API state directory in p, from
// "/tfstate/v1/organization/" up to and including the "/state/" before the
// file name, when p ends in /tfstate/v1/organization/{org}/workspace/{ws}/state/{file}.json
// and none of those segments is empty, "." or "..".
func apiStateDir(p string) (string, bool) {
	const prefix = "/tfstate/v1/organization/"
	_, after, ok := strings.Cut(p, prefix)
	if !ok {
		return "", false
	}
	segs := strings.Split(after, "/") // {org} workspace {ws} state {file}
	if len(segs) != 5 || segs[1] != "workspace" || segs[3] != "state" || !strings.HasSuffix(segs[4], ".json") {
		return "", false
	}
	for _, seg := range segs {
		if seg == "" || seg == "." || seg == ".." {
			return "", false
		}
	}
	return prefix + strings.Join(segs[:4], "/") + "/", true
}

// Rollback makes the state version h the workspace's current state. It sends
// PUT /tfstate/v1/organization/{orgID}/workspace/{wsID}/rollback/{name}.json to
// the configured endpoint, where name is StateName(h). The server does not
// check that h belongs to the workspace, so the caller should.
// It returns a *ValidationError if orgID, wsID or h.Output is empty and a
// *APIError on server errors; a 404 "Raw Terraform state not found" means the
// version is too old to roll back to.
func (s *StateService) Rollback(ctx context.Context, orgID, wsID string, h *History) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", wsID); err != nil {
		return err
	}
	name, err := StateName(h)
	if err != nil {
		return err
	}

	p := path.Join("/tfstate/v1/organization", orgID, "workspace", wsID, "rollback", name+".json")
	resp, err := s.client.requestStream(ctx, http.MethodPut, p, nil)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxErrorBody)) // the body is a short text confirmation
	return resp.Body.Close()
}

// StateName returns the file name of the state version h: the last path
// segment of h.Output without ".json" (and without ".raw" before it, should
// h.Output name the raw file). Executor runs use a random UUID as the name,
// not the history ID.
// It returns a *ValidationError if h or h.Output is empty or does not end in a
// file name.
func StateName(h *History) (string, error) {
	if h == nil || h.Output == "" {
		return "", &ValidationError{Field: "History.Output", Message: "must not be empty"}
	}
	u, err := url.Parse(h.Output)
	if err != nil {
		return "", &ValidationError{Field: "History.Output", Message: "is not a URL: " + err.Error()}
	}
	if u.Path == "" || strings.HasSuffix(u.Path, "/") {
		return "", &ValidationError{Field: "History.Output", Message: "does not end in a file name"}
	}
	name := strings.TrimSuffix(path.Base(u.Path), ".json")
	if strings.HasSuffix(path.Base(u.Path), ".raw.json") {
		name = strings.TrimSuffix(name, ".raw")
	}
	if name == "" || name == "." || name == ".." {
		return "", &ValidationError{Field: "History.Output", Message: "does not end in a file name"}
	}
	return name, nil
}

// nonEmptyBody returns the body of resp, or ErrStateNotFound when it is empty.
func nonEmptyBody(resp *http.Response) (io.ReadCloser, error) {
	br := bufio.NewReader(resp.Body)
	if _, err := br.Peek(1); err != nil {
		resp.Body.Close() //nolint:errcheck,gosec // the body is discarded
		if errors.Is(err, io.EOF) {
			return nil, ErrStateNotFound
		}
		return nil, err
	}
	return struct {
		io.Reader
		io.Closer
	}{br, resp.Body}, nil
}
