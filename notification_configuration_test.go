package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestNotificationConfigurationService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/notificationConfiguration", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Slack notifications"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.NotificationConfiguration{
			{ID: "cfg-1", Name: "slack-alerts", Description: &desc, ChannelType: "SLACK", DestinationURL: "https://hooks.slack.com/services/1", Active: true, MessageStyle: "DETAILED"},
			{ID: "cfg-2", Name: "teams-alerts", ChannelType: "TEAMS", DestinationURL: "https://outlook.office.com/webhook/2", Active: false, MessageStyle: "SIMPLE"},
		})
	})

	client := newTestClient(t, srv)
	configs, err := client.NotificationConfigurations.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("got %d configs, want 2", len(configs))
	}
	if configs[0].Name != "slack-alerts" {
		t.Errorf("Name = %q, want %q", configs[0].Name, "slack-alerts")
	}
	if configs[0].ChannelType != "SLACK" {
		t.Errorf("ChannelType = %q, want %q", configs[0].ChannelType, "SLACK")
	}
}

func TestNotificationConfigurationService_ListByWorkspace(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/notificationConfiguration", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.NotificationConfiguration{
			{ID: "cfg-1", Name: "ws-alerts", ChannelType: "WEBHOOK", DestinationURL: "https://example.com/webhook", Active: true},
		})
	})

	client := newTestClient(t, srv)
	configs, err := client.NotificationConfigurations.ListByWorkspace(context.Background(), "org-1", "ws-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("got %d configs, want 1", len(configs))
	}
	if configs[0].ID != "cfg-1" {
		t.Errorf("ID = %q, want %q", configs[0].ID, "cfg-1")
	}
}

func TestNotificationConfigurationService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/notificationConfiguration/cfg-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.NotificationConfiguration{
			ID:             "cfg-1",
			Name:           "slack-alerts",
			ChannelType:    "SLACK",
			DestinationURL: "https://hooks.slack.com/services/1",
			Active:         true,
		})
	})

	client := newTestClient(t, srv)
	cfg, err := client.NotificationConfigurations.Get(context.Background(), "org-1", "cfg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ID != "cfg-1" {
		t.Errorf("ID = %q, want %q", cfg.ID, "cfg-1")
	}
}

func TestNotificationConfigurationService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/notificationConfiguration", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.NotificationConfiguration{
			ID:             "cfg-new",
			Name:           "new-alerts",
			ChannelType:    "SLACK",
			DestinationURL: "https://hooks.slack.com/services/new",
			Active:         true,
		})
	})

	client := newTestClient(t, srv)
	created, err := client.NotificationConfigurations.Create(context.Background(), "org-1", &terrakube.NotificationConfiguration{
		Name:           "new-alerts",
		ChannelType:    "SLACK",
		DestinationURL: "https://hooks.slack.com/services/new",
		Active:         true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "cfg-new" {
		t.Errorf("ID = %q, want %q", created.ID, "cfg-new")
	}
}

func TestNotificationConfigurationService_CreateForWorkspace(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/notificationConfiguration", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.NotificationConfiguration{
			ID:             "cfg-ws-new",
			Name:           "ws-alerts",
			ChannelType:    "WEBHOOK",
			DestinationURL: "https://example.com/hook",
			Active:         true,
		})
	})

	client := newTestClient(t, srv)
	created, err := client.NotificationConfigurations.CreateForWorkspace(context.Background(), "org-1", "ws-1", &terrakube.NotificationConfiguration{
		Name:           "ws-alerts",
		ChannelType:    "WEBHOOK",
		DestinationURL: "https://example.com/hook",
		Active:         true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "cfg-ws-new" {
		t.Errorf("ID = %q, want %q", created.ID, "cfg-ws-new")
	}
}

func TestNotificationConfigurationService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/notificationConfiguration/cfg-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.NotificationConfiguration{
			ID:             "cfg-1",
			Name:           "updated-alerts",
			ChannelType:    "SLACK",
			DestinationURL: "https://hooks.slack.com/services/updated",
			Active:         false,
		})
	})

	client := newTestClient(t, srv)
	updated, err := client.NotificationConfigurations.Update(context.Background(), "org-1", &terrakube.NotificationConfiguration{
		ID:             "cfg-1",
		Name:           "updated-alerts",
		ChannelType:    "SLACK",
		DestinationURL: "https://hooks.slack.com/services/updated",
		Active:         false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "updated-alerts" {
		t.Errorf("Name = %q, want %q", updated.Name, "updated-alerts")
	}
}

func TestNotificationConfigurationService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/notificationConfiguration/cfg-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.NotificationConfigurations.Delete(context.Background(), "org-1", "cfg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotificationConfigurationService_ValidationErrors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.NotificationConfigurations.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationConfigurations.ListByWorkspace(context.Background(), "", "ws-1", nil)
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationConfigurations.ListByWorkspace(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "workspaceID")

	_, err = client.NotificationConfigurations.Get(context.Background(), "", "cfg-1")
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationConfigurations.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "notificationConfigurationID")

	_, err = client.NotificationConfigurations.Create(context.Background(), "", &terrakube.NotificationConfiguration{})
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationConfigurations.CreateForWorkspace(context.Background(), "", "ws-1", &terrakube.NotificationConfiguration{})
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationConfigurations.CreateForWorkspace(context.Background(), "org-1", "", &terrakube.NotificationConfiguration{})
	assertValidationError(t, err, "workspaceID")

	_, err = client.NotificationConfigurations.Update(context.Background(), "", &terrakube.NotificationConfiguration{ID: "cfg-1"})
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationConfigurations.Update(context.Background(), "org-1", &terrakube.NotificationConfiguration{})
	assertValidationError(t, err, "notificationConfigurationID")

	err = client.NotificationConfigurations.Delete(context.Background(), "", "cfg-1")
	assertValidationError(t, err, "organizationID")

	err = client.NotificationConfigurations.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "notificationConfigurationID")
}
