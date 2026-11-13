package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestPolicyAttachment() *terrakube.PolicyAttachment {
	return &terrakube.PolicyAttachment{
		ID: "policyattachment-1",
	}
}

func TestPolicyAttachmentService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/attachments", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicyAttachment{newTestPolicyAttachment()})
	})

	client := newTestClient(t, srv)

	items, err := client.PolicyAttachments.List(context.Background(), "org-1", "policySet-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ID != "policyattachment-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "policyattachment-1")
	}
}

func TestPolicyAttachmentService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/attachments", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[policy_attachment]"); got != "id==x" {
			t.Errorf("filter[policy_attachment] = %q, want %q", got, "id==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicyAttachment{})
	})

	client := newTestClient(t, srv)

	if _, err := client.PolicyAttachments.List(context.Background(), "org-1", "policySet-1", &terrakube.ListOptions{Filter: "id==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicyAttachmentService_List_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/attachments", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyAttachments.List(context.Background(), "org-1", "policySet-1", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestPolicyAttachmentService_List_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyAttachments.List(context.Background(), "", "policySet-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestPolicyAttachmentService_List_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyAttachments.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "policySetID")
}

func TestPolicyAttachmentService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/attachments/policyattachment-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicyAttachment())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyAttachments.Get(context.Background(), "org-1", "policySet-1", "policyattachment-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyattachment-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyattachment-1")
	}
}

func TestPolicyAttachmentService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/attachments/policyattachment-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyAttachments.Get(context.Background(), "org-1", "policySet-1", "policyattachment-1")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestPolicyAttachmentService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err0 := client.PolicyAttachments.Get(context.Background(), "", "policySet-1", "policyattachment-1")
	assertValidationError(t, err0, "organizationID")

	_, err1 := client.PolicyAttachments.Get(context.Background(), "org-1", "", "policyattachment-1")
	assertValidationError(t, err1, "policySetID")

	_, err := client.PolicyAttachments.Get(context.Background(), "org-1", "policySet-1", "")
	assertValidationError(t, err, "policyAttachmentID")
}

func TestPolicyAttachmentService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet/policySet-1/attachments", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestPolicyAttachment())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyAttachments.Create(context.Background(), "org-1", "policySet-1", &terrakube.PolicyAttachment{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyattachment-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyattachment-1")
	}
}

func TestPolicyAttachmentService_Create_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet/policySet-1/attachments", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "invalid")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyAttachments.Create(context.Background(), "org-1", "policySet-1", &terrakube.PolicyAttachment{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestPolicyAttachmentService_Create_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyAttachments.Create(context.Background(), "", "policySet-1", &terrakube.PolicyAttachment{})
	assertValidationError(t, err, "organizationID")
}

func TestPolicyAttachmentService_Create_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyAttachments.Create(context.Background(), "org-1", "", &terrakube.PolicyAttachment{})
	assertValidationError(t, err, "policySetID")
}

func TestPolicyAttachmentService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/policySet/policySet-1/attachments/policyattachment-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicyAttachment())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyAttachments.Update(context.Background(), "org-1", "policySet-1", newTestPolicyAttachment())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyattachment-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyattachment-1")
	}
}

func TestPolicyAttachmentService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyAttachments.Update(context.Background(), "org-1", "policySet-1", &terrakube.PolicyAttachment{})
	assertValidationError(t, err, "policyAttachmentID")
}

func TestPolicyAttachmentService_Update_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyAttachments.Update(context.Background(), "", "policySet-1", newTestPolicyAttachment())
	assertValidationError(t, err, "organizationID")
}

func TestPolicyAttachmentService_Update_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyAttachments.Update(context.Background(), "org-1", "", newTestPolicyAttachment())
	assertValidationError(t, err, "policySetID")
}

func TestPolicyAttachmentService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/policySet/policySet-1/attachments/policyattachment-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	if err := client.PolicyAttachments.Delete(context.Background(), "org-1", "policySet-1", "policyattachment-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicyAttachmentService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicyAttachments.Delete(context.Background(), "org-1", "policySet-1", "")
	assertValidationError(t, err, "policyAttachmentID")
}

func TestPolicyAttachmentService_Delete_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicyAttachments.Delete(context.Background(), "", "policySet-1", "policyattachment-1")
	assertValidationError(t, err, "organizationID")
}

func TestPolicyAttachmentService_Delete_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicyAttachments.Delete(context.Background(), "org-1", "", "policyattachment-1")
	assertValidationError(t, err, "policySetID")
}
