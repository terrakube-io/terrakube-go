package terrakube

import "context"

// Webhook represents a workspace webhook (v1 flat format).
type Webhook struct {
	ID           string  `jsonapi:"primary,webhook"`
	Path         string  `jsonapi:"attr,path"`
	Branch       string  `jsonapi:"attr,branch"`
	TemplateID   string  `jsonapi:"attr,templateId"`
	RemoteHookID string  `jsonapi:"attr,remoteHookId"`
	Event        string  `jsonapi:"attr,event"`
	CreatedBy    *string `jsonapi:"attr,createdBy"`
	CreatedDate  *string `jsonapi:"attr,createdDate"`
	UpdatedBy    *string `jsonapi:"attr,updatedBy"`
	UpdatedDate  *string `jsonapi:"attr,updatedDate"`
}

// WebhookEvent represents a webhook event entity.
type WebhookEvent struct {
	ID          string   `jsonapi:"primary,webhook_event"`
	Branch      string   `jsonapi:"attr,branch"`
	CreatedBy   string   `jsonapi:"attr,createdBy"`
	CreatedDate string   `jsonapi:"attr,createdDate"`
	Event       string   `jsonapi:"attr,event"`
	Path        string   `jsonapi:"attr,path"`
	Priority    int32    `jsonapi:"attr,priority"`
	TemplateID  string   `jsonapi:"attr,templateId"`
	UpdatedBy   string   `jsonapi:"attr,updatedBy"`
	UpdatedDate string   `jsonapi:"attr,updatedDate"`
	Webhook     *Webhook `jsonapi:"relation,webhook,omitempty"`
}

// WebhookService handles communication with the webhook related methods
// of the Terrakube API.
type WebhookService struct {
	crudService[Webhook]
}

// WebhookEventService handles communication with the webhook event related
// methods of the Terrakube API.
type WebhookEventService struct {
	crudService[WebhookEvent]
}

// List returns all webhooks for a workspace.
func (s *WebhookService) List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*Webhook, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook")
	return s.list(ctx, path, opts)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID)
	return s.get(ctx, path)
}

// Create creates a new webhook for a workspace.
func (s *WebhookService) Create(ctx context.Context, orgID, workspaceID string, webhook *Webhook) (*Webhook, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook")
	return s.create(ctx, path, webhook)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhook.ID)
	return s.update(ctx, path, webhook)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID)
	return s.del(ctx, path)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "events")
	return s.list(ctx, path, opts)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event", eventID)
	return s.get(ctx, path)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event")
	return s.create(ctx, path, event)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event", event.ID)
	return s.update(ctx, path, event)
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

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "webhook", webhookID, "event", eventID)
	return s.del(ctx, path)
}
