package terrakube_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

// allPermissions is a PermissionSet body with every flag the server sends.
const allPermissions = `{"manageState":true,"manageWorkspace":true,"manageModule":true,"manageProvider":true,
	"manageVcs":true,"manageTemplate":true,"manageCollection":true,"manageJob":true,"planJob":true,
	"approveJob":true,"managePermission":true,"managePolicies":true}`

func writeAccepted(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // what the server answers on these endpoints
	_, _ = io.WriteString(w, body)
}

func TestAccessService_CurrentTeams(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /access-token/v1/teams/current-teams", func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", auth)
		}
		writeAccepted(w, `{"groups":["TERRAKUBE_ADMIN","DEVELOPERS"]}`)
	})

	teams, err := newTestClient(t, srv).Access.CurrentTeams(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(teams) != 2 || teams[0] != "TERRAKUBE_ADMIN" || teams[1] != "DEVELOPERS" {
		t.Errorf("teams = %q", teams)
	}
}

func TestAccessService_CurrentTeams_Empty(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /access-token/v1/teams/current-teams", func(w http.ResponseWriter, _ *http.Request) {
		writeAccepted(w, `{"groups":null}`)
	})

	teams, err := newTestClient(t, srv).Access.CurrentTeams(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if teams == nil || len(teams) != 0 {
		t.Errorf("teams = %#v, want an empty, non-nil slice", teams)
	}
}

func TestAccessService_Permissions(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /access-token/v1/teams/permissions/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		writeAccepted(w, allPermissions)
	})
	srv.HandleFunc("GET /access-token/v1/teams/permissions/organization/org-2", func(w http.ResponseWriter, _ *http.Request) {
		writeAccepted(w, `{"manageState":false,"manageWorkspace":true,"manageModule":false,"manageProvider":false,
			"manageVcs":false,"manageTemplate":false,"manageCollection":false,"manageJob":false,"planJob":true,
			"approveJob":false,"managePermission":false,"managePolicies":false}`)
	})
	client := newTestClient(t, srv)
	ctx := context.Background()

	p, err := client.Access.Permissions(ctx, "org-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := terrakube.Permissions{
		ManageState: true, ManageWorkspace: true, ManageModule: true, ManageProvider: true,
		ManageVcs: true, ManageTemplate: true, ManageCollection: true, ManageJob: true,
		PlanJob: true, ApproveJob: true, ManagePermission: true, ManagePolicies: true,
	}
	if *p != want {
		t.Errorf("org-1 = %+v, want every permission", *p)
	}

	p, err = client.Access.Permissions(ctx, "org-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *p != (terrakube.Permissions{ManageWorkspace: true, PlanJob: true}) {
		t.Errorf("org-2 = %+v, want only ManageWorkspace and PlanJob", *p)
	}

	_, err = client.Access.Permissions(ctx, "")
	assertValidationError(t, err, "organizationID")
}

func TestAccessService_WorkspacePermissions(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /access-token/v1/teams/permissions/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		writeAccepted(w, `{"manageState":true,"manageWorkspace":false,"manageModule":false,"manageProvider":false,
			"manageVcs":false,"manageTemplate":false,"manageCollection":false,"manageJob":true,"planJob":true,
			"approveJob":true,"managePermission":false,"managePolicies":false}`)
	})
	client := newTestClient(t, srv)
	ctx := context.Background()

	p, err := client.Access.WorkspacePermissions(ctx, "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *p != (terrakube.Permissions{ManageState: true, ManageJob: true, PlanJob: true, ApproveJob: true}) {
		t.Errorf("permissions = %+v", *p)
	}

	_, err = client.Access.WorkspacePermissions(ctx, "", "ws-1")
	assertValidationError(t, err, "organizationID")
	_, err = client.Access.WorkspacePermissions(ctx, "org-1", "")
	assertValidationError(t, err, "workspaceID")
}

func TestAccessService_Errors(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /access-token/v1/teams/current-teams", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	srv.HandleFunc("GET /access-token/v1/teams/permissions/organization/old", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, springNotFound)
	})
	srv.HandleFunc("GET /access-token/v1/teams/permissions/organization/org-1/workspace/denied", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	srv.HandleFunc("GET /access-token/v1/teams/permissions/organization/bad", func(w http.ResponseWriter, _ *http.Request) {
		writeAccepted(w, "[")
	})
	srv.HandleFunc("GET /access-token/v1/teams/permissions/organization/bad/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		writeAccepted(w, "[")
	})
	client := newTestClient(t, srv)
	ctx := context.Background()

	if _, err := client.Access.CurrentTeams(ctx); !terrakube.IsUnauthorized(err) {
		t.Errorf("CurrentTeams: err = %v, want 401", err)
	}
	if _, err := client.Access.Permissions(ctx, "old"); !errors.Is(err, terrakube.ErrUnsupported) {
		t.Errorf("old server: err = %v, want ErrUnsupported", err)
	}
	if _, err := client.Access.WorkspacePermissions(ctx, "org-1", "denied"); !terrakube.IsForbidden(err) {
		t.Errorf("denied: err = %v, want 403", err)
	}
	if _, err := client.Access.Permissions(ctx, "bad"); err == nil {
		t.Error("bad body: err = nil")
	}
	if _, err := client.Access.WorkspacePermissions(ctx, "bad", "ws-1"); err == nil {
		t.Error("bad body: err = nil")
	}
}

func TestAccessService_CurrentTeams_BadBody(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /access-token/v1/teams/current-teams", func(w http.ResponseWriter, _ *http.Request) {
		writeAccepted(w, "[")
	})
	if _, err := newTestClient(t, srv).Access.CurrentTeams(context.Background()); err == nil {
		t.Error("err = nil, want a decoding error")
	}
}

// TestRawServices_RequestError covers a request that cannot be built (a nil
// context): every method returns the error without sending anything.
func TestRawServices_RequestError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)
	var ctx context.Context // nil

	if _, err := client.PATs.Create(ctx, 1, "x"); err == nil {
		t.Error("PATs.Create: err = nil")
	}
	if _, err := client.PATs.List(ctx); err == nil {
		t.Error("PATs.List: err = nil")
	}
	if err := client.PATs.Delete(ctx, "p-1"); err == nil {
		t.Error("PATs.Delete: err = nil")
	}
	if _, err := client.Access.CurrentTeams(ctx); err == nil {
		t.Error("Access.CurrentTeams: err = nil")
	}
	if _, err := client.Access.Permissions(ctx, "org-1"); err == nil {
		t.Error("Access.Permissions: err = nil")
	}
}
