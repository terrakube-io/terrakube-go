package terrakube

import "context"

// Action represents a Terrakube action resource.
type Action struct {
	ID              string  `jsonapi:"primary,action"`
	Action          string  `jsonapi:"attr,action"`
	Active          bool    `jsonapi:"attr,active"`
	Category        string  `jsonapi:"attr,category"`
	Description     *string `jsonapi:"attr,description"`
	DisplayCriteria *string `jsonapi:"attr,displayCriteria"`
	Label           string  `jsonapi:"attr,label"`
	Name            string  `jsonapi:"attr,name"`
	Type            string  `jsonapi:"attr,type"`
	Version         *string `jsonapi:"attr,version"`
	CreatedBy       *string `jsonapi:"attr,createdBy"`
	CreatedDate     *string `jsonapi:"attr,createdDate"`
	UpdatedBy       *string `jsonapi:"attr,updatedBy"`
	UpdatedDate     *string `jsonapi:"attr,updatedDate"`
}

// ActionService handles communication with the action related methods of the
// Terrakube API.
type ActionService struct {
	crudService[Action]
}

// List returns all actions, optionally filtered.
// It returns a *APIError on server errors.
func (s *ActionService) List(ctx context.Context, opts *ListOptions) ([]*Action, error) {
	path := s.client.apiPath("action")
	return s.list(ctx, path, opts)
}

// Get retrieves an action by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *ActionService) Get(ctx context.Context, id string) (*Action, error) {
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("action", id)
	return s.get(ctx, path)
}

// Create creates a new action.
// It returns a *APIError on server errors.
func (s *ActionService) Create(ctx context.Context, action *Action) (*Action, error) {
	path := s.client.apiPath("action")
	return s.create(ctx, path, action)
}

// Update modifies an existing action. The action's ID field must be set.
// It returns a *ValidationError if the ID is empty and a *APIError on server errors.
func (s *ActionService) Update(ctx context.Context, action *Action) (*Action, error) {
	if err := validateID("action ID", action.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("action", action.ID)
	return s.update(ctx, path, action)
}

// Delete removes an action by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *ActionService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("action", id)
	return s.del(ctx, path)
}
