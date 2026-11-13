package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestPolicySetParameter() *terrakube.PolicySetParameter {
	return &terrakube.PolicySetParameter{
		ID:          "policysetparameter-1",
		Description: new("test-description"),
	}
}

func TestPolicySetParameterService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/parameters", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicySetParameter{newTestPolicySetParameter()})
	})

	client := newTestClient(t, srv)

	items, err := client.PolicySetParameters.List(context.Background(), "org-1", "policySet-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ID != "policysetparameter-1" {
		t.Errorf("ID = %q, want %q", items[0].ID, "policysetparameter-1")
	}
}

func TestPolicySetParameterService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/parameters", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[policy_set_parameter]"); got != "id==x" {
			t.Errorf("filter[policy_set_parameter] = %q, want %q", got, "id==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicySetParameter{})
	})

	client := newTestClient(t, srv)

	if _, err := client.PolicySetParameters.List(context.Background(), "org-1", "policySet-1", &terrakube.ListOptions{Filter: "id==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicySetParameterService_List_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/parameters", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicySetParameters.List(context.Background(), "org-1", "policySet-1", nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestPolicySetParameterService_List_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySetParameters.List(context.Background(), "", "policySet-1", nil)
	assertValidationError(t, err, "organizationID")
}

func TestPolicySetParameterService_List_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySetParameters.List(context.Background(), "org-1", "", nil)
	assertValidationError(t, err, "policySetID")
}

func TestPolicySetParameterService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/parameters/policysetparameter-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicySetParameter())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicySetParameters.Get(context.Background(), "org-1", "policySet-1", "policysetparameter-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policysetparameter-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policysetparameter-1")
	}
	if item.Description == nil || *item.Description != "test-description" {
		t.Errorf("Description = %v, want %q", item.Description, "test-description")
	}
}

func TestPolicySetParameterService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/policySet/policySet-1/parameters/policysetparameter-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicySetParameters.Get(context.Background(), "org-1", "policySet-1", "policysetparameter-1")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestPolicySetParameterService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err0 := client.PolicySetParameters.Get(context.Background(), "", "policySet-1", "policysetparameter-1")
	assertValidationError(t, err0, "organizationID")

	_, err1 := client.PolicySetParameters.Get(context.Background(), "org-1", "", "policysetparameter-1")
	assertValidationError(t, err1, "policySetID")

	_, err := client.PolicySetParameters.Get(context.Background(), "org-1", "policySet-1", "")
	assertValidationError(t, err, "policySetParameterID")
}

func TestPolicySetParameterService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet/policySet-1/parameters", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestPolicySetParameter())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicySetParameters.Create(context.Background(), "org-1", "policySet-1", &terrakube.PolicySetParameter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policysetparameter-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policysetparameter-1")
	}
}

func TestPolicySetParameterService_Create_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/policySet/policySet-1/parameters", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "invalid")
	})

	client := newTestClient(t, srv)

	_, err := client.PolicySetParameters.Create(context.Background(), "org-1", "policySet-1", &terrakube.PolicySetParameter{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestPolicySetParameterService_Create_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySetParameters.Create(context.Background(), "", "policySet-1", &terrakube.PolicySetParameter{})
	assertValidationError(t, err, "organizationID")
}

func TestPolicySetParameterService_Create_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySetParameters.Create(context.Background(), "org-1", "", &terrakube.PolicySetParameter{})
	assertValidationError(t, err, "policySetID")
}

func TestPolicySetParameterService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/policySet/policySet-1/parameters/policysetparameter-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestPolicySetParameter())
	})

	client := newTestClient(t, srv)

	item, err := client.PolicySetParameters.Update(context.Background(), "org-1", "policySet-1", newTestPolicySetParameter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "policysetparameter-1" {
		t.Errorf("ID = %q, want %q", item.ID, "policysetparameter-1")
	}
}

func TestPolicySetParameterService_Update_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySetParameters.Update(context.Background(), "org-1", "policySet-1", &terrakube.PolicySetParameter{})
	assertValidationError(t, err, "policySetParameterID")
}

func TestPolicySetParameterService_Update_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySetParameters.Update(context.Background(), "", "policySet-1", newTestPolicySetParameter())
	assertValidationError(t, err, "organizationID")
}

func TestPolicySetParameterService_Update_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.PolicySetParameters.Update(context.Background(), "org-1", "", newTestPolicySetParameter())
	assertValidationError(t, err, "policySetID")
}

func TestPolicySetParameterService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/policySet/policySet-1/parameters/policysetparameter-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	if err := client.PolicySetParameters.Delete(context.Background(), "org-1", "policySet-1", "policysetparameter-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicySetParameterService_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicySetParameters.Delete(context.Background(), "org-1", "policySet-1", "")
	assertValidationError(t, err, "policySetParameterID")
}

func TestPolicySetParameterService_Delete_EmptyOrganizationID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicySetParameters.Delete(context.Background(), "", "policySet-1", "policysetparameter-1")
	assertValidationError(t, err, "organizationID")
}

func TestPolicySetParameterService_Delete_EmptyPolicySetID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.PolicySetParameters.Delete(context.Background(), "org-1", "", "policysetparameter-1")
	assertValidationError(t, err, "policySetID")
}
