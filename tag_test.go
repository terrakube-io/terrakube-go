package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestTagService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/tag", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Tag{
			{ID: "tag-1", Name: "production"},
			{ID: "tag-2", Name: "staging"},
		})
	})

	c := newTestClient(t, srv)

	tags, err := c.Tags.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2", len(tags))
	}
	if tags[0].ID != "tag-1" {
		t.Errorf("tags[0].ID = %q, want %q", tags[0].ID, "tag-1")
	}
	if tags[0].Name != "production" {
		t.Errorf("tags[0].Name = %q, want %q", tags[0].Name, "production")
	}
}

func TestTagService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/tag", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[tag]")
		if filter != "name==production" {
			t.Errorf("filter = %q, want %q", filter, "name==production")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Tag{
			{ID: "tag-1", Name: "production"},
		})
	})

	c := newTestClient(t, srv)

	tags, err := c.Tags.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==production"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tags) != 1 {
		t.Fatalf("got %d tags, want 1", len(tags))
	}
}

func TestTagService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Tags.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestTagService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/tag/tag-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Tag{
			ID:   "tag-1",
			Name: "production",
		})
	})

	c := newTestClient(t, srv)

	tag, err := c.Tags.Get(context.Background(), "org-1", "tag-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tag.ID != "tag-1" {
		t.Errorf("ID = %q, want %q", tag.ID, "tag-1")
	}
	if tag.Name != "production" {
		t.Errorf("Name = %q, want %q", tag.Name, "production")
	}
}

func TestTagService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/tag/nonexistent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "tag not found")
	})

	c := newTestClient(t, srv)

	_, err := c.Tags.Get(context.Background(), "org-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestTagService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Tags.Get(context.Background(), "", "tag-1")
	assertValidationError(t, err, "organizationID")
}

func TestTagService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Tags.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "tagID")
}

func TestTagService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org-1/tag", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Tag{
			ID:   "tag-new",
			Name: "development",
		})
	})

	c := newTestClient(t, srv)

	tag, err := c.Tags.Create(context.Background(), "org-1", &terrakube.Tag{
		Name: "development",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tag.ID != "tag-new" {
		t.Errorf("ID = %q, want %q", tag.ID, "tag-new")
	}
	if tag.Name != "development" {
		t.Errorf("Name = %q, want %q", tag.Name, "development")
	}
}

func TestTagService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Tags.Create(context.Background(), "", &terrakube.Tag{Name: "test"})
	assertValidationError(t, err, "organizationID")
}

func TestTagService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org-1/tag/tag-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Tag{
			ID:   "tag-1",
			Name: "updated-tag",
		})
	})

	c := newTestClient(t, srv)

	tag, err := c.Tags.Update(context.Background(), "org-1", &terrakube.Tag{
		ID:   "tag-1",
		Name: "updated-tag",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tag.Name != "updated-tag" {
		t.Errorf("Name = %q, want %q", tag.Name, "updated-tag")
	}
}

func TestTagService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Tags.Update(context.Background(), "", &terrakube.Tag{ID: "tag-1"})
	assertValidationError(t, err, "organizationID")
}

func TestTagService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Tags.Update(context.Background(), "org-1", &terrakube.Tag{ID: ""})
	assertValidationError(t, err, "tagID")
}

func TestTagService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/tag/tag-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	c := newTestClient(t, srv)

	err := c.Tags.Delete(context.Background(), "org-1", "tag-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.Tags.Delete(context.Background(), "", "tag-1")
	assertValidationError(t, err, "organizationID")
}

func TestTagService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.Tags.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "tagID")
}

func TestTagService_Delete_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/tag/tag-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	c := newTestClient(t, srv)

	err := c.Tags.Delete(context.Background(), "org-1", "tag-1")
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

func TestTagService_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/tag", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Tag{})
	})

	c := newTestClient(t, srv)

	_, err := c.Tags.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
