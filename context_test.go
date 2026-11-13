package terrakube_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestClient_ContextCancellation_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second) // simulate slow server
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Organization{})
	})

	client := newTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.Organizations.List(ctx, nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestClient_ContextCancellation_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Organization{ID: "org-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Organizations.Get(ctx, "org-1")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestClient_ContextCancellation_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Organization{ID: "org-new", Name: "test"})
	})

	client := newTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Organizations.Create(ctx, &terrakube.Organization{Name: "test"})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestClient_ContextCancellation_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Organizations.Delete(ctx, "org-1")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestClient_ContextTimeout(t *testing.T) {
	t.Parallel()

	// synctest runs the test on a fake clock, so the server's five-second delay
	// costs no real time; the in-memory test server keeps it inside the bubble.
	synctest.Test(t, func(t *testing.T) {
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(5 * time.Second)
			testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Organization{ID: "org-1", Name: "test"})
		}))

		client, err := terrakube.NewClient(
			terrakube.WithEndpoint("http://example.com"),
			terrakube.WithToken("test-token"),
			terrakube.WithHTTPClient(srv.Client()),
		)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
		defer cancel()

		_, err = client.Organizations.Get(ctx, "org-1")
		if err == nil {
			t.Fatal("expected error from context timeout")
		}
	})
}
