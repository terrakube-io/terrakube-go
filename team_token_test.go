package terrakube_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestTeamTokenService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /access-token/v1/teams", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSON(t, w, http.StatusOK, []terrakube.TeamToken{
			{ID: "tok-1", Description: "token one", Days: 30, Group: "admins", Value: "secret-1"},
			{ID: "tok-2", Description: "token two", Days: 7, Hours: 12, Group: "devs", Value: "secret-2"},
		})
	})

	c := newTestClient(t, srv)

	tokens, err := c.TeamTokens.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 2 {
		t.Fatalf("got %d tokens, want 2", len(tokens))
	}
	if tokens[0].ID != "tok-1" {
		t.Errorf("tokens[0].ID = %q, want %q", tokens[0].ID, "tok-1")
	}
	if tokens[0].Description != "token one" {
		t.Errorf("tokens[0].Description = %q, want %q", tokens[0].Description, "token one")
	}
	if tokens[0].Days != 30 {
		t.Errorf("tokens[0].Days = %d, want 30", tokens[0].Days)
	}
	if tokens[0].Group != "admins" {
		t.Errorf("tokens[0].Group = %q, want %q", tokens[0].Group, "admins")
	}
	if tokens[0].Value != "secret-1" {
		t.Errorf("tokens[0].Value = %q, want %q", tokens[0].Value, "secret-1")
	}
}

func TestTeamTokenService_List_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /access-token/v1/teams", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSON(t, w, http.StatusOK, []terrakube.TeamToken{})
	})

	c := newTestClient(t, srv)

	_, err := c.TeamTokens.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamTokenService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /access-token/v1/teams", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		var input terrakube.TeamToken
		if err := json.Unmarshal(body, &input); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}
		if input.Description != "ci token" {
			t.Errorf("Description = %q, want %q", input.Description, "ci token")
		}
		if input.Days != 90 {
			t.Errorf("Days = %d, want 90", input.Days)
		}

		testutil.WriteJSON(t, w, http.StatusOK, &terrakube.TeamToken{
			ID:          "tok-new",
			Description: "ci token",
			Days:        90,
			Group:       "ci-team",
			Value:       "generated-secret",
		})
	})

	c := newTestClient(t, srv)

	token, err := c.TeamTokens.Create(context.Background(), &terrakube.TeamToken{
		Description: "ci token",
		Days:        90,
		Group:       "ci-team",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token.ID != "tok-new" {
		t.Errorf("ID = %q, want %q", token.ID, "tok-new")
	}
	if token.Description != "ci token" {
		t.Errorf("Description = %q, want %q", token.Description, "ci token")
	}
	if token.Value != "generated-secret" {
		t.Errorf("Value = %q, want %q", token.Value, "generated-secret")
	}
}

func TestTeamTokenService_Create_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /access-token/v1/teams", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := newTestClient(t, srv)

	_, err := c.TeamTokens.Create(context.Background(), &terrakube.TeamToken{
		Description: "test",
		Days:        1,
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestTeamTokenService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /access-token/v1/teams/tok-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	c := newTestClient(t, srv)

	err := c.TeamTokens.Delete(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamTokenService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.TeamTokens.Delete(context.Background(), "")
	assertValidationError(t, err, "id")
}

func TestTeamTokenService_Delete_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /access-token/v1/teams/tok-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := newTestClient(t, srv)

	err := c.TeamTokens.Delete(context.Background(), "tok-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
