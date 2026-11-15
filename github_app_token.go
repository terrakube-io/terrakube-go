package terrakube

import (
	"context"
	"time"
)

// GithubAppToken represents a Terrakube GitHub App token resource.
type GithubAppToken struct {
	ID             string     `jsonapi:"primary,github_app_token"`
	AppID          string     `jsonapi:"attr,appId"`
	InstallationID string     `jsonapi:"attr,installationId"`
	Owner          string     `jsonapi:"attr,owner"`
	Token          *string    `jsonapi:"attr,token"`
	ExpiresAt      *time.Time `jsonapi:"attr,expiresAt,omitempty,rfc3339"`
	CreatedBy      *string    `jsonapi:"attr,createdBy"`
	CreatedDate    *time.Time `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy      *string    `jsonapi:"attr,updatedBy"`
	UpdatedDate    *time.Time `jsonapi:"attr,updatedDate,rfc3339"`
}

// GithubAppTokenService handles communication with the GitHub App token endpoints.
type GithubAppTokenService struct {
	crudService[GithubAppToken]
}

// List returns all GitHub App tokens.
// It returns a *APIError on server errors.
func (s *GithubAppTokenService) List(ctx context.Context, opts *ListOptions) ([]*GithubAppToken, error) {
	path := s.client.apiPath("github_app_token")
	return s.list(ctx, path, opts)
}

// Get returns a single GitHub App token by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *GithubAppTokenService) Get(ctx context.Context, id string) (*GithubAppToken, error) {
	if err := validateID("id", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("github_app_token", id)
	return s.get(ctx, path)
}

// Create creates a new GitHub App token.
// It returns a *APIError on server errors.
func (s *GithubAppTokenService) Create(ctx context.Context, token *GithubAppToken) (*GithubAppToken, error) {
	path := s.client.apiPath("github_app_token")
	return s.create(ctx, path, token)
}

// CreateFields creates a new GitHub App token, sending only the attributes and
// relationships set in p; the server applies its defaults to everything else.
// It returns a *APIError on server errors.
func (s *GithubAppTokenService) CreateFields(ctx context.Context, p Patch) (*GithubAppToken, error) {
	path := s.client.apiPath("github_app_token")
	return s.createFields(ctx, path, p)
}

// Update modifies an existing GitHub App token. The token's ID field must be set.
// It returns a *ValidationError if the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *GithubAppTokenService) Update(ctx context.Context, token *GithubAppToken) (*GithubAppToken, error) {
	if err := validateID("githubAppTokenID", token.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("github_app_token", token.ID)
	return s.update(ctx, path, token)
}

// UpdateFields modifies only the attributes and relationships set in p on the GitHub App token
// with the given id; a Relationship with a nil ID clears that link.
// It returns a *ValidationError if id is empty or p sets nothing, and a *APIError on server errors.
func (s *GithubAppTokenService) UpdateFields(ctx context.Context, id string, p Patch) (*GithubAppToken, error) {
	if err := validateID("githubAppTokenID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("github_app_token", id)
	return s.updateFields(ctx, path, id, p)
}

// Delete removes a GitHub App token by ID.
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *GithubAppTokenService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	path := s.client.apiPath("github_app_token", id)
	return s.del(ctx, path)
}
