package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

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
	client *Client
}

// List returns all references for the given collection.
func (s *CollectionReferenceService) List(ctx context.Context, orgID, collectionID string, opts *ListOptions) ([]*CollectionReference, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", collectionID, "reference")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[reference]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var refs []*CollectionReference
	_, err = s.client.do(ctx, req, &refs)
	if err != nil {
		return nil, err
	}

	return refs, nil
}

// Get returns a single collection reference by ID using the flat endpoint.
func (s *CollectionReferenceService) Get(ctx context.Context, id string) (*CollectionReference, error) {
	if err := validateID("referenceID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("reference", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	ref := &CollectionReference{}
	_, err = s.client.do(ctx, req, ref)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

// Create creates a new reference in the given collection.
func (s *CollectionReferenceService) Create(ctx context.Context, orgID, collectionID string, ref *CollectionReference) (*CollectionReference, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", collectionID, "reference")

	req, err := s.client.request(ctx, http.MethodPost, p, ref)
	if err != nil {
		return nil, err
	}

	created := &CollectionReference{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing collection reference using the flat endpoint.
// The reference's ID field must be set.
func (s *CollectionReferenceService) Update(ctx context.Context, ref *CollectionReference) (*CollectionReference, error) {
	if err := validateID("referenceID", ref.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("reference", ref.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, ref)
	if err != nil {
		return nil, err
	}

	updated := &CollectionReference{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a collection reference by ID using the flat endpoint.
func (s *CollectionReferenceService) Delete(ctx context.Context, id string) error {
	if err := validateID("referenceID", id); err != nil {
		return err
	}

	p := s.client.apiPath("reference", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
