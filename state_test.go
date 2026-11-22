package terrakube_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

const stateDir = "/tfstate/v1/organization/org-1/workspace/ws-1/state/"

func readAllClose(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	defer func() { _ = rc.Close() }()
	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return string(b)
}

func TestStateService_Current(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stateDir+"terraform.tfstate", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = io.WriteString(w, `{"version":4}`)
	})

	rc, err := newTestClient(t, srv).State.Current(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readAllClose(t, rc); got != `{"version":4}` {
		t.Errorf("body = %q", got)
	}
}

func TestStateService_Current_EmptyBody(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stateDir+"terraform.tfstate", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := newTestClient(t, srv).State.Current(context.Background(), "org-1", "ws-1")
	if !errors.Is(err, terrakube.ErrStateNotFound) {
		t.Fatalf("err = %v, want ErrStateNotFound", err)
	}
}

func TestStateService_Current_Errors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stateDir+"terraform.tfstate", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	client := newTestClient(t, srv)

	_, err := client.State.Current(context.Background(), "org-1", "ws-1")
	if !terrakube.IsForbidden(err) {
		t.Errorf("err = %v, want 403", err)
	}
	_, err = client.State.Current(context.Background(), "", "ws-1")
	assertValidationError(t, err, "organizationID")
	_, err = client.State.Current(context.Background(), "org-1", "")
	assertValidationError(t, err, "workspaceID")
}

func TestStateService_Download(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /tk"+stateDir+"abc.raw.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "raw")
	})
	srv.HandleFunc("GET /tk"+stateDir+"abc.json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = io.WriteString(w, "json")
	})
	client := newTestClientFromURL(t, srv.URL+"/tk")
	h := &terrakube.History{ID: "h-1", Output: srv.URL + "/tk" + stateDir + "abc.json"}

	for format, want := range map[terrakube.StateFormat]string{terrakube.StateRaw: "raw", terrakube.StateJSON: "json"} {
		rc, err := client.State.Download(context.Background(), h, format)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", format, err)
		}
		if got := readAllClose(t, rc); got != want {
			t.Errorf("%s: body = %q, want %q", format, got, want)
		}
	}
}

func TestStateService_Download_Errors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET "+stateDir+"empty.json", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv.HandleFunc("GET "+stateDir+"gone.raw.json", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	client := newTestClient(t, srv)

	_, err := client.State.Download(context.Background(), &terrakube.History{Output: srv.URL + stateDir + "empty.json"}, terrakube.StateJSON)
	if !errors.Is(err, terrakube.ErrStateNotFound) {
		t.Errorf("empty body: err = %v, want ErrStateNotFound", err)
	}
	_, err = client.State.Download(context.Background(), &terrakube.History{Output: srv.URL + stateDir + "gone.json"}, terrakube.StateRaw)
	if !terrakube.IsNotFound(err) {
		t.Errorf("404: err = %v, want 404 *APIError", err)
	}
	_, err = client.State.Download(context.Background(), &terrakube.History{Output: "https://storage.example.com/bucket/a.json"}, terrakube.StateRaw)
	if !errors.Is(err, terrakube.ErrForeignHost) {
		t.Errorf("foreign: err = %v, want ErrForeignHost", err)
	}
	_, err = client.State.Download(context.Background(), &terrakube.History{Output: srv.URL + stateDir + "a.json"}, "xml")
	if ve, ok := errors.AsType[*terrakube.ValidationError](err); !ok || ve.Field != "format" {
		t.Errorf("format: err = %v, want *ValidationError for format", err)
	}
	_, err = client.State.Download(context.Background(), &terrakube.History{}, terrakube.StateRaw)
	assertValidationError(t, err, "History.Output")
}

func TestStateService_Rollback(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PUT /tfstate/v1/organization/org-1/workspace/ws-1/rollback/abc.json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "Rollback completed")
	})
	srv.HandleFunc("PUT /tfstate/v1/organization/org-1/workspace/ws-1/rollback/old.json", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "Raw Terraform state not found")
	})
	client := newTestClient(t, srv)

	h := &terrakube.History{Output: "https://elsewhere.example.com" + stateDir + "abc.json"}
	if err := client.State.Rollback(context.Background(), "org-1", "ws-1", h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := client.State.Rollback(context.Background(), "org-1", "ws-1", &terrakube.History{Output: srv.URL + stateDir + "old.json"})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok || apiErr.StatusCode != http.StatusNotFound || string(apiErr.Body) != "Raw Terraform state not found" {
		t.Errorf("err = %v, want 404 with the server's text", err)
	}

	assertValidationError(t, client.State.Rollback(context.Background(), "", "ws-1", h), "organizationID")
	assertValidationError(t, client.State.Rollback(context.Background(), "org-1", "", h), "workspaceID")
	assertValidationError(t, client.State.Rollback(context.Background(), "org-1", "ws-1", nil), "History.Output")
}

func TestStateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		output  string
		want    string
		wantErr bool
	}{
		{"https://tk.example.com" + stateDir + "0b9c1f1e-3c3a-4f9e-9f57-3b1a2b8c7d6e.json", "0b9c1f1e-3c3a-4f9e-9f57-3b1a2b8c7d6e", false},
		{"https://tk.example.com/tk" + stateDir + "abc.json?x=1", "abc", false},
		{"https://tk.example.com" + stateDir + "abc", "abc", false},
		{"", "", true},
		{"https://tk.example.com", "", true},
		{"https://tk.example.com/", "", true},
		{"https://tk.example.com/.json", "", true},
		{"%zz", "", true},
		{"https://tk.example.com" + stateDir + "abc/", "", true},
		{"https://tk.example.com" + stateDir + "abc.raw.json", "abc", false},
		{"https://tk.example.com/..", "", true},
	}
	for _, tt := range tests {
		got, err := terrakube.StateName(&terrakube.History{Output: tt.output})
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("StateName(%q) = %q, %v; want %q, error %v", tt.output, got, err, tt.want, tt.wantErr)
		}
		if err != nil {
			if _, ok := errors.AsType[*terrakube.ValidationError](err); !ok {
				t.Errorf("StateName(%q) error is %T, want *ValidationError", tt.output, err)
			}
		}
	}
	if _, err := terrakube.StateName(nil); err == nil {
		t.Error("StateName(nil) = nil error")
	}
}

func TestStateService_Download_InClusterHost(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /tk"+stateDir+"abc.raw.json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = io.WriteString(w, "raw")
	})
	srv.HandleFunc("GET /custom/abc.raw.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "custom")
	})
	client := newTestClientFromURL(t, srv.URL+"/tk")

	// The executor writes the in-cluster API URL; the file is read from the endpoint.
	h := &terrakube.History{Output: "http://terrakube-api.terrakube.svc.cluster.local:8080" + stateDir + "abc.json"}
	rc, err := client.State.Download(context.Background(), h, terrakube.StateRaw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readAllClose(t, rc); got != "raw" {
		t.Errorf("body = %q, want %q", got, "raw")
	}

	// Another path on the endpoint is requested as it is.
	clientNoPrefix := newTestClient(t, srv)
	rc, err = clientNoPrefix.State.Download(context.Background(), &terrakube.History{Output: srv.URL + "/custom/abc.json"}, terrakube.StateRaw)
	if err != nil {
		t.Fatalf("same-origin custom path: %v", err)
	}
	if got := readAllClose(t, rc); got != "custom" {
		t.Errorf("body = %q, want %q", got, "custom")
	}
}

func TestStateService_Download_SingleDomain(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	// A single-domain install serves the API under /api and the UI everywhere else.
	srv.HandleFunc("GET /api"+stateDir+"abc.raw.json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = io.WriteString(w, "raw")
	})
	srv.HandleFunc("GET "+stateDir+"abc.raw.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "<html>UI</html>")
	})
	client := newTestClient(t, srv)

	// The executor writes <public-url>/api/tfstate/...; the endpoint is the bare public URL.
	h := &terrakube.History{Output: srv.URL + "/api" + stateDir + "abc.json"}
	rc, err := client.State.Download(context.Background(), h, terrakube.StateRaw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readAllClose(t, rc); got != "raw" {
		t.Errorf("body = %q, want %q (the /api prefix must be kept)", got, "raw")
	}
}

func TestStateService_Download_RewrittenOntoEndpoint(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /tk"+stateDir+"abc.raw.json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = io.WriteString(w, "raw")
	})
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request to %s", r.URL.Path)
		w.WriteHeader(http.StatusTeapot)
	})
	client := newTestClientFromURL(t, srv.URL+"/tk")

	// None of these is on the endpoint, so only the API state path is taken from them.
	for name, output := range map[string]string{
		"outside the endpoint path": srv.URL + stateDir + "abc.json",
		"other scheme":              strings.Replace(srv.URL, "http://", "https://", 1) + "/tk" + stateDir + "abc.json",
		"dot segment":               srv.URL + "/tk/x/.." + stateDir + "abc.json",
	} {
		rc, err := client.State.Download(context.Background(), &terrakube.History{Output: output}, terrakube.StateRaw)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", name, err)
			continue
		}
		if got := readAllClose(t, rc); got != "raw" {
			t.Errorf("%s: body = %q, want %q", name, got, "raw")
		}
	}
}
