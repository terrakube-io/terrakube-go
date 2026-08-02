package terrakube

import "context"

// Provider represents a Terrakube provider resource within an organization.
type Provider struct {
	ID                string             `jsonapi:"primary,provider"`
	Name              string             `jsonapi:"attr,name"`
	Description       *string            `jsonapi:"attr,description"`
	Imported          bool               `jsonapi:"attr,imported"`
	RegistryNamespace *string            `jsonapi:"attr,registryNamespace,omitempty"`
	Organization      *Organization      `jsonapi:"relation,organization,omitempty"`
	Versions          []*ProviderVersion `jsonapi:"relation,version,omitempty"`
}

// ProviderService handles communication with the provider related methods of
// the Terrakube API.
type ProviderService struct {
	crudService[Provider]
}

// List returns all providers for the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *ProviderService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Provider, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider")
	return s.list(ctx, path, opts)
}

// Get returns a single provider by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *ProviderService) Get(ctx context.Context, orgID, id string) (*Provider, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("providerID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", id)
	return s.get(ctx, path)
}

// Create creates a new provider in the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *ProviderService) Create(ctx context.Context, orgID string, provider *Provider) (*Provider, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider")
	return s.create(ctx, path, provider)
}

// Update modifies an existing provider. The provider's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
func (s *ProviderService) Update(ctx context.Context, orgID string, provider *Provider) (*Provider, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("providerID", provider.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", provider.ID)
	return s.update(ctx, path, provider)
}

// Delete removes a provider by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *ProviderService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("providerID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "provider", id)
	return s.del(ctx, path)
}
