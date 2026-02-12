package terrakube_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/google/jsonapi"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func boolPtr(b bool) *bool { return &b }

func newTestOrganizationVariable() *terrakube.OrganizationVariable {
	return &terrakube.OrganizationVariable{
		ID:          "gv-1",
		Key:         "TF_LOG",
		Value:       "DEBUG",
		Description: "Terraform log level",
		Category:    "ENV",
		Sensitive:   boolPtr(false),
		Hcl:         false,
	}
}

func TestOrganizationVariableService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/globalvar", func(w http.ResponseWriter, _ *http.Request) {
		vars := []*terrakube.OrganizationVariable{newTestOrganizationVariable()}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, vars)
	})

	client := newTestClient(t, srv)

	variables, err := client.OrganizationVariables.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("got %d variables, want 1", len(variables))
	}
	if variables[0].Key != "TF_LOG" {
		t.Errorf("Key = %q, want %q", variables[0].Key, "TF_LOG")
	}
}

func TestOrganizationVariableService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/globalvar", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[globalvar]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.OrganizationVariable{})
	})

	client := newTestClient(t, srv)

	_, err := client.OrganizationVariables.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "key==TF_LOG"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrganizationVariableService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.OrganizationVariables.List(context.Background(), "", nil)
	assertValidationError(t, err, "organization ID")
}

func TestOrganizationVariableService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/globalvar/gv-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestOrganizationVariable())
	})

	client := newTestClient(t, srv)

	v, err := client.OrganizationVariables.Get(context.Background(), "org-1", "gv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != "gv-1" {
		t.Errorf("ID = %q, want %q", v.ID, "gv-1")
	}
	if v.Key != "TF_LOG" {
		t.Errorf("Key = %q, want %q", v.Key, "TF_LOG")
	}
}

func TestOrganizationVariableService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/globalvar/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "variable not found")
	})

	client := newTestClient(t, srv)

	_, err := client.OrganizationVariables.Get(context.Background(), "org-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestOrganizationVariableService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		varID string
		field string
	}{
		{"empty org ID", "", "gv-1", "organization ID"},
		{"empty globalvar ID", "org-1", "", "globalvar ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.OrganizationVariables.Get(context.Background(), tt.orgID, tt.varID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestOrganizationVariableService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/globalvar", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestOrganizationVariable())
	})

	client := newTestClient(t, srv)

	input := &terrakube.OrganizationVariable{
		Key:       "TF_LOG",
		Value:     "DEBUG",
		Category:  "ENV",
		Sensitive: boolPtr(false),
		Hcl:       false,
	}
	v, err := client.OrganizationVariables.Create(context.Background(), "org-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != "gv-1" {
		t.Errorf("ID = %q, want %q", v.ID, "gv-1")
	}
}

func TestOrganizationVariableService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.OrganizationVariables.Create(context.Background(), "", &terrakube.OrganizationVariable{})
	assertValidationError(t, err, "organization ID")
}

func TestOrganizationVariableService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/globalvar/gv-1", func(w http.ResponseWriter, _ *http.Request) {
		v := newTestOrganizationVariable()
		v.Value = "TRACE"
		testutil.WriteJSONAPI(t, w, http.StatusOK, v)
	})

	client := newTestClient(t, srv)

	input := &terrakube.OrganizationVariable{
		ID:    "gv-1",
		Key:   "TF_LOG",
		Value: "TRACE",
	}
	v, err := client.OrganizationVariables.Update(context.Background(), "org-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Value != "TRACE" {
		t.Errorf("Value = %q, want %q", v.Value, "TRACE")
	}
}

func TestOrganizationVariableService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.OrganizationVariables.Update(context.Background(), "", &terrakube.OrganizationVariable{ID: "gv-1"})
	assertValidationError(t, err, "organization ID")
}

func TestOrganizationVariableService_Update_EmptyVarID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.OrganizationVariables.Update(context.Background(), "org-1", &terrakube.OrganizationVariable{ID: ""})
	assertValidationError(t, err, "globalvar ID")
}

func TestOrganizationVariableService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/globalvar/gv-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	err := client.OrganizationVariables.Delete(context.Background(), "org-1", "gv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrganizationVariableService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		varID string
		field string
	}{
		{"empty org ID", "", "gv-1", "organization ID"},
		{"empty globalvar ID", "org-1", "", "globalvar ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.OrganizationVariables.Delete(context.Background(), tt.orgID, tt.varID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestOrganizationVariableService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/globalvar/gv-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)

	err := client.OrganizationVariables.Delete(context.Background(), "org-1", "gv-1")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestOrganizationVariable_BooleanFalseNotDropped(t *testing.T) {
	t.Parallel()

	v := &terrakube.OrganizationVariable{
		ID:        "gv-1",
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: boolPtr(false),
		Hcl:       false,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, v); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.Bytes()

	if !bytes.Contains(payload, []byte(`"sensitive"`)) {
		t.Error("Sensitive=*false was dropped from JSON:API payload; non-nil *bool should be serialized even with omitempty")
	}
	if !bytes.Contains(payload, []byte(`"hcl"`)) {
		t.Error("Hcl=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
}

func TestOrganizationVariable_SensitiveNilOmitted(t *testing.T) {
	t.Parallel()

	v := &terrakube.OrganizationVariable{
		ID:        "gv-1",
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: nil,
		Hcl:       false,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, v); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.Bytes()

	if bytes.Contains(payload, []byte(`"sensitive"`)) {
		t.Error("Sensitive=nil should be omitted from JSON:API payload with *bool + omitempty")
	}
}

func TestOrganizationVariable_BooleanTruePreserved(t *testing.T) {
	t.Parallel()

	v := &terrakube.OrganizationVariable{
		ID:        "gv-1",
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: boolPtr(true),
		Hcl:       true,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, v); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.Bytes()

	if !bytes.Contains(payload, []byte(`"sensitive"`)) {
		t.Error("Sensitive=true was dropped from JSON:API payload")
	}
	if !bytes.Contains(payload, []byte(`"hcl"`)) {
		t.Error("Hcl=true was dropped from JSON:API payload")
	}
}

func TestOrganizationVariableService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/globalvar/gv-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestOrganizationVariable())
	})

	client := newTestClient(t, srv)

	_, err := client.OrganizationVariables.Get(context.Background(), "org-1", "gv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
