package terrakube

import (
	"context"
	"time"
)

// Collection represents a Terrakube collection resource.
type Collection struct {
	ID           string                 `jsonapi:"primary,collection"`
	Name         string                 `jsonapi:"attr,name"`
	Description  *string                `jsonapi:"attr,description"`
	Priority     int32                  `jsonapi:"attr,priority"`
	CreatedBy    *string                `jsonapi:"attr,createdBy"`
	CreatedDate  *time.Time             `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy    *string                `jsonapi:"attr,updatedBy"`
	UpdatedDate  *time.Time             `jsonapi:"attr,updatedDate,rfc3339"`
	Organization *Organization          `jsonapi:"relation,organization,omitempty"`
	Items        []*CollectionItem      `jsonapi:"relation,item,omitempty"`
	References   []*CollectionReference `jsonapi:"relation,reference,omitempty"`
}

// CollectionService handles communication with the collection related methods
// of the Terrakube API.
type CollectionService struct {
	crudService[Collection]
}

// List returns all collections for the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *CollectionService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection")
	return s.list(ctx, path, opts)
}

// Get returns a single collection by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *CollectionService) Get(ctx context.Context, orgID, id string) (*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection", id)
	return s.get(ctx, path)
}

// Create creates a new collection in the given organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *CollectionService) Create(ctx context.Context, orgID string, collection *Collection) (*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection")
	return s.create(ctx, path, collection)
}

// Update modifies an existing collection. The collection's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *CollectionService) Update(ctx context.Context, orgID string, collection *Collection) (*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collection.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection", collection.ID)
	return s.update(ctx, path, collection)
}

// Delete removes a collection by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *CollectionService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("collectionID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "collection", id)
	return s.del(ctx, path)
}
