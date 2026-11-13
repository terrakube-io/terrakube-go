package terrakube

import (
	"context"
	"net/http"
)

// atomicType is the media type of the JSON:API Atomic Operations extension.
// The server answers 415 to POST /operations with any other Content-Type.
const atomicType = mediaType + `; ext="https://jsonapi.org/ext/atomic"`

// OperationAction represents the type of atomic operation.
type OperationAction string

// Supported atomic operation actions.
const (
	OperationAdd    OperationAction = "add"
	OperationUpdate OperationAction = "update"
	OperationRemove OperationAction = "remove"
)

// OperationRef identifies the target resource for an atomic operation.
type OperationRef struct {
	Type         string `json:"type"`
	ID           string `json:"id,omitempty"`
	LID          string `json:"lid,omitempty"`
	Relationship string `json:"relationship,omitempty"`
}

// Operation represents a single atomic operation.
type Operation struct {
	Op   OperationAction `json:"op"`
	Ref  OperationRef    `json:"ref,omitzero"` // the server rejects an operation with both ref and href
	Href string          `json:"href,omitempty"`
	Data map[string]any  `json:"data,omitempty"`
	Meta map[string]any  `json:"meta,omitempty"`
}

// AtomicRequest is the request body for POST /operations.
type AtomicRequest struct {
	Operations []Operation `json:"atomic:operations"`
}

// AtomicResult represents the result of a single atomic operation.
type AtomicResult struct {
	Data map[string]any `json:"data,omitempty"`
	Meta map[string]any `json:"meta,omitempty"`
}

// AtomicResponse is the response from POST /operations.
type AtomicResponse struct {
	Results []AtomicResult `json:"atomic:results"`
}

// OperationsService handles atomic batch operations.
type OperationsService struct {
	client *Client
}

// Submit sends an atomic operations batch request.
// It returns a *APIError on server errors.
func (s *OperationsService) Submit(ctx context.Context, ops *AtomicRequest) (*AtomicResponse, error) {
	path := s.client.apiPath("operations")

	req, err := s.client.requestRaw(ctx, http.MethodPost, path, ops)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", atomicType)
	req.Header.Set("Accept", atomicType)

	result := &AtomicResponse{}
	_, err = s.client.doRaw(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
