package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Team represents a Terrakube team resource.
type Team struct {
	ID               string `jsonapi:"primary,team"`
	Name             string `jsonapi:"attr,name"`
	ManageState      bool   `jsonapi:"attr,manageState"`
	ManageWorkspace  bool   `jsonapi:"attr,manageWorkspace"`
	ManageModule     bool   `jsonapi:"attr,manageModule"`
	ManageProvider   bool   `jsonapi:"attr,manageProvider"`
	ManageVcs        bool   `jsonapi:"attr,manageVcs"`
	ManageTemplate   bool   `jsonapi:"attr,manageTemplate"`
	ManageJob        bool   `jsonapi:"attr,manageJob"`
	ManageCollection bool   `jsonapi:"attr,manageCollection"`
}

// TeamService handles communication with the team-related endpoints.
type TeamService struct {
	client *Client
}

// List returns all teams for an organization, with optional filtering.
func (s *TeamService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "team")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[team]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var teams []*Team
	_, err = s.client.do(ctx, req, &teams)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

// Get retrieves a single team by ID within an organization.
func (s *TeamService) Get(ctx context.Context, orgID, id string) (*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "team", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	team := &Team{}
	_, err = s.client.do(ctx, req, team)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// Create creates a new team within an organization.
func (s *TeamService) Create(ctx context.Context, orgID string, team *Team) (*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "team")

	req, err := s.client.request(ctx, http.MethodPost, p, team)
	if err != nil {
		return nil, err
	}

	result := &Team{}
	_, err = s.client.do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update modifies an existing team within an organization.
func (s *TeamService) Update(ctx context.Context, orgID string, team *Team) (*Team, error) {
	if err := validateID("orgID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("id", team.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "team", team.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, team)
	if err != nil {
		return nil, err
	}

	result := &Team{}
	_, err = s.client.do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete removes a team from an organization.
func (s *TeamService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("orgID", orgID); err != nil {
		return err
	}
	if err := validateID("id", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "team", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
