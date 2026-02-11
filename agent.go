package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Agent represents an agent in Terrakube.
type Agent struct {
	ID          string `jsonapi:"primary,agent"`
	Name        string `jsonapi:"attr,name"`
	Description string `jsonapi:"attr,description"`
	URL         string `jsonapi:"attr,url"`
}

// AgentService handles communication with the Agent related methods of the Terrakube API.
type AgentService struct {
	client *Client
}

func (s *AgentService) basePath(orgID string) string {
	return s.client.apiPath("organization", orgID, "agent")
}

func (s *AgentService) resourcePath(orgID, id string) string {
	return s.client.apiPath("organization", orgID, "agent", id)
}

// List returns all agents for an organization.
func (s *AgentService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[agent]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, s.basePath(orgID), params, nil)
	if err != nil {
		return nil, err
	}

	var items []*Agent
	_, err = s.client.do(ctx, req, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// Get returns a single agent by ID.
func (s *AgentService) Get(ctx context.Context, orgID, id string) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("agentID", id); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodGet, s.resourcePath(orgID, id), nil)
	if err != nil {
		return nil, err
	}

	v := &Agent{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Create creates a new agent in an organization.
func (s *AgentService) Create(ctx context.Context, orgID string, agent *Agent) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPost, s.basePath(orgID), agent)
	if err != nil {
		return nil, err
	}

	v := &Agent{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Update modifies an existing agent.
func (s *AgentService) Update(ctx context.Context, orgID string, agent *Agent) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("agentID", agent.ID); err != nil {
		return nil, err
	}

	req, err := s.client.request(ctx, http.MethodPatch, s.resourcePath(orgID, agent.ID), agent)
	if err != nil {
		return nil, err
	}

	v := &Agent{}
	_, err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Delete removes an agent by ID.
func (s *AgentService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("agentID", id); err != nil {
		return err
	}

	req, err := s.client.request(ctx, http.MethodDelete, s.resourcePath(orgID, id), nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
