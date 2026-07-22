package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestFederatedService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/federated", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Federated{
			{ID: "fed-1", Name: "GitHubOIDC", IssuerURL: "https://token.actions.githubusercontent.com", Audience: "terrakube"},
			{ID: "fed-2", Name: "GitLabOIDC", IssuerURL: "https://gitlab.com", Audience: "terrakube"},
		})
	})

	client := newTestClient(t, srv)
	feds, err := client.Federated.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feds) != 2 {
		t.Fatalf("got %d federated configs, want 2", len(feds))
	}
	if feds[0].Name != "GitHubOIDC" {
		t.Errorf("Name = %q, want %q", feds[0].Name, "GitHubOIDC")
	}
}

func TestFederatedService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/federated/fed-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Federated{
			ID: "fed-1", Name: "GitHubOIDC", IssuerURL: "https://token.actions.githubusercontent.com", Audience: "terrakube",
		})
	})

	client := newTestClient(t, srv)
	fed, err := client.Federated.Get(context.Background(), "fed-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fed.ID != "fed-1" {
		t.Errorf("ID = %q, want %q", fed.ID, "fed-1")
	}
	if fed.IssuerURL != "https://token.actions.githubusercontent.com" {
		t.Errorf("IssuerURL = %q, want %q", fed.IssuerURL, "https://token.actions.githubusercontent.com")
	}
}

func TestFederatedService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Federated.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty id")
	}
	assertValidationError(t, err, "id")
}

func TestFederatedService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/federated", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Federated{
			ID: "fed-new", Name: "VaultOIDC", IssuerURL: "https://vault.example.com", Audience: "terrakube",
		})
	})

	client := newTestClient(t, srv)
	fed, err := client.Federated.Create(context.Background(), &terrakube.Federated{
		Name: "VaultOIDC", IssuerURL: "https://vault.example.com", Audience: "terrakube",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fed.ID != "fed-new" {
		t.Errorf("ID = %q, want %q", fed.ID, "fed-new")
	}
}

func TestFederatedService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/federated/fed-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Federated{
			ID: "fed-1", Name: "UpdatedGitHubOIDC", IssuerURL: "https://token.actions.githubusercontent.com", Audience: "terrakube-v2",
		})
	})

	client := newTestClient(t, srv)
	fed, err := client.Federated.Update(context.Background(), &terrakube.Federated{
		ID: "fed-1", Name: "UpdatedGitHubOIDC", IssuerURL: "https://token.actions.githubusercontent.com", Audience: "terrakube-v2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fed.Audience != "terrakube-v2" {
		t.Errorf("Audience = %q, want %q", fed.Audience, "terrakube-v2")
	}
}

func TestFederatedService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/federated/fed-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Federated.Delete(context.Background(), "fed-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
