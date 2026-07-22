package terrakube

import "context"

// ProjectAccess represents a Terrakube project access resource.
type ProjectAccess struct {
	ID              string   `jsonapi:"primary,project_access"`
	Name            string   `jsonapi:"attr,name"`
	ManageState     bool     `jsonapi:"attr,manageState"`
	ManageJob       bool     `jsonapi:"attr,manageJob"`
	ManageWorkspace bool     `jsonapi:"attr,manageWorkspace"`
	PlanJob         bool     `jsonapi:"attr,planJob"`
	ApproveJob      bool     `jsonapi:"attr,approveJob"`
	Role            string   `jsonapi:"attr,role"`
	Project         *Project `jsonapi:"relation,project,omitempty"`
}

// ProjectAccessService handles communication with the project access related methods of the Terrakube API.
type ProjectAccessService struct {
	crudService[ProjectAccess]
}

// List returns all project access rules for a project within an organization.
// It returns a *ValidationError if orgID or projectID is empty and a *APIError on server errors.
func (s *ProjectAccessService) List(ctx context.Context, orgID, projectID string, opts *ListOptions) ([]*ProjectAccess, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("project ID", projectID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project", projectID, "projectAccess")
	return s.list(ctx, path, opts)
}

// Get retrieves a project access rule by ID.
// It returns a *ValidationError if orgID, projectID, or id is empty and a *APIError on server errors.
func (s *ProjectAccessService) Get(ctx context.Context, orgID, projectID, id string) (*ProjectAccess, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("project ID", projectID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project", projectID, "projectAccess", id)
	return s.get(ctx, path)
}

// Create creates a new project access rule for a project within an organization.
// It returns a *ValidationError if orgID or projectID is empty and a *APIError on server errors.
func (s *ProjectAccessService) Create(ctx context.Context, orgID, projectID string, access *ProjectAccess) (*ProjectAccess, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("project ID", projectID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project", projectID, "projectAccess")
	return s.create(ctx, path, access)
}

// Update modifies an existing project access rule. The project access ID must be set.
// It returns a *ValidationError if orgID, projectID, or access ID is empty and a *APIError on server errors.
func (s *ProjectAccessService) Update(ctx context.Context, orgID, projectID string, access *ProjectAccess) (*ProjectAccess, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("project ID", projectID); err != nil {
		return nil, err
	}
	if err := validateID("projectAccess ID", access.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "project", projectID, "projectAccess", access.ID)
	return s.update(ctx, path, access)
}

// Delete removes a project access rule by ID.
// It returns a *ValidationError if orgID, projectID, or id is empty and a *APIError on server errors.
func (s *ProjectAccessService) Delete(ctx context.Context, orgID, projectID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("project ID", projectID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "project", projectID, "projectAccess", id)
	return s.del(ctx, path)
}
