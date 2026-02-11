package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// History represents a Terrakube workspace history resource.
type History struct {
	ID           string `jsonapi:"primary,history"`
	JobReference string `jsonapi:"attr,jobReference"`
	Output       string `jsonapi:"attr,output"`
}

// HistoryService handles communication with the history-related endpoints.
type HistoryService struct {
	client *Client
}

// List returns all history entries for the given workspace.
func (s *HistoryService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[history]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var histories []*History
	if _, err := s.client.do(ctx, req, &histories); err != nil {
		return nil, err
	}
	return histories, nil
}

// Get returns a single history entry by ID within the given workspace.
func (s *HistoryService) Get(ctx context.Context, orgID, workspaceID, id string) (*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("historyID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	h := &History{}
	if _, err := s.client.do(ctx, req, h); err != nil {
		return nil, err
	}
	return h, nil
}

// Create creates a new history entry in the given workspace.
func (s *HistoryService) Create(ctx context.Context, orgID, workspaceID string, h *History) (*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history")

	req, err := s.client.request(ctx, http.MethodPost, p, h)
	if err != nil {
		return nil, err
	}

	created := &History{}
	if _, err := s.client.do(ctx, req, created); err != nil {
		return nil, err
	}
	return created, nil
}

// Update modifies an existing history entry in the given workspace.
func (s *HistoryService) Update(ctx context.Context, orgID, workspaceID string, h *History) (*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("historyID", h.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history", h.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, h)
	if err != nil {
		return nil, err
	}

	updated := &History{}
	if _, err := s.client.do(ctx, req, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete removes a history entry from the given workspace.
func (s *HistoryService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("historyID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	if _, err := s.client.do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}
