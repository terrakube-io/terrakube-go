package terrakube

import (
	"context"
	"time"
)

// Job represents a Terrakube job resource.
//
// The attributes marked omitempty are set by the server (status, approval,
// dispatch and plan artifact data). It rejects a create request that mentions
// any of them, even as null, so they are sent only when set.
type Job struct {
	ID                             string              `jsonapi:"primary,job"`
	Output                         string              `jsonapi:"attr,output"`
	Status                         string              `jsonapi:"attr,status,omitempty"`
	ApprovalTeam                   *string             `jsonapi:"attr,approvalTeam,omitempty"`
	Comments                       *string             `jsonapi:"attr,comments"`
	CommitID                       *string             `jsonapi:"attr,commitId"`
	OverrideBranch                 *string             `jsonapi:"attr,overrideBranch"`
	OverrideSource                 *string             `jsonapi:"attr,overrideSource,omitempty"`
	PlanChanges                    bool                `jsonapi:"attr,planChanges"`
	Refresh                        bool                `jsonapi:"attr,refresh"`
	RefreshOnly                    bool                `jsonapi:"attr,refreshOnly"`
	PRNumber                       *int                `jsonapi:"attr,prNumber,omitempty"`
	PRCommentError                 *string             `jsonapi:"attr,prCommentError,omitempty"`
	TemplateReference              *string             `jsonapi:"attr,templateReference"`
	TerraformPlan                  *string             `jsonapi:"attr,terraformPlan"`
	Via                            *string             `jsonapi:"attr,via"`
	TargetAddrs                    []string            `jsonapi:"attr,targetAddrs,omitempty"`
	ReplaceAddrs                   []string            `jsonapi:"attr,replaceAddrs,omitempty"`
	CreatedBy                      *string             `jsonapi:"attr,createdBy"`
	CreatedDate                    *time.Time          `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy                      *string             `jsonapi:"attr,updatedBy"`
	UpdatedDate                    *time.Time          `jsonapi:"attr,updatedDate,rfc3339"`
	ApprovedAt                     *time.Time          `jsonapi:"attr,approvedAt,omitempty,rfc3339"`
	ApprovedBy                     *string             `jsonapi:"attr,approvedBy,omitempty"`
	CascadeDepth                   *int32              `jsonapi:"attr,cascadeDepth,omitempty"`
	DispatchFailureCount           *int32              `jsonapi:"attr,dispatchFailureCount,omitempty"`
	DispatchFirstFailureAt         *time.Time          `jsonapi:"attr,dispatchFirstFailureAt,omitempty,rfc3339"`
	DispatchLastAttemptAt          *time.Time          `jsonapi:"attr,dispatchLastAttemptAt,omitempty,rfc3339"`
	DispatchLastError              *string             `jsonapi:"attr,dispatchLastError,omitempty"`
	DispatchNextRetryAt            *time.Time          `jsonapi:"attr,dispatchNextRetryAt,omitempty,rfc3339"`
	EncryptedVariables             *string             `jsonapi:"attr,encryptedVariables,omitempty"`
	ResolvedTerraformVersion       *string             `jsonapi:"attr,resolvedTerraformVersion,omitempty"`
	TerraformPlanArtifacts         *string             `jsonapi:"attr,terraformPlanArtifacts,omitempty"`
	TerraformPlanArtifactsChecksum *string             `jsonapi:"attr,terraformPlanArtifactsChecksum,omitempty"`
	TriggeredByJobID               *int32              `jsonapi:"attr,triggeredByJobId,omitempty"`
	VariableNames                  []string            `jsonapi:"attr,variableNames,omitempty"`
	Variables                      map[string]any      `jsonapi:"attr,variables"`
	VariablesAsArguments           bool                `jsonapi:"attr,variablesAsArguments,omitempty"`
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
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
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
