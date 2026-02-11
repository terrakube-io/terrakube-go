package terrakube_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
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
	var ve *terrakube.ValidationError
	if !errors.As(err, &ve) {
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
			Attributes map[string]interface{} `json:"attributes"`
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
