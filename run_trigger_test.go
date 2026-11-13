package terrakube_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestRunTrigger() *terrakube.RunTrigger {
	return &terrakube.RunTrigger{
		ID: "runtrigger-1",
	}
}

func TestRunTriggerService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/workspace-1/runTriggers", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.RunTrigger{newTestRunTrigger()})
	})

	client := newTestClient(t, srv)

	items, err := client.RunTriggers.List(context.Background(), "org-1", "workspace-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ID != "runtrigger-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "runtrigger-1")
	}
}

func TestRunTriggerService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/workspace-1/runTriggers", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[runTrigger]"); got != "id==x" {
			t.Errorf("filter[runTrigger] = %q, want %q", got, "id==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.RunTrigger{})
	})

	client := newTestClient(t, srv)

	if _, err := client.RunTriggers.List(context.Background(), "org-1", "workspace-1", &terrakube.ListOptions{Filter: "id==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunTriggerService_List_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/workspace-1/runTriggers", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)

	_, err := client.RunTriggers.List(context.Background(), "org-1", "workspace-1", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestRunTriggerService_List_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.RunTriggers.List(context.Background(), "", "workspace-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestRunTriggerService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.RunTriggers.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "workspaceID")
}

func TestRunTriggerService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/workspace-1/runTriggers/runtrigger-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestRunTrigger())
	})

	client := newTestClient(t, srv)

	item, err := client.RunTriggers.Get(context.Background(), "org-1", "workspace-1", "runtrigger-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "runtrigger-1" {
		t.Errorf("ID = %q, want %q", item.ID, "runtrigger-1")
	}
}

func TestRunTriggerService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/workspace-1/runTriggers/runtrigger-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.RunTriggers.Get(context.Background(), "org-1", "workspace-1", "runtrigger-1")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestRunTriggerService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err0 := client.RunTriggers.Get(context.Background(), "", "workspace-1", "runtrigger-1")
	assertValidationError(t, err0, "organizationID")

	_, err1 := client.RunTriggers.Get(context.Background(), "org-1", "", "runtrigger-1")
	assertValidationError(t, err1, "workspaceID")

	_, err := client.RunTriggers.Get(context.Background(), "org-1", "workspace-1", "")
	assertValidationError(t, err, "runTriggerID")
}

func TestRunTriggerService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/workspace-1/runTriggers", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestRunTrigger())
	})

	client := newTestClient(t, srv)

	item, err := client.RunTriggers.Create(context.Background(), "org-1", "workspace-1", &terrakube.RunTrigger{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "runtrigger-1" {
		t.Errorf("ID = %q, want %q", item.ID, "runtrigger-1")
	}
}

func TestRunTriggerService_Create_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/workspace-1/runTriggers", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "invalid")
	})

	client := newTestClient(t, srv)

	_, err := client.RunTriggers.Create(context.Background(), "org-1", "workspace-1", &terrakube.RunTrigger{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestRunTriggerService_Create_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.RunTriggers.Create(context.Background(), "", "workspace-1", &terrakube.RunTrigger{})
	assertValidationError(t, err, "organizationID")
}

func TestRunTriggerService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.RunTriggers.Create(context.Background(), "org-1", "", &terrakube.RunTrigger{})
	assertValidationError(t, err, "workspaceID")
}

func TestRunTrigger_BooleanFieldsSerialization(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/workspace-1/runTriggers", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}
		assertJSONBoolField(t, body, "enabled", false)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestRunTrigger())
	})

	client := newTestClient(t, srv)

	if _, err := client.RunTriggers.Create(context.Background(), "org-1", "workspace-1", &terrakube.RunTrigger{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunTriggerService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/workspace-1/runTriggers/runtrigger-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestRunTrigger())
	})

	client := newTestClient(t, srv)

	item, err := client.RunTriggers.Update(context.Background(), "org-1", "workspace-1", newTestRunTrigger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "runtrigger-1" {
		t.Errorf("ID = %q, want %q", item.ID, "runtrigger-1")
	}
}

func TestRunTriggerService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.RunTriggers.Update(context.Background(), "org-1", "workspace-1", &terrakube.RunTrigger{})
	assertValidationError(t, err, "runTriggerID")
}

func TestRunTriggerService_Update_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.RunTriggers.Update(context.Background(), "", "workspace-1", newTestRunTrigger())
	assertValidationError(t, err, "organizationID")
}

func TestRunTriggerService_Update_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.RunTriggers.Update(context.Background(), "org-1", "", newTestRunTrigger())
	assertValidationError(t, err, "workspaceID")
}

func TestRunTriggerService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/workspace-1/runTriggers/runtrigger-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	if err := client.RunTriggers.Delete(context.Background(), "org-1", "workspace-1", "runtrigger-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunTriggerService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.RunTriggers.Delete(context.Background(), "org-1", "workspace-1", "")
	assertValidationError(t, err, "runTriggerID")
}

func TestRunTriggerService_Delete_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.RunTriggers.Delete(context.Background(), "", "workspace-1", "runtrigger-1")
	assertValidationError(t, err, "organizationID")
}

func TestRunTriggerService_Delete_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.RunTriggers.Delete(context.Background(), "org-1", "", "runtrigger-1")
	assertValidationError(t, err, "workspaceID")
}
