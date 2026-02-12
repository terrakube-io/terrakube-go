package terrakube

import "context"

// ModuleVersion represents a Terrakube module version resource.
type ModuleVersion struct {
	ID          string  `jsonapi:"primary,version"`
	Version     string  `jsonapi:"attr,version"`
	Commit      *string `jsonapi:"attr,commit"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// ModuleVersionService handles communication with the module version endpoints.
type ModuleVersionService struct {
	crudService[ModuleVersion]
}

// List returns all versions for a module.
// It returns a *ValidationError if orgID or moduleID is empty and a *APIError on server errors.
func (s *ModuleVersionService) List(ctx context.Context, orgID, moduleID string, opts *ListOptions) ([]*ModuleVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("module ID", moduleID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", moduleID, "version")
	return s.list(ctx, path, opts)
}

// Get returns a single module version by ID.
// It returns a *ValidationError if orgID, moduleID, or id is empty and a *APIError on server errors.
func (s *ModuleVersionService) Get(ctx context.Context, orgID, moduleID, id string) (*ModuleVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("module ID", moduleID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", moduleID, "version", id)
	return s.get(ctx, path)
}

// Create creates a new version for a module.
// It returns a *ValidationError if orgID or moduleID is empty and a *APIError on server errors.
func (s *ModuleVersionService) Create(ctx context.Context, orgID, moduleID string, version *ModuleVersion) (*ModuleVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("module ID", moduleID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", moduleID, "version")
	return s.create(ctx, path, version)
}

// Update modifies an existing module version. The version's ID field must be set.
// It returns a *ValidationError if orgID, moduleID, or the ID is empty and a *APIError on server errors.
func (s *ModuleVersionService) Update(ctx context.Context, orgID, moduleID string, version *ModuleVersion) (*ModuleVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("module ID", moduleID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", version.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", moduleID, "version", version.ID)
	return s.update(ctx, path, version)
}

// Delete removes a module version by ID.
// It returns a *ValidationError if orgID, moduleID, or id is empty and a *APIError on server errors.
func (s *ModuleVersionService) Delete(ctx context.Context, orgID, moduleID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("module ID", moduleID); err != nil {
		return err
	}
	if err := validateID("version ID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "module", moduleID, "version", id)
	return s.del(ctx, path)
}
