package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// SSH represents an SSH key in Terrakube.
type SSH struct {
	ID          string  `jsonapi:"primary,ssh"`
	Name        string  `jsonapi:"attr,name"`
	Description *string `jsonapi:"attr,description"`
	PrivateKey  string  `jsonapi:"attr,privateKey"`
	SSHType     string  `jsonapi:"attr,sshType"`
}

// SSHService handles communication with the SSH related methods of the Terrakube API.
type SSHService struct {
	client *Client
}

func (s *SSHService) basePath(orgID string) string {
	return s.client.apiPath("organization", orgID, "ssh")
}

func (s *SSHService) resourcePath(orgID, id string) string {
	return s.client.apiPath("organization", orgID, "ssh", id)
}

// List returns all SSH keys for an organization.
func (s *SSHService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[ssh]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, s.basePath(orgID), params, nil)
	if err != nil {
		return nil, err
	}

	var items []*SSH
	_, err = s.client.do(ctx, req, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// Get returns a single SSH key by ID.
func (s *SSHService) Get(ctx context.Context, orgID, id string) (*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("sshID", id); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodGet, s.resourcePath(orgID, id), nil)
	if err != nil {
		return nil, err
	}

	v := &SSH{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Create creates a new SSH key in an organization.
func (s *SSHService) Create(ctx context.Context, orgID string, ssh *SSH) (*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPost, s.basePath(orgID), ssh)
	if err != nil {
		return nil, err
	}

	v := &SSH{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Update modifies an existing SSH key.
func (s *SSHService) Update(ctx context.Context, orgID string, ssh *SSH) (*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("sshID", ssh.ID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPatch, s.resourcePath(orgID, ssh.ID), ssh)
	if err != nil {
		return nil, err
	}

	v := &SSH{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Delete removes an SSH key by ID.
func (s *SSHService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("sshID", id); err != nil {
		return err
	}

	req, err := s.client.request(ctx, http.MethodDelete, s.resourcePath(orgID, id), nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
