package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// WorkspaceAccess represents access control settings for a workspace.
type WorkspaceAccess struct {
	ID              string `jsonapi:"primary,access"`
	ManageState     bool   `jsonapi:"attr,manageState"`
	ManageWorkspace bool   `jsonapi:"attr,manageWorkspace"`
	ManageJob       bool   `jsonapi:"attr,manageJob"`
	Name            string `jsonapi:"attr,name"`
}

// WorkspaceAccessService handles communication with the workspace access related
// methods of the Terrakube API.
type WorkspaceAccessService struct {
	client *Client
}

// List returns all access entries for the given workspace.
func (s *WorkspaceAccessService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[access]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var access []*WorkspaceAccess
	_, err = s.client.do(ctx, req, &access)
	if err != nil {
		return nil, err
	}

	return access, nil
}

// Get returns a single workspace access entry by ID.
func (s *WorkspaceAccessService) Get(ctx context.Context, orgID, workspaceID, id string) (*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("accessID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	a := &WorkspaceAccess{}
	_, err = s.client.do(ctx, req, a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Create creates a new access entry for the given workspace.
func (s *WorkspaceAccessService) Create(ctx context.Context, orgID, workspaceID string, access *WorkspaceAccess) (*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access")

	req, err := s.client.request(ctx, http.MethodPost, p, access)
	if err != nil {
		return nil, err
	}

	created := &WorkspaceAccess{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing workspace access entry. The access entry's ID field must be set.
func (s *WorkspaceAccessService) Update(ctx context.Context, orgID, workspaceID string, access *WorkspaceAccess) (*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("accessID", access.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access", access.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, access)
	if err != nil {
		return nil, err
	}

	updated := &WorkspaceAccess{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a workspace access entry by ID.
func (s *WorkspaceAccessService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("accessID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
