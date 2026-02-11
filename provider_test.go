package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func TestProviderService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider", func(w http.ResponseWriter, _ *http.Request) {
		desc := "AWS provider"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Provider{
			{ID: "prov-1", Name: "aws", Description: &desc},
			{ID: "prov-2", Name: "gcp"},
		})
	})

	client := newTestClient(t, srv)
	providers, err := client.Providers.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("got %d providers, want 2", len(providers))
	}
	if providers[0].Name != "aws" {
		t.Errorf("Name = %q, want %q", providers[0].Name, "aws")
	}
}

func TestProviderService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Provider{
			{ID: "prov-1", Name: "aws"},
		})
	})

	client := newTestClient(t, srv)
	providers, err := client.Providers.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==aws"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("got %d providers, want 1", len(providers))
	}
}

func TestProviderService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Providers.List(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty organization ID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestProviderService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "AWS provider"
		createdBy := "admin"
		createdDate := "2024-01-01"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Provider{
			ID: "prov-1", Name: "aws", Description: &desc,
			CreatedBy: &createdBy, CreatedDate: &createdDate,
		})
	})

	client := newTestClient(t, srv)
	provider, err := client.Providers.Get(context.Background(), "org-1", "prov-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.ID != "prov-1" {
		t.Errorf("ID = %q, want %q", provider.ID, "prov-1")
	}
	if provider.Name != "aws" {
		t.Errorf("Name = %q, want %q", provider.Name, "aws")
	}
	if provider.Description == nil || *provider.Description != "AWS provider" {
		t.Errorf("Description = %v, want %q", provider.Description, "AWS provider")
	}
	if provider.CreatedBy == nil || *provider.CreatedBy != "admin" {
		t.Errorf("CreatedBy = %v, want %q", provider.CreatedBy, "admin")
	}
	if provider.CreatedDate == nil || *provider.CreatedDate != "2024-01-01" {
		t.Errorf("CreatedDate = %v, want %q", provider.CreatedDate, "2024-01-01")
	}
}

func TestProviderService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		id    string
		field string
	}{
		{"empty org ID", "", "prov-1", "organization ID"},
		{"empty provider ID", "org-1", "", "provider ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.Providers.Get(context.Background(), tt.orgID, tt.id)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestProviderService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "provider not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Providers.Get(context.Background(), "org-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestProviderService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/provider", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		desc := "New provider"
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Provider{
			ID: "prov-new", Name: "azure", Description: &desc,
		})
	})

	client := newTestClient(t, srv)
	desc := "New provider"
	provider, err := client.Providers.Create(context.Background(), "org-1", &terrakube.Provider{
		Name: "azure", Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.ID != "prov-new" {
		t.Errorf("ID = %q, want %q", provider.ID, "prov-new")
	}
	if provider.Name != "azure" {
		t.Errorf("Name = %q, want %q", provider.Name, "azure")
	}
}

func TestProviderService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Providers.Create(context.Background(), "", &terrakube.Provider{})
	if err == nil {
		t.Fatal("expected validation error for empty organization ID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestProviderService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/provider/prov-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Updated"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Provider{
			ID: "prov-1", Name: "aws-updated", Description: &desc,
		})
	})

	client := newTestClient(t, srv)
	desc := "Updated"
	provider, err := client.Providers.Update(context.Background(), "org-1", &terrakube.Provider{
		ID: "prov-1", Name: "aws-updated", Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.Name != "aws-updated" {
		t.Errorf("Name = %q, want %q", provider.Name, "aws-updated")
	}
}

func TestProviderService_Update_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name     string
		orgID    string
		provider *terrakube.Provider
		field    string
	}{
		{"empty org ID", "", &terrakube.Provider{ID: "prov-1"}, "organization ID"},
		{"empty provider ID", "org-1", &terrakube.Provider{ID: ""}, "provider ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.Providers.Update(context.Background(), tt.orgID, tt.provider)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestProviderService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/provider/prov-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Providers.Delete(context.Background(), "org-1", "prov-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		id    string
		field string
	}{
		{"empty org ID", "", "prov-1", "organization ID"},
		{"empty provider ID", "org-1", "", "provider ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.Providers.Delete(context.Background(), tt.orgID, tt.id)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestProviderService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/provider/prov-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Providers.Delete(context.Background(), "org-1", "prov-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestProviderService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Provider{ID: "prov-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Providers.Get(context.Background(), "org-1", "prov-1")
}
