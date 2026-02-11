package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func TestVCSService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/vcs", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.VCS{
			{
				ID:             "vcs-1",
				Name:           "github-app",
				Description:    "GitHub App connection",
				VcsType:        "GITHUB",
				ConnectionType: "app",
				ClientID:       "client-id-1",
				ClientSecret:   "client-secret-1",
				PrivateKey:     "private-key-1",
				Endpoint:       "https://github.com",
				APIURL:         "https://api.github.com",
				Status:         "active",
			},
			{
				ID:             "vcs-2",
				Name:           "gitlab-oauth",
				Description:    "GitLab OAuth connection",
				VcsType:        "GITLAB",
				ConnectionType: "oauth",
				ClientID:       "client-id-2",
				ClientSecret:   "client-secret-2",
				PrivateKey:     "",
				Endpoint:       "https://gitlab.com",
				APIURL:         "https://gitlab.com/api/v4",
				Status:         "pending",
			},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.VCS.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "vcs-1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "vcs-1")
	}
	if items[0].Name != "github-app" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "github-app")
	}
	if items[0].Description != "GitHub App connection" {
		t.Errorf("items[0].Description = %q, want %q", items[0].Description, "GitHub App connection")
	}
	if items[0].VcsType != "GITHUB" {
		t.Errorf("items[0].VcsType = %q, want %q", items[0].VcsType, "GITHUB")
	}
	if items[0].ConnectionType != "app" {
		t.Errorf("items[0].ConnectionType = %q, want %q", items[0].ConnectionType, "app")
	}
	if items[0].ClientID != "client-id-1" {
		t.Errorf("items[0].ClientID = %q, want %q", items[0].ClientID, "client-id-1")
	}
	if items[0].ClientSecret != "client-secret-1" {
		t.Errorf("items[0].ClientSecret = %q, want %q", items[0].ClientSecret, "client-secret-1")
	}
	if items[0].PrivateKey != "private-key-1" {
		t.Errorf("items[0].PrivateKey = %q, want %q", items[0].PrivateKey, "private-key-1")
	}
	if items[0].Endpoint != "https://github.com" {
		t.Errorf("items[0].Endpoint = %q, want %q", items[0].Endpoint, "https://github.com")
	}
	if items[0].APIURL != "https://api.github.com" {
		t.Errorf("items[0].APIURL = %q, want %q", items[0].APIURL, "https://api.github.com")
	}
	if items[0].Status != "active" {
		t.Errorf("items[0].Status = %q, want %q", items[0].Status, "active")
	}
}

func TestVCSService_List_WithFilter(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/vcs", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[vcs]")
		if filter != "name==github-app" {
			t.Errorf("filter = %q, want %q", filter, "name==github-app")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.VCS{
			{ID: "vcs-1", Name: "github-app"},
		})
	})

	c := newTestClient(t, srv)

	items, err := c.VCS.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==github-app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
}

func TestVCSService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.VCS.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestVCSService_Get(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/vcs/vcs-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.VCS{
			ID:             "vcs-1",
			Name:           "github-app",
			Description:    "GitHub App connection",
			VcsType:        "GITHUB",
			ConnectionType: "app",
			ClientID:       "client-id-1",
			ClientSecret:   "client-secret-1",
			PrivateKey:     "private-key-1",
			Endpoint:       "https://github.com",
			APIURL:         "https://api.github.com",
			Status:         "active",
		})
	})

	c := newTestClient(t, srv)

	vcs, err := c.VCS.Get(context.Background(), "org-1", "vcs-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vcs.ID != "vcs-1" {
		t.Errorf("ID = %q, want %q", vcs.ID, "vcs-1")
	}
	if vcs.Name != "github-app" {
		t.Errorf("Name = %q, want %q", vcs.Name, "github-app")
	}
	if vcs.Description != "GitHub App connection" {
		t.Errorf("Description = %q, want %q", vcs.Description, "GitHub App connection")
	}
	if vcs.VcsType != "GITHUB" {
		t.Errorf("VcsType = %q, want %q", vcs.VcsType, "GITHUB")
	}
	if vcs.ConnectionType != "app" {
		t.Errorf("ConnectionType = %q, want %q", vcs.ConnectionType, "app")
	}
	if vcs.ClientID != "client-id-1" {
		t.Errorf("ClientID = %q, want %q", vcs.ClientID, "client-id-1")
	}
	if vcs.ClientSecret != "client-secret-1" {
		t.Errorf("ClientSecret = %q, want %q", vcs.ClientSecret, "client-secret-1")
	}
	if vcs.PrivateKey != "private-key-1" {
		t.Errorf("PrivateKey = %q, want %q", vcs.PrivateKey, "private-key-1")
	}
	if vcs.Endpoint != "https://github.com" {
		t.Errorf("Endpoint = %q, want %q", vcs.Endpoint, "https://github.com")
	}
	if vcs.APIURL != "https://api.github.com" {
		t.Errorf("APIURL = %q, want %q", vcs.APIURL, "https://api.github.com")
	}
	if vcs.Status != "active" {
		t.Errorf("Status = %q, want %q", vcs.Status, "active")
	}
}

func TestVCSService_Get_NotFound(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/vcs/nonexistent", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "vcs not found")
	})

	c := newTestClient(t, srv)

	_, err := c.VCS.Get(context.Background(), "org-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestVCSService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.VCS.Get(context.Background(), "", "vcs-1")
	assertValidationError(t, err, "organizationID")
}

func TestVCSService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.VCS.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "vcsID")
}

func TestVCSService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/organization/org-1/vcs", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.VCS{
			ID:             "vcs-new",
			Name:           "new-vcs",
			Description:    "freshly created",
			VcsType:        "GITHUB",
			ConnectionType: "app",
			ClientID:       "new-client-id",
			ClientSecret:   "new-client-secret",
			PrivateKey:     "new-private-key",
			Endpoint:       "https://github.com",
			APIURL:         "https://api.github.com",
			Status:         "pending",
		})
	})

	c := newTestClient(t, srv)

	vcs, err := c.VCS.Create(context.Background(), "org-1", &terrakube.VCS{
		Name:           "new-vcs",
		Description:    "freshly created",
		VcsType:        "GITHUB",
		ConnectionType: "app",
		ClientID:       "new-client-id",
		ClientSecret:   "new-client-secret",
		PrivateKey:     "new-private-key",
		Endpoint:       "https://github.com",
		APIURL:         "https://api.github.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vcs.ID != "vcs-new" {
		t.Errorf("ID = %q, want %q", vcs.ID, "vcs-new")
	}
	if vcs.Name != "new-vcs" {
		t.Errorf("Name = %q, want %q", vcs.Name, "new-vcs")
	}
	if vcs.Status != "pending" {
		t.Errorf("Status = %q, want %q", vcs.Status, "pending")
	}
}

func TestVCSService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.VCS.Create(context.Background(), "", &terrakube.VCS{Name: "test"})
	assertValidationError(t, err, "organizationID")
}

func TestVCSService_Update(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("PATCH /api/v1/organization/org-1/vcs/vcs-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.VCS{
			ID:             "vcs-1",
			Name:           "updated-vcs",
			Description:    "updated desc",
			VcsType:        "GITLAB",
			ConnectionType: "oauth",
			ClientID:       "updated-client-id",
			ClientSecret:   "updated-client-secret",
			PrivateKey:     "updated-private-key",
			Endpoint:       "https://gitlab.com",
			APIURL:         "https://gitlab.com/api/v4",
			Status:         "active",
		})
	})

	c := newTestClient(t, srv)

	vcs, err := c.VCS.Update(context.Background(), "org-1", &terrakube.VCS{
		ID:             "vcs-1",
		Name:           "updated-vcs",
		Description:    "updated desc",
		VcsType:        "GITLAB",
		ConnectionType: "oauth",
		ClientID:       "updated-client-id",
		ClientSecret:   "updated-client-secret",
		PrivateKey:     "updated-private-key",
		Endpoint:       "https://gitlab.com",
		APIURL:         "https://gitlab.com/api/v4",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vcs.Name != "updated-vcs" {
		t.Errorf("Name = %q, want %q", vcs.Name, "updated-vcs")
	}
	if vcs.VcsType != "GITLAB" {
		t.Errorf("VcsType = %q, want %q", vcs.VcsType, "GITLAB")
	}
}

func TestVCSService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.VCS.Update(context.Background(), "", &terrakube.VCS{ID: "vcs-1"})
	assertValidationError(t, err, "organizationID")
}

func TestVCSService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	_, err := c.VCS.Update(context.Background(), "org-1", &terrakube.VCS{ID: ""})
	assertValidationError(t, err, "vcsID")
}

func TestVCSService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/vcs/vcs-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	c := newTestClient(t, srv)

	err := c.VCS.Delete(context.Background(), "org-1", "vcs-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVCSService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.VCS.Delete(context.Background(), "", "vcs-1")
	assertValidationError(t, err, "organizationID")
}

func TestVCSService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	c := newTestClientFromURL(t, "https://example.com")

	err := c.VCS.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "vcsID")
}

func TestVCSService_Delete_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("DELETE /api/v1/organization/org-1/vcs/vcs-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	c := newTestClient(t, srv)

	err := c.VCS.Delete(context.Background(), "org-1", "vcs-1")
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

func TestVCSService_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /api/v1/organization/org-1/vcs", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.VCS{})
	})

	c := newTestClient(t, srv)

	_, err := c.VCS.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
