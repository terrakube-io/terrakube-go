package terrakube

import "context"

// OrganizationVariable represents a Terrakube organization-level global variable.
type OrganizationVariable struct {
	ID           string        `jsonapi:"primary,globalvar"`
	Key          string        `jsonapi:"attr,key"`
	Value        string        `jsonapi:"attr,value"`
	Description  string        `jsonapi:"attr,description"`
	Category     string        `jsonapi:"attr,category"`
	Sensitive    *bool         `jsonapi:"attr,sensitive,omitempty"`
	Hcl          bool          `jsonapi:"attr,hcl"`
	Organization *Organization `jsonapi:"relation,organization,omitempty"`
}

// OrganizationVariableService handles communication with the organization global variable endpoints.
type OrganizationVariableService struct {
	crudService[OrganizationVariable]
}

// List returns all global variables for an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *OrganizationVariableService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*OrganizationVariable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar")
	return s.list(ctx, path, opts)
}

// Get returns a single organization variable by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *OrganizationVariableService) Get(ctx context.Context, orgID, id string) (*OrganizationVariable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("globalvarID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar", id)
	return s.get(ctx, path)
}

// Create creates a new global variable in the organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *OrganizationVariableService) Create(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar")
	return s.create(ctx, path, variable)
}

// CreateFields creates a new global variable in the organization, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID is empty and a *APIError on server errors.
func (s *OrganizationVariableService) CreateFields(ctx context.Context, orgID string, p Patch) (*OrganizationVariable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing organization variable. The variable's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *OrganizationVariableService) Update(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("globalvarID", variable.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar", variable.ID)
	return s.update(ctx, path, variable)
}

// UpdateFields modifies only the attributes and relationships set in p on the organization variable
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID or id is empty or p sets nothing, and a *APIError on server errors.
func (s *OrganizationVariableService) UpdateFields(ctx context.Context, orgID, id string, p Patch) (*OrganizationVariable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("globalvarID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "globalvar", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes an organization variable by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *OrganizationVariableService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("globalvarID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "globalvar", id)
	return s.del(ctx, path)
}
