package terrakube

import (
	"context"
	"time"
)

// PolicyAttachment represents a Terrakube policy attachment resource.
type PolicyAttachment struct {
	ID          string     `jsonapi:"primary,policy_attachment"`
	CreatedBy   *string    `jsonapi:"attr,createdBy"`
	CreatedDate *time.Time `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy   *string    `jsonapi:"attr,updatedBy"`
	UpdatedDate *time.Time `jsonapi:"attr,updatedDate,rfc3339"`
	PolicySet   *PolicySet `jsonapi:"relation,policySet,omitempty"`
	Project     *Project   `jsonapi:"relation,project,omitempty"`
	Tag         *Tag       `jsonapi:"relation,tag,omitempty"`
	Workspace   *Workspace `jsonapi:"relation,workspace,omitempty"`
}

// PolicyAttachmentService handles communication with the policy attachment related
// methods of the Terrakube API.
type PolicyAttachmentService struct {
	crudService[PolicyAttachment]
}

// List returns all policy attachments for a policy set.
// It returns a *ValidationError if orgID or policySetID is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) List(ctx context.Context, orgID, policySetID string, opts *ListOptions) ([]*PolicyAttachment, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "attachments")
	return s.list(ctx, path, opts)
}

// Get retrieves a single policy attachment by ID.
// It returns a *ValidationError if orgID, policySetID, or policyAttachmentID is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) Get(ctx context.Context, orgID, policySetID, policyAttachmentID string) (*PolicyAttachment, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policyAttachmentID", policyAttachmentID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "attachments", policyAttachmentID)
	return s.get(ctx, path)
}

// Create creates a new policy attachment.
// It returns a *ValidationError if orgID or policySetID is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) Create(ctx context.Context, orgID, policySetID string, entity *PolicyAttachment) (*PolicyAttachment, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "attachments")
	return s.create(ctx, path, entity)
}

// CreateFields creates a new policy attachment, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID or policySetID is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) CreateFields(ctx context.Context, orgID, policySetID string, p Patch) (*PolicyAttachment, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "attachments")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing policy attachment. The policy attachment's ID field must be set.
// It returns a *ValidationError if orgID, policySetID, or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *PolicyAttachmentService) Update(ctx context.Context, orgID, policySetID string, entity *PolicyAttachment) (*PolicyAttachment, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policyAttachmentID", entity.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "attachments", entity.ID)
	return s.update(ctx, path, entity)
}

// UpdateFields modifies only the attributes and relationships set in p on the policy attachment
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID, policySetID, or id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicyAttachmentService) UpdateFields(ctx context.Context, orgID, policySetID, id string, p Patch) (*PolicyAttachment, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policyAttachmentID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "attachments", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes a policy attachment.
// It returns a *ValidationError if orgID, policySetID, or policyAttachmentID is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) Delete(ctx context.Context, orgID, policySetID, policyAttachmentID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("policySetID", policySetID); err != nil {
		return err
	}
	if err := validateID("policyAttachmentID", policyAttachmentID); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "policySet", policySetID, "attachments", policyAttachmentID)
	return s.del(ctx, path)
}

// ListByPolicySet returns the policy attachments of a policy set, from the top-level
// /policy_set/{id}/attachments path, for callers that do not know its organization.
// It returns a *ValidationError if policySetID is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) ListByPolicySet(ctx context.Context, policySetID string, opts *ListOptions) ([]*PolicyAttachment, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	return s.list(ctx, s.client.apiPath("policy_set", policySetID, "attachments"), opts)
}

// GetByPolicySet retrieves a policy attachment of a policy set from the top-level
// /policy_set/{id}/attachments/{id} path.
// It returns a *ValidationError if policySetID or id is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) GetByPolicySet(ctx context.Context, policySetID, id string) (*PolicyAttachment, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policyAttachmentID", id); err != nil {
		return nil, err
	}

	return s.get(ctx, s.client.apiPath("policy_set", policySetID, "attachments", id))
}

// CreateFieldsByPolicySet creates a policy attachment in a policy set through the top-level
// /policy_set/{id}/attachments path, sending only the attributes and relationships set in p.
// It returns a *ValidationError if policySetID is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) CreateFieldsByPolicySet(ctx context.Context, policySetID string, p Patch) (*PolicyAttachment, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}

	return s.createFields(ctx, s.client.apiPath("policy_set", policySetID, "attachments"), p)
}

// UpdateFieldsByPolicySet modifies only the attributes and relationships set in p on a
// policy attachment of a policy set, through the top-level /policy_set/{id}/attachments/{id} path; a Relationship with a nil ID
// clears that link.
// It returns a *ValidationError if policySetID or id is empty or p sets nothing, and a *APIError on server errors.
func (s *PolicyAttachmentService) UpdateFieldsByPolicySet(ctx context.Context, policySetID, id string, p Patch) (*PolicyAttachment, error) {
	if err := validateID("policySetID", policySetID); err != nil {
		return nil, err
	}
	if err := validateID("policyAttachmentID", id); err != nil {
		return nil, err
	}

	return s.updateFields(ctx, s.client.apiPath("policy_set", policySetID, "attachments", id), id, p)
}

// DeleteByPolicySet removes a policy attachment of a policy set through the top-level
// /policy_set/{id}/attachments/{id} path.
// It returns a *ValidationError if policySetID or id is empty and a *APIError on server errors.
func (s *PolicyAttachmentService) DeleteByPolicySet(ctx context.Context, policySetID, id string) error {
	if err := validateID("policySetID", policySetID); err != nil {
		return err
	}
	if err := validateID("policyAttachmentID", id); err != nil {
		return err
	}

	return s.del(ctx, s.client.apiPath("policy_set", policySetID, "attachments", id))
}
