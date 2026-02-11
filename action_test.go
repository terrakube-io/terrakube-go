package terrakube_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func TestActionService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/action", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Run plan"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Action{
			{ID: "act-1", Name: "plan", Label: "Plan", Action: "plan", Category: "terraform", Type: "system", Active: true, Description: &desc},
			{ID: "act-2", Name: "apply", Label: "Apply", Action: "apply", Category: "terraform", Type: "system", Active: false},
		})
	})

	client := newTestClient(t, srv)
	actions, err := client.Actions.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("got %d actions, want 2", len(actions))
	}
	if actions[0].Name != "plan" {
		t.Errorf("Name = %q, want %q", actions[0].Name, "plan")
	}
}

func TestActionService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/action", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Action{
			{ID: "act-1", Name: "plan"},
		})
	})

	client := newTestClient(t, srv)
	actions, err := client.Actions.List(context.Background(), &terrakube.ListOptions{Filter: "name==plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1", len(actions))
	}
}

func TestActionService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/action/act-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Run plan"
		ver := "1.0.0"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Action{
			ID: "act-1", Name: "plan", Label: "Plan", Action: "plan",
			Category: "terraform", Type: "system", Active: true,
			Description: &desc, Version: &ver,
		})
	})

	client := newTestClient(t, srv)
	action, err := client.Actions.Get(context.Background(), "act-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.ID != "act-1" {
		t.Errorf("ID = %q, want %q", action.ID, "act-1")
	}
	if action.Name != "plan" {
		t.Errorf("Name = %q, want %q", action.Name, "plan")
	}
	if action.Label != "Plan" {
		t.Errorf("Label = %q, want %q", action.Label, "Plan")
	}
	if action.Action != "plan" {
		t.Errorf("Action = %q, want %q", action.Action, "plan")
	}
	if action.Category != "terraform" {
		t.Errorf("Category = %q, want %q", action.Category, "terraform")
	}
	if action.Type != "system" {
		t.Errorf("Type = %q, want %q", action.Type, "system")
	}
	if !action.Active {
		t.Error("Active = false, want true")
	}
	if action.Description == nil || *action.Description != "Run plan" {
		t.Errorf("Description = %v, want %q", action.Description, "Run plan")
	}
	if action.Version == nil || *action.Version != "1.0.0" {
		t.Errorf("Version = %v, want %q", action.Version, "1.0.0")
	}
}

func TestActionService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Actions.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestActionService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/action/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "action not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Actions.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestActionService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/action", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Action{
			ID: "act-new", Name: "custom", Label: "Custom", Action: "custom",
			Category: "user", Type: "custom", Active: true,
		})
	})

	client := newTestClient(t, srv)
	action, err := client.Actions.Create(context.Background(), &terrakube.Action{
		Name: "custom", Label: "Custom", Action: "custom",
		Category: "user", Type: "custom", Active: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.ID != "act-new" {
		t.Errorf("ID = %q, want %q", action.ID, "act-new")
	}
	if action.Name != "custom" {
		t.Errorf("Name = %q, want %q", action.Name, "custom")
	}
}

func TestActionService_Create_BooleanFalse(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/action", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assertJSONBoolField(t, body, "active", false)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Action{ID: "act-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, err := client.Actions.Create(context.Background(), &terrakube.Action{Name: "test", Active: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActionService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/action/act-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Action{
			ID: "act-1", Name: "updated", Label: "Updated", Active: false,
		})
	})

	client := newTestClient(t, srv)
	action, err := client.Actions.Update(context.Background(), &terrakube.Action{
		ID: "act-1", Name: "updated", Label: "Updated", Active: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Name != "updated" {
		t.Errorf("Name = %q, want %q", action.Name, "updated")
	}
}

func TestActionService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Actions.Update(context.Background(), &terrakube.Action{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "action ID")
}

func TestActionService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/action/act-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Actions.Delete(context.Background(), "act-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActionService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Actions.Delete(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestActionService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/action/act-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Actions.Delete(context.Background(), "act-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestActionService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/action/act-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Action{ID: "act-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Actions.Get(context.Background(), "act-1")
}
