package terrakube

import "context"

// CollectionReference represents a reference within a Terrakube collection.
type CollectionReference struct {
	ID          string      `jsonapi:"primary,reference"`
	Description *string     `jsonapi:"attr,description"`
	CreatedBy   *string     `jsonapi:"attr,createdBy"`
	CreatedDate *string     `jsonapi:"attr,createdDate"`
	UpdatedBy   *string     `jsonapi:"attr,updatedBy"`
	UpdatedDate *string     `jsonapi:"attr,updatedDate"`
	Workspace   *Workspace  `jsonapi:"relation,workspace,omitempty"`
	Collection  *Collection `jsonapi:"relation,collection,omitempty"`
}

// CollectionReferenceService handles communication with the collection reference
// related methods of the Terrakube API.
type CollectionReferenceService struct {
	crudService[CollectionReference]
}

// List returns all references for the given collection.
func (s *CollectionReferenceService) List(ctx context.Context, orgID, collectionID string, opts *ListOptions) ([]*CollectionReference, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "reference")
	return s.list(ctx, path, opts)
}

// Get returns a single collection reference by ID using the flat endpoint.
func (s *CollectionReferenceService) Get(ctx context.Context, id string) (*CollectionReference, error) {
	if err := validateID("referenceID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("reference", id)
	return s.get(ctx, path)
}

// Create creates a new reference in the given collection.
func (s *CollectionReferenceService) Create(ctx context.Context, orgID, collectionID string, ref *CollectionReference) (*CollectionReference, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "reference")
	return s.create(ctx, path, ref)
}

// Update modifies an existing collection reference using the flat endpoint.
// The reference's ID field must be set.
func (s *CollectionReferenceService) Update(ctx context.Context, ref *CollectionReference) (*CollectionReference, error) {
	if err := validateID("referenceID", ref.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("reference", ref.ID)
	return s.update(ctx, path, ref)
}

// Delete removes a collection reference by ID using the flat endpoint.
func (s *CollectionReferenceService) Delete(ctx context.Context, id string) error {
	if err := validateID("referenceID", id); err != nil {
		return err
	}

	path := s.client.apiPath("reference", id)
	return s.del(ctx, path)
}
