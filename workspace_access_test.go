package terrakube_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/google/jsonapi"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func newTestWorkspaceAccess() *terrakube.WorkspaceAccess {
	return &terrakube.WorkspaceAccess{
		ID:              "access-1",
		ManageState:     true,
		ManageWorkspace: false,
		ManageJob:       true,
		Name:            "team-admins",
	}
}

// --- List ---

func TestWorkspaceAccessService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/access", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.WorkspaceAccess{newTestWorkspaceAccess()})
	})

	client := newTestClient(t, srv)
	access, err := client.WorkspaceAccess.List(context.Background(), "org-1", "ws-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(access) != 1 {
		t.Fatalf("got %d access entries, want 1", len(access))
	}
	if access[0].Name != "team-admins" {
		t.Errorf("Name = %q, want %q", access[0].Name, "team-admins")
	}
}

func TestWorkspaceAccessService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/access", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[access]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.WorkspaceAccess{})
	})

	client := newTestClient(t, srv)
	_, err := client.WorkspaceAccess.List(context.Background(), "org-1", "ws-1", &terrakube.ListOptions{Filter: "name==admins"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceAccessService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceAccess.List(context.Background(), "", "ws-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestWorkspaceAccessService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceAccess.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "workspaceID")
}

// --- Get ---

func TestWorkspaceAccessService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/access/access-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestWorkspaceAccess())
	})

	client := newTestClient(t, srv)
	a, err := client.WorkspaceAccess.Get(context.Background(), "org-1", "ws-1", "access-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ID != "access-1" {
		t.Errorf("ID = %q, want %q", a.ID, "access-1")
	}
	if a.Name != "team-admins" {
		t.Errorf("Name = %q, want %q", a.Name, "team-admins")
	}
	if !a.ManageState {
		t.Error("ManageState = false, want true")
	}
}

func TestWorkspaceAccessService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/access/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "access not found")
	})

	client := newTestClient(t, srv)
	_, err := client.WorkspaceAccess.Get(context.Background(), "org-1", "ws-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestWorkspaceAccessService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		wsID  string
		id    string
		field string
	}{
		{"empty org ID", "", "ws-1", "access-1", "organizationID"},
		{"empty workspace ID", "org-1", "", "access-1", "workspaceID"},
		{"empty access ID", "org-1", "ws-1", "", "accessID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.WorkspaceAccess.Get(context.Background(), tt.orgID, tt.wsID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

// --- Create ---

func TestWorkspaceAccessService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/access", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestWorkspaceAccess())
	})

	client := newTestClient(t, srv)
	input := &terrakube.WorkspaceAccess{
		Name:            "team-admins",
		ManageState:     true,
		ManageWorkspace: false,
		ManageJob:       true,
	}
	a, err := client.WorkspaceAccess.Create(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ID != "access-1" {
		t.Errorf("ID = %q, want %q", a.ID, "access-1")
	}
}

func TestWorkspaceAccessService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceAccess.Create(context.Background(), "", "ws-1", &terrakube.WorkspaceAccess{})
	assertValidationError(t, err, "organizationID")
}

func TestWorkspaceAccessService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceAccess.Create(context.Background(), "org-1", "", &terrakube.WorkspaceAccess{})
	assertValidationError(t, err, "workspaceID")
}

// --- Update ---

func TestWorkspaceAccessService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1/access/access-1", func(w http.ResponseWriter, _ *http.Request) {
		updated := newTestWorkspaceAccess()
		updated.ManageWorkspace = true
		testutil.WriteJSONAPI(t, w, http.StatusOK, updated)
	})

	client := newTestClient(t, srv)
	input := &terrakube.WorkspaceAccess{
		ID:              "access-1",
		Name:            "team-admins",
		ManageState:     true,
		ManageWorkspace: true,
		ManageJob:       true,
	}
	a, err := client.WorkspaceAccess.Update(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.ManageWorkspace {
		t.Error("ManageWorkspace = false, want true")
	}
}

func TestWorkspaceAccessService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceAccess.Update(context.Background(), "", "ws-1", &terrakube.WorkspaceAccess{ID: "access-1"})
	assertValidationError(t, err, "organizationID")
}

func TestWorkspaceAccessService_Update_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceAccess.Update(context.Background(), "org-1", "", &terrakube.WorkspaceAccess{ID: "access-1"})
	assertValidationError(t, err, "workspaceID")
}

func TestWorkspaceAccessService_Update_EmptyAccessID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.WorkspaceAccess.Update(context.Background(), "org-1", "ws-1", &terrakube.WorkspaceAccess{ID: ""})
	assertValidationError(t, err, "accessID")
}

// --- Delete ---

func TestWorkspaceAccessService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/access/access-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.WorkspaceAccess.Delete(context.Background(), "org-1", "ws-1", "access-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkspaceAccessService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		wsID  string
		id    string
		field string
	}{
		{"empty org ID", "", "ws-1", "access-1", "organizationID"},
		{"empty workspace ID", "org-1", "", "access-1", "workspaceID"},
		{"empty access ID", "org-1", "ws-1", "", "accessID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.WorkspaceAccess.Delete(context.Background(), tt.orgID, tt.wsID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestWorkspaceAccessService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/access/access-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.WorkspaceAccess.Delete(context.Background(), "org-1", "ws-1", "access-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- Auth ---

func TestWorkspaceAccessService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/access/access-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestWorkspaceAccess())
	})

	client := newTestClient(t, srv)
	_, _ = client.WorkspaceAccess.Get(context.Background(), "org-1", "ws-1", "access-1")
}

// --- Boolean serialization ---

func TestWorkspaceAccess_BooleanFalseNotDropped(t *testing.T) {
	t.Parallel()

	a := &terrakube.WorkspaceAccess{
		ID:              "access-1",
		ManageState:     false,
		ManageWorkspace: false,
		ManageJob:       false,
		Name:            "team-readonly",
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, a); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.Bytes()
	if !bytes.Contains(payload, []byte(`"manageState"`)) {
		t.Error("ManageState=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
	if !bytes.Contains(payload, []byte(`"manageWorkspace"`)) {
		t.Error("ManageWorkspace=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
	if !bytes.Contains(payload, []byte(`"manageJob"`)) {
		t.Error("ManageJob=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
}

func TestWorkspaceAccess_BooleanTruePreserved(t *testing.T) {
	t.Parallel()

	a := &terrakube.WorkspaceAccess{
		ID:              "access-1",
		ManageState:     true,
		ManageWorkspace: true,
		ManageJob:       true,
		Name:            "team-admins",
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, a); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.Bytes()
	if !bytes.Contains(payload, []byte(`"manageState"`)) {
		t.Error("ManageState=true was dropped from JSON:API payload")
	}
	if !bytes.Contains(payload, []byte(`"manageWorkspace"`)) {
		t.Error("ManageWorkspace=true was dropped from JSON:API payload")
	}
	if !bytes.Contains(payload, []byte(`"manageJob"`)) {
		t.Error("ManageJob=true was dropped from JSON:API payload")
	}
}

func TestWorkspaceAccessService_Create_BooleanFalseSerialized(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/access", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		assertJSONBoolField(t, body, "manageState", false)
		assertJSONBoolField(t, body, "manageWorkspace", false)
		assertJSONBoolField(t, body, "manageJob", false)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.WorkspaceAccess{
			ID:   "access-new",
			Name: "readonly",
		})
	})

	client := newTestClient(t, srv)
	_, err := client.WorkspaceAccess.Create(context.Background(), "org-1", "ws-1", &terrakube.WorkspaceAccess{
		Name:            "readonly",
		ManageState:     false,
		ManageWorkspace: false,
		ManageJob:       false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
