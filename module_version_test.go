package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestModuleVersionService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/mod-1/version", func(w http.ResponseWriter, _ *http.Request) {
		commit := "abc123"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.ModuleVersion{
			{ID: "ver-1", Version: "1.0.0", Commit: &commit},
			{ID: "ver-2", Version: "2.0.0"},
		})
	})

	client := newTestClient(t, srv)
	versions, err := client.ModuleVersions.List(context.Background(), "org-1", "mod-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("got %d versions, want 2", len(versions))
	}
	if versions[0].Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", versions[0].Version, "1.0.0")
	}
}

func TestModuleVersionService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/mod-1/version", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[version]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.ModuleVersion{
			{ID: "ver-1", Version: "1.0.0"},
		})
	})

	client := newTestClient(t, srv)
	versions, err := client.ModuleVersions.List(context.Background(), "org-1", "mod-1", &terrakube.ListOptions{Filter: "version==1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("got %d versions, want 1", len(versions))
	}
}

func TestModuleVersionService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ModuleVersions.List(context.Background(), "", "mod-1", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleVersionService_List_EmptyModuleID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ModuleVersions.List(context.Background(), "org-1", "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty moduleID")
	}
	assertValidationError(t, err, "module ID")
}

func TestModuleVersionService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/mod-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		commit := "abc123"
		createdBy := "admin"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ModuleVersion{
			ID: "ver-1", Version: "1.0.0", Commit: &commit, CreatedBy: &createdBy,
		})
	})

	client := newTestClient(t, srv)
	ver, err := client.ModuleVersions.Get(context.Background(), "org-1", "mod-1", "ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.ID != "ver-1" {
		t.Errorf("ID = %q, want %q", ver.ID, "ver-1")
	}
	if ver.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", ver.Version, "1.0.0")
	}
	if ver.Commit == nil || *ver.Commit != "abc123" {
		t.Errorf("Commit = %v, want %q", ver.Commit, "abc123")
	}
	if ver.CreatedBy == nil || *ver.CreatedBy != "admin" {
		t.Errorf("CreatedBy = %v, want %q", ver.CreatedBy, "admin")
	}
}

func TestModuleVersionService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ModuleVersions.Get(context.Background(), "", "mod-1", "ver-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleVersionService_Get_EmptyModuleID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ModuleVersions.Get(context.Background(), "org-1", "", "ver-1")
	if err == nil {
		t.Fatal("expected validation error for empty moduleID")
	}
	assertValidationError(t, err, "module ID")
}

func TestModuleVersionService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ModuleVersions.Get(context.Background(), "org-1", "mod-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "version ID")
}

func TestModuleVersionService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/mod-1/version/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "version not found")
	})

	client := newTestClient(t, srv)
	_, err := client.ModuleVersions.Get(context.Background(), "org-1", "mod-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestModuleVersionService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/module/mod-1/version", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.ModuleVersion{
			ID: "ver-new", Version: "3.0.0",
		})
	})

	client := newTestClient(t, srv)
	ver, err := client.ModuleVersions.Create(context.Background(), "org-1", "mod-1", &terrakube.ModuleVersion{
		Version: "3.0.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.ID != "ver-new" {
		t.Errorf("ID = %q, want %q", ver.ID, "ver-new")
	}
	if ver.Version != "3.0.0" {
		t.Errorf("Version = %q, want %q", ver.Version, "3.0.0")
	}
}

func TestModuleVersionService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/module/mod-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		commit := "def456"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ModuleVersion{
			ID: "ver-1", Version: "1.1.0", Commit: &commit,
		})
	})

	client := newTestClient(t, srv)
	ver, err := client.ModuleVersions.Update(context.Background(), "org-1", "mod-1", &terrakube.ModuleVersion{
		ID: "ver-1", Version: "1.1.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.Version != "1.1.0" {
		t.Errorf("Version = %q, want %q", ver.Version, "1.1.0")
	}
}

func TestModuleVersionService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ModuleVersions.Update(context.Background(), "org-1", "mod-1", &terrakube.ModuleVersion{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "version ID")
}

func TestModuleVersionService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/module/mod-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.ModuleVersions.Delete(context.Background(), "org-1", "mod-1", "ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModuleVersionService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.ModuleVersions.Delete(context.Background(), "", "mod-1", "ver-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleVersionService_Delete_EmptyModuleID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.ModuleVersions.Delete(context.Background(), "org-1", "", "ver-1")
	if err == nil {
		t.Fatal("expected validation error for empty moduleID")
	}
	assertValidationError(t, err, "module ID")
}

func TestModuleVersionService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.ModuleVersions.Delete(context.Background(), "org-1", "mod-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "version ID")
}

func TestModuleVersionService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/module/mod-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.ModuleVersions.Delete(context.Background(), "org-1", "mod-1", "ver-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestModuleVersionService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/mod-1/version/ver-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ModuleVersion{ID: "ver-1", Version: "1.0.0"})
	})

	client := newTestClient(t, srv)
	_, _ = client.ModuleVersions.Get(context.Background(), "org-1", "mod-1", "ver-1")
}
