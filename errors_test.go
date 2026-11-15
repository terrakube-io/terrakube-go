package terrakube_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *terrakube.APIError
		want string
	}{
		{
			name: "with error details",
			err: &terrakube.APIError{
				StatusCode: 404,
				Method:     "GET",
				Path:       "/api/v1/organization/123",
				Errors:     []terrakube.ErrorDetail{{Detail: "not found"}},
			},
			want: "GET /api/v1/organization/123: 404 not found",
		},
		{
			name: "without error details",
			err: &terrakube.APIError{
				StatusCode: 500,
				Method:     "POST",
				Path:       "/api/v1/organization",
			},
			want: "POST /api/v1/organization: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	t.Parallel()
	err := &terrakube.ValidationError{Field: "id", Message: "must not be empty"}
	want := "validation error: id must not be empty"
	if got := err.Error(); got != want {
		t.Errorf("ValidationError.Error() = %q, want %q", got, want)
	}
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"404 api error", &terrakube.APIError{StatusCode: 404}, true},
		{"500 api error", &terrakube.APIError{StatusCode: 500}, false},
		{"wrapped 404", fmt.Errorf("wrap: %w", &terrakube.APIError{StatusCode: 404}), true},
		{"non-api error", errors.New("some error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := terrakube.IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConflict(t *testing.T) {
	t.Parallel()

	if !terrakube.IsConflict(&terrakube.APIError{StatusCode: 409}) {
		t.Error("IsConflict should return true for 409")
	}
	if terrakube.IsConflict(&terrakube.APIError{StatusCode: 404}) {
		t.Error("IsConflict should return false for 404")
	}
	if terrakube.IsConflict(errors.New("other")) {
		t.Error("IsConflict should return false for non-API error")
	}
}

func TestIsUnauthorized(t *testing.T) {
	t.Parallel()

	if !terrakube.IsUnauthorized(&terrakube.APIError{StatusCode: 401}) {
		t.Error("IsUnauthorized should return true for 401")
	}
	if terrakube.IsUnauthorized(&terrakube.APIError{StatusCode: 200}) {
		t.Error("IsUnauthorized should return false for 200")
	}
	if terrakube.IsUnauthorized(errors.New("other")) {
		t.Error("IsUnauthorized should return false for non-API error")
	}
}

func TestErrUnsupported(t *testing.T) {
	t.Parallel()

	springMissing := []byte(`{"timestamp":"2026-11-15T10:00:00.000+00:00","status":404,"error":"Not Found","path":"/x"}`)
	springNotAllowed := []byte(`{"timestamp":"2026-11-15T10:00:00.000+00:00","status":405,"error":"Method Not Allowed","path":"/x"}`)

	tests := []struct {
		name string
		err  *terrakube.APIError
		want bool
	}{
		{"unknown collection", &terrakube.APIError{StatusCode: 404, Path: "/api/v1/organization/o/workspace/w/runTriggers", Errors: []terrakube.ErrorDetail{{Detail: "Unknown collection runTriggers"}}}, true},
		{"unknown collection behind prefix", &terrakube.APIError{StatusCode: 404, Path: "/tk/api/v1/policySet", Errors: []terrakube.ErrorDetail{{Detail: "Unknown collection policySet"}}}, true},
		{"jsonapi entity not found", &terrakube.APIError{StatusCode: 404, Path: "/api/v1/organization/o", Errors: []terrakube.ErrorDetail{{Detail: "Unknown identifier o for organization"}}}, false},
		{"unknown collection detail on 400", &terrakube.APIError{StatusCode: 400, Path: "/api/v1/x", Errors: []terrakube.ErrorDetail{{Detail: "Unknown collection x"}}}, false},
		{"missing route", &terrakube.APIError{StatusCode: 404, Path: "/pat/v1", Body: springMissing}, true},
		{"missing route behind prefix", &terrakube.APIError{StatusCode: 404, Path: "/terrakube/pat/v1", Body: springMissing}, true},
		{"method not allowed", &terrakube.APIError{StatusCode: 405, Path: "/pat/v1", Body: springNotAllowed}, true},
		{"method not allowed, empty body", &terrakube.APIError{StatusCode: 405, Path: "/admin/v1/scheduler/reconciliation"}, true},
		{"problem detail", &terrakube.APIError{StatusCode: 404, Path: "/migrate/v1/workspace", Body: []byte(`{"type":"about:blank","title":"Not Found","status":404,"detail":"No static resource migrate/v1/workspace.","instance":"/migrate/v1/workspace"}`)}, true},
		{"spring message names route", &terrakube.APIError{StatusCode: 404, Path: "/access-token/v1/teams/current-teams", Body: []byte(`{"status":404,"error":"Not Found","path":"/x","message":"No static resource access-token/v1/teams/current-teams."}`)}, true},
		{"stream family", &terrakube.APIError{StatusCode: 404, Path: "/tfoutput/v1/organization/o/job/1/step/s/stream", Body: springMissing}, true},
		{"vcs groups", &terrakube.APIError{StatusCode: 404, Path: "/vcs/v1/abc/groups", Body: springMissing}, true},
		{"vcs repositories", &terrakube.APIError{StatusCode: 405, Path: "/vcs/v1/abc/repositories", Body: springNotAllowed}, true},
		{"policy override", &terrakube.APIError{StatusCode: 404, Path: "/remote/tfe/v2/policy-checks/pe-1/actions/override", Body: springMissing}, true},
		{"policy evaluation", &terrakube.APIError{StatusCode: 404, Path: "/policy/v1/organization/o/workspace/w/evaluation", Body: springMissing}, true},
		{"team permissions", &terrakube.APIError{StatusCode: 404, Path: "/access-token/v1/teams/permissions/organization/o", Body: springMissing}, true},
		{"importer", &terrakube.APIError{StatusCode: 404, Path: "/importer/tfcloud/workspaces", Body: springMissing}, true},
		{"notification", &terrakube.APIError{StatusCode: 404, Path: "/notification/v1/configuration/c/test", Body: springMissing}, true},

		// Entity-not-found answers on listed endpoints.
		{"vcs groups, unknown vcs (empty body)", &terrakube.APIError{StatusCode: 404, Path: "/vcs/v1/abc/groups"}, false},
		{"policy evaluation, unknown workspace", &terrakube.APIError{StatusCode: 404, Path: "/policy/v1/organization/o/workspace/w/evaluation", Body: []byte(`{"message":"Workspace not found"}`)}, false},
		{"stream, unknown step, message exposed", &terrakube.APIError{StatusCode: 404, Path: "/tfoutput/v1/organization/o/job/1/step/s/stream", Body: []byte(`{"status":404,"error":"Not Found","path":"/x","message":"Job or step not found"}`)}, false},
		{"notification, jsonapi errors", &terrakube.APIError{StatusCode: 404, Path: "/notification/v1/configuration/c/test", Body: []byte(`{"errors":[{"detail":"configuration not found"}]}`)}, false},
		{"plain text", &terrakube.APIError{StatusCode: 404, Path: "/migrate/v1/workspace", Body: []byte("Workspace not found")}, false},
		{"405 plain text", &terrakube.APIError{StatusCode: 405, Path: "/migrate/v1/workspace", Body: []byte("no")}, false},

		// Known limit (documented on ErrUnsupported): Terrakube 2.34.1 answers an
		// unknown job or step on the stream endpoint with Spring's default body
		// and no message, which looks exactly like a missing route.
		{"stream, unknown step, 2.34.1 body", &terrakube.APIError{StatusCode: 404, Path: "/tfoutput/v1/organization/o/job/1/step/s/stream", Body: []byte(`{"timestamp":"2026-11-15T10:00:00.000+00:00","status":404,"error":"Not Found","path":"/tfoutput/v1/organization/o/job/1/step/s/stream"}`)}, true},

		// Not a listed endpoint, or not a 404/405.
		{"pat by id", &terrakube.APIError{StatusCode: 404, Path: "/pat/v1/123", Body: springMissing}, false},
		{"jsonapi path", &terrakube.APIError{StatusCode: 404, Path: "/api/v1/organization", Body: springMissing}, false},
		{"tfstate", &terrakube.APIError{StatusCode: 404, Path: "/tfstate/v1/organization/o/workspace/w/state/terraform.tfstate", Body: springMissing}, false},
		{"server error", &terrakube.APIError{StatusCode: 500, Path: "/pat/v1", Body: springMissing}, false},
	}
	for _, tt := range tests {
		if got := errors.Is(tt.err, terrakube.ErrUnsupported); got != tt.want {
			t.Errorf("%s: errors.Is(err, ErrUnsupported) = %v, want %v", tt.name, got, tt.want)
		}
	}

	wrapped := fmt.Errorf("listing: %w", tests[0].err)
	if !errors.Is(wrapped, terrakube.ErrUnsupported) {
		t.Error("wrapped error does not match ErrUnsupported")
	}
	if errors.Is(errors.New("other"), terrakube.ErrUnsupported) {
		t.Error("unrelated error matches ErrUnsupported")
	}
	if !terrakube.IsNotFound(tests[0].err) {
		t.Error("IsNotFound is false for an unsupported 404")
	}
}

func TestErrUnsupported_FromServer(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "Unknown collection policySet")
	})

	_, err := newTestClient(t, srv).PolicySets.List(context.Background(), "org-1", nil)
	if !errors.Is(err, terrakube.ErrUnsupported) {
		t.Fatalf("err = %v, want ErrUnsupported", err)
	}
}
