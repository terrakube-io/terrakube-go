package terrakube

import (
	"context"
	"net/http"
)

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
	Op   OperationAction        `json:"op"`
	Ref  OperationRef           `json:"ref,omitempty"`
	Href string                 `json:"href,omitempty"`
	Data map[string]interface{} `json:"data,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// AtomicRequest is the request body for POST /operations.
type AtomicRequest struct {
	Operations []Operation `json:"atomic:operations"`
}

// AtomicResult represents the result of a single atomic operation.
type AtomicResult struct {
	Data map[string]interface{} `json:"data,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"`
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

	result := &AtomicResponse{}
	_, err = s.client.doRaw(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
