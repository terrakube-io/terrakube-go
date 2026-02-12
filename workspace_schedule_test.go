package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestWorkspaceSchedule() *terrakube.WorkspaceSchedule {
	return &terrakube.WorkspaceSchedule{
		ID:         "sched-1",
		Schedule:   "0 */6 * * *",
		TemplateID: "tmpl-1",
	}
}

// --- List ---

func TestWorkspaceScheduleService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/workspace/ws-1/schedule", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.WorkspaceSchedule{newTestWorkspaceSchedule()})
	})

	client := newTestClient(t, srv)
	schedules, err := client.WorkspaceSchedules.List(context.Background(), "ws-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("got %d schedules, want 1", len(schedules))
	}
	if schedules[0].Schedule != "0 */6 * * *" {
		t.Errorf("Schedule = %q, want %q", schedules[0].Schedule, "0 */6 * * *")
	}
}

func TestWorkspaceScheduleService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/workspace/ws-1/schedule", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[schedule]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.WorkspaceSchedule{})
	})

	client := newTestClient(t, srv)
	_, err := client.WorkspaceSchedules.List(context.Background(), "ws-1", &terrakube.ListOptions{Filter: "id==sched-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceScheduleService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceSchedules.List(context.Background(), "", nil)
	assertValidationError(t, err, "workspaceID")
}

// --- Get ---

func TestWorkspaceScheduleService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/workspace/ws-1/schedule/sched-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestWorkspaceSchedule())
	})

	client := newTestClient(t, srv)
	s, err := client.WorkspaceSchedules.Get(context.Background(), "ws-1", "sched-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "sched-1" {
		t.Errorf("ID = %q, want %q", s.ID, "sched-1")
	}
	if s.TemplateID != "tmpl-1" {
		t.Errorf("TemplateID = %q, want %q", s.TemplateID, "tmpl-1")
	}
}

func TestWorkspaceScheduleService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/workspace/ws-1/schedule/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "schedule not found")
	})

	client := newTestClient(t, srv)
	_, err := client.WorkspaceSchedules.Get(context.Background(), "ws-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestWorkspaceScheduleService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		wsID  string
		id    string
		field string
	}{
		{"empty workspace ID", "", "sched-1", "workspaceID"},
		{"empty schedule ID", "ws-1", "", "scheduleID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.WorkspaceSchedules.Get(context.Background(), tt.wsID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

// --- Create ---

func TestWorkspaceScheduleService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/workspace/ws-1/schedule", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestWorkspaceSchedule())
	})

	client := newTestClient(t, srv)
	input := &terrakube.WorkspaceSchedule{
		Schedule:   "0 */6 * * *",
		TemplateID: "tmpl-1",
	}
	s, err := client.WorkspaceSchedules.Create(context.Background(), "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "sched-1" {
		t.Errorf("ID = %q, want %q", s.ID, "sched-1")
	}
}

func TestWorkspaceScheduleService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceSchedules.Create(context.Background(), "", &terrakube.WorkspaceSchedule{})
	assertValidationError(t, err, "workspaceID")
}

// --- Update ---

func TestWorkspaceScheduleService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/workspace/ws-1/schedule/sched-1", func(w http.ResponseWriter, _ *http.Request) {
		updated := newTestWorkspaceSchedule()
		updated.Schedule = "0 0 * * *"
		testutil.WriteJSONAPI(t, w, http.StatusOK, updated)
	})

	client := newTestClient(t, srv)
	input := &terrakube.WorkspaceSchedule{
		ID:         "sched-1",
		Schedule:   "0 0 * * *",
		TemplateID: "tmpl-1",
	}
	s, err := client.WorkspaceSchedules.Update(context.Background(), "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Schedule != "0 0 * * *" {
		t.Errorf("Schedule = %q, want %q", s.Schedule, "0 0 * * *")
	}
}

func TestWorkspaceScheduleService_Update_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceSchedules.Update(context.Background(), "", &terrakube.WorkspaceSchedule{ID: "sched-1"})
	assertValidationError(t, err, "workspaceID")
}

func TestWorkspaceScheduleService_Update_EmptyScheduleID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceSchedules.Update(context.Background(), "ws-1", &terrakube.WorkspaceSchedule{ID: ""})
	assertValidationError(t, err, "scheduleID")
}

// --- Delete ---

func TestWorkspaceScheduleService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/workspace/ws-1/schedule/sched-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.WorkspaceSchedules.Delete(context.Background(), "ws-1", "sched-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceScheduleService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		wsID  string
		id    string
		field string
	}{
		{"empty workspace ID", "", "sched-1", "workspaceID"},
		{"empty schedule ID", "ws-1", "", "scheduleID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.WorkspaceSchedules.Delete(context.Background(), tt.wsID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestWorkspaceScheduleService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/workspace/ws-1/schedule/sched-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.WorkspaceSchedules.Delete(context.Background(), "ws-1", "sched-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- Auth ---

func TestWorkspaceScheduleService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/workspace/ws-1/schedule/sched-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestWorkspaceSchedule())
	})

	client := newTestClient(t, srv)
	_, _ = client.WorkspaceSchedules.Get(context.Background(), "ws-1", "sched-1")
}
