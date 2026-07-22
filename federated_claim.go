package terrakube

import "context"

// FederatedClaim represents a Terrakube federated claim resource.
type FederatedClaim struct {
	ID         string     `jsonapi:"primary,federated_claim"`
	ClaimKey   string     `jsonapi:"attr,claimKey"`
	ClaimValue string     `jsonapi:"attr,claimValue"`
	Federated  *Federated `jsonapi:"relation,federated,omitempty"`
}

// FederatedClaimService handles communication with the federated claim related methods of the Terrakube API.
type FederatedClaimService struct {
	crudService[FederatedClaim]
}

// List returns all federated claims for a federated identity configuration.
// It returns a *ValidationError if federatedID is empty and a *APIError on server errors.
func (s *FederatedClaimService) List(ctx context.Context, federatedID string, opts *ListOptions) ([]*FederatedClaim, error) {
	if err := validateID("federated ID", federatedID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("federated", federatedID, "claims")
	return s.list(ctx, path, opts)
}

// Get retrieves a federated claim by ID under a federated identity configuration.
// It returns a *ValidationError if federatedID or id is empty and a *APIError on server errors.
func (s *FederatedClaimService) Get(ctx context.Context, federatedID, id string) (*FederatedClaim, error) {
	if err := validateID("federated ID", federatedID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("federated", federatedID, "claims", id)
	return s.get(ctx, path)
}

// Create creates a new federated claim for a federated identity configuration.
// It returns a *ValidationError if federatedID is empty and a *APIError on server errors.
func (s *FederatedClaimService) Create(ctx context.Context, federatedID string, claim *FederatedClaim) (*FederatedClaim, error) {
	if err := validateID("federated ID", federatedID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("federated", federatedID, "claims")
	return s.create(ctx, path, claim)
}

// Update modifies an existing federated claim. The claim ID must be set.
// It returns a *ValidationError if federatedID or claim ID is empty and a *APIError on server errors.
func (s *FederatedClaimService) Update(ctx context.Context, federatedID string, claim *FederatedClaim) (*FederatedClaim, error) {
	if err := validateID("federated ID", federatedID); err != nil {
		return nil, err
	}
	if err := validateID("federatedClaim ID", claim.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("federated", federatedID, "claims", claim.ID)
	return s.update(ctx, path, claim)
}

// Delete removes a federated claim by ID under a federated identity configuration.
// It returns a *ValidationError if federatedID or id is empty and a *APIError on server errors.
func (s *FederatedClaimService) Delete(ctx context.Context, federatedID, id string) error {
	if err := validateID("federated ID", federatedID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("federated", federatedID, "claims", id)
	return s.del(ctx, path)
}
