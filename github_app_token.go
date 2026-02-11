package terrakube

import "context"

// GithubAppToken represents a Terrakube GitHub App token resource.
type GithubAppToken struct {
	ID             string  `jsonapi:"primary,github_app_token"`
	AppID          string  `jsonapi:"attr,appId"`
	InstallationID string  `jsonapi:"attr,installationId"`
	Owner          string  `jsonapi:"attr,owner"`
	Token          *string `jsonapi:"attr,token"`
	CreatedBy      *string `jsonapi:"attr,createdBy"`
	CreatedDate    *string `jsonapi:"attr,createdDate"`
	UpdatedBy      *string `jsonapi:"attr,updatedBy"`
	UpdatedDate    *string `jsonapi:"attr,updatedDate"`
}

// GithubAppTokenService handles communication with the GitHub App token endpoints.
type GithubAppTokenService struct {
	crudService[GithubAppToken]
}

// List returns all GitHub App tokens.
func (s *GithubAppTokenService) List(ctx context.Context, opts *ListOptions) ([]*GithubAppToken, error) {
	path := s.client.apiPath("github_app_token")
	return s.list(ctx, path, opts)
}

// Get returns a single GitHub App token by ID.
func (s *GithubAppTokenService) Get(ctx context.Context, id string) (*GithubAppToken, error) {
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("github_app_token", id)
	return s.get(ctx, path)
}

// Create creates a new GitHub App token.
func (s *GithubAppTokenService) Create(ctx context.Context, token *GithubAppToken) (*GithubAppToken, error) {
	path := s.client.apiPath("github_app_token")
	return s.create(ctx, path, token)
}

// Update modifies an existing GitHub App token. The token's ID field must be set.
func (s *GithubAppTokenService) Update(ctx context.Context, token *GithubAppToken) (*GithubAppToken, error) {
	if err := validateID("github app token ID", token.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("github_app_token", token.ID)
	return s.update(ctx, path, token)
}

// Delete removes a GitHub App token by ID.
func (s *GithubAppTokenService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("github_app_token", id)
	return s.del(ctx, path)
}
