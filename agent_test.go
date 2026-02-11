package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func TestAgentService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/agent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Agent{
			{ID: "agent-1", Name: "local-agent", Description: "local runner", URL: "https://agent1.example.com"},
			{ID: "agent-2", Name: "cloud-agent", Description: "cloud runner", URL: "https://agent2.example.com"},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.Agents.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "agent-1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "agent-1")
	}
	if items[0].Name != "local-agent" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "local-agent")
	}
	if items[0].Description != "local runner" {
		t.Errorf("items[0].Description = %q, want %q", items[0].Description, "local runner")
	}
	if items[0].URL != "https://agent1.example.com" {
		t.Errorf("items[0].URL = %q, want %q", items[0].URL, "https://agent1.example.com")
	}
}

func TestAgentService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/agent", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[agent]")
		if filter != "name==local-agent" {
			t.Errorf("filter = %q, want %q", filter, "name==local-agent")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Agent{
			{ID: "agent-1", Name: "local-agent"},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.Agents.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==local-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
}

func TestAgentService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Agents.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestAgentService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/agent/agent-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Agent{
			ID:          "agent-1",
			Name:        "local-agent",
			Description: "local runner",
			URL:         "https://agent1.example.com",
		})
	})

	c := newTestClient(t, srv)

	agent, err := c.Agents.Get(context.Background(), "org-1", "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.ID != "agent-1" {
		t.Errorf("ID = %q, want %q", agent.ID, "agent-1")
	}
	if agent.Name != "local-agent" {
		t.Errorf("Name = %q, want %q", agent.Name, "local-agent")
	}
	if agent.URL != "https://agent1.example.com" {
		t.Errorf("URL = %q, want %q", agent.URL, "https://agent1.example.com")
	}
}

func TestAgentService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/agent/nonexistent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "agent not found")
	})

	c := newTestClient(t, srv)

	_, err := c.Agents.Get(context.Background(), "org-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestAgentService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Agents.Get(context.Background(), "", "agent-1")
	assertValidationError(t, err, "organizationID")
}

func TestAgentService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Agents.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "agentID")
}

func TestAgentService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org-1/agent", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Agent{
			ID:          "agent-new",
			Name:        "new-agent",
			Description: "freshly created",
			URL:         "https://new-agent.example.com",
		})
	})

	c := newTestClient(t, srv)

	agent, err := c.Agents.Create(context.Background(), "org-1", &terrakube.Agent{
		Name:        "new-agent",
		Description: "freshly created",
		URL:         "https://new-agent.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.ID != "agent-new" {
		t.Errorf("ID = %q, want %q", agent.ID, "agent-new")
	}
	if agent.Name != "new-agent" {
		t.Errorf("Name = %q, want %q", agent.Name, "new-agent")
	}
}

func TestAgentService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Agents.Create(context.Background(), "", &terrakube.Agent{Name: "test"})
	assertValidationError(t, err, "organizationID")
}

func TestAgentService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org-1/agent/agent-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Agent{
			ID:          "agent-1",
			Name:        "updated-agent",
			Description: "updated desc",
			URL:         "https://updated.example.com",
		})
	})

	c := newTestClient(t, srv)

	agent, err := c.Agents.Update(context.Background(), "org-1", &terrakube.Agent{
		ID:          "agent-1",
		Name:        "updated-agent",
		Description: "updated desc",
		URL:         "https://updated.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.Name != "updated-agent" {
		t.Errorf("Name = %q, want %q", agent.Name, "updated-agent")
	}
}

func TestAgentService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Agents.Update(context.Background(), "", &terrakube.Agent{ID: "agent-1"})
	assertValidationError(t, err, "organizationID")
}

func TestAgentService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.Agents.Update(context.Background(), "org-1", &terrakube.Agent{ID: ""})
	assertValidationError(t, err, "agentID")
}

func TestAgentService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/agent/agent-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	c := newTestClient(t, srv)

	err := c.Agents.Delete(context.Background(), "org-1", "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.Agents.Delete(context.Background(), "", "agent-1")
	assertValidationError(t, err, "organizationID")
}

func TestAgentService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.Agents.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "agentID")
}

func TestAgentService_Delete_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/agent/agent-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	c := newTestClient(t, srv)

	err := c.Agents.Delete(context.Background(), "org-1", "agent-1")
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

func TestAgentService_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/agent", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Agent{})
	})

	c := newTestClient(t, srv)

	_, err := c.Agents.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
