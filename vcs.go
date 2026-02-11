package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// VCS represents a version control system connection in Terrakube.
type VCS struct {
	ID             string `jsonapi:"primary,vcs"`
	Name           string `jsonapi:"attr,name"`
	Description    string `jsonapi:"attr,description"`
	VcsType        string `jsonapi:"attr,vcsType"`
	ConnectionType string `jsonapi:"attr,connectionType"`
	ClientID       string `jsonapi:"attr,clientId"`
	ClientSecret   string `jsonapi:"attr,clientSecret"`
	PrivateKey     string `jsonapi:"attr,privateKey"`
	Endpoint       string `jsonapi:"attr,endpoint"`
	APIURL         string `jsonapi:"attr,apiUrl"`
	Status         string `jsonapi:"attr,status"`
}

// VCSService handles communication with the VCS related methods of the Terrakube API.
type VCSService struct {
	client *Client
}

func (s *VCSService) basePath(orgID string) string {
	return s.client.apiPath("organization", orgID, "vcs")
}

func (s *VCSService) resourcePath(orgID, id string) string {
	return s.client.apiPath("organization", orgID, "vcs", id)
}

// List returns all VCS connections for an organization.
func (s *VCSService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[vcs]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, s.basePath(orgID), params, nil)
	if err != nil {
		return nil, err
	}

	var items []*VCS
	_, err = s.client.do(ctx, req, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// Get returns a single VCS connection by ID.
func (s *VCSService) Get(ctx context.Context, orgID, id string) (*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("vcsID", id); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodGet, s.resourcePath(orgID, id), nil)
	if err != nil {
		return nil, err
	}

	v := &VCS{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Create creates a new VCS connection in an organization.
func (s *VCSService) Create(ctx context.Context, orgID string, vcs *VCS) (*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPost, s.basePath(orgID), vcs)
	if err != nil {
		return nil, err
	}

	v := &VCS{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Update modifies an existing VCS connection.
func (s *VCSService) Update(ctx context.Context, orgID string, vcs *VCS) (*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("vcsID", vcs.ID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPatch, s.resourcePath(orgID, vcs.ID), vcs)
	if err != nil {
		return nil, err
	}

	v := &VCS{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Delete removes a VCS connection by ID.
func (s *VCSService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("vcsID", id); err != nil {
		return err
	}

	req, err := s.client.request(ctx, http.MethodDelete, s.resourcePath(orgID, id), nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
