package terrakube

import "context"

// Agent represents an agent in Terrakube.
type Agent struct {
	ID          string  `jsonapi:"primary,agent"`
	Name        string  `jsonapi:"attr,name"`
	Description string  `jsonapi:"attr,description"`
	URL         string  `jsonapi:"attr,url"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// AgentService handles communication with the Agent related methods of the Terrakube API.
type AgentService struct {
	crudService[Agent]
}

// List returns all agents for an organization.
func (s *AgentService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent")
	return s.list(ctx, path, opts)
}

// Get returns a single agent by ID.
func (s *AgentService) Get(ctx context.Context, orgID, id string) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("agentID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent", id)
	return s.get(ctx, path)
}

// Create creates a new agent in an organization.
func (s *AgentService) Create(ctx context.Context, orgID string, agent *Agent) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent")
	return s.create(ctx, path, agent)
}

// Update modifies an existing agent.
func (s *AgentService) Update(ctx context.Context, orgID string, agent *Agent) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("agentID", agent.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent", agent.ID)
	return s.update(ctx, path, agent)
}

// Delete removes an agent by ID.
func (s *AgentService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("agentID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "agent", id)
	return s.del(ctx, path)
}
