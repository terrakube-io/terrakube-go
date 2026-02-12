package terrakube

import "context"

// Team represents a Terrakube team resource.
type Team struct {
	ID               string  `jsonapi:"primary,team"`
	Name             string  `jsonapi:"attr,name"`
	ManageState      bool    `jsonapi:"attr,manageState"`
	ManageWorkspace  bool    `jsonapi:"attr,manageWorkspace"`
	ManageModule     bool    `jsonapi:"attr,manageModule"`
	ManageProvider   bool    `jsonapi:"attr,manageProvider"`
	ManageVcs        bool    `jsonapi:"attr,manageVcs"`
	ManageTemplate   bool    `jsonapi:"attr,manageTemplate"`
	ManageJob        bool    `jsonapi:"attr,manageJob"`
	ManageCollection bool    `jsonapi:"attr,manageCollection"`
	CreatedBy        *string `jsonapi:"attr,createdBy"`
	CreatedDate      *string `jsonapi:"attr,createdDate"`
	UpdatedBy        *string `jsonapi:"attr,updatedBy"`
	UpdatedDate      *string `jsonapi:"attr,updatedDate"`
}

// TeamService handles communication with the team-related endpoints.
type TeamService struct {
	crudService[Team]
}

// List returns all teams for an organization, with optional filtering.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *TeamService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "team")
	return s.list(ctx, path, opts)
}

// Get retrieves a single team by ID within an organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *TeamService) Get(ctx context.Context, orgID, id string) (*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "team", id)
	return s.get(ctx, path)
}

// Create creates a new team within an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *TeamService) Create(ctx context.Context, orgID string, team *Team) (*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "team")
	return s.create(ctx, path, team)
}

// Update modifies an existing team within an organization. The team's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
func (s *TeamService) Update(ctx context.Context, orgID string, team *Team) (*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", team.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "team", team.ID)
	return s.update(ctx, path, team)
}

// Delete removes a team from an organization.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *TeamService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("orgID", orgID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "team", id)
	return s.del(ctx, path)
}
