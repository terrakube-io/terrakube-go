package terrakube_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTestJob() *terrakube.Job {
	return &terrakube.Job{
		ID:     "job-1",
		Output: "Apply complete!",
		Status: "completed",
	}
}

func TestJobService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job", func(w http.ResponseWriter, _ *http.Request) {
		jobs := []*terrakube.Job{newTestJob()}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, jobs)
	})

	client := newTestClient(t, srv)

	jobs, err := client.Jobs.List(context.Background(), "org-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1", len(jobs))
	}
	if jobs[0].Output != "Apply complete!" {
		t.Errorf("Output = %q, want %q", jobs[0].Output, "Apply complete!")
	}
}

func TestJobService_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter[job]")
		if filter == "" {
			t.Error("expected filter query parameter")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.Job{})
	})

	client := newTestClient(t, srv)

	_, err := client.Jobs.List(context.Background(), "org-1", &terrakube.ListOptions{Filter: "status==completed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJobService_List_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.Jobs.List(context.Background(), "", nil)
	assertValidationError(t, err, "organizationID")
}

func TestJobService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestJob())
	})

	client := newTestClient(t, srv)

	job, err := client.Jobs.Get(context.Background(), "org-1", "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("ID = %q, want %q", job.ID, "job-1")
	}
	if job.Output != "Apply complete!" {
		t.Errorf("Output = %q, want %q", job.Output, "Apply complete!")
	}
}

func TestJobService_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/missing", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "job not found")
	})

	client := newTestClient(t, srv)

	_, err := client.Jobs.Get(context.Background(), "org-1", "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestJobService_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		jobID string
		field string
	}{
		{"empty org ID", "", "job-1", "organizationID"},
		{"empty job ID", "org-1", "", "jobID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.Jobs.Get(context.Background(), tt.orgID, tt.jobID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestJobService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/job", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTestJob())
	})

	client := newTestClient(t, srv)

	input := &terrakube.Job{
		Status: "pending",
	}
	job, err := client.Jobs.Create(context.Background(), "org-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("ID = %q, want %q", job.ID, "job-1")
	}
}

func TestJobService_Create_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.Jobs.Create(context.Background(), "", &terrakube.Job{})
	assertValidationError(t, err, "organizationID")
}

func TestJobService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1/job/job-1", func(w http.ResponseWriter, _ *http.Request) {
		job := newTestJob()
		job.Status = "cancelled"
		testutil.WriteJSONAPI(t, w, http.StatusOK, job)
	})

	client := newTestClient(t, srv)

	input := &terrakube.Job{
		ID:     "job-1",
		Status: "cancelled",
	}
	job, err := client.Jobs.Update(context.Background(), "org-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", job.Status, "cancelled")
	}
}

func TestJobService_Update_EmptyOrgID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.Jobs.Update(context.Background(), "", &terrakube.Job{ID: "job-1"})
	assertValidationError(t, err, "organizationID")
}

func TestJobService_Update_EmptyJobID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.Jobs.Update(context.Background(), "org-1", &terrakube.Job{ID: ""})
	assertValidationError(t, err, "jobID")
}

func TestJobService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/job/job-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	err := client.Jobs.Delete(context.Background(), "org-1", "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJobService_Delete_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	tests := []struct {
		name  string
		orgID string
		jobID string
		field string
	}{
		{"empty org ID", "", "job-1", "organizationID"},
		{"empty job ID", "org-1", "", "jobID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := client.Jobs.Delete(context.Background(), tt.orgID, tt.jobID)
			assertValidationError(t, err, tt.field)
		})
	}
}

func TestJobService_Delete_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1/job/job-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "server error")
	})

	client := newTestClient(t, srv)

	err := client.Jobs.Delete(context.Background(), "org-1", "job-1")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestJobService_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization/org-1/job/job-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTestJob())
	})

	client := newTestClient(t, srv)

	_, err := client.Jobs.Get(context.Background(), "org-1", "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJobService_TargetAndReplaceAddrs(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/job", func(w http.ResponseWriter, _ *http.Request) {
		job := newTestJob()
		job.TargetAddrs = []string{"module.vpc.aws_subnet.public[0]"}
		job.ReplaceAddrs = []string{"aws_instance.web"}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, job)
	})

	client := newTestClient(t, srv)
	created, err := client.Jobs.Create(context.Background(), "org-1", &terrakube.Job{
		Status:       "pending",
		TargetAddrs:  []string{"module.vpc.aws_subnet.public[0]"},
		ReplaceAddrs: []string{"aws_instance.web"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created.TargetAddrs) != 1 || created.TargetAddrs[0] != "module.vpc.aws_subnet.public[0]" {
		t.Errorf("TargetAddrs = %v, want [module.vpc.aws_subnet.public[0]]", created.TargetAddrs)
	}
	if len(created.ReplaceAddrs) != 1 || created.ReplaceAddrs[0] != "aws_instance.web" {
		t.Errorf("ReplaceAddrs = %v, want [aws_instance.web]", created.ReplaceAddrs)
	}
}

// TestJobService_Create_OmitsServerOwnedAttributes: the server answers 403 to
// a create request that mentions an attribute only it may set, even as null,
// and 400 to an empty status.
func TestJobService_Create_OmitsServerOwnedAttributes(t *testing.T) {
	t.Parallel()

	var sent map[string]any
	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization/org-1/job", func(w http.ResponseWriter, r *http.Request) {
		var doc struct {
			Data struct {
				Attributes map[string]any `json:"attributes"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		sent = doc.Data.Attributes
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.Job{ID: "1", Status: "pending"})
	})

	tpl := "tpl-1"
	_, err := newTestClient(t, srv).Jobs.Create(context.Background(), "org-1", &terrakube.Job{TemplateReference: &tpl})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	for _, attr := range []string{
		"status", "approvalTeam", "approvedAt", "approvedBy", "cascadeDepth", "dispatchFailureCount",
		"dispatchFirstFailureAt", "dispatchLastAttemptAt", "dispatchLastError", "dispatchNextRetryAt",
		"encryptedVariables", "overrideSource", "resolvedTerraformVersion", "terraformPlanArtifacts",
		"terraformPlanArtifactsChecksum", "triggeredByJobId", "variableNames", "variablesAsArguments",
		"prNumber", "prCommentError",
	} {
		if _, ok := sent[attr]; ok {
			t.Errorf("create request mentions %q, which only the server may set", attr)
		}
	}
	if sent["templateReference"] != "tpl-1" {
		t.Errorf("templateReference = %v, want tpl-1", sent["templateReference"])
	}
	// Booleans a caller controls are still always sent.
	for _, attr := range []string{"refresh", "planChanges", "refreshOnly"} {
		if _, ok := sent[attr]; !ok {
			t.Errorf("create request does not mention %q", attr)
		}
	}
}
