package terrakube_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestRequestStream_DoesNotBuffer(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /tfoutput/v1/x", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Errorf("Accept = %q, want the caller's header", got)
		}
		_, _ = io.WriteString(w, "first\n")
		w.(http.Flusher).Flush()
		select { // the rest is written only after the client has read the first line
		case <-release:
		case <-time.After(5 * time.Second):
			t.Error("client did not read the first line before the body ended")
		}
		_, _ = io.WriteString(w, "second\n")
	})

	resp, err := newTestClient(t, srv).RequestStream(context.Background(), http.MethodGet, "/tfoutput/v1/x",
		http.Header{"Accept": {"text/event-stream"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	buf := make([]byte, len("first\n"))
	if _, err := io.ReadFull(resp.Body, buf); err != nil || string(buf) != "first\n" {
		t.Fatalf("first read = %q, %v", buf, err)
	}
	close(release)
	rest, err := io.ReadAll(resp.Body)
	if err != nil || string(rest) != "second\n" {
		t.Fatalf("rest = %q, %v", rest, err)
	}
}

func TestRequestStream_EndpointPrefix(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /terrakube/tfstate/v1/x", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})

	resp, err := newTestClientFromURL(t, srv.URL+"/terrakube").RequestStream(context.Background(), http.MethodGet, "/tfstate/v1/x", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()
}

// closeTracker records whether a response body was closed.
type closeTracker struct {
	io.ReadCloser
	closed *atomic.Bool
}

func (c closeTracker) Close() error {
	c.closed.Store(true)
	return c.ReadCloser.Close()
}

type trackingTransport struct{ closed *atomic.Bool }

func (tt trackingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err == nil {
		resp.Body = closeTracker{resp.Body, tt.closed}
	}
	return resp, err
}

func TestRequestStream_ErrorStatus(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /pat/v1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusForbidden, "denied")
	})

	var closed atomic.Bool
	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("test-token"),
		terrakube.WithHTTPClient(&http.Client{Transport: trackingTransport{&closed}}),
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.RequestStream(context.Background(), http.MethodGet, "/pat/v1", nil)
	if resp != nil {
		t.Errorf("resp = %v, want nil on error", resp)
	}
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Path != "/pat/v1" || len(apiErr.Errors) != 1 || apiErr.Errors[0].Detail != "denied" {
		t.Errorf("APIError = %+v", apiErr)
	}
	if !closed.Load() {
		t.Error("error response body was not closed")
	}
}

func TestRequestStreamURL_ForeignHost(t *testing.T) {
	t.Parallel()

	var foreignCalls atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { foreignCalls.Add(1) }))
	t.Cleanup(foreign.Close)

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.RequestStreamURL(context.Background(), http.MethodGet, foreign.URL+"/tfstate/v1/x.json", nil)
	if !errors.Is(err, terrakube.ErrForeignHost) {
		t.Fatalf("err = %v, want ErrForeignHost", err)
	}
	if n := foreignCalls.Load(); n != 0 {
		t.Errorf("foreign host got %d requests, want 0", n)
	}
}

func TestRequestStream_RedirectToForeignHost(t *testing.T) {
	t.Parallel()

	var leaked atomic.Bool
	foreign := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			leaked.Store(true)
		}
	}))
	t.Cleanup(foreign.Close)

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /tfstate/v1/x", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, foreign.URL+"/steal", http.StatusFound)
	})
	srv.HandleFunc("GET /tfstate/v1/moved", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tfstate/v1/here", http.StatusFound)
	})
	srv.HandleFunc("GET /tfstate/v1/here", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "here")
	})
	client := newTestClient(t, srv)

	_, err := client.RequestStream(context.Background(), http.MethodGet, "/tfstate/v1/x", nil)
	if !errors.Is(err, terrakube.ErrForeignHost) {
		t.Fatalf("err = %v, want ErrForeignHost", err)
	}
	if leaked.Load() {
		t.Error("token sent to the redirect target")
	}

	resp, err := client.RequestStream(context.Background(), http.MethodGet, "/tfstate/v1/moved", nil)
	if err != nil {
		t.Fatalf("same-origin redirect: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if b, _ := io.ReadAll(resp.Body); !strings.EqualFold(string(b), "here") {
		t.Errorf("body = %q", b)
	}
}
