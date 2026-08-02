package terrakube

import "context"

// Variable represents a Terrakube workspace variable.
type Variable struct {
	ID          string     `jsonapi:"primary,variable"`
	Key         string     `jsonapi:"attr,key"`
	Value       string     `jsonapi:"attr,value"`
	Description string     `jsonapi:"attr,description,omitempty"`
	Category    string     `jsonapi:"attr,category"`
	Sensitive   bool       `jsonapi:"attr,sensitive"`
	Hcl         bool       `jsonapi:"attr,hcl"`
	Incomplete  bool       `jsonapi:"attr,incomplete,omitempty"`
	CreatedBy   *string    `jsonapi:"attr,createdBy,omitempty"`
	CreatedDate *string    `jsonapi:"attr,createdDate,omitempty"`
	UpdatedBy   *string    `jsonapi:"attr,updatedBy,omitempty"`
	UpdatedDate *string    `jsonapi:"attr,updatedDate,omitempty"`
	Workspace   *Workspace `jsonapi:"relation,workspace,omitempty"`
}

// VariableService handles communication with the workspace variable endpoints.
type VariableService struct {
	crudService[Variable]
}

// List returns all variables for a workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *VariableService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*Variable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable")
	return s.list(ctx, path, opts)
}

// Get returns a single variable by ID.
// It returns a *ValidationError if orgID, workspaceID, or id is empty and a *APIError on server errors.
func (s *VariableService) Get(ctx context.Context, orgID, workspaceID, id string) (*Variable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("variableID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable", id)
	return s.get(ctx, path)
}

// Create creates a new variable in the workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *VariableService) Create(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable")
	return s.create(ctx, path, variable)
}

// Update modifies an existing variable. The variable's ID field must be set.
// It returns a *ValidationError if orgID, workspaceID, or the ID is empty and a *APIError on server errors.
func (s *VariableService) Update(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("variableID", variable.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable", variable.ID)
	return s.update(ctx, path, variable)
}

// Delete removes a variable by ID.
// It returns a *ValidationError if orgID, workspaceID, or id is empty and a *APIError on server errors.
func (s *VariableService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("variableID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "variable", id)
	return s.del(ctx, path)
}
