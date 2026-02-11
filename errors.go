package terrakube

import (
	"errors"
	"fmt"
)

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

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("%s %s: %d %s", e.Method, e.Path, e.StatusCode, e.Errors[0].Detail)
	}
	return fmt.Sprintf("%s %s: %d", e.Method, e.Path, e.StatusCode)
}

// ValidationError represents a client-side validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s %s", e.Field, e.Message)
}

// IsNotFound returns true if the error is a 404 API error.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsConflict returns true if the error is a 409 API error.
func IsConflict(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 409
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 API error.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}
