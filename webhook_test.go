package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

// ---------------------------------------------------------------------------
// WebhookService tests
// ---------------------------------------------------------------------------

func TestWebhookService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Webhook{
			{ID: "wh1", Path: "/src", Branch: "main", TemplateID: "tpl1", Event: "push"},
			{ID: "wh2", Path: "/lib", Branch: "dev", TemplateID: "tpl2", Event: "tag"},
		})
	})

	client := newTestClient(t, srv)
	webhooks, err := client.Webhooks.List(context.Background(), "org1", "ws1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(webhooks) != 2 {
		t.Fatalf("got %d webhooks, want 2", len(webhooks))
	}
	if webhooks[0].ID != "wh1" {
		t.Errorf("webhooks[0].ID = %q, want %q", webhooks[0].ID, "wh1")
	}
	if webhooks[1].Branch != "dev" {
		t.Errorf("webhooks[1].Branch = %q, want %q", webhooks[1].Branch, "dev")
	}
}

func TestWebhookService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[webhook]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Webhook{
			{ID: "wh1", Path: "/src", Branch: "main"},
		})
	})

	client := newTestClient(t, srv)
	webhooks, err := client.Webhooks.List(context.Background(), "org1", "ws1", &terrakube.ListOptions{Filter: "branch==main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(webhooks) != 1 {
		t.Fatalf("got %d webhooks, want 1", len(webhooks))
	}
}

func TestWebhookService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.List(context.Background(), "", "ws1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestWebhookService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.List(context.Background(), "org1", "", nil)
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/wh1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Webhook{
			ID: "wh1", Path: "/src", Branch: "main", TemplateID: "tpl1", RemoteHookID: "rh1", Event: "push",
		})
	})

	client := newTestClient(t, srv)
	wh, err := client.Webhooks.Get(context.Background(), "org1", "ws1", "wh1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wh.ID != "wh1" {
		t.Errorf("ID = %q, want %q", wh.ID, "wh1")
	}
	if wh.Path != "/src" {
		t.Errorf("Path = %q, want %q", wh.Path, "/src")
	}
	if wh.Branch != "main" {
		t.Errorf("Branch = %q, want %q", wh.Branch, "main")
	}
	if wh.TemplateID != "tpl1" {
		t.Errorf("TemplateID = %q, want %q", wh.TemplateID, "tpl1")
	}
	if wh.RemoteHookID != "rh1" {
		t.Errorf("RemoteHookID = %q, want %q", wh.RemoteHookID, "rh1")
	}
	if wh.Event != "push" {
		t.Errorf("Event = %q, want %q", wh.Event, "push")
	}
}

func TestWebhookService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "webhook not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Webhooks.Get(context.Background(), "org1", "ws1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestWebhookService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Get(context.Background(), "", "ws1", "wh1")
	assertValidationError(t, err, "organizationID")
}

func TestWebhookService_Get_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Get(context.Background(), "org1", "", "wh1")
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookService_Get_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Get(context.Background(), "org1", "ws1", "")
	assertValidationError(t, err, "webhookID")
}

func TestWebhookService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org1/workspace/ws1/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", r.Header.Get("Content-Type"), "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Webhook{
			ID: "wh-new", Path: "/app", Branch: "main", TemplateID: "tpl1", Event: "push",
		})
	})

	client := newTestClient(t, srv)
	wh, err := client.Webhooks.Create(context.Background(), "org1", "ws1", &terrakube.Webhook{
		Path: "/app", Branch: "main", TemplateID: "tpl1", Event: "push",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wh.ID != "wh-new" {
		t.Errorf("ID = %q, want %q", wh.ID, "wh-new")
	}
	if wh.Path != "/app" {
		t.Errorf("Path = %q, want %q", wh.Path, "/app")
	}
}

func TestWebhookService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Create(context.Background(), "", "ws1", &terrakube.Webhook{Path: "/app"})
	assertValidationError(t, err, "organizationID")
}

func TestWebhookService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Create(context.Background(), "org1", "", &terrakube.Webhook{Path: "/app"})
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookService_Create_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org1/workspace/ws1/webhook", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)
	_, err := client.Webhooks.Create(context.Background(), "org1", "ws1", &terrakube.Webhook{Path: "/app"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	var apiErr *terrakube.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestWebhookService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org1/workspace/ws1/webhook/wh1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Webhook{
			ID: "wh1", Path: "/updated", Branch: "release", TemplateID: "tpl2", Event: "tag",
		})
	})

	client := newTestClient(t, srv)
	wh, err := client.Webhooks.Update(context.Background(), "org1", "ws1", &terrakube.Webhook{
		ID: "wh1", Path: "/updated", Branch: "release", TemplateID: "tpl2", Event: "tag",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wh.Path != "/updated" {
		t.Errorf("Path = %q, want %q", wh.Path, "/updated")
	}
	if wh.Branch != "release" {
		t.Errorf("Branch = %q, want %q", wh.Branch, "release")
	}
}

func TestWebhookService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Update(context.Background(), "", "ws1", &terrakube.Webhook{ID: "wh1"})
	assertValidationError(t, err, "organizationID")
}

func TestWebhookService_Update_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Update(context.Background(), "org1", "", &terrakube.Webhook{ID: "wh1"})
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookService_Update_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.Webhooks.Update(context.Background(), "org1", "ws1", &terrakube.Webhook{ID: ""})
	assertValidationError(t, err, "webhookID")
}

func TestWebhookService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org1/workspace/ws1/webhook/wh1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Webhooks.Delete(context.Background(), "org1", "ws1", "wh1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWebhookService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	err := client.Webhooks.Delete(context.Background(), "", "ws1", "wh1")
	assertValidationError(t, err, "organizationID")
}

func TestWebhookService_Delete_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	err := client.Webhooks.Delete(context.Background(), "org1", "", "wh1")
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookService_Delete_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	err := client.Webhooks.Delete(context.Background(), "org1", "ws1", "")
	assertValidationError(t, err, "webhookID")
}

func TestWebhookService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org1/workspace/ws1/webhook/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "webhook not found")
	})

	client := newTestClient(t, srv)
	err := client.Webhooks.Delete(context.Background(), "org1", "ws1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestWebhookService_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/wh1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Webhook{ID: "wh1"})
	})

	client := newTestClient(t, srv)
	_, err := client.Webhooks.Get(context.Background(), "org1", "ws1", "wh1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// WebhookEventService tests
// ---------------------------------------------------------------------------

func TestWebhookEventService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/wh1/event", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.WebhookEvent{
			{ID: "evt1", Branch: "main", Event: "push", Priority: 1},
			{ID: "evt2", Branch: "dev", Event: "tag", Priority: 2},
		})
	})

	client := newTestClient(t, srv)
	events, err := client.WebhookEvents.List(context.Background(), "org1", "ws1", "wh1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].ID != "evt1" {
		t.Errorf("events[0].ID = %q, want %q", events[0].ID, "evt1")
	}
	if events[1].Branch != "dev" {
		t.Errorf("events[1].Branch = %q, want %q", events[1].Branch, "dev")
	}
}

func TestWebhookEventService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/wh1/event", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[webhook_event]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.WebhookEvent{
			{ID: "evt1", Branch: "main", Event: "push"},
		})
	})

	client := newTestClient(t, srv)
	events, err := client.WebhookEvents.List(context.Background(), "org1", "ws1", "wh1", &terrakube.ListOptions{Filter: "branch==main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
}

func TestWebhookEventService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.List(context.Background(), "", "ws1", "wh1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestWebhookEventService_List_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.List(context.Background(), "org1", "", "wh1", nil)
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookEventService_List_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.List(context.Background(), "org1", "ws1", "", nil)
	assertValidationError(t, err, "webhookID")
}

func TestWebhookEventService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/wh1/event/evt1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.WebhookEvent{
			ID:          "evt1",
			Branch:      "main",
			CreatedBy:   "admin",
			CreatedDate: "2025-01-15T10:00:00Z",
			Event:       "push",
			Path:        "/src",
			Priority:    1,
			TemplateID:  "tpl1",
			UpdatedBy:   "admin",
			UpdatedDate: "2025-01-15T10:00:00Z",
		})
	})

	client := newTestClient(t, srv)
	evt, err := client.WebhookEvents.Get(context.Background(), "org1", "ws1", "wh1", "evt1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.ID != "evt1" {
		t.Errorf("ID = %q, want %q", evt.ID, "evt1")
	}
	if evt.Branch != "main" {
		t.Errorf("Branch = %q, want %q", evt.Branch, "main")
	}
	if evt.CreatedBy != "admin" {
		t.Errorf("CreatedBy = %q, want %q", evt.CreatedBy, "admin")
	}
	if evt.Event != "push" {
		t.Errorf("Event = %q, want %q", evt.Event, "push")
	}
	if evt.Path != "/src" {
		t.Errorf("Path = %q, want %q", evt.Path, "/src")
	}
	if evt.Priority != 1 {
		t.Errorf("Priority = %d, want 1", evt.Priority)
	}
	if evt.TemplateID != "tpl1" {
		t.Errorf("TemplateID = %q, want %q", evt.TemplateID, "tpl1")
	}
}

func TestWebhookEventService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/wh1/event/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "event not found")
	})

	client := newTestClient(t, srv)
	_, err := client.WebhookEvents.Get(context.Background(), "org1", "ws1", "wh1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestWebhookEventService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Get(context.Background(), "", "ws1", "wh1", "evt1")
	assertValidationError(t, err, "organizationID")
}

func TestWebhookEventService_Get_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Get(context.Background(), "org1", "", "wh1", "evt1")
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookEventService_Get_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Get(context.Background(), "org1", "ws1", "", "evt1")
	assertValidationError(t, err, "webhookID")
}

func TestWebhookEventService_Get_EmptyEventID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Get(context.Background(), "org1", "ws1", "wh1", "")
	assertValidationError(t, err, "eventID")
}

func TestWebhookEventService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org1/workspace/ws1/webhook/wh1/event", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", r.Header.Get("Content-Type"), "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.WebhookEvent{
			ID: "evt-new", Branch: "main", Event: "push", Path: "/src", Priority: 1, TemplateID: "tpl1",
		})
	})

	client := newTestClient(t, srv)
	evt, err := client.WebhookEvents.Create(context.Background(), "org1", "ws1", "wh1", &terrakube.WebhookEvent{
		Branch: "main", Event: "push", Path: "/src", Priority: 1, TemplateID: "tpl1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.ID != "evt-new" {
		t.Errorf("ID = %q, want %q", evt.ID, "evt-new")
	}
	if evt.Branch != "main" {
		t.Errorf("Branch = %q, want %q", evt.Branch, "main")
	}
}

func TestWebhookEventService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Create(context.Background(), "", "ws1", "wh1", &terrakube.WebhookEvent{})
	assertValidationError(t, err, "organizationID")
}

func TestWebhookEventService_Create_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Create(context.Background(), "org1", "", "wh1", &terrakube.WebhookEvent{})
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookEventService_Create_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Create(context.Background(), "org1", "ws1", "", &terrakube.WebhookEvent{})
	assertValidationError(t, err, "webhookID")
}

func TestWebhookEventService_Create_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org1/workspace/ws1/webhook/wh1/event", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)
	_, err := client.WebhookEvents.Create(context.Background(), "org1", "ws1", "wh1", &terrakube.WebhookEvent{Branch: "main"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	var apiErr *terrakube.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestWebhookEventService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org1/workspace/ws1/webhook/wh1/event/evt1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.WebhookEvent{
			ID: "evt1", Branch: "release", Event: "tag", Priority: 5,
		})
	})

	client := newTestClient(t, srv)
	evt, err := client.WebhookEvents.Update(context.Background(), "org1", "ws1", "wh1", &terrakube.WebhookEvent{
		ID: "evt1", Branch: "release", Event: "tag", Priority: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.Branch != "release" {
		t.Errorf("Branch = %q, want %q", evt.Branch, "release")
	}
	if evt.Priority != 5 {
		t.Errorf("Priority = %d, want 5", evt.Priority)
	}
}

func TestWebhookEventService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Update(context.Background(), "", "ws1", "wh1", &terrakube.WebhookEvent{ID: "evt1"})
	assertValidationError(t, err, "organizationID")
}

func TestWebhookEventService_Update_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Update(context.Background(), "org1", "", "wh1", &terrakube.WebhookEvent{ID: "evt1"})
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookEventService_Update_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Update(context.Background(), "org1", "ws1", "", &terrakube.WebhookEvent{ID: "evt1"})
	assertValidationError(t, err, "webhookID")
}

func TestWebhookEventService_Update_EmptyEventID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	_, err := client.WebhookEvents.Update(context.Background(), "org1", "ws1", "wh1", &terrakube.WebhookEvent{ID: ""})
	assertValidationError(t, err, "eventID")
}

func TestWebhookEventService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org1/workspace/ws1/webhook/wh1/event/evt1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.WebhookEvents.Delete(context.Background(), "org1", "ws1", "wh1", "evt1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWebhookEventService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	err := client.WebhookEvents.Delete(context.Background(), "", "ws1", "wh1", "evt1")
	assertValidationError(t, err, "organizationID")
}

func TestWebhookEventService_Delete_EmptyWorkspaceID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	err := client.WebhookEvents.Delete(context.Background(), "org1", "", "wh1", "evt1")
	assertValidationError(t, err, "workspaceID")
}

func TestWebhookEventService_Delete_EmptyWebhookID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	err := client.WebhookEvents.Delete(context.Background(), "org1", "ws1", "", "evt1")
	assertValidationError(t, err, "webhookID")
}

func TestWebhookEventService_Delete_EmptyEventID(t *testing.T) {
	t.Parallel()
	client := newTestClientFromURL(t, "https://example.com")
	err := client.WebhookEvents.Delete(context.Background(), "org1", "ws1", "wh1", "")
	assertValidationError(t, err, "eventID")
}

func TestWebhookEventService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org1/workspace/ws1/webhook/wh1/event/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "event not found")
	})

	client := newTestClient(t, srv)
	err := client.WebhookEvents.Delete(context.Background(), "org1", "ws1", "wh1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestWebhookEventService_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org1/workspace/ws1/webhook/wh1/event/evt1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.WebhookEvent{ID: "evt1"})
	})

	client := newTestClient(t, srv)
	_, err := client.WebhookEvents.Get(context.Background(), "org1", "ws1", "wh1", "evt1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

