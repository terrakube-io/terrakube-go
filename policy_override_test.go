package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestPolicyOverride() *terrakube.PolicyOverride {
	return &terrakube.PolicyOverride{
		ID:            "policyoverride-1",
		Justification: new("test-justification"),
	}
}

func TestPolicyOverrideService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/policyEvaluation/policyEvaluation-1/override/policyoverride-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicyOverride())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyOverrides.Get(context.Background(), "org-1", "job-1", "policyEvaluation-1", "policyoverride-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyoverride-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyoverride-1")
	}
	if item.Justification == nil || *item.Justification != "test-justification" {
		t.Errorf("Justification = %v, want %q", item.Justification, "test-justification")
	}
}

func TestPolicyOverrideService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/policyEvaluation/policyEvaluation-1/override/policyoverride-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyOverrides.Get(context.Background(), "org-1", "job-1", "policyEvaluation-1", "policyoverride-1")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestPolicyOverrideService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err0 := client.PolicyOverrides.Get(context.Background(), "", "job-1", "policyEvaluation-1", "policyoverride-1")
	assertValidationError(t, err0, "organizationID")

	_, err1 := client.PolicyOverrides.Get(context.Background(), "org-1", "", "policyEvaluation-1", "policyoverride-1")
	assertValidationError(t, err1, "jobID")

	_, err2 := client.PolicyOverrides.Get(context.Background(), "org-1", "job-1", "", "policyoverride-1")
	assertValidationError(t, err2, "policyEvaluationID")

	_, err := client.PolicyOverrides.Get(context.Background(), "org-1", "job-1", "policyEvaluation-1", "")
	assertValidationError(t, err, "policyOverrideID")
}

func TestPolicyOverrideService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/job/job-1/policyEvaluation/policyEvaluation-1/override", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestPolicyOverride())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyOverrides.Create(context.Background(), "org-1", "job-1", "policyEvaluation-1", &terrakube.PolicyOverride{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyoverride-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyoverride-1")
	}
}

func TestPolicyOverrideService_Create_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/job/job-1/policyEvaluation/policyEvaluation-1/override", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "invalid")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicyOverrides.Create(context.Background(), "org-1", "job-1", "policyEvaluation-1", &terrakube.PolicyOverride{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestPolicyOverrideService_Create_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyOverrides.Create(context.Background(), "", "job-1", "policyEvaluation-1", &terrakube.PolicyOverride{})
	assertValidationError(t, err, "organizationID")
}

func TestPolicyOverrideService_Create_EmptyJobID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyOverrides.Create(context.Background(), "org-1", "", "policyEvaluation-1", &terrakube.PolicyOverride{})
	assertValidationError(t, err, "jobID")
}

func TestPolicyOverrideService_Create_EmptyPolicyEvaluationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyOverrides.Create(context.Background(), "org-1", "job-1", "", &terrakube.PolicyOverride{})
	assertValidationError(t, err, "policyEvaluationID")
}

func TestPolicyOverrideService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/job/job-1/policyEvaluation/policyEvaluation-1/override/policyoverride-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicyOverride())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicyOverrides.Update(context.Background(), "org-1", "job-1", "policyEvaluation-1", newTestPolicyOverride())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policyoverride-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policyoverride-1")
	}
}

func TestPolicyOverrideService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyOverrides.Update(context.Background(), "org-1", "job-1", "policyEvaluation-1", &terrakube.PolicyOverride{})
	assertValidationError(t, err, "policyOverrideID")
}

func TestPolicyOverrideService_Update_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyOverrides.Update(context.Background(), "", "job-1", "policyEvaluation-1", newTestPolicyOverride())
	assertValidationError(t, err, "organizationID")
}

func TestPolicyOverrideService_Update_EmptyJobID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyOverrides.Update(context.Background(), "org-1", "", "policyEvaluation-1", newTestPolicyOverride())
	assertValidationError(t, err, "jobID")
}

func TestPolicyOverrideService_Update_EmptyPolicyEvaluationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicyOverrides.Update(context.Background(), "org-1", "job-1", "", newTestPolicyOverride())
	assertValidationError(t, err, "policyEvaluationID")
}
