package terrakube

import (
	"context"
	"time"
)

// PolicySet represents a Terrakube policy set resource.
type PolicySet struct {
	ID                        string                     `jsonapi:"primary,policy_set"`
	Branch                    *string                    `jsonapi:"attr,branch"`
	Description               *string                    `jsonapi:"attr,description"`
	EnforcementLevel          *string                    `jsonapi:"attr,enforcementLevel"`
	Folder                    *string                    `jsonapi:"attr,folder"`
	Global                    bool                       `jsonapi:"attr,global"`
	Name                      *string                    `jsonapi:"attr,name"`
	OpaVersion                *string                    `jsonapi:"attr,opaVersion"`
	OverrideTeam              *string                    `jsonapi:"attr,overrideTeam"`
	Repository                *string                    `jsonapi:"attr,repository"`
	ShadowEnforcementLevel    *string                    `jsonapi:"attr,shadowEnforcementLevel"`
	CreatedBy                 *string                    `jsonapi:"attr,createdBy"`
	CreatedDate               *time.Time                 `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy                 *string                    `jsonapi:"attr,updatedBy"`
	UpdatedDate               *time.Time                 `jsonapi:"attr,updatedDate,rfc3339"`
	Attachments               []*PolicyAttachment        `jsonapi:"relation,attachments,omitempty"`
	Exemptions                []*PolicyExemption         `jsonapi:"relation,exemptions,omitempty"`
	NotificationConfiguration *NotificationConfiguration `jsonapi:"relation,notificationConfiguration,omitempty"`
	Organization              *Organization              `jsonapi:"relation,organization,omitempty"`
	Parameters                []*PolicySetParameter      `jsonapi:"relation,parameters,omitempty"`
	VCS                       *VCS                       `jsonapi:"relation,vcs,omitempty"`
}

// PolicySetService handles communication with the policy set related
// methods of the Terrakube API.
type PolicySetService struct {
	crudService[PolicySet]
}

// List returns all policy sets for an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *PolicySetService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*PolicySet, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet")
	return s.list(ctx, path, opts)
}

// Get retrieves a single policy set by ID.
// It returns a *ValidationError if orgID or policySetID is empty and a *APIError on server errors.
func (s *PolicySetService) Get(ctx context.Context, orgID, policySetID string) (*PolicySet, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID)
	return s.get(ctx, path)
}

// Create creates a new policy set.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *PolicySetService) Create(ctx context.Context, orgID string, entity *PolicySet) (*PolicySet, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet")
	return s.create(ctx, path, entity)
}

// CreateFields creates a new policy set, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID is empty and a *APIError on server errors.
func (s *PolicySetService) CreateFields(ctx context.Context, orgID string, p Patch) (*PolicySet, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing policy set. The policy set's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *PolicySetService) Update(ctx context.Context, orgID string, entity *PolicySet) (*PolicySet, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", entity.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", entity.ID)
	return s.update(ctx, path, entity)
}

// UpdateFields modifies only the attributes and relationships set in p on the policy set
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID or id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicySetService) UpdateFields(ctx context.Context, orgID, id string, p Patch) (*PolicySet, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes a policy set.
// It returns a *ValidationError if orgID or policySetID is empty and a *APIError on server errors.
func (s *PolicySetService) Delete(ctx context.Context, orgID, policySetID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID)
	return s.del(ctx, path)
}

// GetByID retrieves a policy set by its ID alone, from the top-level /policy_set/{id} path,
// for callers that do not know its parent resources.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *PolicySetService) GetByID(ctx context.Context, id string) (*PolicySet, error) {
	if err := validateID("policySetID", id); err != nil {
		return nil, err
	}

	return s.get(ctx, s.client.apiPath("policy_set", id))
}

// UpdateFieldsByID modifies only the attributes and relationships set in p on the policy set
// with the given id, through the top-level /policy_set/{id} path; a Relationship with a nil ID
// clears that link.
// It returns a *ValidationError if id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicySetService) UpdateFieldsByID(ctx context.Context, id string, p Patch) (*PolicySet, error) {
	if err := validateID("policySetID", id); err != nil {
		return nil, err
	}

	return s.updateFields(ctx, s.client.apiPath("policy_set", id), id, p)
}

// DeleteByID removes the policy set with the given id through the top-level /policy_set/{id} path.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *PolicySetService) DeleteByID(ctx context.Context, id string) error {
	if err := validateID("policySetID", id); err != nil {
		return err
	}

	return s.del(ctx, s.client.apiPath("policy_set", id))
}
