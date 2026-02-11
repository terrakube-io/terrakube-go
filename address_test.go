package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func TestAddressService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/address", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Address{
			{ID: "addr-1", Name: "aws_instance.web", Type: "resource"},
			{ID: "addr-2", Name: "module.vpc", Type: "module"},
		})
	})

	client := newTestClient(t, srv)
	addrs, err := client.Addresses.List(context.Background(), "org-1", "job-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addrs) != 2 {
		t.Fatalf("got %d addresses, want 2", len(addrs))
	}
	if addrs[0].Name != "aws_instance.web" {
		t.Errorf("Name = %q, want %q", addrs[0].Name, "aws_instance.web")
	}
}

func TestAddressService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/address", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Address{
			{ID: "addr-1", Name: "aws_instance.web", Type: "resource"},
		})
	})

	client := newTestClient(t, srv)
	addrs, err := client.Addresses.List(context.Background(), "org-1", "job-1", &terrakube.ListOptions{Filter: "type==resource"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("got %d addresses, want 1", len(addrs))
	}
}

func TestAddressService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Addresses.List(context.Background(), "", "job-1", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestAddressService_List_EmptyJobID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Addresses.List(context.Background(), "org-1", "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty jobID")
	}
	assertValidationError(t, err, "job ID")
}

func TestAddressService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/address/addr-1", func(w http.ResponseWriter, _ *http.Request) {
		createdBy := "admin"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Address{
			ID: "addr-1", Name: "aws_instance.web", Type: "resource", CreatedBy: &createdBy,
		})
	})

	client := newTestClient(t, srv)
	addr, err := client.Addresses.Get(context.Background(), "org-1", "job-1", "addr-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr.ID != "addr-1" {
		t.Errorf("ID = %q, want %q", addr.ID, "addr-1")
	}
	if addr.Name != "aws_instance.web" {
		t.Errorf("Name = %q, want %q", addr.Name, "aws_instance.web")
	}
	if addr.Type != "resource" {
		t.Errorf("Type = %q, want %q", addr.Type, "resource")
	}
	if addr.CreatedBy == nil || *addr.CreatedBy != "admin" {
		t.Errorf("CreatedBy = %v, want %q", addr.CreatedBy, "admin")
	}
}

func TestAddressService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Addresses.Get(context.Background(), "", "job-1", "addr-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestAddressService_Get_EmptyJobID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Addresses.Get(context.Background(), "org-1", "", "addr-1")
	if err == nil {
		t.Fatal("expected validation error for empty jobID")
	}
	assertValidationError(t, err, "job ID")
}

func TestAddressService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Addresses.Get(context.Background(), "org-1", "job-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "address ID")
}

func TestAddressService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/address/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "address not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Addresses.Get(context.Background(), "org-1", "job-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestAddressService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/job/job-1/address", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Address{
			ID: "addr-new", Name: "aws_s3_bucket.data", Type: "resource",
		})
	})

	client := newTestClient(t, srv)
	addr, err := client.Addresses.Create(context.Background(), "org-1", "job-1", &terrakube.Address{
		Name: "aws_s3_bucket.data", Type: "resource",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr.ID != "addr-new" {
		t.Errorf("ID = %q, want %q", addr.ID, "addr-new")
	}
	if addr.Name != "aws_s3_bucket.data" {
		t.Errorf("Name = %q, want %q", addr.Name, "aws_s3_bucket.data")
	}
}

func TestAddressService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/job/job-1/address/addr-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Address{
			ID: "addr-1", Name: "aws_instance.updated", Type: "resource",
		})
	})

	client := newTestClient(t, srv)
	addr, err := client.Addresses.Update(context.Background(), "org-1", "job-1", &terrakube.Address{
		ID: "addr-1", Name: "aws_instance.updated", Type: "resource",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr.Name != "aws_instance.updated" {
		t.Errorf("Name = %q, want %q", addr.Name, "aws_instance.updated")
	}
}

func TestAddressService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Addresses.Update(context.Background(), "org-1", "job-1", &terrakube.Address{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "address ID")
}

func TestAddressService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/job/job-1/address/addr-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Addresses.Delete(context.Background(), "org-1", "job-1", "addr-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddressService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Addresses.Delete(context.Background(), "", "job-1", "addr-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestAddressService_Delete_EmptyJobID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Addresses.Delete(context.Background(), "org-1", "", "addr-1")
	if err == nil {
		t.Fatal("expected validation error for empty jobID")
	}
	assertValidationError(t, err, "job ID")
}

func TestAddressService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Addresses.Delete(context.Background(), "org-1", "job-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "address ID")
}

func TestAddressService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/job/job-1/address/addr-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Addresses.Delete(context.Background(), "org-1", "job-1", "addr-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestAddressService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/address/addr-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Address{
			ID: "addr-1", Name: "aws_instance.web", Type: "resource",
		})
	})

	client := newTestClient(t, srv)
	_, _ = client.Addresses.Get(context.Background(), "org-1", "job-1", "addr-1")
}
