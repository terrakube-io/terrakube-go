package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestStepService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/step", func(w http.ResponseWriter, _ *http.Request) {
		out := "done"
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Step{
			{ID: "step-1", Name: "plan", Status: "completed", StepNumber: 1, Output: &out},
			{ID: "step-2", Name: "apply", Status: "pending", StepNumber: 2},
		})
	})

	client := newTestClient(t, srv)
	steps, err := client.Steps.List(context.Background(), "org-1", "job-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("got %d steps, want 2", len(steps))
	}
	if steps[0].Name != "plan" {
		t.Errorf("Name = %q, want %q", steps[0].Name, "plan")
	}
}

func TestStepService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/step", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[step]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Step{
			{ID: "step-1", Name: "plan"},
		})
	})

	client := newTestClient(t, srv)
	steps, err := client.Steps.List(context.Background(), "org-1", "job-1", &terrakube.ListOptions{Filter: "status==completed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("got %d steps, want 1", len(steps))
	}
}

func TestStepService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Steps.List(context.Background(), "", "job-1", nil)
	if err == nil {
		t.Fatal("expected validation error for empty organization ID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestStepService_List_EmptyJobID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Steps.List(context.Background(), "org-1", "", nil)
	if err == nil {
		t.Fatal("expected validation error for empty job ID")
	}
	assertValidationError(t, err, "job ID")
}

func TestStepService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/step/step-1", func(w http.ResponseWriter, _ *http.Request) {
		out := "Terraform plan output"
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Step{
			ID: "step-1", Name: "plan", Status: "completed", StepNumber: 1, Output: &out,
		})
	})

	client := newTestClient(t, srv)
	step, err := client.Steps.Get(context.Background(), "org-1", "job-1", "step-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step.ID != "step-1" {
		t.Errorf("ID = %q, want %q", step.ID, "step-1")
	}
	if step.Name != "plan" {
		t.Errorf("Name = %q, want %q", step.Name, "plan")
	}
	if step.Status != "completed" {
		t.Errorf("Status = %q, want %q", step.Status, "completed")
	}
	if step.StepNumber != 1 {
		t.Errorf("StepNumber = %d, want %d", step.StepNumber, 1)
	}
	if step.Output == nil || *step.Output != "Terraform plan output" {
		t.Errorf("Output = %v, want %q", step.Output, "Terraform plan output")
	}
}

func TestStepService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		jobID string
		id    string
		field string
	}{
		{"empty org ID", "", "job-1", "step-1", "organization ID"},
		{"empty job ID", "org-1", "", "step-1", "job ID"},
		{"empty step ID", "org-1", "job-1", "", "step ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.Steps.Get(context.Background(), tt.orgID, tt.jobID, tt.id)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestStepService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/step/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "step not found")
	})

	client := newTestClient(t, srv)
	_, err := client.Steps.Get(context.Background(), "org-1", "job-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false")
	}
}

func TestStepService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/job/job-1/step", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Step{
			ID: "step-new", Name: "plan", Status: "pending", StepNumber: 1,
		})
	})

	client := newTestClient(t, srv)
	step, err := client.Steps.Create(context.Background(), "org-1", "job-1", &terrakube.Step{
		Name: "plan", Status: "pending", StepNumber: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step.ID != "step-new" {
		t.Errorf("ID = %q, want %q", step.ID, "step-new")
	}
	if step.Name != "plan" {
		t.Errorf("Name = %q, want %q", step.Name, "plan")
	}
}

func TestStepService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Steps.Create(context.Background(), "", "job-1", &terrakube.Step{})
	if err == nil {
		t.Fatal("expected validation error for empty organization ID")
	}
	assertValidationError(t, err, "organization ID")
}

func TestStepService_Create_EmptyJobID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Steps.Create(context.Background(), "org-1", "", &terrakube.Step{})
	if err == nil {
		t.Fatal("expected validation error for empty job ID")
	}
	assertValidationError(t, err, "job ID")
}

func TestStepService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/job/job-1/step/step-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Step{
			ID: "step-1", Name: "plan", Status: "completed", StepNumber: 1,
		})
	})

	client := newTestClient(t, srv)
	step, err := client.Steps.Update(context.Background(), "org-1", "job-1", &terrakube.Step{
		ID: "step-1", Name: "plan", Status: "completed", StepNumber: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step.Status != "completed" {
		t.Errorf("Status = %q, want %q", step.Status, "completed")
	}
}

func TestStepService_Update_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		jobID string
		step  *terrakube.Step
		field string
	}{
		{"empty org ID", "", "job-1", &terrakube.Step{ID: "step-1"}, "organization ID"},
		{"empty job ID", "org-1", "", &terrakube.Step{ID: "step-1"}, "job ID"},
		{"empty step ID", "org-1", "job-1", &terrakube.Step{ID: ""}, "step ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.Steps.Update(context.Background(), tt.orgID, tt.jobID, tt.step)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestStepService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/job/job-1/step/step-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.Steps.Delete(context.Background(), "org-1", "job-1", "step-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStepService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	tests := []struct {
		name  string
		orgID string
		jobID string
		id    string
		field string
	}{
		{"empty org ID", "", "job-1", "step-1", "organization ID"},
		{"empty job ID", "org-1", "", "step-1", "job ID"},
		{"empty step ID", "org-1", "job-1", "", "step ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.Steps.Delete(context.Background(), tt.orgID, tt.jobID, tt.id)
			if err == nil {
				t.Fatal("expected validation error")
			}
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestStepService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/job/job-1/step/step-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)
	err := client.Steps.Delete(context.Background(), "org-1", "job-1", "step-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestStepService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1/step/step-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.Step{ID: "step-1", Name: "test"})
	})

	client := newTestClient(t, srv)
	_, _ = client.Steps.Get(context.Background(), "org-1", "job-1", "step-1")
}
