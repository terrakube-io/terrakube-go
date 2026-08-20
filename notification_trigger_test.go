package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestNotificationTriggerService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/notificationConfiguration/cfg-1/triggers", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.NotificationTrigger{
			{ID: "trg-1", JobStatus: "completed"},
			{ID: "trg-2", JobStatus: "failed"},
		})
	})

	client := newTestClient(t, srv)
	triggers, err := client.NotificationTriggers.List(context.Background(), "org-1", "cfg-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(triggers) != 2 {
		t.Fatalf("got %d triggers, want 2", len(triggers))
	}
	if triggers[0].JobStatus != "completed" {
		t.Errorf("JobStatus = %q, want %q", triggers[0].JobStatus, "completed")
	}
}

func TestNotificationTriggerService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/notificationConfiguration/cfg-1/triggers/trg-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.NotificationTrigger{
			ID:        "trg-1",
			JobStatus: "running",
		})
	})

	client := newTestClient(t, srv)
	trg, err := client.NotificationTriggers.Get(context.Background(), "org-1", "cfg-1", "trg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trg.ID != "trg-1" {
		t.Errorf("ID = %q, want %q", trg.ID, "trg-1")
	}
	if trg.JobStatus != "running" {
		t.Errorf("JobStatus = %q, want %q", trg.JobStatus, "running")
	}
}

func TestNotificationTriggerService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/notificationConfiguration/cfg-1/triggers", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.NotificationTrigger{
			ID:        "trg-new",
			JobStatus: "failed",
		})
	})

	client := newTestClient(t, srv)
	created, err := client.NotificationTriggers.Create(context.Background(), "org-1", "cfg-1", &terrakube.NotificationTrigger{
		JobStatus: "failed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "trg-new" {
		t.Errorf("ID = %q, want %q", created.ID, "trg-new")
	}
}

func TestNotificationTriggerService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/notificationConfiguration/cfg-1/triggers/trg-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.NotificationTrigger{
			ID:        "trg-1",
			JobStatus: "cancelled",
		})
	})

	client := newTestClient(t, srv)
	updated, err := client.NotificationTriggers.Update(context.Background(), "org-1", "cfg-1", &terrakube.NotificationTrigger{
		ID:        "trg-1",
		JobStatus: "cancelled",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.JobStatus != "cancelled" {
		t.Errorf("JobStatus = %q, want %q", updated.JobStatus, "cancelled")
	}
}

func TestNotificationTriggerService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/notificationConfiguration/cfg-1/triggers/trg-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.NotificationTriggers.Delete(context.Background(), "org-1", "cfg-1", "trg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotificationTriggerService_ValidationErrors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.NotificationTriggers.List(context.Background(), "", "cfg-1", nil)
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationTriggers.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "notificationConfigurationID")

	_, err = client.NotificationTriggers.Get(context.Background(), "", "cfg-1", "trg-1")
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationTriggers.Get(context.Background(), "org-1", "", "trg-1")
	assertValidationError(t, err, "notificationConfigurationID")

	_, err = client.NotificationTriggers.Get(context.Background(), "org-1", "cfg-1", "")
	assertValidationError(t, err, "notificationTriggerID")

	_, err = client.NotificationTriggers.Create(context.Background(), "", "cfg-1", &terrakube.NotificationTrigger{})
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationTriggers.Create(context.Background(), "org-1", "", &terrakube.NotificationTrigger{})
	assertValidationError(t, err, "notificationConfigurationID")

	_, err = client.NotificationTriggers.Update(context.Background(), "", "cfg-1", &terrakube.NotificationTrigger{ID: "trg-1"})
	assertValidationError(t, err, "organizationID")

	_, err = client.NotificationTriggers.Update(context.Background(), "org-1", "", &terrakube.NotificationTrigger{ID: "trg-1"})
	assertValidationError(t, err, "notificationConfigurationID")

	_, err = client.NotificationTriggers.Update(context.Background(), "org-1", "cfg-1", &terrakube.NotificationTrigger{})
	assertValidationError(t, err, "notificationTriggerID")

	err = client.NotificationTriggers.Delete(context.Background(), "", "cfg-1", "trg-1")
	assertValidationError(t, err, "organizationID")

	err = client.NotificationTriggers.Delete(context.Background(), "org-1", "", "trg-1")
	assertValidationError(t, err, "notificationConfigurationID")

	err = client.NotificationTriggers.Delete(context.Background(), "org-1", "cfg-1", "")
	assertValidationError(t, err, "notificationTriggerID")
}
