package speccheck

import (
	"net/http"
	"os"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
)

const realSpecPath = "../../../../openapi/v2.34.1.json"

func recordingClient(t *testing.T) (*terrakube.Client, *Recorder) {
	t.Helper()
	rec := &Recorder{}
	client, err := terrakube.NewClient(
		terrakube.WithEndpoint("https://speccheck.invalid"),
		terrakube.WithToken("t"),
		terrakube.WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, rec
}

func TestCapturePaths_RealClient(t *testing.T) {
	t.Parallel()

	spec, err := LoadFile(realSpecPath)
	if err != nil {
		t.Skipf("real document not available: %v", err)
	}
	client, rec := recordingClient(t)
	models, err := DiscoverModels(client)
	if err != nil {
		t.Fatalf("DiscoverModels: %v", err)
	}
	byName := map[string]ModelInfo{}
	for _, m := range models {
		byName[m.ServiceName] = m
	}

	tests := []struct {
		name           string
		service        string
		wantCollection string
		wantItem       string
	}{
		{"nested org-scoped", "Workspaces", "/organization/id0/workspace", "/organization/id0/workspace/id1"},
		{"flat root", "Actions", "/action", "/action/id0"},
		{"no org", "WorkspaceSchedules", "/workspace/id0/schedule", "/workspace/id0/schedule/id1"},
		{"deep", "Implementations", "/organization/id0/provider/id1/version/id2/implementation", "/organization/id0/provider/id1/version/id2/implementation/id3"},
		{"mixed", "CollectionReferences", "/organization/id0/collection/id1/reference", "/reference/id0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp, err := CapturePaths(client, tt.service, rec)
			if err != nil {
				t.Fatalf("CapturePaths: %v", err)
			}
			if sp.Collection != tt.wantCollection || sp.Item != tt.wantItem {
				t.Errorf("paths = %+v, want collection %q item %q", sp, tt.wantCollection, tt.wantItem)
			}
			if issues := ValidatePaths(spec, byName[tt.service], sp); len(issues) != 0 {
				t.Errorf("ValidatePaths issues = %+v, want none", issues)
			}
		})
	}
}

func TestValidatePaths_RealDocumentFindings(t *testing.T) {
	t.Parallel()

	spec, err := LoadFile(realSpecPath)
	if err != nil {
		t.Skipf("real document not available: %v", err)
	}
	client, rec := recordingClient(t)
	models, err := DiscoverModels(client)
	if err != nil {
		t.Fatalf("DiscoverModels: %v", err)
	}
	var webhooks ModelInfo
	for _, m := range models {
		if m.ServiceName == "Webhooks" {
			webhooks = m
		}
	}

	t.Run("to-one relationship listed as a collection is a type mismatch", func(t *testing.T) {
		sp, err := CapturePaths(client, "Webhooks", rec)
		if err != nil {
			t.Fatalf("CapturePaths: %v", err)
		}
		issues := ValidatePaths(spec, webhooks, sp)
		if len(issues) != 1 || issues[0].Kind != PathTypeMismatch {
			t.Fatalf("issues = %+v, want one PathTypeMismatch", issues)
		}
		if issues[0].GoType != "collection of webhook" || issues[0].SpecType != "webhook" {
			t.Errorf("issue = %+v", issues[0])
		}
	})

	t.Run("made-up path is missing", func(t *testing.T) {
		issues := ValidatePaths(spec, webhooks, ServicePaths{Collection: "/nothing/id0/at/all", Item: "/nothing/id0/at/all/id1"})
		if len(issues) != 2 || issues[0].Kind != PathMissing || issues[1].Kind != PathMissing {
			t.Fatalf("issues = %+v, want two PathMissing", issues)
		}
		if issues[0].Detail != "/nothing/id0/at/all" {
			t.Errorf("Detail = %q", issues[0].Detail)
		}
	})
}

func TestCapturePaths_Errors(t *testing.T) {
	t.Parallel()

	client, rec := recordingClient(t)
	if _, err := CapturePaths(client, "NoSuchService", rec); err == nil {
		t.Error("CapturePaths(unknown service) returned nil error")
	}
	// Operations has no List method.
	if _, err := CapturePaths(client, "Operations", rec); err == nil {
		t.Error("CapturePaths(service without List) returned nil error")
	}
}

func TestRecorder_AnswersNotFoundAndRecordsPath(t *testing.T) {
	t.Parallel()

	rec := &Recorder{}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://speccheck.invalid/api/v1/x", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := rec.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	if got := rec.take(); len(got) != 1 || got[0] != "/api/v1/x" {
		t.Errorf("recorded = %v", got)
	}
	if got := rec.take(); got != nil {
		t.Errorf("second take = %v, want nil", got)
	}
}

// Elide documents the nested GET of a to-one relationship as an array; the
// relationship endpoint is authoritative. policy_evaluation.override is
// @OneToOne, so a List on it would fail against a real server.
func TestResolvePath_DocumentedToOneIsNotACollection(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(realSpecPath); err != nil {
		t.Skipf("vendored document not available: %v", err)
	}
	spec, err := LoadFile(realSpecPath)
	if err != nil {
		t.Fatal(err)
	}
	typ, collection, via, ok := spec.ResolvePath("/organization/id0/job/id1/policyEvaluation/id2/override")
	if !ok || typ != "policy_override" || collection || via != "path" {
		t.Errorf("ResolvePath = %q, collection %v, via %q, ok %v; want policy_override, false, path, true", typ, collection, via, ok)
	}
	typ, collection, _, ok = spec.ResolvePath("/organization/id0/policySet")
	if !ok || typ != "policy_set" || !collection {
		t.Errorf("ResolvePath(policySet) = %q, collection %v, ok %v; want policy_set, true, true", typ, collection, ok)
	}
}
