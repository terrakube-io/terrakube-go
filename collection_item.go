package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// CollectionItem represents a key/value item within a Terrakube collection.
type CollectionItem struct {
	ID          string  `jsonapi:"primary,item"`
	Key         string  `jsonapi:"attr,key"`
	Value       string  `jsonapi:"attr,value"`
	Description *string `jsonapi:"attr,description"`
	Category    string  `jsonapi:"attr,category"`
	Sensitive   bool    `jsonapi:"attr,sensitive"`
	Hcl         bool    `jsonapi:"attr,hcl"`
}

// CollectionItemService handles communication with the collection item related
// methods of the Terrakube API.
type CollectionItemService struct {
	client *Client
}

// List returns all items for the given collection.
func (s *CollectionItemService) List(ctx context.Context, orgID, collectionID string, opts *ListOptions) ([]*CollectionItem, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", collectionID, "item")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[item]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var items []*CollectionItem
	_, err = s.client.do(ctx, req, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// Get returns a single collection item by ID.
func (s *CollectionItemService) Get(ctx context.Context, orgID, collectionID, id string) (*CollectionItem, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}
	if err := validateID("itemID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", collectionID, "item", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	item := &CollectionItem{}
	_, err = s.client.do(ctx, req, item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// Create creates a new item in the given collection.
func (s *CollectionItemService) Create(ctx context.Context, orgID, collectionID string, item *CollectionItem) (*CollectionItem, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", collectionID, "item")

	req, err := s.client.request(ctx, http.MethodPost, p, item)
	if err != nil {
		return nil, err
	}

	created := &CollectionItem{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing collection item. The item's ID field must be set.
func (s *CollectionItemService) Update(ctx context.Context, orgID, collectionID string, item *CollectionItem) (*CollectionItem, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}
	if err := validateID("itemID", item.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "collection", collectionID, "item", item.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, item)
	if err != nil {
		return nil, err
	}

	updated := &CollectionItem{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a collection item by ID.
func (s *CollectionItemService) Delete(ctx context.Context, orgID, collectionID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return err
	}
	if err := validateID("itemID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "collection", collectionID, "item", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
