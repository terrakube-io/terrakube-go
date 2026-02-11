package terrakube_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/jsonapi"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func newTestVariable() *terrakube.Variable {
	return &terrakube.Variable{
		ID:          "var-1",
		Key:         "AWS_REGION",
		Value:       "us-east-1",
		Description: "AWS region",
		Category:    "ENV",
		Sensitive:   false,
		Hcl:         false,
	}
}

func TestVariableService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/variable", func(w http.ResponseWriter, _ *http.Request) {
		vars := []*terrakube.Variable{newTestVariable()}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, vars)
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	variables, err := client.Variables.List(context.Background(), "org-1", "ws-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("got %d variables, want 1", len(variables))
	}
	if variables[0].Key != "AWS_REGION" {
		t.Errorf("Key = %q, want %q", variables[0].Key, "AWS_REGION")
	}
}

func TestVariableService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/variable", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[variable]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Variable{})
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.List(context.Background(), "org-1", "ws-1", &terrakube.ListOptions{Filter: "key==foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVariableService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.List(context.Background(), "", "ws-1", nil)
	if err == nil {
		t.Fatal("expected error for empty org ID")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "organization ID" {
		t.Errorf("Field = %q, want %q", valErr.Field, "organization ID")
	}
}

func TestVariableService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.List(context.Background(), "org-1", "", nil)
	if err == nil {
		t.Fatal("expected error for empty workspace ID")
	}
	var valErr *terrakube.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "workspace ID" {
		t.Errorf("Field = %q, want %q", valErr.Field, "workspace ID")
	}
}

func TestVariableService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/variable/var-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestVariable())
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, err := client.Variables.Get(context.Background(), "org-1", "ws-1", "var-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != "var-1" {
		t.Errorf("ID = %q, want %q", v.ID, "var-1")
	}
	if v.Key != "AWS_REGION" {
		t.Errorf("Key = %q, want %q", v.Key, "AWS_REGION")
	}
}

func TestVariableService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name   string
		orgID  string
		wsID   string
		varID  string
		field  string
	}{
		{"empty org ID", "", "ws-1", "var-1", "organization ID"},
		{"empty workspace ID", "org-1", "", "var-1", "workspace ID"},
		{"empty variable ID", "org-1", "ws-1", "", "variable ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.Variables.Get(context.Background(), tt.orgID, tt.wsID, tt.varID)
			if err == nil {
				t.Fatal("expected error")
			}
			var valErr *terrakube.ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
			if valErr.Field != tt.field {
				t.Errorf("Field = %q, want %q", valErr.Field, tt.field)
			}
		})
	}
}

func TestVariableService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/variable/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "variable not found")
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.Get(context.Background(), "org-1", "ws-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestVariableService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/variable", func(w http.ResponseWriter, _ *http.Request) {
		v := newTestVariable()
		testutil.WriteJSONAPI(t, w, http.StatusCreated, v)
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := &terrakube.Variable{
		Key:       "AWS_REGION",
		Value:     "us-east-1",
		Category:  "ENV",
		Sensitive: false,
		Hcl:       false,
	}
	v, err := client.Variables.Create(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != "var-1" {
		t.Errorf("ID = %q, want %q", v.ID, "var-1")
	}
}

func TestVariableService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.Create(context.Background(), "", "ws-1", &terrakube.Variable{})
	if err == nil {
		t.Fatal("expected error for empty org ID")
	}
}

func TestVariableService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.Create(context.Background(), "org-1", "", &terrakube.Variable{})
	if err == nil {
		t.Fatal("expected error for empty workspace ID")
	}
}

func TestVariableService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1/variable/var-1", func(w http.ResponseWriter, _ *http.Request) {
		v := newTestVariable()
		v.Value = "us-west-2"
		testutil.WriteJSONAPI(t, w, http.StatusOK, v)
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := &terrakube.Variable{
		ID:    "var-1",
		Key:   "AWS_REGION",
		Value: "us-west-2",
	}
	v, err := client.Variables.Update(context.Background(), "org-1", "ws-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Value != "us-west-2" {
		t.Errorf("Value = %q, want %q", v.Value, "us-west-2")
	}
}

func TestVariableService_Update_EmptyVariableID(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.Update(context.Background(), "org-1", "ws-1", &terrakube.Variable{ID: ""})
	if err == nil {
		t.Fatal("expected error for empty variable ID")
	}
}

func TestVariableService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/variable/var-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = client.Variables.Delete(context.Background(), "org-1", "ws-1", "var-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVariableService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.com"), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		orgID string
		wsID  string
		varID string
	}{
		{"empty org ID", "", "ws-1", "var-1"},
		{"empty workspace ID", "org-1", "", "var-1"},
		{"empty variable ID", "org-1", "ws-1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.Variables.Delete(context.Background(), tt.orgID, tt.wsID, tt.varID)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestVariableService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1/variable/var-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = client.Variables.Delete(context.Background(), "org-1", "ws-1", "var-1")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestVariable_BooleanFalseNotDropped(t *testing.T) {
	t.Parallel()

	v := &terrakube.Variable{
		ID:        "var-1",
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: false,
		Hcl:       false,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, v); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.String()

	if !bytes.Contains([]byte(payload), []byte(`"sensitive"`)) {
		t.Error("Sensitive=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
	if !bytes.Contains([]byte(payload), []byte(`"hcl"`)) {
		t.Error("Hcl=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
}

func TestVariable_BooleanTruePreserved(t *testing.T) {
	t.Parallel()

	v := &terrakube.Variable{
		ID:        "var-1",
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: true,
		Hcl:       true,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, v); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.String()

	if !bytes.Contains([]byte(payload), []byte(`"sensitive"`)) {
		t.Error("Sensitive=true was dropped from JSON:API payload")
	}
	if !bytes.Contains([]byte(payload), []byte(`"hcl"`)) {
		t.Error("Hcl=true was dropped from JSON:API payload")
	}
}

func TestVariableService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/variable/var-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestVariable())
	})

	client, err := terrakube.NewClient(terrakube.WithEndpoint(srv.URL), terrakube.WithToken("test-token"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Variables.Get(context.Background(), "org-1", "ws-1", "var-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
