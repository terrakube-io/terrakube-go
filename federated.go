package terrakube

import "context"

// Federated represents a Terrakube federated identity provider resource.
type Federated struct {
	ID        string           `jsonapi:"primary,federated"`
	Name      string           `jsonapi:"attr,name"`
	IssuerURL string           `jsonapi:"attr,issuerUrl"`
	Audience  string           `jsonapi:"attr,audience"`
	Claims    []*FederatedClaim `jsonapi:"relation,claims,omitempty"`
}

// FederatedService handles communication with the federated identity related methods of the Terrakube API.
type FederatedService struct {
	crudService[Federated]
}

// List returns all federated identity configurations.
// It returns a *APIError on server errors.
func (s *FederatedService) List(ctx context.Context, opts *ListOptions) ([]*Federated, error) {
	path := s.client.apiPath("federated")
	return s.list(ctx, path, opts)
}

// Get retrieves a federated identity configuration by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *FederatedService) Get(ctx context.Context, id string) (*Federated, error) {
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("federated", id)
	return s.get(ctx, path)
}

// Create creates a new federated identity configuration.
// It returns a *APIError on server errors.
func (s *FederatedService) Create(ctx context.Context, fed *Federated) (*Federated, error) {
	path := s.client.apiPath("federated")
	return s.create(ctx, path, fed)
}

// Update modifies an existing federated identity configuration. The ID field must be set.
// It returns a *ValidationError if ID is empty and a *APIError on server errors.
func (s *FederatedService) Update(ctx context.Context, fed *Federated) (*Federated, error) {
	if err := validateID("federatedID", fed.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("federated", fed.ID)
	return s.update(ctx, path, fed)
}

// Delete removes a federated identity configuration by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *FederatedService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("federated", id)
	return s.del(ctx, path)
}
