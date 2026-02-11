package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Tag represents a Terrakube tag resource.
type Tag struct {
	ID   string `jsonapi:"primary,tag"`
	Name string `jsonapi:"attr,name"`
}

// TagService handles communication with the tag-related endpoints.
type TagService struct {
	client *Client
}

// List returns all tags for the given organization.
func (s *TagService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "tag")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[tag]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var tags []*Tag
	if _, err := s.client.do(ctx, req, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// Get returns a single tag by ID within the given organization.
func (s *TagService) Get(ctx context.Context, orgID, id string) (*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("tagID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "tag", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	tag := &Tag{}
	if _, err := s.client.do(ctx, req, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// Create creates a new tag in the given organization.
func (s *TagService) Create(ctx context.Context, orgID string, tag *Tag) (*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "tag")

	req, err := s.client.request(ctx, http.MethodPost, p, tag)
	if err != nil {
		return nil, err
	}

	created := &Tag{}
	if _, err := s.client.do(ctx, req, created); err != nil {
		return nil, err
	}
	return created, nil
}

// Update modifies an existing tag in the given organization.
func (s *TagService) Update(ctx context.Context, orgID string, tag *Tag) (*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("tagID", tag.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "tag", tag.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, tag)
	if err != nil {
		return nil, err
	}

	updated := &Tag{}
	if _, err := s.client.do(ctx, req, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete removes a tag from the given organization.
func (s *TagService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("tagID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "tag", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	if _, err := s.client.do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}
