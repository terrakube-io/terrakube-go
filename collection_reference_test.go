package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func newTestCollectionReference() *terrakube.CollectionReference {
	desc := "test reference"
	return &terrakube.CollectionReference{
		ID:          "ref-1",
		Description: &desc,
	}
}

// --- List ---

func TestCollectionReferenceService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/reference", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.CollectionReference{newTestCollectionReference()})
	})

	client := newTestClient(t, srv)
	refs, err := client.CollectionReferences.List(context.Background(), "org-1", "col-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("got %d references, want 1", len(refs))
	}
	if refs[0].ID != "ref-1" {
		t.Errorf("ID = %q, want %q", refs[0].ID, "ref-1")
	}
}

func TestCollectionReferenceService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/reference", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[reference]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.CollectionReference{})
	})

	client := newTestClient(t, srv)
	_, err := client.CollectionReferences.List(context.Background(), "org-1", "col-1", &terrakube.ListOptions{Filter: "id==ref-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectionReferenceService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionReferences.List(context.Background(), "", "col-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestCollectionReferenceService_List_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionReferences.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "collectionID")
}

// --- Get ---

func TestCollectionReferenceService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/reference/ref-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestCollectionReference())
	})

	client := newTestClient(t, srv)
	ref, err := client.CollectionReferences.Get(context.Background(), "org-1", "col-1", "ref-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "ref-1" {
		t.Errorf("ID = %q, want %q", ref.ID, "ref-1")
	}
	if ref.Description == nil || *ref.Description != "test reference" {
		t.Errorf("Description = %v, want %q", ref.Description, "test reference")
	}
}

func TestCollectionReferenceService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/reference/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "reference not found")
	})

	client := newTestClient(t, srv)
	_, err := client.CollectionReferences.Get(context.Background(), "org-1", "col-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestCollectionReferenceService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		colID string
		id    string
		field string
	}{
		{"empty org ID", "", "col-1", "ref-1", "organizationID"},
		{"empty collection ID", "org-1", "", "ref-1", "collectionID"},
		{"empty reference ID", "org-1", "col-1", "", "referenceID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.CollectionReferences.Get(context.Background(), tt.orgID, tt.colID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

// --- Create ---

func TestCollectionReferenceService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/collection/col-1/reference", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestCollectionReference())
	})

	client := newTestClient(t, srv)
	desc := "new ref"
	ref, err := client.CollectionReferences.Create(context.Background(), "org-1", "col-1", &terrakube.CollectionReference{
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "ref-1" {
		t.Errorf("ID = %q, want %q", ref.ID, "ref-1")
	}
}

func TestCollectionReferenceService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionReferences.Create(context.Background(), "", "col-1", &terrakube.CollectionReference{})
	assertValidationError(t, err, "organizationID")
}

func TestCollectionReferenceService_Create_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionReferences.Create(context.Background(), "org-1", "", &terrakube.CollectionReference{})
	assertValidationError(t, err, "collectionID")
}

// --- Update ---

func TestCollectionReferenceService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/collection/col-1/reference/ref-1", func(w http.ResponseWriter, _ *http.Request) {
		updated := newTestCollectionReference()
		desc := "updated desc"
		updated.Description = &desc
		testutil.WriteJSONAPI(t, w, http.StatusOK, updated)
	})

	client := newTestClient(t, srv)
	desc := "updated desc"
	ref, err := client.CollectionReferences.Update(context.Background(), "org-1", "col-1", &terrakube.CollectionReference{
		ID:          "ref-1",
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Description == nil || *ref.Description != "updated desc" {
		t.Errorf("Description = %v, want %q", ref.Description, "updated desc")
	}
}

func TestCollectionReferenceService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionReferences.Update(context.Background(), "", "col-1", &terrakube.CollectionReference{ID: "ref-1"})
	assertValidationError(t, err, "organizationID")
}

func TestCollectionReferenceService_Update_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionReferences.Update(context.Background(), "org-1", "", &terrakube.CollectionReference{ID: "ref-1"})
	assertValidationError(t, err, "collectionID")
}

func TestCollectionReferenceService_Update_EmptyReferenceID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionReferences.Update(context.Background(), "org-1", "col-1", &terrakube.CollectionReference{ID: ""})
	assertValidationError(t, err, "referenceID")
}

// --- Delete ---

func TestCollectionReferenceService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/collection/col-1/reference/ref-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.CollectionReferences.Delete(context.Background(), "org-1", "col-1", "ref-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectionReferenceService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		colID string
		id    string
		field string
	}{
		{"empty org ID", "", "col-1", "ref-1", "organizationID"},
		{"empty collection ID", "org-1", "", "ref-1", "collectionID"},
		{"empty reference ID", "org-1", "col-1", "", "referenceID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.CollectionReferences.Delete(context.Background(), tt.orgID, tt.colID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestCollectionReferenceService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/collection/col-1/reference/ref-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.CollectionReferences.Delete(context.Background(), "org-1", "col-1", "ref-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- Auth ---

func TestCollectionReferenceService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/reference/ref-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestCollectionReference())
	})

	client := newTestClient(t, srv)
	_, _ = client.CollectionReferences.Get(context.Background(), "org-1", "col-1", "ref-1")
}
