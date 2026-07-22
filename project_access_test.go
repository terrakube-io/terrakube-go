package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestProjectAccessService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/project/proj-1/projectAccess", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.ProjectAccess{
			{ID: "pa-1", Name: "DevTeam", Role: "admin", ManageJob: true},
			{ID: "pa-2", Name: "QATeam", Role: "user", ManageJob: false},
		})
	})

	client := newTestClient(t, srv)
	accesses, err := client.ProjectAccess.List(context.Background(), "org-1", "proj-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accesses) != 2 {
		t.Fatalf("got %d project access rules, want 2", len(accesses))
	}
	if accesses[0].Role != "admin" {
		t.Errorf("Role = %q, want %q", accesses[0].Role, "admin")
	}
}

func TestProjectAccessService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/project/proj-1/projectAccess/pa-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ProjectAccess{
			ID: "pa-1", Name: "DevTeam", Role: "admin", ManageWorkspace: true, PlanJob: true, ApproveJob: true,
		})
	})

	client := newTestClient(t, srv)
	pa, err := client.ProjectAccess.Get(context.Background(), "org-1", "proj-1", "pa-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pa.ID != "pa-1" {
		t.Errorf("ID = %q, want %q", pa.ID, "pa-1")
	}
	if !pa.ManageWorkspace {
		t.Errorf("ManageWorkspace = false, want true")
	}
}

func TestProjectAccessService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/project/proj-1/projectAccess", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.ProjectAccess{
			ID: "pa-new", Name: "OpsTeam", Role: "operator",
		})
	})

	client := newTestClient(t, srv)
	pa, err := client.ProjectAccess.Create(context.Background(), "org-1", "proj-1", &terrakube.ProjectAccess{
		Name: "OpsTeam", Role: "operator",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pa.ID != "pa-new" {
		t.Errorf("ID = %q, want %q", pa.ID, "pa-new")
	}
}

func TestProjectAccessService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/project/proj-1/projectAccess/pa-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ProjectAccess{
			ID: "pa-1", Name: "DevTeam", Role: "maintainer",
		})
	})

	client := newTestClient(t, srv)
	pa, err := client.ProjectAccess.Update(context.Background(), "org-1", "proj-1", &terrakube.ProjectAccess{
		ID: "pa-1", Name: "DevTeam", Role: "maintainer",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pa.Role != "maintainer" {
		t.Errorf("Role = %q, want %q", pa.Role, "maintainer")
	}
}

func TestProjectAccessService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/project/proj-1/projectAccess/pa-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.ProjectAccess.Delete(context.Background(), "org-1", "proj-1", "pa-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
