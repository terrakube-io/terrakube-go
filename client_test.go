package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
)

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

	var apiErr *terrakube.APIError
	if errors.As(err, &apiErr) {
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
