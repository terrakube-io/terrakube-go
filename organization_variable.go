package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// OrganizationVariable represents a Terrakube organization-level global variable.
type OrganizationVariable struct {
	ID          string `jsonapi:"primary,globalvar"`
	Key         string `jsonapi:"attr,key"`
	Value       string `jsonapi:"attr,value"`
	Description string `jsonapi:"attr,description"`
	Category    string `jsonapi:"attr,category"`
	Sensitive   bool   `jsonapi:"attr,sensitive"`
	Hcl         bool   `jsonapi:"attr,hcl"`
}

// OrganizationVariableService handles communication with the organization global variable endpoints.
type OrganizationVariableService struct {
	client *Client
}

func (s *OrganizationVariableService) basePath(orgID string) string {
	return s.client.apiPath("organization", orgID, "globalvar")
}

// List returns all global variables for an organization.
func (s *OrganizationVariableService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[globalvar]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, s.basePath(orgID), params, nil)
	if err != nil {
		return nil, err
	}

	var variables []*OrganizationVariable
	_, err = s.client.do(ctx, req, &variables)
	return variables, err
}

// Get returns a single organization variable by ID.
func (s *OrganizationVariableService) Get(ctx context.Context, orgID, id string) (*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("globalvar ID", id); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodGet, s.basePath(orgID)+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	v := &OrganizationVariable{}
	_, err = s.client.do(ctx, req, v)
	return v, err
}

// Create creates a new global variable in the organization.
func (s *OrganizationVariableService) Create(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPost, s.basePath(orgID), variable)
	if err != nil {
		return nil, err
	}

	v := &OrganizationVariable{}
	_, err = s.client.do(ctx, req, v)
	return v, err
}

// Update modifies an existing organization variable. The variable's ID field must be set.
func (s *OrganizationVariableService) Update(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("globalvar ID", variable.ID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPatch, s.basePath(orgID)+"/"+variable.ID, variable)
	if err != nil {
		return nil, err
	}

	v := &OrganizationVariable{}
	_, err = s.client.do(ctx, req, v)
	return v, err
}

// Delete removes an organization variable by ID.
func (s *OrganizationVariableService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("globalvar ID", id); err != nil {
		return err
	}

	req, err := s.client.request(ctx, http.MethodDelete, s.basePath(orgID)+"/"+id, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
