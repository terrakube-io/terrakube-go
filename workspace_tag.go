package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// WorkspaceTag represents a tag association on a workspace.
type WorkspaceTag struct {
	ID    string `jsonapi:"primary,workspacetag"`
	TagID string `jsonapi:"attr,tagId"`
}

// WorkspaceTagService handles communication with the workspace tag related
// methods of the Terrakube API.
type WorkspaceTagService struct {
	client *Client
}

// List returns all tags for a workspace.
func (s *WorkspaceTagService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[workspacetag]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var tags []*WorkspaceTag
	_, err = s.client.do(ctx, req, &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
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

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag", tagID)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	tag := &WorkspaceTag{}
	_, err = s.client.do(ctx, req, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// Create creates a new tag association on a workspace.
func (s *WorkspaceTagService) Create(ctx context.Context, orgID, workspaceID string, tag *WorkspaceTag) (*WorkspaceTag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag")

	req, err := s.client.request(ctx, http.MethodPost, p, tag)
	if err != nil {
		return nil, err
	}

	created := &WorkspaceTag{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
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

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag", tag.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, tag)
	if err != nil {
		return nil, err
	}

	updated := &WorkspaceTag{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
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

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "workspaceTag", tagID)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
