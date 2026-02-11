package terrakube

import "context"

// Variable represents a Terrakube workspace variable.
type Variable struct {
	ID          string  `jsonapi:"primary,variable"`
	Key         string  `jsonapi:"attr,key"`
	Value       string  `jsonapi:"attr,value"`
	Description string  `jsonapi:"attr,description"`
	Category    string  `jsonapi:"attr,category"`
	Sensitive   bool    `jsonapi:"attr,sensitive"`
	Hcl         bool    `jsonapi:"attr,hcl"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// VariableService handles communication with the workspace variable endpoints.
type VariableService struct {
	crudService[Variable]
}

// List returns all variables for a workspace.
func (s *VariableService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable")
	return s.list(ctx, path, opts)
}

// Get returns a single variable by ID.
func (s *VariableService) Get(ctx context.Context, orgID, workspaceID, id string) (*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("variable ID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable", id)
	return s.get(ctx, path)
}

// Create creates a new variable in the workspace.
func (s *VariableService) Create(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable")
	return s.create(ctx, path, variable)
}

// Update modifies an existing variable. The variable's ID field must be set.
func (s *VariableService) Update(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("variable ID", variable.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable", variable.ID)
	return s.update(ctx, path, variable)
}

// Delete removes a variable by ID.
func (s *VariableService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("workspace ID", workspaceID); err != nil {
		return err
	}
	if err := validateID("variable ID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable", id)
	return s.del(ctx, path)
}
