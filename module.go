package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Module represents a Terrakube module resource.
type Module struct {
	ID          string  `jsonapi:"primary,module"`
	Name        string  `jsonapi:"attr,name"`
	Description string  `jsonapi:"attr,description"`
	Provider    string  `jsonapi:"attr,provider"`
	Source      string  `jsonapi:"attr,source"`
	Folder      *string `jsonapi:"attr,folder"`
	TagPrefix   *string `jsonapi:"attr,tagPrefix"`
}

// ModuleService handles communication with the module related
// methods of the Terrakube API.
type ModuleService struct {
	client *Client
}

// List returns all modules for an organization, optionally filtered.
func (s *ModuleService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module")

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

	var modules []*Module
	_, err = s.client.do(ctx, req, &modules)
	if err != nil {
		return nil, err
	}

	return modules, nil
}

// Get retrieves a module by ID within an organization.
func (s *ModuleService) Get(ctx context.Context, orgID, id string) (*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", id)
	req, err := s.client.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	mod := &Module{}
	_, err = s.client.do(ctx, req, mod)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

// Create creates a new module within an organization.
func (s *ModuleService) Create(ctx context.Context, orgID string, mod *Module) (*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module")
	req, err := s.client.request(ctx, http.MethodPost, path, mod)
	if err != nil {
		return nil, err
	}

	created := &Module{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing module within an organization.
func (s *ModuleService) Update(ctx context.Context, orgID string, mod *Module) (*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("module ID", mod.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", mod.ID)
	req, err := s.client.request(ctx, http.MethodPatch, path, mod)
	if err != nil {
		return nil, err
	}

	updated := &Module{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a module by ID within an organization.
func (s *ModuleService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "module", id)
	req, err := s.client.request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
