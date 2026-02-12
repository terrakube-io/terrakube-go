package terrakube_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestOrganizationService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		desc := "First org"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Organization{
			{ID: "org-1", Name: "Alpha", Description: &desc, ExecutionMode: "remote"},
			{ID: "org-2", Name: "Beta", ExecutionMode: "local"},
		})
	})

	client := newTestClient(t, srv)
	orgs, err := client.Organizations.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgs) != 2 {
		t.Fatalf("got %d organizations, want 2", len(orgs))
	}
	if orgs[0].Name != "Alpha" {
		t.Errorf("Name = %q, want %q", orgs[0].Name, "Alpha")
	}
}

func TestOrganizationService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Organization{
			{ID: "org-1", Name: "Filtered"},
		})
	})

	client := newTestClient(t, srv)
	orgs, err := client.Organizations.List(context.Background(), &terrakube.ListOptions{Filter: "name==Filtered"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgs) != 1 {
		t.Fatalf("got %d organizations, want 1", len(orgs))
	}
}

func TestOrganizationService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Test org"
		icon := "icon-url"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Organization{
			ID: "org-1", Name: "Alpha", Description: &desc, ExecutionMode: "remote", Disabled: false, Icon: &icon,
		})
	})

	client := newTestClient(t, srv)
	org, err := client.Organizations.Get(context.Background(), "org-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.ID != "org-1" {
		t.Errorf("ID = %q, want %q", org.ID, "org-1")
	}
	if org.Name != "Alpha" {
		t.Errorf("Name = %q, want %q", org.Name, "Alpha")
	}
	if org.ExecutionMode != "remote" {
		t.Errorf("ExecutionMode = %q, want %q", org.ExecutionMode, "remote")
	}
	if org.Description == nil || *org.Description != "Test org" {
		t.Errorf("Description = %v, want %q", org.Description, "Test org")
	}
	if org.Icon == nil || *org.Icon != "icon-url" {
		t.Errorf("Icon = %v, want %q", org.Icon, "icon-url")
	}
}

func TestOrganizationService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Organizations.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestOrganizationService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "organization not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Organizations.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestOrganizationService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		desc := "New org"
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Organization{
			ID: "org-new", Name: "Created", Description: &desc, ExecutionMode: "remote",
		})
	})

	client := newTestClient(t, srv)
	desc := "New org"
	org, err := client.Organizations.Create(context.Background(), &terrakube.Organization{
		Name: "Created", Description: &desc, ExecutionMode: "remote",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.ID != "org-new" {
		t.Errorf("ID = %q, want %q", org.ID, "org-new")
	}
	if org.Name != "Created" {
		t.Errorf("Name = %q, want %q", org.Name, "Created")
	}
}

func TestOrganizationService_Create_BooleanFalse(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assertJSONBoolField(t, body, "disabled", false)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Organization{ID: "org-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, err := client.Organizations.Create(context.Background(), &terrakube.Organization{Name: "test", Disabled: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrganizationService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Updated"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Organization{
			ID: "org-1", Name: "Updated", Description: &desc, ExecutionMode: "local",
		})
	})

	client := newTestClient(t, srv)
	desc := "Updated"
	org, err := client.Organizations.Update(context.Background(), &terrakube.Organization{
		ID: "org-1", Name: "Updated", Description: &desc, ExecutionMode: "local",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Name != "Updated" {
		t.Errorf("Name = %q, want %q", org.Name, "Updated")
	}
	if org.ExecutionMode != "local" {
		t.Errorf("ExecutionMode = %q, want %q", org.ExecutionMode, "local")
	}
}

func TestOrganizationService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Organizations.Update(context.Background(), &terrakube.Organization{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestOrganizationService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Organizations.Delete(context.Background(), "org-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrganizationService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Organizations.Delete(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "id")
}

func TestOrganizationService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Organizations.Delete(context.Background(), "org-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestOrganizationService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Organization{ID: "org-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Organizations.Get(context.Background(), "org-1")
}
