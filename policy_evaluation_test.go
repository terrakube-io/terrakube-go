package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestPolicyEvaluation() *terrakube.PolicyEvaluation {
	return &terrakube.PolicyEvaluation{
		ID:     "policyevaluation-1",
		Status: new("test-status"),
	}
}

func TestPolicyEvaluationService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/policyEvaluation", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicyEvaluation{newTestPolicyEvaluation()})
	})

	client := newTestClient(t, srv)

	items, err := client.PolicyEvaluations.List(context.Background(), "org-1", "job-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ID != "policyevaluation-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "policyevaluation-1")
	}
}

func TestPolicyEvaluationService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/policyEvaluation", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[policy_evaluation]"); got != "id==x" {
			t.Errorf("filter[policy_evaluation] = %q, want %q", got, "id==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicyEvaluation{})
	})

	client := newTestClient(t, srv)

	if _, err := client.PolicyEvaluations.List(context.Background(), "org-1", "job-1", &terrakube.ListOptions{Filter: "id==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicyEvaluationService_List_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/policyEvaluation", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyEvaluations.List(context.Background(), "org-1", "job-1", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestPolicyEvaluationService_List_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyEvaluations.List(context.Background(), "", "job-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestPolicyEvaluationService_List_EmptyJobID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyEvaluations.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "jobID")
}

func TestPolicyEvaluationService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/policyEvaluation/policyevaluation-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicyEvaluation())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyEvaluations.Get(context.Background(), "org-1", "job-1", "policyevaluation-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyevaluation-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyevaluation-1")
	}
	if item.Status == nil || *item.Status != "test-status" {
		t.Errorf("Status = %v, want %q", item.Status, "test-status")
	}
}

func TestPolicyEvaluationService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/policyEvaluation/policyevaluation-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyEvaluations.Get(context.Background(), "org-1", "job-1", "policyevaluation-1")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestPolicyEvaluationService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err0 := client.PolicyEvaluations.Get(context.Background(), "", "job-1", "policyevaluation-1")
	assertValidationError(t, err0, "organizationID")

	_, err1 := client.PolicyEvaluations.Get(context.Background(), "org-1", "", "policyevaluation-1")
	assertValidationError(t, err1, "jobID")

	_, err := client.PolicyEvaluations.Get(context.Background(), "org-1", "job-1", "")
	assertValidationError(t, err, "policyEvaluationID")
}
