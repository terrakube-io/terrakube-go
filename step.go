package terrakube

import "context"

// Step represents a Terrakube step resource within a job.
type Step struct {
	ID                string              `jsonapi:"primary,step"`
	Name              string              `jsonapi:"attr,name"`
	Output            *string             `jsonapi:"attr,output"`
	Status            string              `jsonapi:"attr,status"`
	LogStatus         *string             `jsonapi:"attr,logStatus,omitempty"` // DEPRECATED (absent from the Terrakube API 2.34.1 OpenAPI document; kept for compatibility).
	StepNumber        int                 `jsonapi:"attr,stepNumber"`
	CreatedBy         *string             `jsonapi:"attr,createdBy"`   // DEPRECATED (absent from the Terrakube API 2.34.1 OpenAPI document; kept for compatibility).
	CreatedDate       *string             `jsonapi:"attr,createdDate"` // DEPRECATED (absent from the Terrakube API 2.34.1 OpenAPI document; kept for compatibility).
	UpdatedBy         *string             `jsonapi:"attr,updatedBy"`   // DEPRECATED (absent from the Terrakube API 2.34.1 OpenAPI document; kept for compatibility).
	UpdatedDate       *string             `jsonapi:"attr,updatedDate"` // DEPRECATED (absent from the Terrakube API 2.34.1 OpenAPI document; kept for compatibility).
	DeliveryRevision  *int64              `jsonapi:"attr,deliveryRevision"`
	DeliveryState     *string             `jsonapi:"attr,deliveryState"`
	DeliveryTimestamp *int64              `jsonapi:"attr,deliveryTimestamp"`
	Job               *Job                `jsonapi:"relation,job,omitempty"`
	PolicyEvaluation  []*PolicyEvaluation `jsonapi:"relation,policyEvaluation,omitempty"`
}

// StepService handles communication with the step related methods of the
// Terrakube API.
type StepService struct {
	crudService[Step]
}

// List returns all steps for the given job within an organization.
// It returns a *ValidationError if orgID or jobID is empty and a *APIError on server errors.
func (s *StepService) List(ctx context.Context, orgID, jobID string, opts *ListOptions) ([]*Step, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step")
	return s.list(ctx, path, opts)
}

// Get returns a single step by ID.
// It returns a *ValidationError if orgID, jobID, or id is empty and a *APIError on server errors.
func (s *StepService) Get(ctx context.Context, orgID, jobID, id string) (*Step, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("stepID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step", id)
	return s.get(ctx, path)
}

// Create creates a new step in the given job.
// It returns a *ValidationError if orgID or jobID is empty and a *APIError on server errors.
func (s *StepService) Create(ctx context.Context, orgID, jobID string, step *Step) (*Step, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step")
	return s.create(ctx, path, step)
}

// Update modifies an existing step. The step's ID field must be set.
// It returns a *ValidationError if orgID, jobID, or the ID is empty and a *APIError on server errors.
func (s *StepService) Update(ctx context.Context, orgID, jobID string, step *Step) (*Step, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("stepID", step.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step", step.ID)
	return s.update(ctx, path, step)
}

// Delete removes a step by ID.
// It returns a *ValidationError if orgID, jobID, or id is empty and a *APIError on server errors.
func (s *StepService) Delete(ctx context.Context, orgID, jobID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("jobID", jobID); err != nil {
		return err
	}
	if err := validateID("stepID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step", id)
	return s.del(ctx, path)
}
