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
