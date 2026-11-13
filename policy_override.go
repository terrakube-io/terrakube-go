package terrakube

import "context"

// PolicyOverride represents a Terrakube policy override resource.
type PolicyOverride struct {
	ID            string            `jsonapi:"primary,policy_override"`
	Justification *string           `jsonapi:"attr,justification"`
	OverriddenAt  *string           `jsonapi:"attr,overriddenAt"`
	OverriddenBy  *string           `jsonapi:"attr,overriddenBy"`
	CreatedBy     *string           `jsonapi:"attr,createdBy"`
	CreatedDate   *string           `jsonapi:"attr,createdDate"`
	UpdatedBy     *string           `jsonapi:"attr,updatedBy"`
	UpdatedDate   *string           `jsonapi:"attr,updatedDate"`
	Evaluation    *PolicyEvaluation `jsonapi:"relation,evaluation,omitempty"`
}

// PolicyOverrideService handles communication with the policy override related
// methods of the Terrakube API.
type PolicyOverrideService struct {
	crudService[PolicyOverride]
}

// Get retrieves a single policy override by ID.
// It returns a *ValidationError if orgID, jobID, policyEvaluationID, or policyOverrideID is empty and a *APIError on server errors.
func (s *PolicyOverrideService) Get(ctx context.Context, orgID, jobID, policyEvaluationID, policyOverrideID string) (*PolicyOverride, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("policyEvaluationID", policyEvaluationID); err != nil {
		return nil, err
	}
	if err := validateID("policyOverrideID", policyOverrideID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "policyEvaluation", policyEvaluationID, "override", policyOverrideID)
	return s.get(ctx, path)
}

// Create creates a new policy override.
// It returns a *ValidationError if orgID, jobID, or policyEvaluationID is empty and a *APIError on server errors.
func (s *PolicyOverrideService) Create(ctx context.Context, orgID, jobID, policyEvaluationID string, entity *PolicyOverride) (*PolicyOverride, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("policyEvaluationID", policyEvaluationID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "policyEvaluation", policyEvaluationID, "override")
	return s.create(ctx, path, entity)
}

// Update modifies an existing policy override. The policy override's ID field must be set.
// It returns a *ValidationError if orgID, jobID, policyEvaluationID, or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *PolicyOverrideService) Update(ctx context.Context, orgID, jobID, policyEvaluationID string, entity *PolicyOverride) (*PolicyOverride, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("policyEvaluationID", policyEvaluationID); err != nil {
		return nil, err
	}
	if err := validateID("policyOverrideID", entity.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "policyEvaluation", policyEvaluationID, "override", entity.ID)
	return s.update(ctx, path, entity)
}
