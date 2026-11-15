package terrakube

import "context"

// Agent represents an agent in Terrakube.
type Agent struct {
	ID           string        `jsonapi:"primary,agent"`
	Name         string        `jsonapi:"attr,name"`
	Description  string        `jsonapi:"attr,description"`
	URL          string        `jsonapi:"attr,url"`
	Organization *Organization `jsonapi:"relation,organization,omitempty"`
}

// AgentService handles communication with the Agent related methods of the Terrakube API.
type AgentService struct {
	crudService[Agent]
}

// List returns all agents for an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *AgentService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent")
	return s.list(ctx, path, opts)
}

// Get returns a single agent by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
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
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *AgentService) Create(ctx context.Context, orgID string, agent *Agent) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent")
	return s.create(ctx, path, agent)
}

// CreateFields creates a new agent in an organization, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *ValidationError if organizationID is empty and a *APIError on server errors.
func (s *AgentService) CreateFields(ctx context.Context, orgID string, p Patch) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing agent. The agent's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
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

// UpdateFields modifies only the attributes and relationships set in p on the agent
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if organizationID or id is empty or p sets nothing, and a *APIError on server errors.
func (s *AgentService) UpdateFields(ctx context.Context, orgID, id string, p Patch) (*Agent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("agentID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "agent", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes an agent by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
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
