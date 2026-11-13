package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestPolicyExemption() *terrakube.PolicyExemption {
	return &terrakube.PolicyExemption{
		ID:        "policyexemption-1",
		ExpiresAt: new("test-expiresAt"),
	}
}

func TestPolicyExemptionService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policyExemption", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicyExemption{newTestPolicyExemption()})
	})

	client := newTestClient(t, srv)

	items, err := client.PolicyExemptions.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ID != "policyexemption-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "policyexemption-1")
	}
}

func TestPolicyExemptionService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policyExemption", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[policy_exemption]"); got != "id==x" {
			t.Errorf("filter[policy_exemption] = %q, want %q", got, "id==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicyExemption{})
	})

	client := newTestClient(t, srv)

	if _, err := client.PolicyExemptions.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "id==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicyExemptionService_List_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policyExemption", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyExemptions.List(context.Background(), "org-1", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestPolicyExemptionService_List_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyExemptions.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestPolicyExemptionService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policyExemption/policyexemption-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicyExemption())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyExemptions.Get(context.Background(), "org-1", "policyexemption-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyexemption-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyexemption-1")
	}
	if item.ExpiresAt == nil || *item.ExpiresAt != "test-expiresAt" {
		t.Errorf("ExpiresAt = %v, want %q", item.ExpiresAt, "test-expiresAt")
	}
}

func TestPolicyExemptionService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policyExemption/policyexemption-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyExemptions.Get(context.Background(), "org-1", "policyexemption-1")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestPolicyExemptionService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err0 := client.PolicyExemptions.Get(context.Background(), "", "policyexemption-1")
	assertValidationError(t, err0, "organizationID")

	_, err := client.PolicyExemptions.Get(context.Background(), "org-1", "")
	assertValidationError(t, err, "policyExemptionID")
}

func TestPolicyExemptionService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policyExemption", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestPolicyExemption())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyExemptions.Create(context.Background(), "org-1", &terrakube.PolicyExemption{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyexemption-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyexemption-1")
	}
}

func TestPolicyExemptionService_Create_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policyExemption", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "invalid")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyExemptions.Create(context.Background(), "org-1", &terrakube.PolicyExemption{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestPolicyExemptionService_Create_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyExemptions.Create(context.Background(), "", &terrakube.PolicyExemption{})
	assertValidationError(t, err, "organizationID")
}

func TestPolicyExemptionService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/policyExemption/policyexemption-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicyExemption())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyExemptions.Update(context.Background(), "org-1", newTestPolicyExemption())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyexemption-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyexemption-1")
	}
}

func TestPolicyExemptionService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyExemptions.Update(context.Background(), "org-1", &terrakube.PolicyExemption{})
	assertValidationError(t, err, "policyExemptionID")
}

func TestPolicyExemptionService_Update_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyExemptions.Update(context.Background(), "", newTestPolicyExemption())
	assertValidationError(t, err, "organizationID")
}

func TestPolicyExemptionService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/policyExemption/policyexemption-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	if err := client.PolicyExemptions.Delete(context.Background(), "org-1", "policyexemption-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicyExemptionService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicyExemptions.Delete(context.Background(), "org-1", "")
	assertValidationError(t, err, "policyExemptionID")
}

func TestPolicyExemptionService_Delete_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicyExemptions.Delete(context.Background(), "", "policyexemption-1")
	assertValidationError(t, err, "organizationID")
}
