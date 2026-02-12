package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestCollectionService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection", func(w http.ResponseWriter, _ *http.Request) {
		desc := "First collection"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Collection{
			{ID: "col-1", Name: "Vars", Description: &desc, Priority: 1},
			{ID: "col-2", Name: "Secrets", Priority: 2},
		})
	})

	client := newTestClient(t, srv)
	collections, err := client.Collections.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collections) != 2 {
		t.Fatalf("got %d collections, want 2", len(collections))
	}
	if collections[0].Name != "Vars" {
		t.Errorf("Name = %q, want %q", collections[0].Name, "Vars")
	}
}

func TestCollectionService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[collection]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Collection{
			{ID: "col-1", Name: "Filtered"},
		})
	})

	client := newTestClient(t, srv)
	collections, err := client.Collections.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "name==Filtered"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collections) != 1 {
		t.Fatalf("got %d collections, want 1", len(collections))
	}
}

func TestCollectionService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Collections.List(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organizationID")
}

func TestCollectionService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Test collection"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Collection{
			ID: "col-1", Name: "Vars", Description: &desc, Priority: 5,
		})
	})

	client := newTestClient(t, srv)
	col, err := client.Collections.Get(context.Background(), "org-1", "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.ID != "col-1" {
		t.Errorf("ID = %q, want %q", col.ID, "col-1")
	}
	if col.Name != "Vars" {
		t.Errorf("Name = %q, want %q", col.Name, "Vars")
	}
	if col.Priority != 5 {
		t.Errorf("Priority = %d, want %d", col.Priority, 5)
	}
	if col.Description == nil || *col.Description != "Test collection" {
		t.Errorf("Description = %v, want %q", col.Description, "Test collection")
	}
}

func TestCollectionService_Get_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Collections.Get(context.Background(), "", "col-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organizationID")
}

func TestCollectionService_Get_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Collections.Get(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "collectionID")
}

func TestCollectionService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "collection not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Collections.Get(context.Background(), "org-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestCollectionService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/collection", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		desc := "New collection"
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Collection{
			ID: "col-new", Name: "Created", Description: &desc, Priority: 3,
		})
	})

	client := newTestClient(t, srv)
	desc := "New collection"
	col, err := client.Collections.Create(context.Background(), "org-1", &terrakube.Collection{
		Name: "Created", Description: &desc, Priority: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.ID != "col-new" {
		t.Errorf("ID = %q, want %q", col.ID, "col-new")
	}
	if col.Name != "Created" {
		t.Errorf("Name = %q, want %q", col.Name, "Created")
	}
}

func TestCollectionService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Collections.Create(context.Background(), "", &terrakube.Collection{Name: "test"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organizationID")
}

func TestCollectionService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/collection/col-1", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Updated"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Collection{
			ID: "col-1", Name: "Updated", Description: &desc, Priority: 10,
		})
	})

	client := newTestClient(t, srv)
	desc := "Updated"
	col, err := client.Collections.Update(context.Background(), "org-1", &terrakube.Collection{
		ID: "col-1", Name: "Updated", Description: &desc, Priority: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Name != "Updated" {
		t.Errorf("Name = %q, want %q", col.Name, "Updated")
	}
	if col.Priority != 10 {
		t.Errorf("Priority = %d, want %d", col.Priority, 10)
	}
}

func TestCollectionService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Collections.Update(context.Background(), "", &terrakube.Collection{ID: "col-1"})
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organizationID")
}

func TestCollectionService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Collections.Update(context.Background(), "org-1", &terrakube.Collection{ID: ""})
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "collectionID")
}

func TestCollectionService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/collection/col-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Collections.Delete(context.Background(), "org-1", "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectionService_Delete_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Collections.Delete(context.Background(), "", "col-1")
	if err == nil {
		t.Fatal("expected validation error for empty orgID")
	}
	assertValidationError(t, err, "organizationID")
}

func TestCollectionService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	err := client.Collections.Delete(context.Background(), "org-1", "")
	if err == nil {
		t.Fatal("expected validation error for empty ID")
	}
	assertValidationError(t, err, "collectionID")
}

func TestCollectionService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/collection/col-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Collections.Delete(context.Background(), "org-1", "col-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestCollectionService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Collection{ID: "col-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Collections.Get(context.Background(), "org-1", "col-1")
}

