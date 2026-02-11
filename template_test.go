package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func TestTemplateService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	desc := "default template"
	version := "1.0.0"
	srv.HandleFunc("GET /api/v1/organization/org-1/template", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Template{
			{ID: "tmpl-1", Name: "plan-only", Description: &desc, Version: &version, Content: "plan"},
			{ID: "tmpl-2", Name: "apply-all", Content: "apply"},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.Templates.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "tmpl-1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "tmpl-1")
	}
	if items[0].Name != "plan-only" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "plan-only")
	}
	if items[0].Content != "plan" {
		t.Errorf("items[0].Content = %q, want %q", items[0].Content, "plan")
	}
}

func TestTemplateService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/template", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[template]")
		if filter != "name==plan-only" {
			t.Errorf("filter = %q, want %q", filter, "name==plan-only")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Template{
			{ID: "tmpl-1", Name: "plan-only"},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.Templates.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==plan-only"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
}

func TestTemplateService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Templates.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestTemplateService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	desc := "my template"
	version := "2.0.0"
	srv.HandleFunc("GET /api/v1/organization/org-1/template/tmpl-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Template{
			ID:          "tmpl-1",
			Name:        "plan-only",
			Description: &desc,
			Version:     &version,
			Content:     "plan",
		})
	})

	c := newTestClient(t, srv)

	tmpl, err := c.Templates.Get(context.Background(), "org-1", "tmpl-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tmpl.ID != "tmpl-1" {
		t.Errorf("ID = %q, want %q", tmpl.ID, "tmpl-1")
	}
	if tmpl.Name != "plan-only" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "plan-only")
	}
	if tmpl.Content != "plan" {
		t.Errorf("Content = %q, want %q", tmpl.Content, "plan")
	}
	if tmpl.Description == nil || *tmpl.Description != "my template" {
		t.Errorf("Description = %v, want %q", tmpl.Description, "my template")
	}
	if tmpl.Version == nil || *tmpl.Version != "2.0.0" {
		t.Errorf("Version = %v, want %q", tmpl.Version, "2.0.0")
	}
}

func TestTemplateService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/template/nonexistent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "template not found")
	})

	c := newTestClient(t, srv)

	_, err := c.Templates.Get(context.Background(), "org-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestTemplateService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Templates.Get(context.Background(), "", "tmpl-1")
	assertValidationError(t, err, "organizationID")
}

func TestTemplateService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Templates.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "templateID")
}

func TestTemplateService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org-1/template", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Template{
			ID:      "tmpl-new",
			Name:    "new-template",
			Content: "apply -auto-approve",
		})
	})

	c := newTestClient(t, srv)

	tmpl, err := c.Templates.Create(context.Background(), "org-1", &terrakube.Template{
		Name:    "new-template",
		Content: "apply -auto-approve",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tmpl.ID != "tmpl-new" {
		t.Errorf("ID = %q, want %q", tmpl.ID, "tmpl-new")
	}
	if tmpl.Name != "new-template" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "new-template")
	}
}

func TestTemplateService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Templates.Create(context.Background(), "", &terrakube.Template{Name: "test"})
	assertValidationError(t, err, "organizationID")
}

func TestTemplateService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org-1/template/tmpl-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Template{
			ID:      "tmpl-1",
			Name:    "updated-template",
			Content: "plan -detailed-exitcode",
		})
	})

	c := newTestClient(t, srv)

	tmpl, err := c.Templates.Update(context.Background(), "org-1", &terrakube.Template{
		ID:      "tmpl-1",
		Name:    "updated-template",
		Content: "plan -detailed-exitcode",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tmpl.Name != "updated-template" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "updated-template")
	}
}

func TestTemplateService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Templates.Update(context.Background(), "", &terrakube.Template{ID: "tmpl-1"})
	assertValidationError(t, err, "organizationID")
}

func TestTemplateService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Templates.Update(context.Background(), "org-1", &terrakube.Template{ID: ""})
	assertValidationError(t, err, "templateID")
}

func TestTemplateService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/template/tmpl-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	c := newTestClient(t, srv)

	err := c.Templates.Delete(context.Background(), "org-1", "tmpl-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTemplateService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.Templates.Delete(context.Background(), "", "tmpl-1")
	assertValidationError(t, err, "organizationID")
}

func TestTemplateService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.Templates.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "templateID")
}

func TestTemplateService_Delete_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/template/tmpl-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	c := newTestClient(t, srv)

	err := c.Templates.Delete(context.Background(), "org-1", "tmpl-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	var apiErr *terrakube.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestTemplateService_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/template", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Template{})
	})

	c := newTestClient(t, srv)

	_, err := c.Templates.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
