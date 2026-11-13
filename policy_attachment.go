package terrakube

import "context"

// PolicyAttachment represents a Terrakube policy attachment resource.
type PolicyAttachment struct {
	ID          string     `jsonapi:"primary,policy_attachment"`
	CreatedBy   *string    `jsonapi:"attr,createdBy"`
	CreatedDate *string    `jsonapi:"attr,createdDate"`
	UpdatedBy   *string    `jsonapi:"attr,updatedBy"`
	UpdatedDate *string    `jsonapi:"attr,updatedDate"`
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

// Update modifies an existing policy attachment. The policy attachment's ID field must be set.
// It returns a *ValidationError if orgID, policySetID, or the ID is empty and a *APIError on server errors.
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
