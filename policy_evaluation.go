package terrakube

import "context"

// PolicyEvaluation represents a Terrakube policy evaluation resource.
type PolicyEvaluation struct {
	ID                      string          `jsonapi:"primary,policy_evaluation"`
	HardMandatoryViolations *int32          `jsonapi:"attr,hardMandatoryViolations"`
	PassedRules             *int32          `jsonapi:"attr,passedRules"`
	ShadowHardViolations    *int32          `jsonapi:"attr,shadowHardViolations"`
	ShadowSoftViolations    *int32          `jsonapi:"attr,shadowSoftViolations"`
	SoftMandatoryViolations *int32          `jsonapi:"attr,softMandatoryViolations"`
	Status                  *string         `jsonapi:"attr,status"`
	StorageURI              *string         `jsonapi:"attr,storageUri"`
	WarningRules            *int32          `jsonapi:"attr,warningRules"`
	CreatedBy               *string         `jsonapi:"attr,createdBy"`
	CreatedDate             *string         `jsonapi:"attr,createdDate"`
	UpdatedBy               *string         `jsonapi:"attr,updatedBy"`
	UpdatedDate             *string         `jsonapi:"attr,updatedDate"`
	Job                     *Job            `jsonapi:"relation,job,omitempty"`
	Override                *PolicyOverride `jsonapi:"relation,override,omitempty"`
	Step                    *Step           `jsonapi:"relation,step,omitempty"`
}

// PolicyEvaluationService handles communication with the policy evaluation related
// methods of the Terrakube API.
type PolicyEvaluationService struct {
	crudService[PolicyEvaluation]
}

// List returns all policy evaluations for a job.
// It returns a *ValidationError if orgID or jobID is empty and a *APIError on server errors.
func (s *PolicyEvaluationService) List(ctx context.Context, orgID, jobID string, opts *ListOptions) ([]*PolicyEvaluation, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "policyEvaluation")
	return s.list(ctx, path, opts)
}

// Get retrieves a single policy evaluation by ID.
// It returns a *ValidationError if orgID, jobID, or policyEvaluationID is empty and a *APIError on server errors.
func (s *PolicyEvaluationService) Get(ctx context.Context, orgID, jobID, policyEvaluationID string) (*PolicyEvaluation, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("policyEvaluationID", policyEvaluationID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "policyEvaluation", policyEvaluationID)
	return s.get(ctx, path)
}
