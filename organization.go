package terrakube

import "context"

// Organization represents a Terrakube organization resource.
type Organization struct {
	ID            string  `jsonapi:"primary,organization"`
	Name          string  `jsonapi:"attr,name"`
	Description   *string `jsonapi:"attr,description"`
	ExecutionMode string  `jsonapi:"attr,executionMode"`
	Disabled      bool    `jsonapi:"attr,disabled"`
	Icon          *string `jsonapi:"attr,icon"`
	CreatedBy     *string `jsonapi:"attr,createdBy"`
	CreatedDate   *string `jsonapi:"attr,createdDate"`
	UpdatedBy     *string `jsonapi:"attr,updatedBy"`
	UpdatedDate   *string `jsonapi:"attr,updatedDate"`
}

// OrganizationService handles communication with the organization related
// methods of the Terrakube API.
type OrganizationService struct {
	crudService[Organization]
}

// List returns all organizations, optionally filtered.
// It returns a *APIError on server errors.
func (s *OrganizationService) List(ctx context.Context, opts *ListOptions) ([]*Organization, error) {
	path := s.client.apiPath("organization")
	return s.list(ctx, path, opts)
}

// Get retrieves an organization by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *OrganizationService) Get(ctx context.Context, id string) (*Organization, error) {
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", id)
	return s.get(ctx, path)
}

// Create creates a new organization.
// It returns a *APIError on server errors.
func (s *OrganizationService) Create(ctx context.Context, org *Organization) (*Organization, error) {
	path := s.client.apiPath("organization")
	return s.create(ctx, path, org)
}

// Update modifies an existing organization. The organization's ID field must be set.
// It returns a *ValidationError if the ID is empty and a *APIError on server errors.
func (s *OrganizationService) Update(ctx context.Context, org *Organization) (*Organization, error) {
	if err := validateID("organization ID", org.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", org.ID)
	return s.update(ctx, path, org)
}

// Delete removes an organization by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *OrganizationService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", id)
	return s.del(ctx, path)
}
