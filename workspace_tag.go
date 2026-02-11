package terrakube

import "context"

// WorkspaceTag represents a tag association on a workspace.
type WorkspaceTag struct {
	ID          string  `jsonapi:"primary,workspacetag"`
	TagID       string  `jsonapi:"attr,tagId"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// WorkspaceTagService handles communication with the workspace tag related
// methods of the Terrakube API.
type WorkspaceTagService struct {
	crudService[WorkspaceTag]
}

// List returns all tags for a workspace.
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

// Update modifies an existing workspace tag.
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

// Delete removes a tag association from a workspace.
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
