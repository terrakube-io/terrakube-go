package terrakube

import "context"

// CollectionReference represents a reference within a Terrakube collection.
type CollectionReference struct {
	ID          string      `jsonapi:"primary,reference"`
	Description *string     `jsonapi:"attr,description"`
	Workspace   *Workspace  `jsonapi:"relation,workspace,omitempty"`
	Collection  *Collection `jsonapi:"relation,collection,omitempty"`
}

// CollectionReferenceService handles communication with the collection reference
// related methods of the Terrakube API.
type CollectionReferenceService struct {
	crudService[CollectionReference]
}

// List returns all references for the given collection.
// It returns a *ValidationError if orgID or collectionID is empty and a *APIError on server errors.
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
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *CollectionReferenceService) Get(ctx context.Context, id string) (*CollectionReference, error) {
	if err := validateID("referenceID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("reference", id)
	return s.get(ctx, path)
}

// Create creates a new reference in the given collection.
// It returns a *ValidationError if orgID or collectionID is empty and a *APIError on server errors.
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

// CreateFields creates a new reference in the given collection, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID or collectionID is empty and a *APIError on server errors.
func (s *CollectionReferenceService) CreateFields(ctx context.Context, orgID, collectionID string, p Patch) (*CollectionReference, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "reference")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing collection reference using the flat endpoint.
// The reference's ID field must be set.
// It returns a *ValidationError if the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *CollectionReferenceService) Update(ctx context.Context, ref *CollectionReference) (*CollectionReference, error) {
	if err := validateID("referenceID", ref.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("reference", ref.ID)
	return s.update(ctx, path, ref)
}

// UpdateFields modifies only the attributes and relationships set in p on the collection reference
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if id is empty or p sets nothing, and a *APIError on server errors.
func (s *CollectionReferenceService) UpdateFields(ctx context.Context, id string, p Patch) (*CollectionReference, error) {
	if err := validateID("referenceID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("reference", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes a collection reference by ID using the flat endpoint.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *CollectionReferenceService) Delete(ctx context.Context, id string) error {
	if err := validateID("referenceID", id); err != nil {
		return err
	}

	path := s.client.apiPath("reference", id)
	return s.del(ctx, path)
}
