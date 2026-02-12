package terrakube

import "context"

// SSH represents an SSH key in Terrakube.
type SSH struct {
	ID          string  `jsonapi:"primary,ssh"`
	Name        string  `jsonapi:"attr,name"`
	Description *string `jsonapi:"attr,description"`
	PrivateKey  string  `jsonapi:"attr,privateKey"`
	SSHType     string  `jsonapi:"attr,sshType"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
}

// SSHService handles communication with the SSH related methods of the Terrakube API.
type SSHService struct {
	crudService[SSH]
}

// List returns all SSH keys for an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *SSHService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "ssh")
	return s.list(ctx, path, opts)
}

// Get returns a single SSH key by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *SSHService) Get(ctx context.Context, orgID, id string) (*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("sshID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "ssh", id)
	return s.get(ctx, path)
}

// Create creates a new SSH key in an organization.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *SSHService) Create(ctx context.Context, orgID string, ssh *SSH) (*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "ssh")
	return s.create(ctx, path, ssh)
}

// Update modifies an existing SSH key. The SSH key's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
func (s *SSHService) Update(ctx context.Context, orgID string, ssh *SSH) (*SSH, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("sshID", ssh.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "ssh", ssh.ID)
	return s.update(ctx, path, ssh)
}

// Delete removes an SSH key by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *SSHService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("sshID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "ssh", id)
	return s.del(ctx, path)
}
