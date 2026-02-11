package terrakube_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func TestWorkspaceService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, _ *http.Request) {
		desc := "First workspace"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Workspace{
			{ID: "ws-1", Name: "Dev", Description: &desc, Source: "https://github.com/ex/repo", Branch: "main"},
			{ID: "ws-2", Name: "Prod", Source: "https://github.com/ex/repo", Branch: "prod"},
		})
	})

	client := newTestClient(t, srv)
	workspaces, err := client.Workspaces.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workspaces) != 2 {
		t.Fatalf("got %d workspaces, want 2", len(workspaces))
	}
	if workspaces[0].Name != "Dev" {
		t.Errorf("Name = %q, want %q", workspaces[0].Name, "Dev")
	}
}

func TestWorkspaceService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[workspace]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Workspace{
			{ID: "ws-1", Name: "Filtered"},
		})
	})

	client := newTestClient(t, srv)
	workspaces, err := client.Workspaces.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==Filtered"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("got %d workspaces, want 1", len(workspaces))
	}
}

func TestWorkspaceService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Workspaces.List(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestWorkspaceService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Test workspace"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Workspace{
			ID: "ws-1", Name: "Dev", Description: &desc, Source: "https://github.com/ex/repo",
			Branch: "main", Folder: "/", IaCType: "terraform", IaCVersion: "1.5.0",
			ExecutionMode: "remote",
		})
	})

	client := newTestClient(t, srv)
	ws, err := client.Workspaces.Get(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.ID != "ws-1" {
		t.Errorf("ID = %q, want %q", ws.ID, "ws-1")
	}
	if ws.Name != "Dev" {
		t.Errorf("Name = %q, want %q", ws.Name, "Dev")
	}
	if ws.Source != "https://github.com/ex/repo" {
		t.Errorf("Source = %q, want %q", ws.Source, "https://github.com/ex/repo")
	}
	if ws.Description == nil || *ws.Description != "Test workspace" {
		t.Errorf("Description = %v, want %q", ws.Description, "Test workspace")
	}
	if ws.IaCType != "terraform" {
		t.Errorf("IaCType = %q, want %q", ws.IaCType, "terraform")
	}
}

func TestWorkspaceService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Workspaces.Get(context.Background(), "", "ws-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestWorkspaceService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Workspaces.Get(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestWorkspaceService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "workspace not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Workspaces.Get(context.Background(), "org-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestWorkspaceService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		desc := "New workspace"
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Workspace{
			ID: "ws-new", Name: "Created", Description: &desc, Source: "https://github.com/ex/repo",
			Branch: "main", IaCType: "terraform", IaCVersion: "1.5.0", ExecutionMode: "remote",
		})
	})

	client := newTestClient(t, srv)
	desc := "New workspace"
	ws, err := client.Workspaces.Create(context.Background(), "org-1", &terrakube.Workspace{
		Name: "Created", Description: &desc, Source: "https://github.com/ex/repo",
		Branch: "main", IaCType: "terraform", IaCVersion: "1.5.0", ExecutionMode: "remote",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.ID != "ws-new" {
		t.Errorf("ID = %q, want %q", ws.ID, "ws-new")
	}
	if ws.Name != "Created" {
		t.Errorf("Name = %q, want %q", ws.Name, "Created")
	}
}

func TestWorkspaceService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Workspaces.Create(context.Background(), "", &terrakube.Workspace{Name: "test"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestWorkspaceService_Create_BooleanFalse(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assertJSONBoolField(t, body, "deleted", false)
		assertJSONBoolField(t, body, "allowRemoteApply", false)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Workspace{ID: "ws-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, err := client.Workspaces.Create(context.Background(), "org-1", &terrakube.Workspace{Name: "test", Deleted: false, AllowRemoteApply: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Updated"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Workspace{
			ID: "ws-1", Name: "Updated", Description: &desc, Source: "https://github.com/ex/repo",
			Branch: "develop", ExecutionMode: "local",
		})
	})

	client := newTestClient(t, srv)
	desc := "Updated"
	ws, err := client.Workspaces.Update(context.Background(), "org-1", &terrakube.Workspace{
		ID: "ws-1", Name: "Updated", Description: &desc, Branch: "develop", ExecutionMode: "local",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Name != "Updated" {
		t.Errorf("Name = %q, want %q", ws.Name, "Updated")
	}
	if ws.Branch != "develop" {
		t.Errorf("Branch = %q, want %q", ws.Branch, "develop")
	}
}

func TestWorkspaceService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Workspaces.Update(context.Background(), "", &terrakube.Workspace{ID: "ws-1"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestWorkspaceService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Workspaces.Update(context.Background(), "org-1", &terrakube.Workspace{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "workspace ID")
}

func TestWorkspaceService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Workspaces.Delete(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Workspaces.Delete(context.Background(), "", "ws-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestWorkspaceService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Workspaces.Delete(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestWorkspaceService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Workspaces.Delete(context.Background(), "org-1", "ws-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestWorkspaceService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Workspace{ID: "ws-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Workspaces.Get(context.Background(), "org-1", "ws-1")
}
