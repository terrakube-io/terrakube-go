package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Organization represents a Terrakube organization resource.
type Organization struct {
	ID            string  `jsonapi:"primary,organization"`
	Name          string  `jsonapi:"attr,name"`
	Description   *string `jsonapi:"attr,description"`
	ExecutionMode string  `jsonapi:"attr,executionMode"`
	Disabled      bool    `jsonapi:"attr,disabled"`
	Icon          *string `jsonapi:"attr,icon"`
}

// OrganizationService handles communication with the organization related
// methods of the Terrakube API.
type OrganizationService struct {
	client *Client
}

// List returns all organizations, optionally filtered.
func (s *OrganizationService) List(ctx context.Context, opts *ListOptions) ([]*Organization, error) {
	path := s.client.apiPath("organization")

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

	var orgs []*Organization
	_, err = s.client.do(ctx, req, &orgs)
	if err != nil {
		return nil, err
	}

	return orgs, nil
}

// Get retrieves an organization by ID.
func (s *OrganizationService) Get(ctx context.Context, id string) (*Organization, error) {
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", id)
	req, err := s.client.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	org := &Organization{}
	_, err = s.client.do(ctx, req, org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// Create creates a new organization.
func (s *OrganizationService) Create(ctx context.Context, org *Organization) (*Organization, error) {
	path := s.client.apiPath("organization")
	req, err := s.client.request(ctx, http.MethodPost, path, org)
	if err != nil {
		return nil, err
	}

	created := &Organization{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing organization.
func (s *OrganizationService) Update(ctx context.Context, org *Organization) (*Organization, error) {
	if err := validateID("organization ID", org.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", org.ID)
	req, err := s.client.request(ctx, http.MethodPatch, path, org)
	if err != nil {
		return nil, err
	}

	updated := &Organization{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes an organization by ID.
func (s *OrganizationService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", id)
	req, err := s.client.request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
