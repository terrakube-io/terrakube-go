package terrakube

import (
	"context"
	"time"
)

// Template represents a Terrakube template resource.
type Template struct {
	ID           string        `jsonapi:"primary,template"`
	Name         string        `jsonapi:"attr,name"`
	Description  *string       `jsonapi:"attr,description"`
	Version      *string       `jsonapi:"attr,version"`
	Content      string        `jsonapi:"attr,tcl"`
	CreatedBy    *string       `jsonapi:"attr,createdBy"`
	CreatedDate  *time.Time    `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy    *string       `jsonapi:"attr,updatedBy"`
	UpdatedDate  *time.Time    `jsonapi:"attr,updatedDate,rfc3339"`
	Organization *Organization `jsonapi:"relation,organization,omitempty"`
}

// TemplateService handles communication with the template-related endpoints.
type TemplateService struct {
	crudService[Template]
}

// List returns all templates for the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *TemplateService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "template")
	return s.list(ctx, path, opts)
}

// Get returns a single template by ID within the given organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *TemplateService) Get(ctx context.Context, orgID, id string) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("templateID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "template", id)
	return s.get(ctx, path)
}

// Create creates a new template in the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *TemplateService) Create(ctx context.Context, orgID string, tmpl *Template) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "template")
	return s.create(ctx, path, tmpl)
}

// CreateFields creates a new template in the given organization, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID is empty and a *APIError on server errors.
func (s *TemplateService) CreateFields(ctx context.Context, orgID string, p Patch) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "template")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing template in the given organization. The template's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *TemplateService) Update(ctx context.Context, orgID string, tmpl *Template) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("templateID", tmpl.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "template", tmpl.ID)
	return s.update(ctx, path, tmpl)
}

// UpdateFields modifies only the attributes and relationships set in p on the template
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID or id is empty or p sets nothing, and a *APIError on server errors.
func (s *TemplateService) UpdateFields(ctx context.Context, orgID, id string, p Patch) (*Template, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("templateID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "template", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes a template from the given organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *TemplateService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("templateID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "template", id)
	return s.del(ctx, path)
}
