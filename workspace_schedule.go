package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// WorkspaceSchedule represents a scheduled job for a workspace.
type WorkspaceSchedule struct {
	ID         string `jsonapi:"primary,schedule"`
	Schedule   string `jsonapi:"attr,cron"`
	TemplateID string `jsonapi:"attr,templateReference"`
}

// WorkspaceScheduleService handles communication with the workspace schedule
// related methods of the Terrakube API.
type WorkspaceScheduleService struct {
	client *Client
}

// List returns all schedules for the given workspace.
func (s *WorkspaceScheduleService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*WorkspaceSchedule, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "schedule")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[schedule]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var schedules []*WorkspaceSchedule
	_, err = s.client.do(ctx, req, &schedules)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

// Get returns a single workspace schedule by ID.
func (s *WorkspaceScheduleService) Get(ctx context.Context, orgID, workspaceID, id string) (*WorkspaceSchedule, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("scheduleID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "schedule", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	schedule := &WorkspaceSchedule{}
	_, err = s.client.do(ctx, req, schedule)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

// Create creates a new schedule for the given workspace.
func (s *WorkspaceScheduleService) Create(ctx context.Context, orgID, workspaceID string, schedule *WorkspaceSchedule) (*WorkspaceSchedule, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "schedule")

	req, err := s.client.request(ctx, http.MethodPost, p, schedule)
	if err != nil {
		return nil, err
	}

	created := &WorkspaceSchedule{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing workspace schedule. The schedule's ID field must be set.
func (s *WorkspaceScheduleService) Update(ctx context.Context, orgID, workspaceID string, schedule *WorkspaceSchedule) (*WorkspaceSchedule, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("scheduleID", schedule.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "schedule", schedule.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, schedule)
	if err != nil {
		return nil, err
	}

	updated := &WorkspaceSchedule{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a workspace schedule by ID.
func (s *WorkspaceScheduleService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("scheduleID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "schedule", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
