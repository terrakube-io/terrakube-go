package terrakube_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
	"github.com/google/jsonapi"
)

func TestTeamService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/team", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Team{
			{ID: "team-1", Name: "admins", ManageState: true},
			{ID: "team-2", Name: "readers", ManageState: false},
		})
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	teams, err := c.Teams.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(teams) != 2 {
		t.Fatalf("got %d teams, want 2", len(teams))
	}
	if teams[0].ID != "team-1" {
		t.Errorf("teams[0].ID = %q, want %q", teams[0].ID, "team-1")
	}
	if teams[0].Name != "admins" {
		t.Errorf("teams[0].Name = %q, want %q", teams[0].Name, "admins")
	}
	if !teams[0].ManageState {
		t.Error("teams[0].ManageState should be true")
	}
}

func TestTeamService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/team", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[team]")
		if filter != "name==admins" {
			t.Errorf("filter = %q, want %q", filter, "name==admins")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Team{
			{ID: "team-1", Name: "admins"},
		})
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	teams, err := c.Teams.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==admins"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(teams) != 1 {
		t.Fatalf("got %d teams, want 1", len(teams))
	}
}

func TestTeamService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.List(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}

	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "orgID" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "orgID")
	}
}

func TestTeamService_List_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/team", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.List(context.Background(), "org-1", nil)
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

func TestTeamService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/team/team-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Team{
			ID:              "team-1",
			Name:            "admins",
			ManageState:     true,
			ManageWorkspace: true,
		})
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	team, err := c.Teams.Get(context.Background(), "org-1", "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.ID != "team-1" {
		t.Errorf("ID = %q, want %q", team.ID, "team-1")
	}
	if team.Name != "admins" {
		t.Errorf("Name = %q, want %q", team.Name, "admins")
	}
	if !team.ManageState {
		t.Error("ManageState should be true")
	}
	if !team.ManageWorkspace {
		t.Error("ManageWorkspace should be true")
	}
}

func TestTeamService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.Get(context.Background(), "", "team-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "orgID" {
		t.Errorf("Field = %q, want %q", valErr.Field, "orgID")
	}
}

func TestTeamService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.Get(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty id")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "id" {
		t.Errorf("Field = %q, want %q", valErr.Field, "id")
	}
}

func TestTeamService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/team/nonexistent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "team not found")
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.Get(context.Background(), "org-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestTeamService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org-1/team", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Team{
			ID:          "team-new",
			Name:        "developers",
			ManageState: true,
		})
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	team, err := c.Teams.Create(context.Background(), "org-1", &terrakube.Team{
		Name:        "developers",
		ManageState: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.ID != "team-new" {
		t.Errorf("ID = %q, want %q", team.ID, "team-new")
	}
	if team.Name != "developers" {
		t.Errorf("Name = %q, want %q", team.Name, "developers")
	}
}

func TestTeamService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.Create(context.Background(), "", &terrakube.Team{Name: "test"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "orgID" {
		t.Errorf("Field = %q, want %q", valErr.Field, "orgID")
	}
}

func TestTeamService_Create_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org-1/team", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusConflict, "team already exists")
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.Create(context.Background(), "org-1", &terrakube.Team{Name: "test"})
	if err == nil {
		t.Fatal("expected error for 409 response")
	}
	if !terrakube.IsConflict(err) {
		t.Errorf("IsConflict() = false, want true")
	}
}

func TestTeamService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org-1/team/team-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Team{
			ID:          "team-1",
			Name:        "updated-admins",
			ManageState: false,
		})
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	team, err := c.Teams.Update(context.Background(), "org-1", &terrakube.Team{
		ID:          "team-1",
		Name:        "updated-admins",
		ManageState: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.Name != "updated-admins" {
		t.Errorf("Name = %q, want %q", team.Name, "updated-admins")
	}
}

func TestTeamService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.Update(context.Background(), "", &terrakube.Team{ID: "team-1"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "orgID" {
		t.Errorf("Field = %q, want %q", valErr.Field, "orgID")
	}
}

func TestTeamService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Teams.Update(context.Background(), "org-1", &terrakube.Team{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty id")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "id" {
		t.Errorf("Field = %q, want %q", valErr.Field, "id")
	}
}

func TestTeamService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/team/team-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Teams.Delete(context.Background(), "org-1", "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Teams.Delete(context.Background(), "", "team-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "orgID" {
		t.Errorf("Field = %q, want %q", valErr.Field, "orgID")
	}
}

func TestTeamService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	c, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Teams.Delete(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty id")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "id" {
		t.Errorf("Field = %q, want %q", valErr.Field, "id")
	}
}

func TestTeamService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/team/nonexistent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "team not found")
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Teams.Delete(context.Background(), "org-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

// TestTeam_BooleanFieldsSerialization verifies that boolean fields set to false
// are NOT dropped during JSON:API serialization. This is the critical fix for
// the known CLI bug where omitempty on booleans prevents disabling permissions.
func TestTeam_BooleanFieldsSerialization(t *testing.T) {
	t.Parallel()

	team := &terrakube.Team{
		ID:               "team-1",
		Name:             "restricted",
		ManageState:      false,
		ManageWorkspace:  false,
		ManageModule:     false,
		ManageProvider:   false,
		ManageVcs:        false,
		ManageTemplate:   false,
		ManageJob:        false,
		ManageCollection: false,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, team); err != nil {
		t.Fatalf("failed to marshal team: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	var data struct {
		Attributes map[string]interface{} `json:"attributes"`
	}
	if err := json.Unmarshal(payload["data"], &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	boolFields := []string{
		"manageState",
		"manageWorkspace",
		"manageModule",
		"manageProvider",
		"manageVcs",
		"manageTemplate",
		"manageJob",
		"manageCollection",
	}

	for _, field := range boolFields {
		val, exists := data.Attributes[field]
		if !exists {
			t.Errorf("boolean field %q was dropped from serialization (omitempty bug)", field)
			continue
		}
		if boolVal, ok := val.(bool); !ok {
			t.Errorf("field %q is not a bool: %T", field, val)
		} else if boolVal != false {
			t.Errorf("field %q = %v, want false", field, boolVal)
		}
	}
}

// TestTeam_BooleanFieldsTrueRoundTrip verifies that boolean fields set to true
// survive a marshal/unmarshal round trip through JSON:API.
func TestTeam_BooleanFieldsTrueRoundTrip(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/team/team-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Team{
			ID:               "team-1",
			Name:             "full-access",
			ManageState:      true,
			ManageWorkspace:  true,
			ManageModule:     true,
			ManageProvider:   true,
			ManageVcs:        true,
			ManageTemplate:   true,
			ManageJob:        true,
			ManageCollection: true,
		})
	})

	c, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	team, err := c.Teams.Get(context.Background(), "org-1", "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !team.ManageState {
		t.Error("ManageState should be true")
	}
	if !team.ManageWorkspace {
		t.Error("ManageWorkspace should be true")
	}
	if !team.ManageModule {
		t.Error("ManageModule should be true")
	}
	if !team.ManageProvider {
		t.Error("ManageProvider should be true")
	}
	if !team.ManageVcs {
		t.Error("ManageVcs should be true")
	}
	if !team.ManageTemplate {
		t.Error("ManageTemplate should be true")
	}
	if !team.ManageJob {
		t.Error("ManageJob should be true")
	}
	if !team.ManageCollection {
		t.Error("ManageCollection should be true")
	}
}
