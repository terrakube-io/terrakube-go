package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Variable represents a Terrakube workspace variable.
type Variable struct {
	ID          string `jsonapi:"primary,variable"`
	Key         string `jsonapi:"attr,key"`
	Value       string `jsonapi:"attr,value"`
	Description string `jsonapi:"attr,description"`
	Category    string `jsonapi:"attr,category"`
	Sensitive   bool   `jsonapi:"attr,sensitive"`
	Hcl         bool   `jsonapi:"attr,hcl"`
}

// VariableService handles communication with the workspace variable endpoints.
type VariableService struct {
	client *Client
}

func (s *VariableService) basePath(orgID, workspaceID string) string {
	return s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable")
}

// List returns all variables for a workspace.
func (s *VariableService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[variable]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, s.basePath(orgID, workspaceID), params, nil)
	if err != nil {
		return nil, err
	}

	var variables []*Variable
	_, err = s.client.do(ctx, req, &variables)
	return variables, err
}

// Get returns a single variable by ID.
func (s *VariableService) Get(ctx context.Context, orgID, workspaceID, id string) (*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("variable ID", id); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodGet, s.basePath(orgID, workspaceID)+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	v := &Variable{}
	_, err = s.client.do(ctx, req, v)
	return v, err
}

// Create creates a new variable in the workspace.
func (s *VariableService) Create(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPost, s.basePath(orgID, workspaceID), variable)
	if err != nil {
		return nil, err
	}

	v := &Variable{}
	_, err = s.client.do(ctx, req, v)
	return v, err
}

// Update modifies an existing variable. The variable's ID field must be set.
func (s *VariableService) Update(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("variable ID", variable.ID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPatch, s.basePath(orgID, workspaceID)+"/"+variable.ID, variable)
	if err != nil {
		return nil, err
	}

	v := &Variable{}
	_, err = s.client.do(ctx, req, v)
	return v, err
}

// Delete removes a variable by ID.
func (s *VariableService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return err
	}
	if err := validateID("variable ID", id); err != nil {
		return err
	}

	req, err := s.client.request(ctx, http.MethodDelete, s.basePath(orgID, workspaceID)+"/"+id, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
