package terrakube

import "context"

// WorkspaceAccess represents access control settings for a workspace.
type WorkspaceAccess struct {
	ID              string  `jsonapi:"primary,access"`
	ManageState     bool    `jsonapi:"attr,manageState"`
	ManageWorkspace bool    `jsonapi:"attr,manageWorkspace"`
	ManageJob       bool    `jsonapi:"attr,manageJob"`
	Name            string  `jsonapi:"attr,name"`
	CreatedBy       *string `jsonapi:"attr,createdBy"`
	CreatedDate     *string `jsonapi:"attr,createdDate"`
	UpdatedBy       *string `jsonapi:"attr,updatedBy"`
	UpdatedDate     *string `jsonapi:"attr,updatedDate"`
}

// WorkspaceAccessService handles communication with the workspace access related
// methods of the Terrakube API.
type WorkspaceAccessService struct {
	crudService[WorkspaceAccess]
}

// List returns all access entries for the given workspace.
func (s *WorkspaceAccessService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access")
	return s.list(ctx, path, opts)
}

// Get returns a single workspace access entry by ID.
func (s *WorkspaceAccessService) Get(ctx context.Context, orgID, workspaceID, id string) (*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("accessID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access", id)
	return s.get(ctx, path)
}

// Create creates a new access entry for the given workspace.
func (s *WorkspaceAccessService) Create(ctx context.Context, orgID, workspaceID string, access *WorkspaceAccess) (*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access")
	return s.create(ctx, path, access)
}

// Update modifies an existing workspace access entry. The access entry's ID field must be set.
func (s *WorkspaceAccessService) Update(ctx context.Context, orgID, workspaceID string, access *WorkspaceAccess) (*WorkspaceAccess, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("accessID", access.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access", access.ID)
	return s.update(ctx, path, access)
}

// Delete removes a workspace access entry by ID.
func (s *WorkspaceAccessService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("accessID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "access", id)
	return s.del(ctx, path)
}
