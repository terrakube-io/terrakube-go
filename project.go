package terrakube

import "context"

// Project represents a Terrakube project resource.
type Project struct {
	ID            string           `jsonapi:"primary,project"`
	Name          string           `jsonapi:"attr,name"`
	Description   *string          `jsonapi:"attr,description"`
	CreatedBy     *string          `jsonapi:"attr,createdBy"`
	CreatedDate   *string          `jsonapi:"attr,createdDate"`
	UpdatedBy     *string          `jsonapi:"attr,updatedBy"`
	UpdatedDate   *string          `jsonapi:"attr,updatedDate"`
	Organization  *Organization    `jsonapi:"relation,organization,omitempty"`
	ProjectAccess []*ProjectAccess `jsonapi:"relation,projectAccess,omitempty"`
}

// ProjectService handles communication with the project-related endpoints of the Terrakube API.
type ProjectService struct {
	crudService[Project]
}

// List returns all projects for an organization, optionally filtered.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *ProjectService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Project, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project")
	return s.list(ctx, path, opts)
}

// Get retrieves a project by ID within an organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *ProjectService) Get(ctx context.Context, orgID, id string) (*Project, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project", id)
	return s.get(ctx, path)
}

// Create creates a new project within an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *ProjectService) Create(ctx context.Context, orgID string, project *Project) (*Project, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project")
	return s.create(ctx, path, project)
}

// Update modifies an existing project within an organization. The project's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
func (s *ProjectService) Update(ctx context.Context, orgID string, project *Project) (*Project, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("projectID", project.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project", project.ID)
	return s.update(ctx, path, project)
}

// Delete removes a project by ID within an organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *ProjectService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "project", id)
	return s.del(ctx, path)
}
