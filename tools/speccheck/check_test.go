package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/tools/speccheck/internal/speccheck"
)

// realRoot returns the repository root and skips the test when the vendored
// document for the client's API version is not there.
func realRoot(t *testing.T) string {
	t.Helper()
	root, err := resolveRoot("")
	if err != nil {
		t.Fatalf("resolveRoot: %v", err)
	}
	if _, err := os.Stat(specFile(root, terrakube.APIVersion)); err != nil {
		t.Skipf("no vendored document for %s: %v", terrakube.APIVersion, err)
	}
	return root
}

func TestRunCheck_FailOnMissingPassesForVendoredVersion(t *testing.T) {
	t.Parallel()

	root := realRoot(t)
	var stdout, stderr strings.Builder
	if err := runCheck([]string{"--fail-on-missing", "--root", root}, &stdout, &stderr); err != nil {
		t.Fatalf("runCheck: %v\n%s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "0 failing") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestRunCheck_JSONReportHasNoFailingIssues(t *testing.T) {
	t.Parallel()

	root := realRoot(t)
	var stdout, stderr strings.Builder
	if err := runCheck([]string{"--json", "--root", root}, &stdout, &stderr); err != nil {
		t.Fatalf("runCheck: %v", err)
	}
	var rep checkReport
	if err := json.Unmarshal([]byte(stdout.String()), &rep); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if rep.Version != terrakube.APIVersion {
		t.Errorf("Version = %q, want %q", rep.Version, terrakube.APIVersion)
	}
	if len(rep.MissingResources) != 0 {
		t.Errorf("MissingResources = %v", rep.MissingResources)
	}
	if len(rep.Models) != 42 {
		t.Errorf("got %d models, want 42", len(rep.Models))
	}
	for _, m := range rep.Models {
		for _, i := range m.Issues {
			if i.Kind.Failing() {
				t.Errorf("%s: failing issue %+v", m.Service, i)
			}
		}
	}
}

func TestRunCheck_VerboseListsServices(t *testing.T) {
	t.Parallel()

	root := realRoot(t)
	var stdout, stderr strings.Builder
	if err := runCheck([]string{"--verbose", "--root", root}, &stdout, &stderr); err != nil {
		t.Fatalf("runCheck: %v", err)
	}
	out := stdout.String()
	for _, want := range []string{"ok    Organizations (organization)", "Webhooks (webhook)", "known: "} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
	if strings.Contains(out, "use --verbose") {
		t.Error("verbose output should not suggest --verbose")
	}
}

func TestRunCheck_QuietHidesInformationalIssues(t *testing.T) {
	t.Parallel()

	root := realRoot(t)
	var stdout, stderr strings.Builder
	if err := runCheck([]string{"--root", root}, &stdout, &stderr); err != nil {
		t.Fatalf("runCheck: %v", err)
	}
	if !strings.Contains(stdout.String(), "(use --verbose to list them)") || strings.Contains(stdout.String(), "ok    ") {
		t.Errorf("output = %q", stdout.String())
	}
}

func TestBuildCheckReport_WebhooksPathMismatchIsKnownIssue(t *testing.T) {
	t.Parallel()

	root := realRoot(t)
	spec, err := speccheck.LoadFile(specFile(root, terrakube.APIVersion))
	if err != nil {
		t.Fatal(err)
	}
	rep, err := buildCheckReport(spec, terrakube.APIVersion)
	if err != nil {
		t.Fatalf("buildCheckReport: %v", err)
	}
	var found bool
	for _, m := range rep.Models {
		if m.Service != "Webhooks" {
			continue
		}
		for _, i := range m.Issues {
			if i.Kind == speccheck.PathTypeMismatch {
				t.Errorf("PathTypeMismatch not converted: %+v", i)
			}
			if i.Kind == speccheck.KnownIssue {
				found = true
				if !strings.Contains(i.Detail, "terrakube-go#38") {
					t.Errorf("Detail = %q, want the tracking reference", i.Detail)
				}
			}
		}
		if m.Failing() {
			t.Error("Webhooks report is failing, want non-failing known issue")
		}
	}
	if !found {
		t.Error("no KnownIssue reported for Webhooks")
	}
	if rep.failing() {
		t.Error("report is failing")
	}
	if failing, info := rep.counts(); failing != 0 || info == 0 {
		t.Errorf("counts = %d failing, %d info", failing, info)
	}
}

// tinySpec has one resource that the client does not cover and none of the
// client's resources, so a check against it fails everywhere.
const tinySpec = `{"openapi":"3.0.1","info":{"title":"t","version":"1"},
"paths":{"/gadget":{"get":{"responses":{"200":{"description":"ok","content":{"application/vnd.api+json":{"schema":{"type":"object","properties":{"data":{"type":"array","items":{"$ref":"#/components/schemas/gadget"}}}}}}}}}}},
"components":{"schemas":{"gadget":{"type":"object","properties":{"attributes":{"type":"object","properties":{"name":{"type":"string"}}}}}}}}`

func TestRunCheck_FailOnMissingFailsForIncompleteClient(t *testing.T) {
	t.Parallel()

	root := writeRepo(t)
	writeFile(t, filepath.Join(root, "openapi", "v9.9.9.json"), tinySpec)
	code, stdout, stderr := runCLI(t, "check", "--version", "9.9.9", "--root", root, "--fail-on-missing")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	for _, want := range []string{"FAIL  resource gadget has no service in the client", "no schema in spec", "path not in spec"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
	// Without the flag the same report does not fail the command.
	if code, _, _ := runCLI(t, "check", "--version", "9.9.9", "--root", root); code != 0 {
		t.Errorf("exit without --fail-on-missing = %d, want 0", code)
	}
}

func TestRunCheck_Errors(t *testing.T) {
	t.Parallel()

	root := writeRepo(t)
	if code, _, stderr := runCLI(t, "check", "--version", "0.0.1", "--root", root); code != 1 || !strings.Contains(stderr, "loading OpenAPI document") {
		t.Errorf("missing document: exit = %d, stderr = %q", code, stderr)
	}
	if code, _, _ := runCLI(t, "check", "--nope"); code != 1 {
		t.Errorf("bad flag: exit = %d, want 1", code)
	}
}

func TestDescribeIssue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		issue speccheck.Issue
		want  string
	}{
		{speccheck.Issue{Kind: speccheck.MissingInStruct, Field: "a", SpecType: "string"}, "missing in Go: attribute a (string)"},
		{speccheck.Issue{Kind: speccheck.MissingInStruct, Field: "r", Relation: true, SpecType: "x"}, "missing in Go: relationship r (x)"},
		{speccheck.Issue{Kind: speccheck.MissingInSchema, Field: "a", GoType: "*string"}, "not in spec (kept): attribute a (*string)"},
		{speccheck.Issue{Kind: speccheck.TypeMismatch, Field: "a", GoType: "bool", SpecType: "string"}, "type mismatch: attribute a is bool in Go, string in spec"},
		{speccheck.Issue{Kind: speccheck.RelationTargetMismatch, Field: "r", Relation: true, GoType: "x", SpecType: "y"}, "target mismatch: relationship r points to x in Go, y in spec"},
		{speccheck.Issue{Kind: speccheck.SchemaMissing, Detail: "no schema for type t"}, "no schema in spec: no schema for type t"},
		{speccheck.Issue{Kind: speccheck.PathMissing, Detail: "/p"}, "path not in spec: /p"},
		{speccheck.Issue{Kind: speccheck.PathTypeMismatch, Detail: "/p", SpecType: "a", GoType: "b"}, "path /p returns a in spec, model is b"},
		{speccheck.Issue{Kind: speccheck.KnownIssue, Detail: "why"}, "known: why"},
		{speccheck.Issue{Kind: speccheck.FilterKeyMismatch, GoType: "f", SpecType: "g"}, "filter key is f, spec uses g"},
		{speccheck.Issue{Kind: "something_else"}, "something_else"},
	}
	for _, tt := range tests {
		t.Run(string(tt.issue.Kind), func(t *testing.T) {
			t.Parallel()
			if got := describeIssue(tt.issue); got != tt.want {
				t.Errorf("describeIssue = %q, want %q", got, tt.want)
			}
		})
	}
}
