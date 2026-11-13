package terrakube

import "context"

// Job represents a Terrakube job resource.
type Job struct {
	ID                             string              `jsonapi:"primary,job"`
	Command                        string              `jsonapi:"attr,command"` // DEPRECATED (absent from the Terrakube API 2.34.1 OpenAPI document; kept for compatibility).
	Output                         string              `jsonapi:"attr,output"`
	Status                         string              `jsonapi:"attr,status"`
	ApprovalTeam                   *string             `jsonapi:"attr,approvalTeam"`
	Comments                       *string             `jsonapi:"attr,comments"`
	CommitID                       *string             `jsonapi:"attr,commitId"`
	OverrideBranch                 *string             `jsonapi:"attr,overrideBranch"`
	OverrideSource                 *string             `jsonapi:"attr,overrideSource"`
	PlanChanges                    bool                `jsonapi:"attr,planChanges"`
	Refresh                        bool                `jsonapi:"attr,refresh"`
	RefreshOnly                    bool                `jsonapi:"attr,refreshOnly"`
	PRNumber                       *int                `jsonapi:"attr,prNumber,omitempty"`
	PRCommentError                 *string             `jsonapi:"attr,prCommentError,omitempty"`
	Tcl                            *string             `jsonapi:"attr,tcl"` // DEPRECATED (absent from the Terrakube API 2.34.1 OpenAPI document; kept for compatibility).
	TemplateReference              *string             `jsonapi:"attr,templateReference"`
	TerraformPlan                  *string             `jsonapi:"attr,terraformPlan"`
	Via                            *string             `jsonapi:"attr,via"`
	TargetAddrs                    []string            `jsonapi:"attr,targetAddrs,omitempty"`
	ReplaceAddrs                   []string            `jsonapi:"attr,replaceAddrs,omitempty"`
	CreatedBy                      *string             `jsonapi:"attr,createdBy"`
	CreatedDate                    *string             `jsonapi:"attr,createdDate"`
	UpdatedBy                      *string             `jsonapi:"attr,updatedBy"`
	UpdatedDate                    *string             `jsonapi:"attr,updatedDate"`
	ApprovedAt                     *string             `jsonapi:"attr,approvedAt"`
	ApprovedBy                     *string             `jsonapi:"attr,approvedBy"`
	CascadeDepth                   *int32              `jsonapi:"attr,cascadeDepth"`
	DispatchFailureCount           *int32              `jsonapi:"attr,dispatchFailureCount"`
	DispatchFirstFailureAt         *string             `jsonapi:"attr,dispatchFirstFailureAt"`
	DispatchLastAttemptAt          *string             `jsonapi:"attr,dispatchLastAttemptAt"`
	DispatchLastError              *string             `jsonapi:"attr,dispatchLastError"`
	DispatchNextRetryAt            *string             `jsonapi:"attr,dispatchNextRetryAt"`
	EncryptedVariables             *string             `jsonapi:"attr,encryptedVariables"`
	ResolvedTerraformVersion       *string             `jsonapi:"attr,resolvedTerraformVersion"`
	TerraformPlanArtifacts         *string             `jsonapi:"attr,terraformPlanArtifacts"`
	TerraformPlanArtifactsChecksum *string             `jsonapi:"attr,terraformPlanArtifactsChecksum"`
	TriggeredByJobID               *int32              `jsonapi:"attr,triggeredByJobId"`
	VariableNames                  []string            `jsonapi:"attr,variableNames"`
	Variables                      map[string]any      `jsonapi:"attr,variables"`
	VariablesAsArguments           bool                `jsonapi:"attr,variablesAsArguments"`
	Organization                   *Organization       `jsonapi:"relation,organization,omitempty"`
	Workspace                      *Workspace          `jsonapi:"relation,workspace,omitempty"`
	Steps                          []*Step             `jsonapi:"relation,step,omitempty"`
	Addresses                      []*Address          `jsonapi:"relation,address,omitempty"`
	PolicyEvaluation               []*PolicyEvaluation `jsonapi:"relation,policyEvaluation,omitempty"`
}

// JobService handles communication with the job related methods of the
// Terrakube API.
type JobService struct {
	crudService[Job]
}

// List returns all jobs for the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *JobService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job")
	return s.list(ctx, path, opts)
}

// Get returns a single job by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *JobService) Get(ctx context.Context, orgID, id string) (*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", id)
	return s.get(ctx, path)
}

// Create creates a new job in the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *JobService) Create(ctx context.Context, orgID string, job *Job) (*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job")
	return s.create(ctx, path, job)
}

// Update modifies an existing job. The job's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
func (s *JobService) Update(ctx context.Context, orgID string, job *Job) (*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", job.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", job.ID)
	return s.update(ctx, path, job)
}

// Delete removes a job by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *JobService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("jobID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "job", id)
	return s.del(ctx, path)
}
