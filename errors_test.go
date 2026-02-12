package terrakube_test

import (
	"errors"
	"fmt"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
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
