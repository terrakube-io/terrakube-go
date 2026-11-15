package terrakube

import (
	"context"
	"net/http"
)

// Permissions are the caller's effective permissions, as the server computes
// them from the teams the caller belongs to.
type Permissions struct {
	ManageState      bool `json:"manageState"`
	ManageWorkspace  bool `json:"manageWorkspace"`
	ManageModule     bool `json:"manageModule"`
	ManageProvider   bool `json:"manageProvider"`
	ManageVcs        bool `json:"manageVcs"`
	ManageTemplate   bool `json:"manageTemplate"`
	ManageCollection bool `json:"manageCollection"`
	ManageJob        bool `json:"manageJob"`
	PlanJob          bool `json:"planJob"`
	ApproveJob       bool `json:"approveJob"`
	// ManagePermission is true for a team that can manage workspaces, and for
	// the instance owner.
	ManagePermission bool `json:"managePermission"`
	ManagePolicies   bool `json:"managePolicies"`
}

// AccessService reads the caller's teams and effective permissions
// (/access-token/v1/teams/...).
type AccessService struct {
	client *Client
}

// CurrentTeams returns the names of the groups the caller belongs to, as the
// server sees them (for a federated token, the groups of its federated
// credentials).
//
// It returns a *APIError on server errors.
func (s *AccessService) CurrentTeams(ctx context.Context) ([]string, error) {
	req, err := s.client.requestRaw(ctx, http.MethodGet, teamTokenBasePath+"/current-teams", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Groups []string `json:"groups"`
	}
	if _, err := s.client.doRaw(req, &resp); err != nil {
		return nil, err
	}
	if resp.Groups == nil {
		return []string{}, nil
	}
	return resp.Groups, nil
}

// Permissions returns the caller's permissions in an organization: the union
// of the permissions of the organization's teams the caller belongs to. The
// server answers an organization the caller has no team in, or that does not
// exist, with no permissions rather than an error.
//
// It returns a *ValidationError if orgID is empty and a *APIError on server
// errors.
func (s *AccessService) Permissions(ctx context.Context, orgID string) (*Permissions, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	return s.get(ctx, teamTokenBasePath+"/permissions/organization/"+orgID)
}

// WorkspacePermissions returns the caller's permissions in a workspace: those
// of [AccessService.Permissions] plus the state, workspace and job
// permissions granted by the workspace's access entries and its project's
// access entries for the caller's teams. An unknown workspace adds nothing.
//
// It returns a *ValidationError if orgID or wsID is empty and a *APIError on
// server errors.
func (s *AccessService) WorkspacePermissions(ctx context.Context, orgID, wsID string) (*Permissions, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", wsID); err != nil {
		return nil, err
	}
	return s.get(ctx, teamTokenBasePath+"/permissions/organization/"+orgID+"/workspace/"+wsID)
}

func (s *AccessService) get(ctx context.Context, path string) (*Permissions, error) {
	req, err := s.client.requestRaw(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	result := &Permissions{}
	if _, err := s.client.doRaw(req, result); err != nil {
		return nil, err
	}
	return result, nil
}
