package terrakube

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const patBasePath = "/pat/v1"

// PAT is a personal access token of the caller, as listed by the server. The
// token string itself is returned only once, by [PATService.Create].
type PAT struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	// Days is the lifetime in days from CreatedDate; 0 means the token never expires.
	Days    int  `json:"days"`
	Deleted bool `json:"deleted"`
	// Source is how the token was created: "API" or "CLI_LOGIN" (terraform login).
	Source      string     `json:"source"`
	LastUsedAt  *time.Time `json:"lastUsedAt"`
	CreatedDate *time.Time `json:"createdDate"`
	CreatedBy   *string    `json:"createdBy"`
	UpdatedDate *time.Time `json:"updatedDate"`
	UpdatedBy   *string    `json:"updatedBy"`
}

// PATCreated is a newly created personal access token.
type PATCreated struct {
	// Token is the bearer token. The server returns it only once.
	Token string `json:"token"`
	// ID identifies the token for [PATService.Delete]. The server does not
	// return it; it is read from the token's jti claim, and is empty if the
	// token cannot be decoded.
	ID string `json:"-"`
}

// PATService handles communication with the personal access token endpoints
// (/pat/v1). They act on the tokens of the caller.
type PATService struct {
	client *Client
}

// Create creates a personal access token for the caller that expires after
// days days; 0 means it never expires. The server may refuse a lifetime above
// its configured maximum (400, which then also refuses 0), and allows only an
// interactive user to create tokens (403), unless it lets a personal access
// token create shorter-lived ones.
//
// It returns a *ValidationError if days is negative and a *APIError on server
// errors.
func (s *PATService) Create(ctx context.Context, days int, description string) (*PATCreated, error) {
	if days < 0 {
		return nil, &ValidationError{Field: "days", Message: "must not be negative"}
	}

	body := struct {
		Days        int    `json:"days"`
		Description string `json:"description"`
	}{days, description}
	req, err := s.client.requestRaw(ctx, http.MethodPost, patBasePath, body)
	if err != nil {
		return nil, err
	}

	result := &PATCreated{}
	if _, err := s.client.doRaw(req, result); err != nil {
		return nil, err
	}
	// The server answers 201 with an empty token when it fails to sign one.
	if result.Token == "" {
		return nil, errors.New("terrakube: server returned no token")
	}
	result.ID = tokenID(result.Token)
	return result, nil
}

// List returns the caller's personal access tokens that are not deleted,
// expired ones included. A federated identity has none: the server answers
// 403.
//
// It returns a *APIError on server errors.
func (s *PATService) List(ctx context.Context) ([]*PAT, error) {
	req, err := s.client.requestRaw(ctx, http.MethodGet, patBasePath, nil)
	if err != nil {
		return nil, err
	}

	var pats []*PAT
	if _, err := s.client.doRaw(req, &pats); err != nil {
		return nil, err
	}
	return pats, nil
}

// Delete revokes a personal access token. Only its owner or the instance
// owner may revoke it (403 otherwise); an unknown or already deleted id is a
// 404.
//
// It returns a *ValidationError if id is empty and a *APIError on server errors.
func (s *PATService) Delete(ctx context.Context, id string) error {
	if err := validateID("id", id); err != nil {
		return err
	}

	req, err := s.client.requestRaw(ctx, http.MethodDelete, patBasePath+"/"+id, nil)
	if err != nil {
		return err
	}

	_, err = s.client.doRaw(req, nil)
	return err
}

// tokenID returns the jti claim of a JWT, without verifying the token, or ""
// when it has none or cannot be decoded.
func tokenID(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		ID string `json:"jti"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return ""
	}
	return claims.ID
}
