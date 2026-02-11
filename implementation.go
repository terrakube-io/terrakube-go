package terrakube

import "context"

// Implementation represents a Terrakube provider version implementation resource.
type Implementation struct {
	ID                  string  `jsonapi:"primary,implementation"`
	Os                  string  `jsonapi:"attr,os"`
	Arch                string  `jsonapi:"attr,arch"`
	Filename            string  `jsonapi:"attr,filename"`
	DownloadURL         *string `jsonapi:"attr,downloadUrl"`
	ShasumsURL          *string `jsonapi:"attr,shasumsUrl"`
	ShasumsSignatureURL *string `jsonapi:"attr,shasumsSignatureUrl"`
	Shasum              *string `jsonapi:"attr,shasum"`
	KeyID               *string `jsonapi:"attr,keyId"`
	ASCIIArmor          *string `jsonapi:"attr,asciiArmor"`
	TrustSignature      *string `jsonapi:"attr,trustSignature"`
	Source              *string `jsonapi:"attr,source"`
	SourceURL           *string `jsonapi:"attr,sourceUrl"`
	CreatedBy           *string `jsonapi:"attr,createdBy"`
	CreatedDate         *string `jsonapi:"attr,createdDate"`
	UpdatedBy           *string `jsonapi:"attr,updatedBy"`
	UpdatedDate         *string `jsonapi:"attr,updatedDate"`
}

// ImplementationService handles communication with the implementation-related endpoints.
type ImplementationService struct {
	crudService[Implementation]
}

// List returns all implementations for a provider version.
func (s *ImplementationService) List(ctx context.Context, orgID, providerID, versionID string, opts *ListOptions) ([]*Implementation, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", versionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", versionID, "implementation")
	return s.list(ctx, path, opts)
}

// Get returns a single implementation by ID.
func (s *ImplementationService) Get(ctx context.Context, orgID, providerID, versionID, id string) (*Implementation, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", versionID); err != nil {
		return nil, err
	}
	if err := validateID("implementation ID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", versionID, "implementation", id)
	return s.get(ctx, path)
}

// Create creates a new implementation for a provider version.
func (s *ImplementationService) Create(ctx context.Context, orgID, providerID, versionID string, impl *Implementation) (*Implementation, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", versionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", versionID, "implementation")
	return s.create(ctx, path, impl)
}

// Update modifies an existing implementation. The implementation's ID field must be set.
func (s *ImplementationService) Update(ctx context.Context, orgID, providerID, versionID string, impl *Implementation) (*Implementation, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return nil, err
	}
	if err := validateID("version ID", versionID); err != nil {
		return nil, err
	}
	if err := validateID("implementation ID", impl.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", versionID, "implementation", impl.ID)
	return s.update(ctx, path, impl)
}

// Delete removes an implementation by ID.
func (s *ImplementationService) Delete(ctx context.Context, orgID, providerID, versionID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("provider ID", providerID); err != nil {
		return err
	}
	if err := validateID("version ID", versionID); err != nil {
		return err
	}
	if err := validateID("implementation ID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "provider", providerID, "version", versionID, "implementation", id)
	return s.del(ctx, path)
}
