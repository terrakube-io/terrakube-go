package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestProjectService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/project", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Default project"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Project{
			{ID: "proj-1", Name: "Default", Description: &desc},
			{ID: "proj-2", Name: "Staging"},
		})
	})

	client := newTestClient(t, srv)
	projects, err := client.Projects.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(projects))
	}
	if projects[0].Name != "Default" {
		t.Errorf("Name = %q, want %q", projects[0].Name, "Default")
	}
}

func TestProjectService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Projects.List(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organizationID")
}

func TestProjectService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/project/proj-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Test project"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Project{
			ID: "proj-1", Name: "Production", Description: &desc,
		})
	})

	client := newTestClient(t, srv)
	proj, err := client.Projects.Get(context.Background(), "org-1", "proj-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.ID != "proj-1" {
		t.Errorf("ID = %q, want %q", proj.ID, "proj-1")
	}
	if proj.Name != "Production" {
		t.Errorf("Name = %q, want %q", proj.Name, "Production")
	}
}

func TestProjectService_Get_ValidationErrors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Projects.Get(context.Background(), "", "proj-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}

	_, err = client.Projects.Get(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty id")
	}
}

func TestProjectService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/project", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		desc := "New proj"
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Project{
			ID: "proj-new", Name: "NewProj", Description: &desc,
		})
	})

	client := newTestClient(t, srv)
	desc := "New proj"
	proj, err := client.Projects.Create(context.Background(), "org-1", &terrakube.Project{
		Name: "NewProj", Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.ID != "proj-new" {
		t.Errorf("ID = %q, want %q", proj.ID, "proj-new")
	}
}

func TestProjectService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/project/proj-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Project{
			ID: "proj-1", Name: "UpdatedProj",
		})
	})

	client := newTestClient(t, srv)
	proj, err := client.Projects.Update(context.Background(), "org-1", &terrakube.Project{
		ID: "proj-1", Name: "UpdatedProj",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.Name != "UpdatedProj" {
		t.Errorf("Name = %q, want %q", proj.Name, "UpdatedProj")
	}
}

func TestProjectService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/project/proj-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Projects.Delete(context.Background(), "org-1", "proj-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
