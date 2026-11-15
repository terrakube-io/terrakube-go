package terrakube_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

// The Fields interfaces match the services that implement them.
var (
	_ terrakube.WorkspaceFieldsUpdater                          = (*terrakube.WorkspaceService)(nil)
	_ terrakube.WorkspaceFieldsCreator                          = (*terrakube.WorkspaceService)(nil)
	_ terrakube.JobFieldsUpdater                                = (*terrakube.JobService)(nil)
	_ terrakube.NotificationConfigurationWorkspaceFieldsCreator = (*terrakube.NotificationConfigurationService)(nil)
)

// decodeDocument decodes a JSON:API request body into a generic map.
func decodeDocument(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("decoding body %s: %v", body, err)
	}
	return doc
}

func TestUpdateFields_Body(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		patch terrakube.Patch
		want  string
	}{
		{
			name:  "attributes only",
			patch: terrakube.Patch{Attributes: map[string]any{"name": "new", "locked": false, "description": nil}},
			want:  `{"data":{"type":"workspace","id":"ws-1","attributes":{"name":"new","locked":false,"description":null}}}`,
		},
		{
			name:  "relationship set",
			patch: terrakube.Patch{Relationships: map[string]terrakube.Relationship{"vcs": {Type: "vcs", ID: new("vcs-1")}}},
			want:  `{"data":{"type":"workspace","id":"ws-1","relationships":{"vcs":{"data":{"type":"vcs","id":"vcs-1"}}}}}`,
		},
		{
			name:  "relationship cleared",
			patch: terrakube.Patch{Relationships: map[string]terrakube.Relationship{"agent": {Type: "agent"}, "ssh": {}}},
			want:  `{"data":{"type":"workspace","id":"ws-1","relationships":{"agent":{"data":null},"ssh":{"data":null}}}}`,
		},
		{
			name: "to-many set",
			patch: terrakube.Patch{Relationships: map[string]terrakube.Relationship{
				"templates": {Type: "template", Many: true, IDs: []string{"t-1", "t-2"}},
			}},
			want: `{"data":{"type":"workspace","id":"ws-1","relationships":{"templates":{"data":[{"type":"template","id":"t-1"},{"type":"template","id":"t-2"}]}}}}`,
		},
		{
			name: "to-many cleared",
			patch: terrakube.Patch{Relationships: map[string]terrakube.Relationship{
				"templates": {Type: "template", Many: true},
				"policySet": {Type: "policy_set", Many: true, IDs: []string{}},
			}},
			want: `{"data":{"type":"workspace","id":"ws-1","relationships":{"templates":{"data":[]},"policySet":{"data":[]}}}}`,
		},
		{
			name: "attributes and relationships",
			patch: terrakube.Patch{
				Attributes:    map[string]any{"branch": "main"},
				Relationships: map[string]terrakube.Relationship{"project": {Type: "project", ID: new("p-1")}},
			},
			want: `{"data":{"type":"workspace","id":"ws-1","attributes":{"branch":"main"},"relationships":{"project":{"data":{"type":"project","id":"p-1"}}}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := testutil.NewServer(t)
			srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, r *http.Request) {
				if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
					t.Errorf("Content-Type = %q", ct)
				}
				if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
					t.Errorf("Authorization = %q", auth)
				}
				got := decodeDocument(t, r)
				var want map[string]any
				if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					gotJSON, _ := json.Marshal(got)
					t.Errorf("body = %s\nwant   %s", gotJSON, tt.want)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Workspace{ID: "ws-1", Name: "fresh"})
			})

			ws, err := newTestClient(t, srv).Workspaces.UpdateFields(context.Background(), "org-1", "ws-1", tt.patch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ws.Name != "fresh" {
				t.Errorf("Name = %q, want the re-read %q", ws.Name, "fresh")
			}
		})
	}
}

func TestUpdateFields_EmptyPatch(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := testutil.NewServer(t)
	srv.HandleFunc("/", func(http.ResponseWriter, *http.Request) { calls.Add(1) })
	client := newTestClient(t, srv)

	for _, p := range []terrakube.Patch{{}, {Attributes: map[string]any{}, Relationships: map[string]terrakube.Relationship{}}} {
		_, err := client.Workspaces.UpdateFields(context.Background(), "org-1", "ws-1", p)
		ve, ok := errors.AsType[*terrakube.ValidationError](err)
		if !ok {
			t.Fatalf("err = %v, want *ValidationError", err)
		}
		if ve.Field != "Patch" {
			t.Errorf("Field = %q, want %q", ve.Field, "Patch")
		}
	}
	if n := calls.Load(); n != 0 {
		t.Errorf("sent %d requests, want 0", n)
	}
}

func TestUpdateFields_InvalidRelationship(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := map[string]struct {
		rels  map[string]terrakube.Relationship
		field string
	}{
		"missing type":             {map[string]terrakube.Relationship{"vcs": {ID: new("v-1")}}, "Patch.Relationships[vcs].Type"},
		"empty id":                 {map[string]terrakube.Relationship{"vcs": {Type: "vcs", ID: new("")}}, "Patch.Relationships[vcs].ID"},
		"empty name":               {map[string]terrakube.Relationship{"": {}}, "Patch.Relationships"},
		"one and many":             {map[string]terrakube.Relationship{"templates": {Type: "template", ID: new("t-1"), Many: true}}, "Patch.Relationships[templates]"},
		"ids without many":         {map[string]terrakube.Relationship{"templates": {Type: "template", IDs: []string{"t-1"}}}, "Patch.Relationships[templates].IDs"},
		"many without type":        {map[string]terrakube.Relationship{"templates": {Many: true, IDs: []string{"t-1"}}}, "Patch.Relationships[templates].Type"},
		"empty clear without type": {map[string]terrakube.Relationship{"templates": {Many: true}}, "Patch.Relationships[templates].Type"},
		"empty id in many":         {map[string]terrakube.Relationship{"templates": {Type: "template", Many: true, IDs: []string{"t-1", ""}}}, "Patch.Relationships[templates].IDs[1]"},
		"first name in order": {map[string]terrakube.Relationship{
			"zeta": {ID: new("z")}, "alpha": {ID: new("a")}, "mid": {ID: new("m")},
		}, "Patch.Relationships[alpha].Type"},
	}
	for name, tt := range tests {
		for range 5 { // map order varies between runs; the reported field must not
			_, err := client.Workspaces.UpdateFields(context.Background(), "org-1", "ws-1", terrakube.Patch{Relationships: tt.rels})
			ve, ok := errors.AsType[*terrakube.ValidationError](err)
			if !ok || ve.Field != tt.field {
				t.Fatalf("%s: err = %v, want *ValidationError for %s", name, err, tt.field)
			}
		}
		_, err := client.Workspaces.UpdateFields(context.Background(), "org-1", "ws-1", terrakube.Patch{Relationships: tt.rels})
		ve, ok := errors.AsType[*terrakube.ValidationError](err)
		if !ok {
			t.Fatalf("%s: err = %v, want *ValidationError", name, err)
		}
		if ve.Field != tt.field {
			t.Errorf("%s: Field = %q, want %q", name, ve.Field, tt.field)
		}
	}
}

func TestUpdateFields_ResponseBody(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Workspace{ID: "ws-1", Name: "from-patch"})
	})
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1", func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected GET after a PATCH that returned the resource")
	})

	ws, err := newTestClient(t, srv).Workspaces.UpdateFields(context.Background(), "org-1", "ws-1",
		terrakube.Patch{Attributes: map[string]any{"name": "from-patch"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Name != "from-patch" {
		t.Errorf("Name = %q, want %q", ws.Name, "from-patch")
	}
}

func TestUpdateFields_UnreadableAfterUpdate(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})
	client := newTestClient(t, srv)

	org, err := client.Organizations.UpdateFields(context.Background(), "org-1",
		terrakube.Patch{Attributes: map[string]any{"disabled": true, "name": "gone"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.ID != "org-1" || !org.Disabled || org.Name != "gone" {
		t.Errorf("got %+v, want the sent attributes", org)
	}

	// Attributes that do not fit the struct still give the ID.
	org, err = client.Organizations.UpdateFields(context.Background(), "org-1",
		terrakube.Patch{Attributes: map[string]any{"name": 42}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.ID != "org-1" {
		t.Errorf("ID = %q, want %q", org.ID, "org-1")
	}
}

func TestUpdateFields_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "bad")
	})

	_, err := newTestClient(t, srv).Workspaces.UpdateFields(context.Background(), "org-1", "ws-1",
		terrakube.Patch{Attributes: map[string]any{"name": "x"}})
	if apiErr, ok := errors.AsType[*terrakube.APIError](err); !ok || apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("err = %v, want a 422 *APIError", err)
	}
}

func TestCreateFields_SendsOnlyGivenKeys(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/runTriggers", func(w http.ResponseWriter, r *http.Request) {
		got := decodeDocument(t, r)
		want := map[string]any{"data": map[string]any{
			"type":       "runTrigger",
			"attributes": map[string]any{"enabled": false},
			"relationships": map[string]any{
				"sourceWorkspace": map[string]any{"data": map[string]any{"type": "workspace", "id": "ws-2"}},
			},
		}}
		if !reflect.DeepEqual(got, want) {
			gotJSON, _ := json.Marshal(got)
			t.Errorf("body = %s", gotJSON)
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.RunTrigger{ID: "rt-1"})
	})

	rt, err := newTestClient(t, srv).RunTriggers.CreateFields(context.Background(), "org-1", "ws-1", terrakube.Patch{
		Attributes:    map[string]any{"enabled": false},
		Relationships: map[string]terrakube.Relationship{"sourceWorkspace": {Type: "workspace", ID: new("ws-2")}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rt.ID != "rt-1" {
		t.Errorf("ID = %q, want %q", rt.ID, "rt-1")
	}
}

func TestCreateFields_EmptyPatch(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, r *http.Request) {
		got := decodeDocument(t, r)
		want := map[string]any{"data": map[string]any{"type": "policy_set"}}
		if !reflect.DeepEqual(got, want) {
			gotJSON, _ := json.Marshal(got)
			t.Errorf("body = %s", gotJSON)
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.PolicySet{ID: "ps-1"})
	})

	if _, err := newTestClient(t, srv).PolicySets.CreateFields(context.Background(), "org-1", terrakube.Patch{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateFields_Errors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusBadRequest, "bad")
	})
	client := newTestClient(t, srv)

	_, err := client.Workspaces.CreateFields(context.Background(), "org-1", terrakube.Patch{Attributes: map[string]any{"name": "x"}})
	if apiErr, ok := errors.AsType[*terrakube.APIError](err); !ok || apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("err = %v, want a 400 *APIError", err)
	}

	_, err = client.Workspaces.CreateFields(context.Background(), "org-1", terrakube.Patch{
		Relationships: map[string]terrakube.Relationship{"vcs": {ID: new("v-1")}},
	})
	if _, ok := errors.AsType[*terrakube.ValidationError](err); !ok {
		t.Errorf("err = %v, want *ValidationError", err)
	}
}

func TestCreateFields_ToMany(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/notificationConfiguration", func(w http.ResponseWriter, r *http.Request) {
		got := decodeDocument(t, r)
		want := map[string]any{"data": map[string]any{
			"type":       "notification_configuration",
			"attributes": map[string]any{"name": "n"},
			"relationships": map[string]any{
				"templates": map[string]any{"data": []any{
					map[string]any{"type": "template", "id": "t-1"},
				}},
				"triggers": map[string]any{"data": []any{}},
			},
		}}
		if !reflect.DeepEqual(got, want) {
			gotJSON, _ := json.Marshal(got)
			t.Errorf("body = %s", gotJSON)
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.NotificationConfiguration{ID: "nc-1"})
	})

	nc, err := newTestClient(t, srv).NotificationConfigurations.CreateFields(context.Background(), "org-1", terrakube.Patch{
		Attributes: map[string]any{"name": "n"},
		Relationships: map[string]terrakube.Relationship{
			"templates": {Type: "template", Many: true, IDs: []string{"t-1"}},
			"triggers":  {Type: "notification_trigger", Many: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nc.ID != "nc-1" {
		t.Errorf("ID = %q, want %q", nc.ID, "nc-1")
	}
}
