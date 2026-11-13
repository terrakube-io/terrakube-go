package terrakube_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestPolicySet() *terrakube.PolicySet {
	return &terrakube.PolicySet{
		ID:     "policyset-1",
		Branch: new("test-branch"),
	}
}

func TestPolicySetService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicySet{newTestPolicySet()})
	})

	client := newTestClient(t, srv)

	items, err := client.PolicySets.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ID != "policyset-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "policyset-1")
	}
}

func TestPolicySetService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[policy_set]"); got != "id==x" {
			t.Errorf("filter[policy_set] = %q, want %q", got, "id==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicySet{})
	})

	client := newTestClient(t, srv)

	if _, err := client.PolicySets.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "id==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicySetService_List_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicySets.List(context.Background(), "org-1", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestPolicySetService_List_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySets.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestPolicySetService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policyset-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicySet())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicySets.Get(context.Background(), "org-1", "policyset-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyset-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyset-1")
	}
	if item.Branch == nil || *item.Branch != "test-branch" {
		t.Errorf("Branch = %v, want %q", item.Branch, "test-branch")
	}
}

func TestPolicySetService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policyset-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicySets.Get(context.Background(), "org-1", "policyset-1")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestPolicySetService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err0 := client.PolicySets.Get(context.Background(), "", "policyset-1")
	assertValidationError(t, err0, "organizationID")

	_, err := client.PolicySets.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "policySetID")
}

func TestPolicySetService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestPolicySet())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicySets.Create(context.Background(), "org-1", &terrakube.PolicySet{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyset-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyset-1")
	}
}

func TestPolicySetService_Create_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "invalid")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicySets.Create(context.Background(), "org-1", &terrakube.PolicySet{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestPolicySetService_Create_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySets.Create(context.Background(), "", &terrakube.PolicySet{})
	assertValidationError(t, err, "organizationID")
}

func TestPolicySet_BooleanFieldsSerialization(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}
		assertJSONBoolField(t, body, "global", false)
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestPolicySet())
	})

	client := newTestClient(t, srv)

	if _, err := client.PolicySets.Create(context.Background(), "org-1", &terrakube.PolicySet{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicySetService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/policySet/policyset-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicySet())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicySets.Update(context.Background(), "org-1", newTestPolicySet())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyset-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyset-1")
	}
}

func TestPolicySetService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySets.Update(context.Background(), "org-1", &terrakube.PolicySet{})
	assertValidationError(t, err, "policySetID")
}

func TestPolicySetService_Update_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySets.Update(context.Background(), "", newTestPolicySet())
	assertValidationError(t, err, "organizationID")
}

func TestPolicySetService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/policySet/policyset-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	if err := client.PolicySets.Delete(context.Background(), "org-1", "policyset-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicySetService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicySets.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "policySetID")
}

func TestPolicySetService_Delete_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicySets.Delete(context.Background(), "", "policyset-1")
	assertValidationError(t, err, "organizationID")
}
