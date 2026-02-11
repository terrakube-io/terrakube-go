package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Template represents a Terrakube template resource.
type Template struct {
	ID          string  `jsonapi:"primary,template"`
	Name        string  `jsonapi:"attr,name"`
	Description *string `jsonapi:"attr,description"`
	Version     *string `jsonapi:"attr,version"`
	Content     string  `jsonapi:"attr,tcl"`
}

// TemplateService handles communication with the template-related endpoints.
type TemplateService struct {
	client *Client
}

// List returns all templates for the given organization.
func (s *TemplateService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "template")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[template]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var templates []*Template
	if _, err := s.client.do(ctx, req, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// Get returns a single template by ID within the given organization.
func (s *TemplateService) Get(ctx context.Context, orgID, id string) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("templateID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "template", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	tmpl := &Template{}
	if _, err := s.client.do(ctx, req, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Create creates a new template in the given organization.
func (s *TemplateService) Create(ctx context.Context, orgID string, tmpl *Template) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "template")

	req, err := s.client.request(ctx, http.MethodPost, p, tmpl)
	if err != nil {
		return nil, err
	}

	created := &Template{}
	if _, err := s.client.do(ctx, req, created); err != nil {
		return nil, err
	}
	return created, nil
}

// Update modifies an existing template in the given organization.
func (s *TemplateService) Update(ctx context.Context, orgID string, tmpl *Template) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("templateID", tmpl.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "template", tmpl.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, tmpl)
	if err != nil {
		return nil, err
	}

	updated := &Template{}
	if _, err := s.client.do(ctx, req, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete removes a template from the given organization.
func (s *TemplateService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("templateID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "template", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	if _, err := s.client.do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}
