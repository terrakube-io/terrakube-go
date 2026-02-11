package terrakube_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/google/jsonapi"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
)

func newTestCollectionItem() *terrakube.CollectionItem {
	desc := "test item"
	return &terrakube.CollectionItem{
		ID:          "item-1",
		Key:         "DB_HOST",
		Value:       "localhost",
		Description: &desc,
		Category:    "ENV",
		Sensitive:   false,
		Hcl:         false,
	}
}

// --- List ---

func TestCollectionItemService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/item", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.CollectionItem{newTestCollectionItem()})
	})

	client := newTestClient(t, srv)
	items, err := client.CollectionItems.List(context.Background(), "org-1", "col-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Key != "DB_HOST" {
		t.Errorf("Key = %q, want %q", items[0].Key, "DB_HOST")
	}
}

func TestCollectionItemService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/item", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[item]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.CollectionItem{})
	})

	client := newTestClient(t, srv)
	_, err := client.CollectionItems.List(context.Background(), "org-1", "col-1", &terrakube.ListOptions{Filter: "key==foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectionItemService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionItems.List(context.Background(), "", "col-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestCollectionItemService_List_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionItems.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "collectionID")
}

// --- Get ---

func TestCollectionItemService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/item/item-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestCollectionItem())
	})

	client := newTestClient(t, srv)
	item, err := client.CollectionItems.Get(context.Background(), "org-1", "col-1", "item-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "item-1" {
		t.Errorf("ID = %q, want %q", item.ID, "item-1")
	}
	if item.Key != "DB_HOST" {
		t.Errorf("Key = %q, want %q", item.Key, "DB_HOST")
	}
}

func TestCollectionItemService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/item/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "item not found")
	})

	client := newTestClient(t, srv)
	_, err := client.CollectionItems.Get(context.Background(), "org-1", "col-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestCollectionItemService_Get_EmptyIDs(t *testing.T) {
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
		{"empty org ID", "", "col-1", "item-1", "organizationID"},
		{"empty collection ID", "org-1", "", "item-1", "collectionID"},
		{"empty item ID", "org-1", "col-1", "", "itemID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.CollectionItems.Get(context.Background(), tt.orgID, tt.colID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

// --- Create ---

func TestCollectionItemService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/collection/col-1/item", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestCollectionItem())
	})

	client := newTestClient(t, srv)
	input := &terrakube.CollectionItem{
		Key:      "DB_HOST",
		Value:    "localhost",
		Category: "ENV",
	}
	item, err := client.CollectionItems.Create(context.Background(), "org-1", "col-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "item-1" {
		t.Errorf("ID = %q, want %q", item.ID, "item-1")
	}
}

func TestCollectionItemService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionItems.Create(context.Background(), "", "col-1", &terrakube.CollectionItem{})
	assertValidationError(t, err, "organizationID")
}

func TestCollectionItemService_Create_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionItems.Create(context.Background(), "org-1", "", &terrakube.CollectionItem{})
	assertValidationError(t, err, "collectionID")
}

// --- Update ---

func TestCollectionItemService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/collection/col-1/item/item-1", func(w http.ResponseWriter, _ *http.Request) {
		updated := newTestCollectionItem()
		updated.Value = "remotehost"
		testutil.WriteJSONAPI(t, w, http.StatusOK, updated)
	})

	client := newTestClient(t, srv)
	input := &terrakube.CollectionItem{
		ID:    "item-1",
		Key:   "DB_HOST",
		Value: "remotehost",
	}
	item, err := client.CollectionItems.Update(context.Background(), "org-1", "col-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Value != "remotehost" {
		t.Errorf("Value = %q, want %q", item.Value, "remotehost")
	}
}

func TestCollectionItemService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionItems.Update(context.Background(), "", "col-1", &terrakube.CollectionItem{ID: "item-1"})
	assertValidationError(t, err, "organizationID")
}

func TestCollectionItemService_Update_EmptyCollectionID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionItems.Update(context.Background(), "org-1", "", &terrakube.CollectionItem{ID: "item-1"})
	assertValidationError(t, err, "collectionID")
}

func TestCollectionItemService_Update_EmptyItemID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.CollectionItems.Update(context.Background(), "org-1", "col-1", &terrakube.CollectionItem{ID: ""})
	assertValidationError(t, err, "itemID")
}

// --- Delete ---

func TestCollectionItemService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/collection/col-1/item/item-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.CollectionItems.Delete(context.Background(), "org-1", "col-1", "item-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectionItemService_Delete_EmptyIDs(t *testing.T) {
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
		{"empty org ID", "", "col-1", "item-1", "organizationID"},
		{"empty collection ID", "org-1", "", "item-1", "collectionID"},
		{"empty item ID", "org-1", "col-1", "", "itemID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.CollectionItems.Delete(context.Background(), tt.orgID, tt.colID, tt.id)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestCollectionItemService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/collection/col-1/item/item-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.CollectionItems.Delete(context.Background(), "org-1", "col-1", "item-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- Auth ---

func TestCollectionItemService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/collection/col-1/item/item-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestCollectionItem())
	})

	client := newTestClient(t, srv)
	_, _ = client.CollectionItems.Get(context.Background(), "org-1", "col-1", "item-1")
}

// --- Boolean serialization ---

func TestCollectionItem_BooleanFalseNotDropped(t *testing.T) {
	t.Parallel()

	item := &terrakube.CollectionItem{
		ID:        "item-1",
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: false,
		Hcl:       false,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, item); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.Bytes()
	if !bytes.Contains(payload, []byte(`"sensitive"`)) {
		t.Error("Sensitive=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
	if !bytes.Contains(payload, []byte(`"hcl"`)) {
		t.Error("Hcl=false was dropped from JSON:API payload; omitempty must not be used on boolean fields")
	}
}

func TestCollectionItem_BooleanTruePreserved(t *testing.T) {
	t.Parallel()

	item := &terrakube.CollectionItem{
		ID:        "item-1",
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: true,
		Hcl:       true,
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, item); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	payload := buf.Bytes()
	if !bytes.Contains(payload, []byte(`"sensitive"`)) {
		t.Error("Sensitive=true was dropped from JSON:API payload")
	}
	if !bytes.Contains(payload, []byte(`"hcl"`)) {
		t.Error("Hcl=true was dropped from JSON:API payload")
	}
}

func TestCollectionItemService_Create_BooleanFalseSerialized(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/collection/col-1/item", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		assertJSONBoolField(t, body, "sensitive", false)
		assertJSONBoolField(t, body, "hcl", false)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestCollectionItem())
	})

	client := newTestClient(t, srv)
	_, err := client.CollectionItems.Create(context.Background(), "org-1", "col-1", &terrakube.CollectionItem{
		Key:       "MY_VAR",
		Value:     "val",
		Category:  "ENV",
		Sensitive: false,
		Hcl:       false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
