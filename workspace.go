package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Workspace represents a Terrakube workspace resource.
type Workspace struct {
	ID               string  `jsonapi:"primary,workspace"`
	Name             string  `jsonapi:"attr,name"`
	Description      *string `jsonapi:"attr,description"`
	Source           string  `jsonapi:"attr,source"`
	Branch           string  `jsonapi:"attr,branch"`
	Folder           string  `jsonapi:"attr,folder"`
	TemplateID       string  `jsonapi:"attr,defaultTemplate"`
	IaCType          string  `jsonapi:"attr,iacType"`
	IaCVersion       string  `jsonapi:"attr,terraformVersion"`
	ExecutionMode    string  `jsonapi:"attr,executionMode"`
	Deleted          bool    `jsonapi:"attr,deleted"`
	AllowRemoteApply bool    `jsonapi:"attr,allowRemoteApply"`
}

// WorkspaceService handles communication with the workspace related
// methods of the Terrakube API.
type WorkspaceService struct {
	client *Client
}

// List returns all workspaces for an organization, optionally filtered.
func (s *WorkspaceService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace")

	var req *http.Request
	var err error

	if opts != nil && opts.Filter != "" {
		params := url.Values{"filter": {opts.Filter}}
		req, err = s.client.requestWithQuery(ctx, http.MethodGet, path, params, nil)
	} else {
		req, err = s.client.request(ctx, http.MethodGet, path, nil)
	}
	if err != nil {
		return nil, err
	}

	var workspaces []*Workspace
	_, err = s.client.do(ctx, req, &workspaces)
	if err != nil {
		return nil, err
	}

	return workspaces, nil
}

// Get retrieves a workspace by ID within an organization.
func (s *WorkspaceService) Get(ctx context.Context, orgID, id string) (*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", id)
	req, err := s.client.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	ws := &Workspace{}
	_, err = s.client.do(ctx, req, ws)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

// Create creates a new workspace within an organization.
func (s *WorkspaceService) Create(ctx context.Context, orgID string, ws *Workspace) (*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace")
	req, err := s.client.request(ctx, http.MethodPost, path, ws)
	if err != nil {
		return nil, err
	}

	created := &Workspace{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing workspace within an organization.
func (s *WorkspaceService) Update(ctx context.Context, orgID string, ws *Workspace) (*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", ws.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", ws.ID)
	req, err := s.client.request(ctx, http.MethodPatch, path, ws)
	if err != nil {
		return nil, err
	}

	updated := &Workspace{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a workspace by ID within an organization.
func (s *WorkspaceService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", id)
	req, err := s.client.request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
