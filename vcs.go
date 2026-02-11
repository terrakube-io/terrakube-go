package terrakube

import "context"

// VCS represents a version control system connection in Terrakube.
type VCS struct {
	ID             string  `jsonapi:"primary,vcs"`
	Name           string  `jsonapi:"attr,name"`
	Description    string  `jsonapi:"attr,description"`
	VcsType        string  `jsonapi:"attr,vcsType"`
	ConnectionType string  `jsonapi:"attr,connectionType"`
	ClientID       string  `jsonapi:"attr,clientId"`
	ClientSecret   string  `jsonapi:"attr,clientSecret"`
	PrivateKey     string  `jsonapi:"attr,privateKey"`
	Endpoint       string  `jsonapi:"attr,endpoint"`
	APIURL         string  `jsonapi:"attr,apiUrl"`
	Status         string  `jsonapi:"attr,status"`
	Callback       *string `jsonapi:"attr,callback"`
	AccessToken    *string `jsonapi:"attr,accessToken"`
	RedirectURL    *string `jsonapi:"attr,redirectUrl"`
	CreatedBy      *string `jsonapi:"attr,createdBy"`
	CreatedDate    *string `jsonapi:"attr,createdDate"`
	UpdatedBy      *string `jsonapi:"attr,updatedBy"`
	UpdatedDate    *string `jsonapi:"attr,updatedDate"`
}

// VCSService handles communication with the VCS related methods of the Terrakube API.
type VCSService struct {
	crudService[VCS]
}

// List returns all VCS connections for an organization.
func (s *VCSService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "vcs")
	return s.list(ctx, path, opts)
}

// Get returns a single VCS connection by ID.
func (s *VCSService) Get(ctx context.Context, orgID, id string) (*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("vcsID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "vcs", id)
	return s.get(ctx, path)
}

// Create creates a new VCS connection in an organization.
func (s *VCSService) Create(ctx context.Context, orgID string, vcs *VCS) (*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "vcs")
	return s.create(ctx, path, vcs)
}

// Update modifies an existing VCS connection.
func (s *VCSService) Update(ctx context.Context, orgID string, vcs *VCS) (*VCS, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("vcsID", vcs.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "vcs", vcs.ID)
	return s.update(ctx, path, vcs)
}

// Delete removes a VCS connection by ID.
func (s *VCSService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("vcsID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "vcs", id)
	return s.del(ctx, path)
}
