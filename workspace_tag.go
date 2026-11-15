package terrakube

import (
	"context"
	"time"
)

// WorkspaceTag represents a tag association on a workspace.
type WorkspaceTag struct {
	ID          string     `jsonapi:"primary,workspacetag"`
	TagID       string     `jsonapi:"attr,tagId"`
	CreatedBy   *string    `jsonapi:"attr,createdBy"`
	CreatedDate *time.Time `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy   *string    `jsonapi:"attr,updatedBy"`
	UpdatedDate *time.Time `jsonapi:"attr,updatedDate,rfc3339"`
	Value       *string    `jsonapi:"attr,value"`
	Workspace   *Workspace `jsonapi:"relation,workspace,omitempty"`
}

// WorkspaceTagService handles communication with the workspace tag related
// methods of the Terrakube API.
type WorkspaceTagService struct {
	crudService[WorkspaceTag]
}

// List returns all tags for a workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *WorkspaceTagService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag")
	return s.list(ctx, path, opts)
}

// Get retrieves a single workspace tag by ID.
// It returns a *ValidationError if orgID, workspaceID, or tagID is empty and a *APIError on server errors.
func (s *WorkspaceTagService) Get(ctx context.Context, orgID, workspaceID, tagID string) (*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("tagID", tagID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag", tagID)
	return s.get(ctx, path)
}

// Create creates a new tag association on a workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *WorkspaceTagService) Create(ctx context.Context, orgID, workspaceID string, tag *WorkspaceTag) (*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag")
	return s.create(ctx, path, tag)
}

// CreateFields creates a new tag association on a workspace, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID or workspaceID is empty and a *APIError on server errors.
func (s *WorkspaceTagService) CreateFields(ctx context.Context, orgID, workspaceID string, p Patch) (*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing workspace tag. The workspace tag's ID field must be set.
// It returns a *ValidationError if orgID, workspaceID, or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *WorkspaceTagService) Update(ctx context.Context, orgID, workspaceID string, tag *WorkspaceTag) (*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("tagID", tag.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag", tag.ID)
	return s.update(ctx, path, tag)
}

// UpdateFields modifies only the attributes and relationships set in p on the workspace tag
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID, workspaceID, or id is empty or p sets nothing, and a *APIError on server errors.
func (s *WorkspaceTagService) UpdateFields(ctx context.Context, orgID, workspaceID, id string, p Patch) (*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("tagID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes a tag association from a workspace.
// It returns a *ValidationError if orgID, workspaceID, or tagID is empty and a *APIError on server errors.
func (s *WorkspaceTagService) Delete(ctx context.Context, orgID, workspaceID, tagID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("tagID", tagID); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag", tagID)
	return s.del(ctx, path)
}
