package terrakube

import "context"

// NotificationConfiguration represents a Terrakube notification configuration resource.
type NotificationConfiguration struct {
	ID             string                 `jsonapi:"primary,notification_configuration"`
	Name           string                 `jsonapi:"attr,name"`
	Description    *string                `jsonapi:"attr,description,omitempty"`
	ChannelType    string                 `jsonapi:"attr,channelType"`
	DestinationURL string                 `jsonapi:"attr,destinationUrl"`
	SigningSecret  *string                `jsonapi:"attr,signingSecret,omitempty"`
	Active         bool                   `jsonapi:"attr,active"`
	MessageStyle   string                 `jsonapi:"attr,messageStyle,omitempty"`
	CreatedBy      *string                `jsonapi:"attr,createdBy"`
	CreatedDate    *string                `jsonapi:"attr,createdDate"`
	UpdatedBy      *string                `jsonapi:"attr,updatedBy"`
	UpdatedDate    *string                `jsonapi:"attr,updatedDate"`
	Organization   *Organization          `jsonapi:"relation,organization,omitempty"`
	Workspace      *Workspace             `jsonapi:"relation,workspace,omitempty"`
	Triggers       []*NotificationTrigger `jsonapi:"relation,triggers,omitempty"`
	Templates      []*Template            `jsonapi:"relation,templates,omitempty"`
}

// NotificationConfigurationService handles communication with the notification configuration
// related methods of the Terrakube API.
type NotificationConfigurationService struct {
	crudService[NotificationConfiguration]
}

// List returns all organization-level notification configurations.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *NotificationConfigurationService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*NotificationConfiguration, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration")
	return s.list(ctx, path, opts)
}

// ListByWorkspace returns all notification configurations for a specific workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *NotificationConfigurationService) ListByWorkspace(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*NotificationConfiguration, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "notificationConfiguration")
	return s.list(ctx, path, opts)
}

// Get returns a single notification configuration by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *NotificationConfigurationService) Get(ctx context.Context, orgID, id string) (*NotificationConfiguration, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("notificationConfigurationID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", id)
	return s.get(ctx, path)
}

// Create creates a new organization-level notification configuration.
// It returns a *ValidationError if orgID is empty and a *APIError on server errors.
func (s *NotificationConfigurationService) Create(ctx context.Context, orgID string, config *NotificationConfiguration) (*NotificationConfiguration, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration")
	return s.create(ctx, path, config)
}

// CreateForWorkspace creates a new notification configuration for a specific workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError on server errors.
func (s *NotificationConfigurationService) CreateForWorkspace(ctx context.Context, orgID, workspaceID string, config *NotificationConfiguration) (*NotificationConfiguration, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", workspaceID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "workspace", workspaceID, "notificationConfiguration")
	return s.create(ctx, path, config)
}

// Update modifies an existing notification configuration. The configuration's ID field must be set.
// It returns a *ValidationError if orgID or the ID is empty and a *APIError on server errors.
func (s *NotificationConfigurationService) Update(ctx context.Context, orgID string, config *NotificationConfiguration) (*NotificationConfiguration, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("notificationConfigurationID", config.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", config.ID)
	return s.update(ctx, path, config)
}

// Delete removes a notification configuration by ID.
// It returns a *ValidationError if orgID or id is empty and a *APIError on server errors.
func (s *NotificationConfigurationService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("notificationConfigurationID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", id)
	return s.del(ctx, path)
}
