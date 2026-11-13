package terrakube

import (
	"context"
	"time"
)

// RunTrigger represents a Terrakube run trigger resource.
type RunTrigger struct {
	ID                   string     `jsonapi:"primary,runTrigger"`
	Enabled              bool       `jsonapi:"attr,enabled"`
	CreatedBy            *string    `jsonapi:"attr,createdBy"`
	CreatedDate          *time.Time `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy            *string    `jsonapi:"attr,updatedBy"`
	UpdatedDate          *time.Time `jsonapi:"attr,updatedDate,rfc3339"`
	DestinationWorkspace *Workspace `jsonapi:"relation,destinationWorkspace,omitempty"`
	SourceWorkspace      *Workspace `jsonapi:"relation,sourceWorkspace,omitempty"`
	Template             *Template  `jsonapi:"relation,template,omitempty"`
}

// RunTriggerService handles communication with the run trigger related
// methods of the Terrakube API.
type RunTriggerService struct {
	crudService[RunTrigger]
}

// List returns all run triggers for a workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *RunTriggerService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*RunTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "runTriggers")
	return s.list(ctx, path, opts)
}

// Get retrieves a single run trigger by ID.
// It returns a *ValidationError if orgID, workspaceID, or runTriggerID is empty and a *APIError on server errors.
func (s *RunTriggerService) Get(ctx context.Context, orgID, workspaceID, runTriggerID string) (*RunTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("runTriggerID", runTriggerID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "runTriggers", runTriggerID)
	return s.get(ctx, path)
}

// Create creates a new run trigger.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *RunTriggerService) Create(ctx context.Context, orgID, workspaceID string, entity *RunTrigger) (*RunTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "runTriggers")
	return s.create(ctx, path, entity)
}

// Update modifies an existing run trigger. The run trigger's ID field must be set.
// It returns a *ValidationError if orgID, workspaceID, or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *RunTriggerService) Update(ctx context.Context, orgID, workspaceID string, entity *RunTrigger) (*RunTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("runTriggerID", entity.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "runTriggers", entity.ID)
	return s.update(ctx, path, entity)
}

// Delete removes a run trigger.
// It returns a *ValidationError if orgID, workspaceID, or runTriggerID is empty and a *APIError on server errors.
func (s *RunTriggerService) Delete(ctx context.Context, orgID, workspaceID, runTriggerID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("runTriggerID", runTriggerID); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "runTriggers", runTriggerID)
	return s.del(ctx, path)
}
