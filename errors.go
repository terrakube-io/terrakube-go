package terrakube

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// ErrUnsupported is matched with errors.Is against an error from the API. It
// matches a *APIError that shows the server lacks the resource type or
// endpoint, as an older Terrakube release does:
//   - a JSON:API 404 whose detail starts with "Unknown collection";
//   - a 404 or 405 from one of these raw endpoint families, when the body is
//     not an entity-not-found error: /tfoutput/v1/.../stream, /pat/v1,
//     /notification/v1/..., /access-token/v1/teams/current-teams,
//     /access-token/v1/teams/permissions/..., /vcs/v1/{id}/groups,
//     /vcs/v1/{id}/repositories, /importer/tfcloud/..., /migrate/v1/...,
//     /admin/v1/scheduler/reconciliation, /policy/v1/.../evaluation,
//     /remote/tfe/v2/policy-checks/{id}/actions/override.
//
// Known limit: on /tfoutput/v1/.../stream, Terrakube 2.34.1 answers an unknown
// job or step with Spring's default error body and no message, which cannot be
// told apart from a missing route; that answer matches ErrUnsupported too.
var ErrUnsupported = errors.New("terrakube: not supported by this server")

// unsupportedFamilies are the raw endpoint paths on which a missing route
// means the server lacks the endpoint. Any endpoint path prefix comes first.
var unsupportedFamilies = regexp.MustCompile(`^(?:/.*)?(?:` + strings.Join([]string{
	`/tfoutput/v1/.+/stream`,
	`/pat/v1/?`,
	`/notification/v1/.*`,
	`/access-token/v1/teams/current-teams`,
	`/access-token/v1/teams/permissions/.*`,
	`/vcs/v1/[^/]+/groups`,
	`/vcs/v1/[^/]+/repositories`,
	`/importer/tfcloud/.*`,
	`/migrate/v1/.*`,
	`/admin/v1/scheduler/reconciliation`,
	`/policy/v1/.+/evaluation`,
	`/remote/tfe/v2/policy-checks/[^/]+/actions/override`,
}, "|") + `)$`)

// APIError represents an error response from the Terrakube API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       []byte
	Errors     []ErrorDetail
}

// ErrorDetail represents a single error entry in a JSON:API error response.
type ErrorDetail struct {
	Detail string `json:"detail"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

// Error returns a string representation including the HTTP method, path, and status code.
func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("%s %s: %d %s", e.Method, e.Path, e.StatusCode, e.Errors[0].Detail)
	}
	return fmt.Sprintf("%s %s: %d", e.Method, e.Path, e.StatusCode)
}

// Is reports whether e matches target. It makes errors.Is(err, ErrUnsupported)
// true for an error that shows the server lacks a resource type or endpoint.
func (e *APIError) Is(target error) bool {
	return target == ErrUnsupported && e.unsupported() //nolint:errorlint // sentinel comparison inside Is
}

// unsupported reports whether e shows that the server lacks the resource type
// or endpoint the request was for.
func (e *APIError) unsupported() bool {
	if e.StatusCode == http.StatusNotFound {
		for _, d := range e.Errors {
			if strings.HasPrefix(d.Detail, "Unknown collection") {
				return true
			}
		}
	}
	if e.StatusCode != http.StatusNotFound && e.StatusCode != http.StatusMethodNotAllowed {
		return false
	}
	if !unsupportedFamilies.MatchString(e.Path) {
		return false
	}
	return routeMissing(e.StatusCode, e.Body)
}

// routeMissing reports whether a 404 or 405 body says that no route matched,
// rather than that an entity was not found. Terrakube's controllers answer a
// missing entity with an empty body, plain text, {"message": ...} or JSON:API
// errors; Spring answers a missing route with its default error document
// ({"timestamp", "status", "error", "path"}), or a problem detail. A 405
// counts as a missing route unless its body names an entity. A controller that
// throws ResponseStatusException(NOT_FOUND) gets the same default body, without
// its message under Spring's default settings, and so also counts as a missing
// route (see ErrUnsupported).
func routeMissing(status int, body []byte) bool {
	var doc struct {
		Status  *int            `json:"status"`
		Error   *string         `json:"error"`
		Path    *string         `json:"path"`
		Title   *string         `json:"title"`
		Detail  string          `json:"detail"`
		Message string          `json:"message"`
		Errors  json.RawMessage `json:"errors"`
	}
	if json.Unmarshal(body, &doc) != nil {
		return status == http.StatusMethodNotAllowed && len(strings.TrimSpace(string(body))) == 0
	}
	if len(doc.Errors) > 0 || (doc.Message != "" && !routeMessage(doc.Message)) || (doc.Detail != "" && !routeMessage(doc.Detail)) {
		return false
	}
	springDefault := doc.Status != nil && doc.Error != nil && doc.Path != nil
	problemDetail := doc.Status != nil && doc.Title != nil
	return springDefault || problemDetail
}

// routeMessage reports whether a Spring error message describes a missing
// route or method rather than a missing entity.
func routeMessage(msg string) bool {
	for _, prefix := range []string{"No static resource", "No endpoint", "No handler found", "Request method"} {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return false
}

// ValidationError represents a client-side validation failure.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns a string representation of the validation failure.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s %s", e.Field, e.Message)
}

// IsNotFound returns true if the error is a 404 API error.
func IsNotFound(err error) bool {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsConflict returns true if the error is a 409 API error.
func IsConflict(err error) bool {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.StatusCode == 409
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 API error.
func IsUnauthorized(err error) bool {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsForbidden returns true if the error is a 403 API error.
func IsForbidden(err error) bool {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.StatusCode == 403
	}
	return false
}
