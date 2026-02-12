package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestSSHService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	desc := "my ssh key"
	srv.HandleFunc("GET /api/v1/organization/org-1/ssh", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.SSH{
			{ID: "ssh-1", Name: "key-one", Description: &desc, PrivateKey: "pk1", SSHType: "rsa"},
			{ID: "ssh-2", Name: "key-two", PrivateKey: "pk2", SSHType: "ed25519"},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.SSH.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "ssh-1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "ssh-1")
	}
	if items[0].Name != "key-one" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "key-one")
	}
	if items[0].SSHType != "rsa" {
		t.Errorf("items[0].SSHType = %q, want %q", items[0].SSHType, "rsa")
	}
}

func TestSSHService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/ssh", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[ssh]")
		if filter != "name==key-one" {
			t.Errorf("filter = %q, want %q", filter, "name==key-one")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.SSH{
			{ID: "ssh-1", Name: "key-one"},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.SSH.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==key-one"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
}

func TestSSHService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.SSH.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestSSHService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	desc := "test key"
	srv.HandleFunc("GET /api/v1/organization/org-1/ssh/ssh-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.SSH{
			ID:          "ssh-1",
			Name:        "key-one",
			Description: &desc,
			PrivateKey:  "pk1",
			SSHType:     "rsa",
		})
	})

	c := newTestClient(t, srv)

	ssh, err := c.SSH.Get(context.Background(), "org-1", "ssh-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ssh.ID != "ssh-1" {
		t.Errorf("ID = %q, want %q", ssh.ID, "ssh-1")
	}
	if ssh.Name != "key-one" {
		t.Errorf("Name = %q, want %q", ssh.Name, "key-one")
	}
	if ssh.PrivateKey != "pk1" {
		t.Errorf("PrivateKey = %q, want %q", ssh.PrivateKey, "pk1")
	}
}

func TestSSHService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/ssh/nonexistent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "ssh key not found")
	})

	c := newTestClient(t, srv)

	_, err := c.SSH.Get(context.Background(), "org-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestSSHService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.SSH.Get(context.Background(), "", "ssh-1")
	assertValidationError(t, err, "organizationID")
}

func TestSSHService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.SSH.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "sshID")
}

func TestSSHService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org-1/ssh", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.SSH{
			ID:         "ssh-new",
			Name:       "new-key",
			PrivateKey: "pk-new",
			SSHType:    "ed25519",
		})
	})

	c := newTestClient(t, srv)

	ssh, err := c.SSH.Create(context.Background(), "org-1", &terrakube.SSH{
		Name:       "new-key",
		PrivateKey: "pk-new",
		SSHType:    "ed25519",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ssh.ID != "ssh-new" {
		t.Errorf("ID = %q, want %q", ssh.ID, "ssh-new")
	}
	if ssh.Name != "new-key" {
		t.Errorf("Name = %q, want %q", ssh.Name, "new-key")
	}
}

func TestSSHService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.SSH.Create(context.Background(), "", &terrakube.SSH{Name: "test"})
	assertValidationError(t, err, "organizationID")
}

func TestSSHService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org-1/ssh/ssh-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.SSH{
			ID:      "ssh-1",
			Name:    "updated-key",
			SSHType: "rsa",
		})
	})

	c := newTestClient(t, srv)

	ssh, err := c.SSH.Update(context.Background(), "org-1", &terrakube.SSH{
		ID:      "ssh-1",
		Name:    "updated-key",
		SSHType: "rsa",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ssh.Name != "updated-key" {
		t.Errorf("Name = %q, want %q", ssh.Name, "updated-key")
	}
}

func TestSSHService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.SSH.Update(context.Background(), "", &terrakube.SSH{ID: "ssh-1"})
	assertValidationError(t, err, "organizationID")
}

func TestSSHService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.SSH.Update(context.Background(), "org-1", &terrakube.SSH{ID: ""})
	assertValidationError(t, err, "sshID")
}

func TestSSHService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/ssh/ssh-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	c := newTestClient(t, srv)

	err := c.SSH.Delete(context.Background(), "org-1", "ssh-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.SSH.Delete(context.Background(), "", "ssh-1")
	assertValidationError(t, err, "organizationID")
}

func TestSSHService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.SSH.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "sshID")
}

func TestSSHService_Delete_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/ssh/ssh-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	c := newTestClient(t, srv)

	err := c.SSH.Delete(context.Background(), "org-1", "ssh-1")
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

func TestSSHService_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/ssh", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.SSH{})
	})

	c := newTestClient(t, srv)

	_, err := c.SSH.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
