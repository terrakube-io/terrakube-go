package terrakube

import "context"

// Workspace represents a Terrakube workspace resource.
type Workspace struct {
	ID               string  `jsonapi:"primary,workspace"`
	Name             string  `jsonapi:"attr,name"`
	Description      *string `jsonapi:"attr,description"`
	Source           string  `jsonapi:"attr,source"`
	Branch           string  `jsonapi:"attr,branch"`
	Folder           string  `jsonapi:"attr,folder"`
	TemplateID       string  `jsonapi:"attr,defaultTemplate"`
	IaCType          string  `jsonapi:"attr,iacType"`
	IaCVersion       string  `jsonapi:"attr,terraformVersion"`
	ExecutionMode    string  `jsonapi:"attr,executionMode"`
	Deleted          bool    `jsonapi:"attr,deleted"`
	Locked           bool    `jsonapi:"attr,locked"`
	AllowRemoteApply bool    `jsonapi:"attr,allowRemoteApply"`
	LockDescription  *string `jsonapi:"attr,lockDescription"`
	ModuleSSHKey     *string `jsonapi:"attr,moduleSshKey"`
	LastJobStatus    *string `jsonapi:"attr,lastJobStatus"`
	LastJobDate      *string `jsonapi:"attr,lastJobDate"`
	CreatedBy        *string `jsonapi:"attr,createdBy"`
	CreatedDate      *string `jsonapi:"attr,createdDate"`
	UpdatedBy        *string `jsonapi:"attr,updatedBy"`
	UpdatedDate      *string `jsonapi:"attr,updatedDate"`
	Vcs              *VCS    `jsonapi:"relation,vcs,omitempty"`
}

// WorkspaceService handles communication with the workspace related
// methods of the Terrakube API.
type WorkspaceService struct {
	crudService[Workspace]
}

// List returns all workspaces for an organization, optionally filtered.
func (s *WorkspaceService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace")
	return s.list(ctx, path, opts)
}

// Get retrieves a workspace by ID within an organization.
func (s *WorkspaceService) Get(ctx context.Context, orgID, id string) (*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", id)
	return s.get(ctx, path)
}

// Create creates a new workspace within an organization.
func (s *WorkspaceService) Create(ctx context.Context, orgID string, ws *Workspace) (*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace")
	return s.create(ctx, path, ws)
}

// Update modifies an existing workspace within an organization.
func (s *WorkspaceService) Update(ctx context.Context, orgID string, ws *Workspace) (*Workspace, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspace ID", ws.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", ws.ID)
	return s.update(ctx, path, ws)
}

// Delete removes a workspace by ID within an organization.
func (s *WorkspaceService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", id)
	return s.del(ctx, path)
}
