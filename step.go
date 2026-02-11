package terrakube

import "context"

// Step represents a Terrakube step resource within a job.
type Step struct {
	ID          string  `jsonapi:"primary,step"`
	Name        string  `jsonapi:"attr,name"`
	Output      *string `jsonapi:"attr,output"`
	Status      string  `jsonapi:"attr,status"`
	StepNumber  int     `jsonapi:"attr,stepNumber"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
	Job         *Job    `jsonapi:"relation,job,omitempty"`
}

// StepService handles communication with the step related methods of the
// Terrakube API.
type StepService struct {
	crudService[Step]
}

// List returns all steps for the given job within an organization.
func (s *StepService) List(ctx context.Context, orgID, jobID string, opts *ListOptions) ([]*Step, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step")
	return s.list(ctx, path, opts)
}

// Get returns a single step by ID.
func (s *StepService) Get(ctx context.Context, orgID, jobID, id string) (*Step, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("step ID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step", id)
	return s.get(ctx, path)
}

// Create creates a new step in the given job.
func (s *StepService) Create(ctx context.Context, orgID, jobID string, step *Step) (*Step, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step")
	return s.create(ctx, path, step)
}

// Update modifies an existing step. The step's ID field must be set.
func (s *StepService) Update(ctx context.Context, orgID, jobID string, step *Step) (*Step, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("step ID", step.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step", step.ID)
	return s.update(ctx, path, step)
}

// Delete removes a step by ID.
func (s *StepService) Delete(ctx context.Context, orgID, jobID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("job ID", jobID); err != nil {
		return err
	}
	if err := validateID("step ID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "step", id)
	return s.del(ctx, path)
}
