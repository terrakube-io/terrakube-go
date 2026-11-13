package terrakube

import "context"

// PolicyExemption represents a Terrakube policy exemption resource.
type PolicyExemption struct {
	ID              string        `jsonapi:"primary,policy_exemption"`
	ExpiresAt       *string       `jsonapi:"attr,expiresAt"`
	Justification   *string       `jsonapi:"attr,justification"`
	RuleID          *string       `jsonapi:"attr,ruleId"`
	TicketReference *string       `jsonapi:"attr,ticketReference"`
	CreatedBy       *string       `jsonapi:"attr,createdBy"`
	CreatedDate     *string       `jsonapi:"attr,createdDate"`
	UpdatedBy       *string       `jsonapi:"attr,updatedBy"`
	UpdatedDate     *string       `jsonapi:"attr,updatedDate"`
	Organization    *Organization `jsonapi:"relation,organization,omitempty"`
	PolicySet       *PolicySet    `jsonapi:"relation,policySet,omitempty"`
	Project         *Project      `jsonapi:"relation,project,omitempty"`
	Workspace       *Workspace    `jsonapi:"relation,workspace,omitempty"`
}

// PolicyExemptionService handles communication with the policy exemption related
// methods of the Terrakube API.
type PolicyExemptionService struct {
	crudService[PolicyExemption]
}

// List returns all policy exemptions for an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *PolicyExemptionService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*PolicyExemption, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policyExemption")
	return s.list(ctx, path, opts)
}

// Get retrieves a single policy exemption by ID.
// It returns a *ValidationError if orgID or policyExemptionID is empty and a *APIError on server errors.
func (s *PolicyExemptionService) Get(ctx context.Context, orgID, policyExemptionID string) (*PolicyExemption, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policyExemptionID", policyExemptionID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policyExemption", policyExemptionID)
	return s.get(ctx, path)
}

// Create creates a new policy exemption.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *PolicyExemptionService) Create(ctx context.Context, orgID string, entity *PolicyExemption) (*PolicyExemption, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policyExemption")
	return s.create(ctx, path, entity)
}

// Update modifies an existing policy exemption. The policy exemption's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *PolicyExemptionService) Update(ctx context.Context, orgID string, entity *PolicyExemption) (*PolicyExemption, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policyExemptionID", entity.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policyExemption", entity.ID)
	return s.update(ctx, path, entity)
}

// Delete removes a policy exemption.
// It returns a *ValidationError if orgID or policyExemptionID is empty and a *APIError on server errors.
func (s *PolicyExemptionService) Delete(ctx context.Context, orgID, policyExemptionID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("policyExemptionID", policyExemptionID); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "policyExemption", policyExemptionID)
	return s.del(ctx, path)
}
