package terrakube

import (
	"context"
	"time"
)

// PolicySetParameter represents a Terrakube policy set parameter resource.
type PolicySetParameter struct {
	ID          string     `jsonapi:"primary,policy_set_parameter"`
	Description *string    `jsonapi:"attr,description"`
	Key         *string    `jsonapi:"attr,key"`
	Value       *string    `jsonapi:"attr,value"`
	CreatedBy   *string    `jsonapi:"attr,createdBy"`
	CreatedDate *time.Time `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy   *string    `jsonapi:"attr,updatedBy"`
	UpdatedDate *time.Time `jsonapi:"attr,updatedDate,rfc3339"`
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

// CreateFields creates a new policy set parameter, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID or policySetID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) CreateFields(ctx context.Context, orgID, policySetID string, p Patch) (*PolicySetParameter, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "parameters")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing policy set parameter. The policy set parameter's ID field must be set.
// It returns a *ValidationError if orgID, policySetID, or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
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

// UpdateFields modifies only the attributes and relationships set in p on the policy set parameter
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID, policySetID, or id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicySetParameterService) UpdateFields(ctx context.Context, orgID, policySetID, id string, p Patch) (*PolicySetParameter, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policySetParameterID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "parameters", id)
	return s.updateFields(ctx, path, id, p)
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

// ListByPolicySet returns the policy set parameters of a policy set, from the top-level
// /policy_set/{id}/parameters path, for callers that do not know its organization.
// It returns a *ValidationError if policySetID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) ListByPolicySet(ctx context.Context, policySetID string, opts *ListOptions) ([]*PolicySetParameter, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	return s.list(ctx, s.client.apiPath("policy_set", policySetID, "parameters"), opts)
}

// GetByPolicySet retrieves a policy set parameter of a policy set from the top-level
// /policy_set/{id}/parameters/{id} path.
// It returns a *ValidationError if policySetID or id is empty and a *APIError on server errors.
func (s *PolicySetParameterService) GetByPolicySet(ctx context.Context, policySetID, id string) (*PolicySetParameter, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policySetParameterID", id); err != nil {
		return nil, err
	}

	return s.get(ctx, s.client.apiPath("policy_set", policySetID, "parameters", id))
}

// CreateFieldsByPolicySet creates a policy set parameter in a policy set through the top-level
// /policy_set/{id}/parameters path, sending only the attributes and relationships set in p.
// It returns a *ValidationError if policySetID is empty and a *APIError on server errors.
func (s *PolicySetParameterService) CreateFieldsByPolicySet(ctx context.Context, policySetID string, p Patch) (*PolicySetParameter, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	return s.createFields(ctx, s.client.apiPath("policy_set", policySetID, "parameters"), p)
}

// UpdateFieldsByPolicySet modifies only the attributes and relationships set in p on a
// policy set parameter of a policy set, through the top-level /policy_set/{id}/parameters/{id} path; a Relationship with a nil ID
// clears that link.
// It returns a *ValidationError if policySetID or id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicySetParameterService) UpdateFieldsByPolicySet(ctx context.Context, policySetID, id string, p Patch) (*PolicySetParameter, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policySetParameterID", id); err != nil {
		return nil, err
	}

	return s.updateFields(ctx, s.client.apiPath("policy_set", policySetID, "parameters", id), id, p)
}

// DeleteByPolicySet removes a policy set parameter of a policy set through the top-level
// /policy_set/{id}/parameters/{id} path.
// It returns a *ValidationError if policySetID or id is empty and a *APIError on server errors.
func (s *PolicySetParameterService) DeleteByPolicySet(ctx context.Context, policySetID, id string) error {
	if err := validateID("policySetID", policySetID); err != nil {
		return err
	}
	if err := validateID("policySetParameterID", id); err != nil {
		return err
	}

	return s.del(ctx, s.client.apiPath("policy_set", policySetID, "parameters", id))
}
