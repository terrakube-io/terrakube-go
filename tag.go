package terrakube

import "context"

// Tag represents a Terrakube tag resource.
type Tag struct {
	ID          string  `jsonapi:"primary,tag"`
	Name        string  `jsonapi:"attr,name"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// TagService handles communication with the tag-related endpoints.
type TagService struct {
	crudService[Tag]
}

// List returns all tags for the given organization.
func (s *TagService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "tag")
	return s.list(ctx, path, opts)
}

// Get returns a single tag by ID within the given organization.
func (s *TagService) Get(ctx context.Context, orgID, id string) (*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("tagID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "tag", id)
	return s.get(ctx, path)
}

// Create creates a new tag in the given organization.
func (s *TagService) Create(ctx context.Context, orgID string, tag *Tag) (*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "tag")
	return s.create(ctx, path, tag)
}

// Update modifies an existing tag in the given organization.
func (s *TagService) Update(ctx context.Context, orgID string, tag *Tag) (*Tag, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("tagID", tag.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "tag", tag.ID)
	return s.update(ctx, path, tag)
}

// Delete removes a tag from the given organization.
func (s *TagService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("tagID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "tag", id)
	return s.del(ctx, path)
}
