package terrakube

import "context"

// Module represents a Terrakube module resource.
type Module struct {
	ID               string  `jsonapi:"primary,module"`
	Name             string  `jsonapi:"attr,name"`
	Description      string  `jsonapi:"attr,description"`
	Provider         string  `jsonapi:"attr,provider"`
	Source           string  `jsonapi:"attr,source"`
	Folder           *string `jsonapi:"attr,folder"`
	TagPrefix        *string `jsonapi:"attr,tagPrefix"`
	DownloadQuantity int     `jsonapi:"attr,downloadQuantity"`
	LatestVersion    *string `jsonapi:"attr,latestVersion"`
	RegistryPath     *string `jsonapi:"attr,registryPath"`
	CreatedBy        *string `jsonapi:"attr,createdBy"`
	CreatedDate      *string `jsonapi:"attr,createdDate"`
	UpdatedBy        *string `jsonapi:"attr,updatedBy"`
	UpdatedDate      *string `jsonapi:"attr,updatedDate"`
	Vcs              *VCS    `jsonapi:"relation,vcs,omitempty"`
	SSH              *SSH    `jsonapi:"relation,ssh,omitempty"`
}

// ModuleService handles communication with the module related
// methods of the Terrakube API.
type ModuleService struct {
	crudService[Module]
}

// List returns all modules for an organization, optionally filtered.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *ModuleService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module")
	return s.list(ctx, path, opts)
}

// Get retrieves a module by ID within an organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *ModuleService) Get(ctx context.Context, orgID, id string) (*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", id)
	return s.get(ctx, path)
}

// Create creates a new module within an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *ModuleService) Create(ctx context.Context, orgID string, mod *Module) (*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module")
	return s.create(ctx, path, mod)
}

// Update modifies an existing module within an organization. The module's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
func (s *ModuleService) Update(ctx context.Context, orgID string, mod *Module) (*Module, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("module ID", mod.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "module", mod.ID)
	return s.update(ctx, path, mod)
}

// Delete removes a module by ID within an organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *ModuleService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "module", id)
	return s.del(ctx, path)
}
