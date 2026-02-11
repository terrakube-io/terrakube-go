package terrakube

import (
	"context"
	"net/http"
)

const teamTokenBasePath = "/access-token/v1/teams"

// TeamToken represents a Terrakube team access token.
type TeamToken struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Days        int32  `json:"days"`
	Hours       int32  `json:"hours"`
	Minutes     int32  `json:"minutes"`
	Group       string `json:"group"`
	Value       string `json:"token"`
}

// TeamTokenService handles communication with the team token endpoints.
type TeamTokenService struct {
	client *Client
}

// Create generates a new team token.
func (s *TeamTokenService) Create(ctx context.Context, token *TeamToken) (*TeamToken, error) {
	req, err := s.client.requestRaw(ctx, http.MethodPost, teamTokenBasePath, token)
	if err != nil {
		return nil, err
	}

	result := &TeamToken{}
	_, err = s.client.doRaw(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// List returns all team tokens.
func (s *TeamTokenService) List(ctx context.Context) ([]TeamToken, error) {
	req, err := s.client.requestRaw(ctx, http.MethodGet, teamTokenBasePath, nil)
	if err != nil {
		return nil, err
	}

	var tokens []TeamToken
	_, err = s.client.doRaw(ctx, req, &tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Delete removes a team token by ID.
func (s *TeamTokenService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	req, err := s.client.requestRaw(ctx, http.MethodDelete, teamTokenBasePath+"/"+id, nil)
	if err != nil {
		return err
	}

	_, err = s.client.doRaw(ctx, req, nil)
	return err
}
