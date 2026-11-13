package terrakube

import "context"

// PolicySetParameter represents a Terrakube policy set parameter resource.
type PolicySetParameter struct {
	ID          string     `jsonapi:"primary,policy_set_parameter"`
	Description *string    `jsonapi:"attr,description"`
	Key         *string    `jsonapi:"attr,key"`
	Value       *string    `jsonapi:"attr,value"`
	CreatedBy   *string    `jsonapi:"attr,createdBy"`
	CreatedDate *string    `jsonapi:"attr,createdDate"`
	UpdatedBy   *string    `jsonapi:"attr,updatedBy"`
	UpdatedDate *string    `jsonapi:"attr,updatedDate"`
	PolicySet   *PolicySet `jsonapi:"relation,policySet,omitempty"`
}

// PolicySetParameterService handles communication with the policy set parameter related
// methods of the Terrakube API.
type PolicySetParameterService struct {
	crudService[PolicySetParameter]
}

// List returns all policy set parameters for a policy set.
// It returns a *ValidationError if orgID or policySetID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) List(ctx context.Context, orgID, policySetID string, opts *ListOptions) ([]*PolicySetParameter, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "parameters")
	return s.list(ctx, path, opts)
}

// Get retrieves a single policy set parameter by ID.
// It returns a *ValidationError if orgID, policySetID, or policySetParameterID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) Get(ctx context.Context, orgID, policySetID, policySetParameterID string) (*PolicySetParameter, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policySetParameterID", policySetParameterID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "parameters", policySetParameterID)
	return s.get(ctx, path)
}

// Create creates a new policy set parameter.
// It returns a *ValidationError if orgID or policySetID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) Create(ctx context.Context, orgID, policySetID string, entity *PolicySetParameter) (*PolicySetParameter, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "parameters")
	return s.create(ctx, path, entity)
}

// Update modifies an existing policy set parameter. The policy set parameter's ID field must be set.
// It returns a *ValidationError if orgID, policySetID, or the ID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) Update(ctx context.Context, orgID, policySetID string, entity *PolicySetParameter) (*PolicySetParameter, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policySetParameterID", entity.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "parameters", entity.ID)
	return s.update(ctx, path, entity)
}

// Delete removes a policy set parameter.
// It returns a *ValidationError if orgID, policySetID, or policySetParameterID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) Delete(ctx context.Context, orgID, policySetID, policySetParameterID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return err
	}
	if err := validateID("policySetParameterID", policySetParameterID); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "parameters", policySetParameterID)
	return s.del(ctx, path)
}
