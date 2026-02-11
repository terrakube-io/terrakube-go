package terrakube

import "context"

// History represents a Terrakube workspace history resource.
type History struct {
	ID           string  `jsonapi:"primary,history"`
	JobReference string  `jsonapi:"attr,jobReference,omitempty"`
	Output       string  `jsonapi:"attr,output,omitempty"`
	Serial       int     `jsonapi:"attr,serial"`
	Md5          *string `jsonapi:"attr,md5"`
	Lineage      *string `jsonapi:"attr,lineage"`
	CreatedBy    *string `jsonapi:"attr,createdBy"`
	CreatedDate  *string `jsonapi:"attr,createdDate"`
	UpdatedBy    *string `jsonapi:"attr,updatedBy"`
	UpdatedDate  *string `jsonapi:"attr,updatedDate"`
}

// HistoryService handles communication with the history-related endpoints.
type HistoryService struct {
	crudService[History]
}

// List returns all history entries for the given workspace.
func (s *HistoryService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history")
	return s.list(ctx, path, opts)
}

// Get returns a single history entry by ID within the given workspace.
func (s *HistoryService) Get(ctx context.Context, orgID, workspaceID, id string) (*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("historyID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history", id)
	return s.get(ctx, path)
}

// Create creates a new history entry in the given workspace.
func (s *HistoryService) Create(ctx context.Context, orgID, workspaceID string, h *History) (*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history")
	return s.create(ctx, path, h)
}

// Update modifies an existing history entry in the given workspace.
func (s *HistoryService) Update(ctx context.Context, orgID, workspaceID string, h *History) (*History, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("historyID", h.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history", h.ID)
	return s.update(ctx, path, h)
}

// Delete removes a history entry from the given workspace.
func (s *HistoryService) Delete(ctx context.Context, orgID, workspaceID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("historyID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "history", id)
	return s.del(ctx, path)
}
