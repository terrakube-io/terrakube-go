package terrakube

import "context"

// WorkspaceSchedule represents a scheduled job for a workspace.
type WorkspaceSchedule struct {
	ID          string  `jsonapi:"primary,schedule"`
	// Schedule is the cron expression for the schedule (JSON:API attr: "cron").
	Schedule string `jsonapi:"attr,cron"`
	// TemplateID is the template reference (JSON:API attr: "templateReference").
	TemplateID string `jsonapi:"attr,templateReference"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// WorkspaceScheduleService handles communication with the workspace schedule
// related methods of the Terrakube API.
type WorkspaceScheduleService struct {
	crudService[WorkspaceSchedule]
}

// List returns all schedules for the given workspace.
// It returns a *ValidationError if workspaceID is empty and a *APIError on server errors.
func (s *WorkspaceScheduleService) List(ctx context.Context, workspaceID string, opts *ListOptions) ([]*WorkspaceSchedule, error) {
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("workspace", workspaceID, "schedule")
	return s.list(ctx, path, opts)
}

// Get returns a single workspace schedule by ID.
// It returns a *ValidationError if workspaceID or id is empty and a *APIError on server errors.
func (s *WorkspaceScheduleService) Get(ctx context.Context, workspaceID, id string) (*WorkspaceSchedule, error) {
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("scheduleID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("workspace", workspaceID, "schedule", id)
	return s.get(ctx, path)
}

// Create creates a new schedule for the given workspace.
// It returns a *ValidationError if workspaceID is empty and a *APIError on server errors.
func (s *WorkspaceScheduleService) Create(ctx context.Context, workspaceID string, schedule *WorkspaceSchedule) (*WorkspaceSchedule, error) {
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("workspace", workspaceID, "schedule")
	return s.create(ctx, path, schedule)
}

// Update modifies an existing workspace schedule. The schedule's ID field must be set.
// It returns a *ValidationError if workspaceID or the ID is empty and a *APIError on server errors.
func (s *WorkspaceScheduleService) Update(ctx context.Context, workspaceID string, schedule *WorkspaceSchedule) (*WorkspaceSchedule, error) {
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("scheduleID", schedule.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("workspace", workspaceID, "schedule", schedule.ID)
	return s.update(ctx, path, schedule)
}

// Delete removes a workspace schedule by ID.
// It returns a *ValidationError if workspaceID or id is empty and a *APIError on server errors.
func (s *WorkspaceScheduleService) Delete(ctx context.Context, workspaceID, id string) error {
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("scheduleID", id); err != nil {
		return err
	}

	path := s.client.apiPath("workspace", workspaceID, "schedule", id)
	return s.del(ctx, path)
}
