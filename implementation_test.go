package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestImplementationService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Implementation{
			{ID: "impl-1", Os: "linux", Arch: "amd64", Filename: "terraform-provider-aws_5.0.0_linux_amd64.zip"},
			{ID: "impl-2", Os: "darwin", Arch: "arm64", Filename: "terraform-provider-aws_5.0.0_darwin_arm64.zip"},
		})
	})

	client := newTestClient(t, srv)
	impls, err := client.Implementations.List(context.Background(), "org-1", "prov-1", "ver-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(impls) != 2 {
		t.Fatalf("got %d implementations, want 2", len(impls))
	}
	if impls[0].Os != "linux" {
		t.Errorf("Os = %q, want %q", impls[0].Os, "linux")
	}
}

func TestImplementationService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[implementation]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Implementation{
			{ID: "impl-1", Os: "linux", Arch: "amd64", Filename: "filtered.zip"},
		})
	})

	client := newTestClient(t, srv)
	impls, err := client.Implementations.List(context.Background(), "org-1", "prov-1", "ver-1", &terrakube.ListOptions{Filter: "os==linux"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(impls) != 1 {
		t.Fatalf("got %d implementations, want 1", len(impls))
	}
}

func TestImplementationService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.List(context.Background(), "", "prov-1", "ver-1", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestImplementationService_List_EmptyProviderID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.List(context.Background(), "org-1", "", "ver-1", nil)
	if err == nil {
		t.Fatal("expected validation error for empty providerID")
	}
	assertValidationError(t, err, "provider ID")
}

func TestImplementationService_List_EmptyVersionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.List(context.Background(), "org-1", "prov-1", "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty versionID")
	}
	assertValidationError(t, err, "version ID")
}

func TestImplementationService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation/impl-1", func(w http.ResponseWriter, _ *http.Request) {
		dl := "https://example.com/download"
		shasum := "abc123"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Implementation{
			ID: "impl-1", Os: "linux", Arch: "amd64", Filename: "provider.zip",
			DownloadURL: &dl, Shasum: &shasum,
		})
	})

	client := newTestClient(t, srv)
	impl, err := client.Implementations.Get(context.Background(), "org-1", "prov-1", "ver-1", "impl-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if impl.ID != "impl-1" {
		t.Errorf("ID = %q, want %q", impl.ID, "impl-1")
	}
	if impl.Os != "linux" {
		t.Errorf("Os = %q, want %q", impl.Os, "linux")
	}
	if impl.Arch != "amd64" {
		t.Errorf("Arch = %q, want %q", impl.Arch, "amd64")
	}
	if impl.Filename != "provider.zip" {
		t.Errorf("Filename = %q, want %q", impl.Filename, "provider.zip")
	}
	if impl.DownloadURL == nil || *impl.DownloadURL != "https://example.com/download" {
		t.Errorf("DownloadURL = %v, want %q", impl.DownloadURL, "https://example.com/download")
	}
	if impl.Shasum == nil || *impl.Shasum != "abc123" {
		t.Errorf("Shasum = %v, want %q", impl.Shasum, "abc123")
	}
}

func TestImplementationService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.Get(context.Background(), "", "prov-1", "ver-1", "impl-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestImplementationService_Get_EmptyProviderID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.Get(context.Background(), "org-1", "", "ver-1", "impl-1")
	if err == nil {
		t.Fatal("expected validation error for empty providerID")
	}
	assertValidationError(t, err, "provider ID")
}

func TestImplementationService_Get_EmptyVersionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.Get(context.Background(), "org-1", "prov-1", "", "impl-1")
	if err == nil {
		t.Fatal("expected validation error for empty versionID")
	}
	assertValidationError(t, err, "version ID")
}

func TestImplementationService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.Get(context.Background(), "org-1", "prov-1", "ver-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "implementation ID")
}

func TestImplementationService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "implementation not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Implementations.Get(context.Background(), "org-1", "prov-1", "ver-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestImplementationService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Implementation{
			ID: "impl-new", Os: "linux", Arch: "amd64", Filename: "new.zip",
		})
	})

	client := newTestClient(t, srv)
	impl, err := client.Implementations.Create(context.Background(), "org-1", "prov-1", "ver-1", &terrakube.Implementation{
		Os: "linux", Arch: "amd64", Filename: "new.zip",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if impl.ID != "impl-new" {
		t.Errorf("ID = %q, want %q", impl.ID, "impl-new")
	}
	if impl.Os != "linux" {
		t.Errorf("Os = %q, want %q", impl.Os, "linux")
	}
}

func TestImplementationService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation/impl-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Implementation{
			ID: "impl-1", Os: "darwin", Arch: "arm64", Filename: "updated.zip",
		})
	})

	client := newTestClient(t, srv)
	impl, err := client.Implementations.Update(context.Background(), "org-1", "prov-1", "ver-1", &terrakube.Implementation{
		ID: "impl-1", Os: "darwin", Arch: "arm64", Filename: "updated.zip",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if impl.Os != "darwin" {
		t.Errorf("Os = %q, want %q", impl.Os, "darwin")
	}
	if impl.Arch != "arm64" {
		t.Errorf("Arch = %q, want %q", impl.Arch, "arm64")
	}
}

func TestImplementationService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Implementations.Update(context.Background(), "org-1", "prov-1", "ver-1", &terrakube.Implementation{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "implementation ID")
}

func TestImplementationService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation/impl-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Implementations.Delete(context.Background(), "org-1", "prov-1", "ver-1", "impl-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImplementationService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Implementations.Delete(context.Background(), "", "prov-1", "ver-1", "impl-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestImplementationService_Delete_EmptyProviderID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Implementations.Delete(context.Background(), "org-1", "", "ver-1", "impl-1")
	if err == nil {
		t.Fatal("expected validation error for empty providerID")
	}
	assertValidationError(t, err, "provider ID")
}

func TestImplementationService_Delete_EmptyVersionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Implementations.Delete(context.Background(), "org-1", "prov-1", "", "impl-1")
	if err == nil {
		t.Fatal("expected validation error for empty versionID")
	}
	assertValidationError(t, err, "version ID")
}

func TestImplementationService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Implementations.Delete(context.Background(), "org-1", "prov-1", "ver-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "implementation ID")
}

func TestImplementationService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation/impl-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Implementations.Delete(context.Background(), "org-1", "prov-1", "ver-1", "impl-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestImplementationService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/provider/prov-1/version/ver-1/implementation/impl-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Implementation{ID: "impl-1", Os: "linux", Arch: "amd64", Filename: "test.zip"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Implementations.Get(context.Background(), "org-1", "prov-1", "ver-1", "impl-1")
}
