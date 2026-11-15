package terrakube

import (
	"context"
	"time"
)

// PolicyExemption represents a Terrakube policy exemption resource.
type PolicyExemption struct {
	ID              string        `jsonapi:"primary,policy_exemption"`
	ExpiresAt       *time.Time    `jsonapi:"attr,expiresAt,rfc3339"`
	Justification   *string       `jsonapi:"attr,justification"`
	RuleID          *string       `jsonapi:"attr,ruleId"`
	TicketReference *string       `jsonapi:"attr,ticketReference"`
	CreatedBy       *string       `jsonapi:"attr,createdBy"`
	CreatedDate     *time.Time    `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy       *string       `jsonapi:"attr,updatedBy"`
	UpdatedDate     *time.Time    `jsonapi:"attr,updatedDate,rfc3339"`
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

// CreateFields creates a new policy exemption, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID is empty and a *APIError on server errors.
func (s *PolicyExemptionService) CreateFields(ctx context.Context, orgID string, p Patch) (*PolicyExemption, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policyExemption")
	return s.createFields(ctx, path, p)
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

// UpdateFields modifies only the attributes and relationships set in p on the policy exemption
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID or id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicyExemptionService) UpdateFields(ctx context.Context, orgID, id string, p Patch) (*PolicyExemption, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policyExemptionID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policyExemption", id)
	return s.updateFields(ctx, path, id, p)
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

// GetByID retrieves a policy exemption by its ID alone, from the top-level /policy_exemption/{id} path,
// for callers that do not know its parent resources.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *PolicyExemptionService) GetByID(ctx context.Context, id string) (*PolicyExemption, error) {
	if err := validateID("policyExemptionID", id); err != nil {
		return nil, err
	}

	return s.get(ctx, s.client.apiPath("policy_exemption", id))
}

// UpdateFieldsByID modifies only the attributes and relationships set in p on the policy exemption
// with the given id, through the top-level /policy_exemption/{id} path; a Relationship with a nil ID
// clears that link.
// It returns a *ValidationError if id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicyExemptionService) UpdateFieldsByID(ctx context.Context, id string, p Patch) (*PolicyExemption, error) {
	if err := validateID("policyExemptionID", id); err != nil {
		return nil, err
	}

	return s.updateFields(ctx, s.client.apiPath("policy_exemption", id), id, p)
}

// DeleteByID removes the policy exemption with the given id through the top-level /policy_exemption/{id} path.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *PolicyExemptionService) DeleteByID(ctx context.Context, id string) error {
	if err := validateID("policyExemptionID", id); err != nil {
		return err
	}

	return s.del(ctx, s.client.apiPath("policy_exemption", id))
}
