package terrakube_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

// newTestClient creates a client pointing at the test server.
func newTestClient(t *testing.T, srv *testutil.Server) *terrakube.Client {
	t.Helper()
	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}

// newTestClientFromURL creates a client from a raw URL string.
func newTestClientFromURL(t *testing.T, url string) *terrakube.Client { //nolint:unparam // Helper accepts URL for flexibility
	t.Helper()
	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(url),
		terrakube.WithToken("test-token"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

// assertValidationError checks that err is a *ValidationError for the given field.
func assertValidationError(t *testing.T, err error, field string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error for %q, got nil", field)
	}
	ve, ok := errors.AsType[*terrakube.ValidationError](err)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	if ve.Field != field {
		t.Errorf("ValidationError.Field = %q, want %q", ve.Field, field)
	}
	if !strings.Contains(ve.Message, "must not be empty") {
		t.Errorf("ValidationError.Message = %q, want it to contain %q", ve.Message, "must not be empty")
	}
}

// assertJSONBoolField checks that a JSON body contains a boolean field set to the expected value.
func assertJSONBoolField(t *testing.T, body []byte, attrName string, expected bool) { //nolint:unparam // Helper accepts expected value for flexibility
	t.Helper()
	var payload struct {
		Data struct {
			Attributes map[string]any `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	val, ok := payload.Data.Attributes[attrName]
	if !ok {
		t.Fatalf("attribute %q not found in JSON body", attrName)
	}
	boolVal, ok := val.(bool)
	if !ok {
		t.Fatalf("attribute %q is %T, not bool", attrName, val)
	}
	if boolVal != expected {
		t.Errorf("attribute %q = %v, want %v", attrName, boolVal, expected)
	}
}

// testPatch is the sparse write the per-service UpdateFields and CreateFields tests send.
var testPatch = terrakube.Patch{Attributes: map[string]any{"description": "x"}}

// checkFieldsBody checks that r carries testPatch as a JSON:API document of jsonapiType with the given id.
func checkFieldsBody(t *testing.T, r *http.Request, jsonapiType, id string) {
	t.Helper()
	var doc struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		t.Errorf("decoding body: %v", err)
		return
	}
	want := map[string]any{"type": jsonapiType, "attributes": map[string]any{"description": "x"}}
	if id != "" {
		want["id"] = id
	}
	if !reflect.DeepEqual(doc.Data, want) {
		t.Errorf("data = %v, want %v", doc.Data, want)
	}
}

// writeResource answers with a JSON:API resource object of jsonapiType and id.
func writeResource(w http.ResponseWriter, status int, jsonapiType, id string) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"type": jsonapiType, "id": id, "attributes": map[string]any{}}})
}

// handleUpdateFields serves a sparse PATCH of testPatch to the resource "id-1" at path, answered
// with 204, and the GET that follows.
func handleUpdateFields(t *testing.T, srv *testutil.Server, path, jsonapiType string) {
	t.Helper()
	const id = "id-1"
	srv.HandleFunc("PATCH "+path, func(w http.ResponseWriter, r *http.Request) {
		checkFieldsBody(t, r, jsonapiType, id)
		w.WriteHeader(http.StatusNoContent)
	})
	srv.HandleFunc("GET "+path, func(w http.ResponseWriter, _ *http.Request) {
		writeResource(w, http.StatusOK, jsonapiType, id)
	})
}

// handleCreateFields serves a sparse POST of testPatch at path, answered with a resource whose ID is "new-1".
func handleCreateFields(t *testing.T, srv *testutil.Server, path, jsonapiType string) {
	t.Helper()
	srv.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		checkFieldsBody(t, r, jsonapiType, "")
		writeResource(w, http.StatusCreated, jsonapiType, "new-1")
	})
}
