package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Webhook represents a workspace webhook (v1 flat format).
type Webhook struct {
	ID           string `jsonapi:"primary,webhook"`
	Path         string `jsonapi:"attr,path"`
	Branch       string `jsonapi:"attr,branch"`
	TemplateID   string `jsonapi:"attr,templateId"`
	RemoteHookID string `jsonapi:"attr,remoteHookId"`
	Event        string `jsonapi:"attr,event"`
}

// WebhookEvent represents a webhook event entity.
type WebhookEvent struct {
	ID          string `jsonapi:"primary,webhook_event"`
	Branch      string `jsonapi:"attr,branch"`
	CreatedBy   string `jsonapi:"attr,createdBy"`
	CreatedDate string `jsonapi:"attr,createdDate"`
	Event       string `jsonapi:"attr,event"`
	Path        string `jsonapi:"attr,path"`
	Priority    int32  `jsonapi:"attr,priority"`
	TemplateID  string `jsonapi:"attr,templateId"`
	UpdatedBy   string `jsonapi:"attr,updatedBy"`
	UpdatedDate string `jsonapi:"attr,updatedDate"`
}

// WebhookService handles communication with the webhook related methods
// of the Terrakube API.
type WebhookService struct {
	client *Client
}

// WebhookEventService handles communication with the webhook event related
// methods of the Terrakube API.
type WebhookEventService struct {
	client *Client
}

// List returns all webhooks for a workspace.
func (s *WebhookService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*Webhook, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[webhook]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var webhooks []*Webhook
	_, err = s.client.do(ctx, req, &webhooks)
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

// Get retrieves a single webhook by ID.
func (s *WebhookService) Get(ctx context.Context, orgID, workspaceID, webhookID string) (*Webhook, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("webhookID", webhookID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	webhook := &Webhook{}
	_, err = s.client.do(ctx, req, webhook)
	if err != nil {
		return nil, err
	}

	return webhook, nil
}

// Create creates a new webhook for a workspace.
func (s *WebhookService) Create(ctx context.Context, orgID, workspaceID string, webhook *Webhook) (*Webhook, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook")

	req, err := s.client.request(ctx, http.MethodPost, p, webhook)
	if err != nil {
		return nil, err
	}

	created := &Webhook{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing webhook.
func (s *WebhookService) Update(ctx context.Context, orgID, workspaceID string, webhook *Webhook) (*Webhook, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("webhookID", webhook.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhook.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, webhook)
	if err != nil {
		return nil, err
	}

	updated := &Webhook{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a webhook.
func (s *WebhookService) Delete(ctx context.Context, orgID, workspaceID, webhookID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("webhookID", webhookID); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}

// List returns all events for a webhook.
func (s *WebhookEventService) List(ctx context.Context, orgID, workspaceID, webhookID string, opts *ListOptions) ([]*WebhookEvent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("webhookID", webhookID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[webhook_event]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var events []*WebhookEvent
	_, err = s.client.do(ctx, req, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// Get retrieves a single webhook event by ID.
func (s *WebhookEventService) Get(ctx context.Context, orgID, workspaceID, webhookID, eventID string) (*WebhookEvent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("webhookID", webhookID); err != nil {
		return nil, err
	}
	if err := validateID("eventID", eventID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event", eventID)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	event := &WebhookEvent{}
	_, err = s.client.do(ctx, req, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// Create creates a new webhook event.
func (s *WebhookEventService) Create(ctx context.Context, orgID, workspaceID, webhookID string, event *WebhookEvent) (*WebhookEvent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("webhookID", webhookID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event")

	req, err := s.client.request(ctx, http.MethodPost, p, event)
	if err != nil {
		return nil, err
	}

	created := &WebhookEvent{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing webhook event.
func (s *WebhookEventService) Update(ctx context.Context, orgID, workspaceID, webhookID string, event *WebhookEvent) (*WebhookEvent, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}
	if err := validateID("webhookID", webhookID); err != nil {
		return nil, err
	}
	if err := validateID("eventID", event.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event", event.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, event)
	if err != nil {
		return nil, err
	}

	updated := &WebhookEvent{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a webhook event.
func (s *WebhookEventService) Delete(ctx context.Context, orgID, workspaceID, webhookID, eventID string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return err
	}
	if err := validateID("webhookID", webhookID); err != nil {
		return err
	}
	if err := validateID("eventID", eventID); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event", eventID)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
