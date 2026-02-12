package terrakube

import "context"

// ProviderVersion represents a Terrakube provider version resource.
type ProviderVersion struct {
	ID            string  `jsonapi:"primary,version"`
	VersionNumber string  `jsonapi:"attr,versionNumber"`
	Protocols     *string `jsonapi:"attr,protocols"`
	CreatedBy     *string `jsonapi:"attr,createdBy"`
	CreatedDate   *string `jsonapi:"attr,createdDate"`
	UpdatedBy     *string `jsonapi:"attr,updatedBy"`
	UpdatedDate   *string `jsonapi:"attr,updatedDate"`
}

// ProviderVersionService handles communication with the provider version
// related methods of the Terrakube API.
type ProviderVersionService struct {
	crudService[ProviderVersion]
}

// List returns all versions for the given provider within an organization.
// It returns a *ValidationError if orgID or providerID is empty and a *APIError on server errors.
func (s *ProviderVersionService) List(ctx context.Context, orgID, providerID string, opts *ListOptions) ([]*ProviderVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version")
	return s.list(ctx, path, opts)
}

// Get returns a single provider version by ID.
// It returns a *ValidationError if orgID, providerID, or id is empty and a *APIError on server errors.
func (s *ProviderVersionService) Get(ctx context.Context, orgID, providerID, id string) (*ProviderVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", id)
	return s.get(ctx, path)
}

// Create creates a new version for the given provider.
// It returns a *ValidationError if orgID or providerID is empty and a *APIError on server errors.
func (s *ProviderVersionService) Create(ctx context.Context, orgID, providerID string, version *ProviderVersion) (*ProviderVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version")
	return s.create(ctx, path, version)
}

// Update modifies an existing provider version. The version's ID field must be set.
// It returns a *ValidationError if orgID, providerID, or the ID is empty and a *APIError on server errors.
func (s *ProviderVersionService) Update(ctx context.Context, orgID, providerID string, version *ProviderVersion) (*ProviderVersion, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", version.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", version.ID)
	return s.update(ctx, path, version)
}

// Delete removes a provider version by ID.
// It returns a *ValidationError if orgID, providerID, or id is empty and a *APIError on server errors.
func (s *ProviderVersionService) Delete(ctx context.Context, orgID, providerID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return err
	}
	if err := validateID("version ID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", id)
	return s.del(ctx, path)
}
