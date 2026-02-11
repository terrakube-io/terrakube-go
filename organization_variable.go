package terrakube

import "context"

// OrganizationVariable represents a Terrakube organization-level global variable.
type OrganizationVariable struct {
	ID          string  `jsonapi:"primary,globalvar"`
	Key         string  `jsonapi:"attr,key"`
	Value       string  `jsonapi:"attr,value"`
	Description string  `jsonapi:"attr,description"`
	Category    string  `jsonapi:"attr,category"`
	Sensitive   *bool   `jsonapi:"attr,sensitive,omitempty"`
	Hcl         bool    `jsonapi:"attr,hcl"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// OrganizationVariableService handles communication with the organization global variable endpoints.
type OrganizationVariableService struct {
	crudService[OrganizationVariable]
}

// List returns all global variables for an organization.
func (s *OrganizationVariableService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar")
	return s.list(ctx, path, opts)
}

// Get returns a single organization variable by ID.
func (s *OrganizationVariableService) Get(ctx context.Context, orgID, id string) (*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("globalvar ID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar", id)
	return s.get(ctx, path)
}

// Create creates a new global variable in the organization.
func (s *OrganizationVariableService) Create(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar")
	return s.create(ctx, path, variable)
}

// Update modifies an existing organization variable. The variable's ID field must be set.
func (s *OrganizationVariableService) Update(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("globalvar ID", variable.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar", variable.ID)
	return s.update(ctx, path, variable)
}

// Delete removes an organization variable by ID.
func (s *OrganizationVariableService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("globalvar ID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "globalvar", id)
	return s.del(ctx, path)
}
