package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"
)

func TestAPIVersion(t *testing.T) {
	t.Parallel()
	if terrakube.APIVersion == "" {
		t.Fatal("APIVersion must not be empty")
	}
}

func TestNewClient_RequiresEndpoint(t *testing.T) {
	t.Parallel()
	_, err := terrakube.NewClient(terrakube.WithToken("tok"))
	if err == nil {
		t.Fatal("expected error when endpoint missing")
	}
}

func TestNewClient_RequiresToken(t *testing.T) {
	t.Parallel()
	_, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"))
	if err == nil {
		t.Fatal("expected error when token missing")
	}
}

func TestNewClient_EmptyEndpoint(t *testing.T) {
	t.Parallel()
	_, err := terrakube.NewClient(terrakube.WithEndpoint(""), terrakube.WithToken("tok"))
	if err == nil {
		t.Fatal("expected error for empty endpoint")
	}
}

func TestNewClient_EmptyToken(t *testing.T) {
	t.Parallel()
	_, err := terrakube.NewClient(terrakube.WithToken(""), terrakube.WithEndpoint("https://example.com"))
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestNewClient_AddsHTTPS(t *testing.T) {
	t.Parallel()
	c, err := terrakube.NewClient(
		terrakube.WithEndpoint("example.com"),
		terrakube.WithToken("tok"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_AllOptions(t *testing.T) {
	t.Parallel()
	c, err := terrakube.NewClient(
		terrakube.WithEndpoint("https://example.com"),
		terrakube.WithToken("tok"),
		terrakube.WithHTTPClient(&http.Client{}),
		terrakube.WithUserAgent("test-agent"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_InsecureTLS(t *testing.T) {
	t.Parallel()
	c, err := terrakube.NewClient(
		terrakube.WithEndpoint("https://example.com"),
		terrakube.WithToken("tok"),
		terrakube.WithInsecureTLS(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClient_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"type":"organization","id":"1","attributes":{"name":"test"}}}`))
	}))
	defer srv.Close()

	c, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Organizations.Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_NonOKStatusReturnsAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"detail":"internal error"}]}`))
	}))
	defer srv.Close()

	c, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("tok"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Organizations.Get(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	if apiErr, ok := errors.AsType[*terrakube.APIError](err); ok {
		if apiErr.StatusCode != 500 {
			t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
		}
		if len(apiErr.Errors) == 0 || apiErr.Errors[0].Detail != "internal error" {
			t.Errorf("expected error detail 'internal error', got %v", apiErr.Errors)
		}
	} else {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
}

func TestClient_UserAgent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "custom-agent" {
			t.Errorf("User-Agent = %q, want %q", ua, "custom-agent")
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"type":"organization","id":"1","attributes":{"name":"test"}}}`))
	}))
	defer srv.Close()

	c, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("tok"),
		terrakube.WithUserAgent("custom-agent"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _ = c.Organizations.Get(context.Background(), "1")
}

func TestClient_EndpointPathPrefix(t *testing.T) {
	t.Parallel()

	var (
		mu    sync.Mutex
		paths []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		if r.URL.Path == "/terrakube/access-token/v1/teams" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"type":"organization","id":"1","attributes":{"name":"test"}}}`))
	}))
	defer srv.Close()

	for _, endpoint := range []string{srv.URL + "/terrakube", srv.URL + "/terrakube/"} {
		mu.Lock()
		paths = nil
		mu.Unlock()
		c, err := terrakube.NewClient(terrakube.WithEndpoint(endpoint), terrakube.WithToken("tok"))
		if err != nil {
			t.Fatalf("NewClient(%q): %v", endpoint, err)
		}
		if _, err := c.Organizations.Get(context.Background(), "1"); err != nil {
			t.Fatalf("Get: %v", err)
		}
		if _, err := c.TeamTokens.List(context.Background()); err != nil {
			t.Fatalf("TeamTokens.List: %v", err)
		}
		want := []string{"/terrakube/api/v1/organization/1", "/terrakube/access-token/v1/teams"}
		mu.Lock()
		got := slices.Clone(paths)
		mu.Unlock()
		if !slices.Equal(got, want) {
			t.Errorf("endpoint %q: request paths = %q, want %q", endpoint, got, want)
		}
	}
}

type passthroughRoundTripper struct{}

func (passthroughRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return http.DefaultTransport.RoundTrip(r)
}

func TestNewClient_InsecureTLS_Behavior(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"type":"organization","id":"1","attributes":{"name":"test"}}}`))
	}))
	t.Cleanup(srv.Close) // not defer: the parallel subtests run after this function returns

	get := func(t *testing.T, opts ...terrakube.Option) error {
		t.Helper()
		opts = append([]terrakube.Option{terrakube.WithEndpoint(srv.URL), terrakube.WithToken("tok")}, opts...)
		c, err := terrakube.NewClient(opts...)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		_, err = c.Organizations.Get(context.Background(), "1")
		return err
	}

	t.Run("verification is on by default", func(t *testing.T) {
		t.Parallel()
		if err := get(t); err == nil {
			t.Fatal("request to a self-signed server succeeded without WithInsecureTLS")
		}
	})

	t.Run("option order does not matter", func(t *testing.T) {
		t.Parallel()
		custom := &http.Client{Timeout: 30 * time.Second}
		if err := get(t, terrakube.WithHTTPClient(custom), terrakube.WithInsecureTLS()); err != nil {
			t.Errorf("WithHTTPClient then WithInsecureTLS: %v", err)
		}
		if err := get(t, terrakube.WithInsecureTLS(), terrakube.WithHTTPClient(custom)); err != nil {
			t.Errorf("WithInsecureTLS then WithHTTPClient: %v", err)
		}
		if custom.Transport != nil {
			t.Error("the caller's http.Client was modified")
		}
	})

	t.Run("custom RoundTripper is an error", func(t *testing.T) {
		t.Parallel()
		_, err := terrakube.NewClient(
			terrakube.WithEndpoint(srv.URL), terrakube.WithToken("tok"),
			terrakube.WithHTTPClient(&http.Client{Transport: passthroughRoundTripper{}}),
			terrakube.WithInsecureTLS(),
		)
		if err == nil {
			t.Fatal("expected an error: the custom transport would be silently dropped")
		}
	})
}

func TestNewClient_NilHTTPClient(t *testing.T) {
	t.Parallel()
	_, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("tok"), terrakube.WithHTTPClient(nil))
	if err == nil {
		t.Fatal("expected an error for a nil http client")
	}
}

func TestSameOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		endpoint, target string
		want             bool
	}{
		{"https://tk.example.com", "https://tk.example.com/tfstate/v1/a.json", true},
		{"https://tk.example.com", "https://TK.example.com:443/tfstate/v1/a.json", true},
		{"http://tk.example.com", "http://tk.example.com:80/x", true},
		{"https://tk.example.com/", "https://tk.example.com/x", true},
		{"https://tk.example.com/terrakube", "https://tk.example.com/terrakube/tfstate/v1/a.json", true},
		{"https://tk.example.com/terrakube/", "https://tk.example.com/terrakube/x", true},
		{"https://tk.example.com/terrakube", "https://tk.example.com/terrakube", true},
		{"https://tk.example.com", "http://tk.example.com/x", false},
		{"https://tk.example.com", "https://evil.example.com/x", false},
		{"https://tk.example.com", "https://tk.example.com.evil.com/x", false},
		{"https://tk.example.com", "https://tk.example.com:8443/x", false},
		{"https://tk.example.com:8443", "https://tk.example.com/x", false},
		{"https://tk.example.com/terrakube", "https://tk.example.com/other/x", false},
		{"https://tk.example.com/terrakube", "https://tk.example.com/terrakube-evil/x", false},
		{"https://tk.example.com/terrakube", "https://tk.example.com/terrakube/../x", false},
		{"https://tk.example.com", "/tfstate/v1/a.json", false},
		{"https://tk.example.com", "mailto:a@tk.example.com", false},
		{"https://kube.example.com", "https://\u212aube.example.com/x", false},
		{"https://tk.example.com", "https://user:pw@tk.example.com/x", false},
		{"https://tk.example.com", "https://user@tk.example.com/x", false},
		{"https://tk.example.com/tk", "https://tk.example.com/tk//x", false},
		{"https://tk.example.com/tk", "https://tk.example.com/tk/..;/x", false},
		{"https://tk.example.com/tk", "https://tk.example.com/tk/.;v=1/x", false},
		{"https://tk.example.com/tk", "https://tk.example.com/tk/%2e%2e/x", false},
		{"https://tk.example.com/tk", "https://tk.example.com/tk/a%2F..%2F..%2Fx", false},
		{"https://tk.example.com/tk", "https://tk.example.com/tk/a;v=1/x", true},
	}
	for _, tt := range tests {
		e, err := url.Parse(tt.endpoint)
		if err != nil {
			t.Fatal(err)
		}
		g, err := url.Parse(tt.target)
		if err != nil {
			t.Fatal(err)
		}
		if got := terrakube.SameOrigin(e, g); got != tt.want {
			t.Errorf("sameOrigin(%q, %q) = %v, want %v", tt.endpoint, tt.target, got, tt.want)
		}
	}
	if terrakube.SameOrigin(nil, &url.URL{}) || terrakube.SameOrigin(&url.URL{}, nil) {
		t.Error("sameOrigin with a nil URL = true, want false")
	}
}

func TestErrForeignHost_NoToken(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	t.Cleanup(foreign.Close)
	var apiCalls atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls.Add(1)
		if r.URL.Path != "/tfstate/v1/organization/o/workspace/w/state/abc.json" || r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("API got %s with Authorization %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(api.Close)

	c, err := terrakube.NewClient(terrakube.WithEndpoint(api.URL), terrakube.WithToken("tok"))
	if err != nil {
		t.Fatal(err)
	}

	// A state path of the API on another host is read from the endpoint.
	h := &terrakube.History{Output: foreign.URL + "/tfstate/v1/organization/o/workspace/w/state/abc.json"}
	rc, err := c.State.Download(context.Background(), h, terrakube.StateJSON)
	if err != nil {
		t.Fatalf("API state path on another host: %v", err)
	}
	_ = rc.Close()
	if n := apiCalls.Load(); n != 1 {
		t.Errorf("API got %d requests, want 1", n)
	}

	// Any other URL on another host is refused.
	for _, out := range []string{
		foreign.URL + "/bucket/states/abc.json",
		foreign.URL + "/tfstate/v1/organization/o/workspace/w/state/../../../../x/abc.json",
		foreign.URL + "/tfstate/v1/organization/o/workspace//state/abc.json",
	} {
		_, err := c.State.Download(context.Background(), &terrakube.History{Output: out}, terrakube.StateJSON)
		if !errors.Is(err, terrakube.ErrForeignHost) {
			t.Errorf("%s: err = %v, want ErrForeignHost", out, err)
		}
	}
	if n := calls.Load(); n != 0 {
		t.Errorf("foreign host got %d requests, want 0", n)
	}
}

func TestListOptions_Sort(t *testing.T) {
	t.Parallel()

	var sorts sync.Map
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sorts.Store(r.URL.Path, r.URL.Query().Get("sort"))
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(srv.Close)

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("tok"))
	if err != nil {
		t.Fatal(err)
	}
	opts := &terrakube.ListOptions{Sort: "-name"}
	if _, err := c.Workspaces.List(context.Background(), "o", opts); err != nil { // paged
		t.Fatal(err)
	}
	if _, err := c.Teams.List(context.Background(), "o", opts); err != nil { // not paged
		t.Fatal(err)
	}
	if _, err := c.Agents.List(context.Background(), "o", nil); err != nil { // paged, default order
		t.Fatal(err)
	}
	for p, want := range map[string]string{
		"/api/v1/organization/o/workspace": "-name,id",
		"/api/v1/organization/o/team":      "-name",
		"/api/v1/organization/o/agent":     "id",
	} {
		if got, _ := sorts.Load(p); got != want {
			t.Errorf("%s: sort = %v, want %q", p, got, want)
		}
	}
}

func TestListOptions_Sort_TieBreak(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = append(got, r.URL.Query().Get("sort"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.URL.Path == "/api/v1/organization/o/agent" && r.URL.Query().Has("page[totals]") {
			w.WriteHeader(http.StatusBadRequest) // an older server that cannot page agents
			_, _ = w.Write([]byte(`{"errors":[{"detail":"Cannot paginate agent"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(srv.Close)

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("tok"))
	if err != nil {
		t.Fatal(err)
	}
	for _, sortExpr := range []string{"-name", "-id", "name, +id", "-createdDate,name"} {
		if _, err := c.Workspaces.List(context.Background(), "o", &terrakube.ListOptions{Sort: sortExpr}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := c.Agents.List(context.Background(), "o", &terrakube.ListOptions{Sort: "-name"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"-name,id", "-id", "name, +id", "-createdDate,name,id", "-name,id", "-name"}
	if !slices.Equal(got, want) {
		t.Errorf("sorts = %q, want %q (the last is the 400 fallback, which keeps the caller's order)", got, want)
	}
}

// redirectingAPI returns a client for an API at prefix on a server that
// answers every request under prefix with a 307 redirect to target(r).
func redirectingAPI(t *testing.T, srv *httptest.Server, prefix string, opts ...terrakube.Option) *terrakube.Client {
	t.Helper()
	c, err := terrakube.NewClient(append([]terrakube.Option{terrakube.WithEndpoint(srv.URL + prefix), terrakube.WithToken("tok")}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRedirect_TokenStaysOnEndpoint(t *testing.T) {
	t.Parallel()

	var leaks atomic.Int32
	catch := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			leaks.Add(1)
		}
	})
	plain := httptest.NewServer(catch) // another scheme (from the TLS server) and another port
	t.Cleanup(plain.Close)

	redirectTo := func(target func(r *http.Request) string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/other/") {
				catch(w, r)
				return
			}
			http.Redirect(w, r, target(r), http.StatusTemporaryRedirect)
		}
	}

	tlsSrv := httptest.NewTLSServer(redirectTo(func(r *http.Request) string { return plain.URL + r.URL.Path }))
	t.Cleanup(tlsSrv.Close)
	portSrv := httptest.NewServer(redirectTo(func(r *http.Request) string { return plain.URL + r.URL.Path }))
	t.Cleanup(portSrv.Close)
	var prefixSrv *httptest.Server
	prefixSrv = httptest.NewServer(redirectTo(func(r *http.Request) string { return prefixSrv.URL + "/other" + r.URL.Path }))
	t.Cleanup(prefixSrv.Close)

	clients := map[string]*terrakube.Client{
		"https to http":       redirectingAPI(t, tlsSrv, "", terrakube.WithInsecureTLS()),
		"another port":        redirectingAPI(t, portSrv, ""),
		"outside path prefix": redirectingAPI(t, prefixSrv, "/tk"),
	}
	patch := terrakube.Patch{Attributes: map[string]any{"name": "x"}}
	for name, c := range clients {
		if _, err := c.Workspaces.UpdateFields(context.Background(), "o", "w", patch); !errors.Is(err, terrakube.ErrForeignHost) {
			t.Errorf("%s: PATCH err = %v, want ErrForeignHost", name, err)
		}
		if _, err := c.Workspaces.Get(context.Background(), "o", "w"); !errors.Is(err, terrakube.ErrForeignHost) {
			t.Errorf("%s: GET err = %v, want ErrForeignHost", name, err)
		}
		if _, err := c.Workspaces.List(context.Background(), "o", nil); !errors.Is(err, terrakube.ErrForeignHost) {
			t.Errorf("%s: list err = %v, want ErrForeignHost", name, err)
		}
		if _, err := c.TeamTokens.List(context.Background()); !errors.Is(err, terrakube.ErrForeignHost) {
			t.Errorf("%s: raw JSON err = %v, want ErrForeignHost", name, err)
		}
	}
	if n := leaks.Load(); n != 0 {
		t.Errorf("token arrived off the endpoint %d times", n)
	}
}

func TestRedirect_SameOriginAndCallerPolicy(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tk/api/v1/organization/old" {
			http.Redirect(w, r, "/tk/api/v1/organization/new", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"type":"organization","id":"new","attributes":{}}}`))
	}))
	t.Cleanup(srv.Close)

	org, err := redirectingAPI(t, srv, "/tk").Organizations.Get(context.Background(), "old")
	if err != nil || org.ID != "new" {
		t.Fatalf("same-origin redirect: org = %v, err = %v", org, err)
	}

	errStop := errors.New("caller stops redirects")
	custom := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return errStop }}
	_, err = redirectingAPI(t, srv, "/tk", terrakube.WithHTTPClient(custom)).Organizations.Get(context.Background(), "old")
	if !errors.Is(err, errStop) {
		t.Errorf("err = %v, want the caller's CheckRedirect error", err)
	}
	if err := custom.CheckRedirect(nil, nil); !errors.Is(err, errStop) {
		t.Error("the caller's http.Client was modified")
	}
}

func TestRedirectAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		endpoint, target string
		want             bool
	}{
		{"http://tk.example.com", "http://tk.example.com/x", true},
		{"http://tk.example.com", "https://tk.example.com/x", true},
		{"http://tk.example.com:8080/tk", "https://tk.example.com:8080/tk/x", true},
		{"http://tk.example.com:8080", "https://tk.example.com/x", false},
		{"http://tk.example.com/tk", "https://tk.example.com/x", false},
		{"https://tk.example.com", "http://tk.example.com/x", false},
		{"http://tk.example.com", "https://evil.example.com/x", false},
	}
	for _, tt := range tests {
		c, err := terrakube.NewClient(terrakube.WithEndpoint(tt.endpoint), terrakube.WithToken("tok"))
		if err != nil {
			t.Fatal(err)
		}
		u, err := url.Parse(tt.target)
		if err != nil {
			t.Fatal(err)
		}
		if got := c.RedirectAllowed(u); got != tt.want {
			t.Errorf("redirect %s -> %s allowed = %v, want %v", tt.endpoint, tt.target, got, tt.want)
		}
	}
}

func TestValidateID_PathSegment(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://tk.example.com")
	for _, id := range []string{"a/b", "../admin", ".", ".."} {
		_, err := c.Workspaces.Get(context.Background(), "org-1", id)
		ve, ok := errors.AsType[*terrakube.ValidationError](err)
		if !ok || ve.Field != "id" {
			t.Errorf("Get(%q): err = %v, want *ValidationError for id", id, err)
		}
	}
}
