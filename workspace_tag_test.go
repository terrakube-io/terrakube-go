package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func newTestWorkspaceTag() *terrakube.WorkspaceTag {
	return &terrakube.WorkspaceTag{
		ID:    "wstag-1",
		TagID: "tag-42",
	}
}

func TestWorkspaceTagService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/tag", func(w http.ResponseWriter, _ *http.Request) {
		tags := []*terrakube.WorkspaceTag{newTestWorkspaceTag()}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, tags)
	})

	client := newTestClient(t, srv)

	tags, err := client.WorkspaceTags.List(context.Background(), "org-1", "ws-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("got %d tags, want 1", len(tags))
	}
	if tags[0].TagID != "tag-42" {
		t.Errorf("TagID = %q, want %q", tags[0].TagID, "tag-42")
	}
}

func TestWorkspaceTagService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/tag", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[workspacetag]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.WorkspaceTag{})
	})

	client := newTestClient(t, srv)

	_, err := client.WorkspaceTags.List(context.Background(), "org-1", "ws-1", &terrakube.ListOptions{Filter: "tagId==tag-42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceTagService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.WorkspaceTags.List(context.Background(), "", "ws-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestWorkspaceTagService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.WorkspaceTags.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "workspaceID")
}

func TestWorkspaceTagService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/tag/wstag-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestWorkspaceTag())
	})

	client := newTestClient(t, srv)

	tag, err := client.WorkspaceTags.Get(context.Background(), "org-1", "ws-1", "wstag-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.ID != "wstag-1" {
		t.Errorf("ID = %q, want %q", tag.ID, "wstag-1")
	}
	if tag.TagID != "tag-42" {
		t.Errorf("TagID = %q, want %q", tag.TagID, "tag-42")
	}
}

func TestWorkspaceTagService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/tag/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "tag not found")
	})

	client := newTestClient(t, srv)

	_, err := client.WorkspaceTags.Get(context.Background(), "org-1", "ws-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestWorkspaceTagService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		wsID  string
		tagID string
		field string
	}{
		{"empty org ID", "", "ws-1", "wstag-1", "organizationID"},
		{"empty workspace ID", "org-1", "", "wstag-1", "workspaceID"},
		{"empty tag ID", "org-1", "ws-1", "", "tagID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.WorkspaceTags.Get(context.Background(), tt.orgID, tt.wsID, tt.tagID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestWorkspaceTagService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/tag", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestWorkspaceTag())
	})

	client := newTestClient(t, srv)

	input := &terrakube.WorkspaceTag{
		TagID: "tag-42",
	}
	tag, err := client.WorkspaceTags.Create(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.ID != "wstag-1" {
		t.Errorf("ID = %q, want %q", tag.ID, "wstag-1")
	}
}

func TestWorkspaceTagService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.WorkspaceTags.Create(context.Background(), "", "ws-1", &terrakube.WorkspaceTag{})
	assertValidationError(t, err, "organizationID")
}

func TestWorkspaceTagService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.WorkspaceTags.Create(context.Background(), "org-1", "", &terrakube.WorkspaceTag{})
	assertValidationError(t, err, "workspaceID")
}

func TestWorkspaceTagService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1/tag/wstag-1", func(w http.ResponseWriter, _ *http.Request) {
		tag := newTestWorkspaceTag()
		tag.TagID = "tag-99"
		testutil.WriteJSONAPI(t, w, http.StatusOK, tag)
	})

	client := newTestClient(t, srv)

	input := &terrakube.WorkspaceTag{
		ID:    "wstag-1",
		TagID: "tag-99",
	}
	tag, err := client.WorkspaceTags.Update(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.TagID != "tag-99" {
		t.Errorf("TagID = %q, want %q", tag.TagID, "tag-99")
	}
}

func TestWorkspaceTagService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.WorkspaceTags.Update(context.Background(), "", "ws-1", &terrakube.WorkspaceTag{ID: "wstag-1"})
	assertValidationError(t, err, "organizationID")
}

func TestWorkspaceTagService_Update_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.WorkspaceTags.Update(context.Background(), "org-1", "", &terrakube.WorkspaceTag{ID: "wstag-1"})
	assertValidationError(t, err, "workspaceID")
}

func TestWorkspaceTagService_Update_EmptyTagID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.WorkspaceTags.Update(context.Background(), "org-1", "ws-1", &terrakube.WorkspaceTag{ID: ""})
	assertValidationError(t, err, "tagID")
}

func TestWorkspaceTagService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/tag/wstag-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	err := client.WorkspaceTags.Delete(context.Background(), "org-1", "ws-1", "wstag-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceTagService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		wsID  string
		tagID string
		field string
	}{
		{"empty org ID", "", "ws-1", "wstag-1", "organizationID"},
		{"empty workspace ID", "org-1", "", "wstag-1", "workspaceID"},
		{"empty tag ID", "org-1", "ws-1", "", "tagID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.WorkspaceTags.Delete(context.Background(), tt.orgID, tt.wsID, tt.tagID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestWorkspaceTagService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/tag/wstag-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)

	err := client.WorkspaceTags.Delete(context.Background(), "org-1", "ws-1", "wstag-1")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestWorkspaceTagService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/tag/wstag-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestWorkspaceTag())
	})

	client := newTestClient(t, srv)

	_, err := client.WorkspaceTags.Get(context.Background(), "org-1", "ws-1", "wstag-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
