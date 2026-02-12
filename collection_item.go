package terrakube

import "context"

// CollectionItem represents a key/value item within a Terrakube collection.
type CollectionItem struct {
	ID          string  `jsonapi:"primary,item"`
	Key         string  `jsonapi:"attr,key"`
	Value       string  `jsonapi:"attr,value"`
	Description *string `jsonapi:"attr,description"`
	Category    string  `jsonapi:"attr,category"`
	Sensitive   bool    `jsonapi:"attr,sensitive"`
	Hcl         bool    `jsonapi:"attr,hcl"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// CollectionItemService handles communication with the collection item related
// methods of the Terrakube API.
type CollectionItemService struct {
	crudService[CollectionItem]
}

// List returns all items for the given collection.
// It returns a *ValidationError if orgID or collectionID is empty and a *APIError on server errors.
func (s *CollectionItemService) List(ctx context.Context, orgID, collectionID string, opts *ListOptions) ([]*CollectionItem, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "item")
	return s.list(ctx, path, opts)
}

// Get returns a single collection item by ID.
// It returns a *ValidationError if orgID, collectionID, or id is empty and a *APIError on server errors.
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

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "item", id)
	return s.get(ctx, path)
}

// Create creates a new item in the given collection.
// It returns a *ValidationError if orgID or collectionID is empty and a *APIError on server errors.
func (s *CollectionItemService) Create(ctx context.Context, orgID, collectionID string, item *CollectionItem) (*CollectionItem, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("collectionID", collectionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "item")
	return s.create(ctx, path, item)
}

// Update modifies an existing collection item. The item's ID field must be set.
// It returns a *ValidationError if orgID, collectionID, or the ID is empty and a *APIError on server errors.
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

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "item", item.ID)
	return s.update(ctx, path, item)
}

// Delete removes a collection item by ID.
// It returns a *ValidationError if orgID, collectionID, or id is empty and a *APIError on server errors.
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

	path := s.client.apiPath("organization", orgID, "collection", collectionID, "item", id)
	return s.del(ctx, path)
}
