package terrakube

import (
	"context"
	"time"
)

// NotificationTrigger represents a Terrakube notification trigger resource.
type NotificationTrigger struct {
	ID            string                     `jsonapi:"primary,notification_trigger"`
	JobStatus     string                     `jsonapi:"attr,jobStatus"`
	CreatedBy     *string                    `jsonapi:"attr,createdBy"`
	CreatedDate   *time.Time                 `jsonapi:"attr,createdDate,rfc3339"`
	UpdatedBy     *string                    `jsonapi:"attr,updatedBy"`
	UpdatedDate   *time.Time                 `jsonapi:"attr,updatedDate,rfc3339"`
	Configuration *NotificationConfiguration `jsonapi:"relation,configuration,omitempty"`
}

// NotificationTriggerService handles communication with the notification trigger
// related methods of the Terrakube API.
type NotificationTriggerService struct {
	crudService[NotificationTrigger]
}

// List returns all triggers for a notification configuration.
// It returns a *ValidationError if orgID or configID is empty and a *APIError on server errors.
func (s *NotificationTriggerService) List(ctx context.Context, orgID, configID string, opts *ListOptions) ([]*NotificationTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("notificationConfigurationID", configID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", configID, "triggers")
	return s.list(ctx, path, opts)
}

// Get returns a single notification trigger by ID.
// It returns a *ValidationError if orgID, configID, or id is empty and a *APIError on server errors.
func (s *NotificationTriggerService) Get(ctx context.Context, orgID, configID, id string) (*NotificationTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("notificationConfigurationID", configID); err != nil {
		return nil, err
	}
	if err := validateID("notificationTriggerID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", configID, "triggers", id)
	return s.get(ctx, path)
}

// Create creates a new trigger for a notification configuration.
// It returns a *ValidationError if orgID or configID is empty and a *APIError on server errors.
func (s *NotificationTriggerService) Create(ctx context.Context, orgID, configID string, trigger *NotificationTrigger) (*NotificationTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("notificationConfigurationID", configID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", configID, "triggers")
	return s.create(ctx, path, trigger)
}

// Update modifies an existing notification trigger. The trigger's ID field must be set.
// It returns a *ValidationError if orgID, configID, or the ID is empty and a *APIError on server errors.
// Update sends every attribute of the struct, so start from the result of Get:
// a field left at its zero value overwrites the value on the server.
func (s *NotificationTriggerService) Update(ctx context.Context, orgID, configID string, trigger *NotificationTrigger) (*NotificationTrigger, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("notificationConfigurationID", configID); err != nil {
		return nil, err
	}
	if err := validateID("notificationTriggerID", trigger.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", configID, "triggers", trigger.ID)
	return s.update(ctx, path, trigger)
}

// Delete removes a notification trigger by ID.
// It returns a *ValidationError if orgID, configID, or id is empty and a *APIError on server errors.
func (s *NotificationTriggerService) Delete(ctx context.Context, orgID, configID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("notificationConfigurationID", configID); err != nil {
		return err
	}
	if err := validateID("notificationTriggerID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "notificationConfiguration", configID, "triggers", id)
	return s.del(ctx, path)
}
