package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestProviderVersionService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version", func(w http.ResponseWriter, _ *http.Request) {
		proto := "5.0"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.ProviderVersion{
			{ID: "ver-1", VersionNumber: "1.0.0", Protocols: &proto},
			{ID: "ver-2", VersionNumber: "2.0.0"},
		})
	})

	client := newTestClient(t, srv)
	versions, err := client.ProviderVersions.List(context.Background(), "org-1", "prov-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("got %d versions, want 2", len(versions))
	}
	if versions[0].VersionNumber != "1.0.0" {
		t.Errorf("VersionNumber = %q, want %q", versions[0].VersionNumber, "1.0.0")
	}
}

func TestProviderVersionService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[version]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.ProviderVersion{
			{ID: "ver-1", VersionNumber: "1.0.0"},
		})
	})

	client := newTestClient(t, srv)
	versions, err := client.ProviderVersions.List(context.Background(), "org-1", "prov-1", &terrakube.ListOptions{Filter: "versionNumber==1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("got %d versions, want 1", len(versions))
	}
}

func TestProviderVersionService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ProviderVersions.List(context.Background(), "", "prov-1", nil)
	if err == nil {
		t.Fatal("expected validation error for empty organization ID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestProviderVersionService_List_EmptyProviderID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ProviderVersions.List(context.Background(), "org-1", "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty provider ID")
	}
	assertValidationError(t, err, "provider ID")
}

func TestProviderVersionService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		proto := "5.0"
		createdBy := "admin"
		createdDate := "2024-01-01"
		updatedBy := "admin"
		updatedDate := "2024-06-01"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ProviderVersion{
			ID: "ver-1", VersionNumber: "1.0.0", Protocols: &proto,
			CreatedBy: &createdBy, CreatedDate: &createdDate,
			UpdatedBy: &updatedBy, UpdatedDate: &updatedDate,
		})
	})

	client := newTestClient(t, srv)
	version, err := client.ProviderVersions.Get(context.Background(), "org-1", "prov-1", "ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.ID != "ver-1" {
		t.Errorf("ID = %q, want %q", version.ID, "ver-1")
	}
	if version.VersionNumber != "1.0.0" {
		t.Errorf("VersionNumber = %q, want %q", version.VersionNumber, "1.0.0")
	}
	if version.Protocols == nil || *version.Protocols != "5.0" {
		t.Errorf("Protocols = %v, want %q", version.Protocols, "5.0")
	}
	if version.CreatedBy == nil || *version.CreatedBy != "admin" {
		t.Errorf("CreatedBy = %v, want %q", version.CreatedBy, "admin")
	}
	if version.CreatedDate == nil || *version.CreatedDate != "2024-01-01" {
		t.Errorf("CreatedDate = %v, want %q", version.CreatedDate, "2024-01-01")
	}
	if version.UpdatedBy == nil || *version.UpdatedBy != "admin" {
		t.Errorf("UpdatedBy = %v, want %q", version.UpdatedBy, "admin")
	}
	if version.UpdatedDate == nil || *version.UpdatedDate != "2024-06-01" {
		t.Errorf("UpdatedDate = %v, want %q", version.UpdatedDate, "2024-06-01")
	}
}

func TestProviderVersionService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name       string
		orgID      string
		providerID string
		id         string
		field      string
	}{
		{"empty org ID", "", "prov-1", "ver-1", "organization ID"},
		{"empty provider ID", "org-1", "", "ver-1", "provider ID"},
		{"empty version ID", "org-1", "prov-1", "", "version ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.ProviderVersions.Get(context.Background(), tt.orgID, tt.providerID, tt.id)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestProviderVersionService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "version not found")
	})

	client := newTestClient(t, srv)
	_, err := client.ProviderVersions.Get(context.Background(), "org-1", "prov-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestProviderVersionService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/provider/prov-1/version", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		proto := "5.0"
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.ProviderVersion{
			ID: "ver-new", VersionNumber: "3.0.0", Protocols: &proto,
		})
	})

	client := newTestClient(t, srv)
	proto := "5.0"
	version, err := client.ProviderVersions.Create(context.Background(), "org-1", "prov-1", &terrakube.ProviderVersion{
		VersionNumber: "3.0.0", Protocols: &proto,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.ID != "ver-new" {
		t.Errorf("ID = %q, want %q", version.ID, "ver-new")
	}
	if version.VersionNumber != "3.0.0" {
		t.Errorf("VersionNumber = %q, want %q", version.VersionNumber, "3.0.0")
	}
}

func TestProviderVersionService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ProviderVersions.Create(context.Background(), "", "prov-1", &terrakube.ProviderVersion{})
	if err == nil {
		t.Fatal("expected validation error for empty organization ID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestProviderVersionService_Create_EmptyProviderID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.ProviderVersions.Create(context.Background(), "org-1", "", &terrakube.ProviderVersion{})
	if err == nil {
		t.Fatal("expected validation error for empty provider ID")
	}
	assertValidationError(t, err, "provider ID")
}

func TestProviderVersionService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/provider/prov-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		proto := "6.0"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ProviderVersion{
			ID: "ver-1", VersionNumber: "1.1.0", Protocols: &proto,
		})
	})

	client := newTestClient(t, srv)
	proto := "6.0"
	version, err := client.ProviderVersions.Update(context.Background(), "org-1", "prov-1", &terrakube.ProviderVersion{
		ID: "ver-1", VersionNumber: "1.1.0", Protocols: &proto,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.VersionNumber != "1.1.0" {
		t.Errorf("VersionNumber = %q, want %q", version.VersionNumber, "1.1.0")
	}
}

func TestProviderVersionService_Update_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name       string
		orgID      string
		providerID string
		version    *terrakube.ProviderVersion
		field      string
	}{
		{"empty org ID", "", "prov-1", &terrakube.ProviderVersion{ID: "ver-1"}, "organization ID"},
		{"empty provider ID", "org-1", "", &terrakube.ProviderVersion{ID: "ver-1"}, "provider ID"},
		{"empty version ID", "org-1", "prov-1", &terrakube.ProviderVersion{ID: ""}, "version ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.ProviderVersions.Update(context.Background(), tt.orgID, tt.providerID, tt.version)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestProviderVersionService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/provider/prov-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.ProviderVersions.Delete(context.Background(), "org-1", "prov-1", "ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderVersionService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name       string
		orgID      string
		providerID string
		id         string
		field      string
	}{
		{"empty org ID", "", "prov-1", "ver-1", "organization ID"},
		{"empty provider ID", "org-1", "", "ver-1", "provider ID"},
		{"empty version ID", "org-1", "prov-1", "", "version ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.ProviderVersions.Delete(context.Background(), tt.orgID, tt.providerID, tt.id)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestProviderVersionService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/provider/prov-1/version/ver-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.ProviderVersions.Delete(context.Background(), "org-1", "prov-1", "ver-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestProviderVersionService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/ver-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.ProviderVersion{ID: "ver-1", VersionNumber: "1.0.0"})
	})

	client := newTestClient(t, srv)
	_, _ = client.ProviderVersions.Get(context.Background(), "org-1", "prov-1", "ver-1")
}
