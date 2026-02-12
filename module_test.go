package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestModuleService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Module{
			{ID: "mod-1", Name: "vpc", Description: "VPC module", Provider: "aws", Source: "https://github.com/ex/vpc"},
			{ID: "mod-2", Name: "ecs", Description: "ECS module", Provider: "aws", Source: "https://github.com/ex/ecs"},
		})
	})

	client := newTestClient(t, srv)
	modules, err := client.Modules.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(modules) != 2 {
		t.Fatalf("got %d modules, want 2", len(modules))
	}
	if modules[0].Name != "vpc" {
		t.Errorf("Name = %q, want %q", modules[0].Name, "vpc")
	}
}

func TestModuleService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[module]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Module{
			{ID: "mod-1", Name: "Filtered"},
		})
	})

	client := newTestClient(t, srv)
	modules, err := client.Modules.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==Filtered"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(modules) != 1 {
		t.Fatalf("got %d modules, want 1", len(modules))
	}
}

func TestModuleService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Modules.List(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/mod-1", func(w http.ResponseWriter, _ *http.Request) {
		folder := "modules/vpc"
		tagPrefix := "v"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Module{
			ID: "mod-1", Name: "vpc", Description: "VPC module", Provider: "aws",
			Source: "https://github.com/ex/vpc", Folder: &folder, TagPrefix: &tagPrefix,
		})
	})

	client := newTestClient(t, srv)
	mod, err := client.Modules.Get(context.Background(), "org-1", "mod-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod.ID != "mod-1" {
		t.Errorf("ID = %q, want %q", mod.ID, "mod-1")
	}
	if mod.Name != "vpc" {
		t.Errorf("Name = %q, want %q", mod.Name, "vpc")
	}
	if mod.Provider != "aws" {
		t.Errorf("Provider = %q, want %q", mod.Provider, "aws")
	}
	if mod.Source != "https://github.com/ex/vpc" {
		t.Errorf("Source = %q, want %q", mod.Source, "https://github.com/ex/vpc")
	}
	if mod.Folder == nil || *mod.Folder != "modules/vpc" {
		t.Errorf("Folder = %v, want %q", mod.Folder, "modules/vpc")
	}
	if mod.TagPrefix == nil || *mod.TagPrefix != "v" {
		t.Errorf("TagPrefix = %v, want %q", mod.TagPrefix, "v")
	}
}

func TestModuleService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Modules.Get(context.Background(), "", "mod-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Modules.Get(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestModuleService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "module not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Modules.Get(context.Background(), "org-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestModuleService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/module", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Module{
			ID: "mod-new", Name: "Created", Description: "New module", Provider: "aws",
			Source: "https://github.com/ex/mod",
		})
	})

	client := newTestClient(t, srv)
	mod, err := client.Modules.Create(context.Background(), "org-1", &terrakube.Module{
		Name: "Created", Description: "New module", Provider: "aws",
		Source: "https://github.com/ex/mod",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod.ID != "mod-new" {
		t.Errorf("ID = %q, want %q", mod.ID, "mod-new")
	}
	if mod.Name != "Created" {
		t.Errorf("Name = %q, want %q", mod.Name, "Created")
	}
}

func TestModuleService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Modules.Create(context.Background(), "", &terrakube.Module{Name: "test"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/module/mod-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Module{
			ID: "mod-1", Name: "Updated", Description: "Updated module", Provider: "gcp",
			Source: "https://github.com/ex/mod",
		})
	})

	client := newTestClient(t, srv)
	mod, err := client.Modules.Update(context.Background(), "org-1", &terrakube.Module{
		ID: "mod-1", Name: "Updated", Description: "Updated module", Provider: "gcp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod.Name != "Updated" {
		t.Errorf("Name = %q, want %q", mod.Name, "Updated")
	}
	if mod.Provider != "gcp" {
		t.Errorf("Provider = %q, want %q", mod.Provider, "gcp")
	}
}

func TestModuleService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Modules.Update(context.Background(), "", &terrakube.Module{ID: "mod-1"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Modules.Update(context.Background(), "org-1", &terrakube.Module{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "module ID")
}

func TestModuleService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/module/mod-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Modules.Delete(context.Background(), "org-1", "mod-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModuleService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Modules.Delete(context.Background(), "", "mod-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestModuleService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Modules.Delete(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestModuleService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/module/mod-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Modules.Delete(context.Background(), "org-1", "mod-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestModuleService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/module/mod-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Module{ID: "mod-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Modules.Get(context.Background(), "org-1", "mod-1")
}
