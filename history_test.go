package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestHistory() *terrakube.History {
	return &terrakube.History{
		ID:           "hist-1",
		JobReference: "job-ref-1",
		Output:       "apply complete",
	}
}

func TestHistoryService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/history", func(w http.ResponseWriter, _ *http.Request) {
		histories := []*terrakube.History{newTestHistory()}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, histories)
	})

	client := newTestClient(t, srv)

	histories, err := client.History.List(context.Background(), "org-1", "ws-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(histories) != 1 {
		t.Fatalf("got %d histories, want 1", len(histories))
	}
	if histories[0].JobReference != "job-ref-1" {
		t.Errorf("JobReference = %q, want %q", histories[0].JobReference, "job-ref-1")
	}
}

func TestHistoryService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/history", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[history]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.History{})
	})

	client := newTestClient(t, srv)

	_, err := client.History.List(context.Background(), "org-1", "ws-1", &terrakube.ListOptions{Filter: "output==success"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHistoryService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.History.List(context.Background(), "", "ws-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestHistoryService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.History.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "workspaceID")
}

func TestHistoryService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/history/hist-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestHistory())
	})

	client := newTestClient(t, srv)

	h, err := client.History.Get(context.Background(), "org-1", "ws-1", "hist-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.ID != "hist-1" {
		t.Errorf("ID = %q, want %q", h.ID, "hist-1")
	}
	if h.JobReference != "job-ref-1" {
		t.Errorf("JobReference = %q, want %q", h.JobReference, "job-ref-1")
	}
}

func TestHistoryService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/history/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "history not found")
	})

	client := newTestClient(t, srv)

	_, err := client.History.Get(context.Background(), "org-1", "ws-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestHistoryService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		wsID  string
		hID   string
		field string
	}{
		{"empty org ID", "", "ws-1", "hist-1", "organizationID"},
		{"empty workspace ID", "org-1", "", "hist-1", "workspaceID"},
		{"empty history ID", "org-1", "ws-1", "", "historyID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.History.Get(context.Background(), tt.orgID, tt.wsID, tt.hID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestHistoryService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/history", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestHistory())
	})

	client := newTestClient(t, srv)

	input := &terrakube.History{
		JobReference: "job-ref-1",
		Output:       "apply complete",
	}
	h, err := client.History.Create(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.ID != "hist-1" {
		t.Errorf("ID = %q, want %q", h.ID, "hist-1")
	}
}

func TestHistoryService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.History.Create(context.Background(), "", "ws-1", &terrakube.History{})
	assertValidationError(t, err, "organizationID")
}

func TestHistoryService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.History.Create(context.Background(), "org-1", "", &terrakube.History{})
	assertValidationError(t, err, "workspaceID")
}

func TestHistoryService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1/history/hist-1", func(w http.ResponseWriter, _ *http.Request) {
		h := newTestHistory()
		h.Output = "updated output"
		testutil.WriteJSONAPI(t, w, http.StatusOK, h)
	})

	client := newTestClient(t, srv)

	input := &terrakube.History{
		ID:     "hist-1",
		Output: "updated output",
	}
	h, err := client.History.Update(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Output != "updated output" {
		t.Errorf("Output = %q, want %q", h.Output, "updated output")
	}
}

func TestHistoryService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.History.Update(context.Background(), "", "ws-1", &terrakube.History{ID: "hist-1"})
	assertValidationError(t, err, "organizationID")
}

func TestHistoryService_Update_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.History.Update(context.Background(), "org-1", "", &terrakube.History{ID: "hist-1"})
	assertValidationError(t, err, "workspaceID")
}

func TestHistoryService_Update_EmptyHistoryID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.History.Update(context.Background(), "org-1", "ws-1", &terrakube.History{ID: ""})
	assertValidationError(t, err, "historyID")
}

func TestHistoryService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/history/hist-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	err := client.History.Delete(context.Background(), "org-1", "ws-1", "hist-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHistoryService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		wsID  string
		hID   string
		field string
	}{
		{"empty org ID", "", "ws-1", "hist-1", "organizationID"},
		{"empty workspace ID", "org-1", "", "hist-1", "workspaceID"},
		{"empty history ID", "org-1", "ws-1", "", "historyID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.History.Delete(context.Background(), tt.orgID, tt.wsID, tt.hID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestHistoryService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/history/hist-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)

	err := client.History.Delete(context.Background(), "org-1", "ws-1", "hist-1")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestHistoryService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/history/hist-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestHistory())
	})

	client := newTestClient(t, srv)

	_, err := client.History.Get(context.Background(), "org-1", "ws-1", "hist-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
