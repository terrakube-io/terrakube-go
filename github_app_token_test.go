package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestGithubAppTokenService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/github_app_token", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.GithubAppToken{
			{ID: "gat-1", AppID: "12345", InstallationID: "67890", Owner: "my-org"},
			{ID: "gat-2", AppID: "11111", InstallationID: "22222", Owner: "other-org"},
		})
	})

	client := newTestClient(t, srv)
	tokens, err := client.GithubAppTokens.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("got %d tokens, want 2", len(tokens))
	}
	if tokens[0].AppID != "12345" {
		t.Errorf("AppID = %q, want %q", tokens[0].AppID, "12345")
	}
}

func TestGithubAppTokenService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/github_app_token", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.GithubAppToken{
			{ID: "gat-1", AppID: "12345", InstallationID: "67890", Owner: "my-org"},
		})
	})

	client := newTestClient(t, srv)
	tokens, err := client.GithubAppTokens.List(context.Background(), &terrakube.ListOptions{Filter: "owner==my-org"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want 1", len(tokens))
	}
}

func TestGithubAppTokenService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/github_app_token/gat-1", func(w http.ResponseWriter, _ *http.Request) {
		token := "ghp_abc123"
		createdBy := "admin"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.GithubAppToken{
			ID: "gat-1", AppID: "12345", InstallationID: "67890", Owner: "my-org",
			Token: &token, CreatedBy: &createdBy,
		})
	})

	client := newTestClient(t, srv)
	gat, err := client.GithubAppTokens.Get(context.Background(), "gat-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gat.ID != "gat-1" {
		t.Errorf("ID = %q, want %q", gat.ID, "gat-1")
	}
	if gat.AppID != "12345" {
		t.Errorf("AppID = %q, want %q", gat.AppID, "12345")
	}
	if gat.InstallationID != "67890" {
		t.Errorf("InstallationID = %q, want %q", gat.InstallationID, "67890")
	}
	if gat.Owner != "my-org" {
		t.Errorf("Owner = %q, want %q", gat.Owner, "my-org")
	}
	if gat.Token == nil || *gat.Token != "ghp_abc123" {
		t.Errorf("Token = %v, want %q", gat.Token, "ghp_abc123")
	}
	if gat.CreatedBy == nil || *gat.CreatedBy != "admin" {
		t.Errorf("CreatedBy = %v, want %q", gat.CreatedBy, "admin")
	}
}

func TestGithubAppTokenService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.GithubAppTokens.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestGithubAppTokenService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/github_app_token/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "github app token not found")
	})

	client := newTestClient(t, srv)
	_, err := client.GithubAppTokens.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestGithubAppTokenService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/github_app_token", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.GithubAppToken{
			ID: "gat-new", AppID: "99999", InstallationID: "88888", Owner: "new-org",
		})
	})

	client := newTestClient(t, srv)
	gat, err := client.GithubAppTokens.Create(context.Background(), &terrakube.GithubAppToken{
		AppID: "99999", InstallationID: "88888", Owner: "new-org",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gat.ID != "gat-new" {
		t.Errorf("ID = %q, want %q", gat.ID, "gat-new")
	}
	if gat.Owner != "new-org" {
		t.Errorf("Owner = %q, want %q", gat.Owner, "new-org")
	}
}

func TestGithubAppTokenService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/github_app_token/gat-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.GithubAppToken{
			ID: "gat-1", AppID: "12345", InstallationID: "67890", Owner: "updated-org",
		})
	})

	client := newTestClient(t, srv)
	gat, err := client.GithubAppTokens.Update(context.Background(), &terrakube.GithubAppToken{
		ID: "gat-1", AppID: "12345", InstallationID: "67890", Owner: "updated-org",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gat.Owner != "updated-org" {
		t.Errorf("Owner = %q, want %q", gat.Owner, "updated-org")
	}
}

func TestGithubAppTokenService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.GithubAppTokens.Update(context.Background(), &terrakube.GithubAppToken{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "github app token ID")
}

func TestGithubAppTokenService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/github_app_token/gat-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.GithubAppTokens.Delete(context.Background(), "gat-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGithubAppTokenService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.GithubAppTokens.Delete(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestGithubAppTokenService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/github_app_token/gat-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.GithubAppTokens.Delete(context.Background(), "gat-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGithubAppTokenService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/github_app_token/gat-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.GithubAppToken{
			ID: "gat-1", AppID: "12345", InstallationID: "67890", Owner: "my-org",
		})
	})

	client := newTestClient(t, srv)
	_, _ = client.GithubAppTokens.Get(context.Background(), "gat-1")
}
