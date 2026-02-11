package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Collection represents a Terrakube collection resource.
type Collection struct {
	ID          string  `jsonapi:"primary,collection"`
	Name        string  `jsonapi:"attr,name"`
	Description *string `jsonapi:"attr,description"`
	Priority    int32   `jsonapi:"attr,priority"`
}

// CollectionService handles communication with the collection related methods
// of the Terrakube API.
type CollectionService struct {
	client *Client
}

// List returns all collections for the given organization.
func (s *CollectionService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[collection]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var collections []*Collection
	_, err = s.client.do(ctx, req, &collections)
	if err != nil {
		return nil, err
	}

	return collections, nil
}

// Get returns a single collection by ID.
func (s *CollectionService) Get(ctx context.Context, orgID, id string) (*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	collection := &Collection{}
	_, err = s.client.do(ctx, req, collection)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

// Create creates a new collection in the given organization.
func (s *CollectionService) Create(ctx context.Context, orgID string, collection *Collection) (*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection")

	req, err := s.client.request(ctx, http.MethodPost, p, collection)
	if err != nil {
		return nil, err
	}

	result := &Collection{}
	_, err = s.client.do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update modifies an existing collection. The collection's ID field must be set.
func (s *CollectionService) Update(ctx context.Context, orgID string, collection *Collection) (*Collection, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collection.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", collection.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, collection)
	if err != nil {
		return nil, err
	}

	result := &Collection{}
	_, err = s.client.do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete removes a collection by ID.
func (s *CollectionService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("collectionID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "collection", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
