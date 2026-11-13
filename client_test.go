package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
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
