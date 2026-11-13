package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func specWithPaths(paths ...string) string {
	ops := make([]string, 0, len(paths))
	for _, p := range paths {
		ops = append(ops, `"`+p+`":{"get":{"operationId":"get`+strings.ReplaceAll(p, "/", "_")+`","responses":{"200":{"description":"ok"}}}}`)
	}
	return `{"openapi":"3.0.1","info":{"title":"t","version":"1"},"paths":{` + strings.Join(ops, ",") + `}}`
}

// diffRepo vendors v1.0.0 (paths /a, /b), v1.1.0 (/a, /c) and v1.0.1 (a copy of v1.0.0).
func diffRepo(t *testing.T) string {
	t.Helper()
	root := writeRepo(t)
	writeFile(t, filepath.Join(root, "openapi", "v1.0.0.json"), specWithPaths("/a", "/b"))
	writeFile(t, filepath.Join(root, "openapi", "v1.0.1.json"), specWithPaths("/a", "/b"))
	writeFile(t, filepath.Join(root, "openapi", "v1.1.0.json"), specWithPaths("/a", "/c"))
	return root
}

func TestRunDiff_TextOutput(t *testing.T) {
	t.Parallel()

	root := diffRepo(t)
	code, stdout, stderr := runCLI(t, "diff", "--from", "1.0.0", "--to", "1.1.0", "--root", root)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr)
	}
	for _, want := range []string{"Terrakube API 1.0.0 -> 1.1.0", "ERR ", "/b", "INFO", "/c", "semver guidance: MAJOR (1 breaking change(s))"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if strings.Index(stdout, "ERR ") > strings.Index(stdout, "INFO") {
		t.Errorf("breaking changes should be listed first:\n%s", stdout)
	}
}

func TestRunDiff_JSONOutput(t *testing.T) {
	t.Parallel()

	root := diffRepo(t)
	code, stdout, stderr := runCLI(t, "diff", "--from", "v1.0.0", "--to", "1.1.0", "--format=json", "--root", root)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr)
	}
	var rep diffReport
	if err := json.Unmarshal([]byte(stdout), &rep); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, stdout)
	}
	if rep.Semver != "MAJOR" || rep.Counts["ERR"] != 1 || rep.Counts["INFO"] < 1 {
		t.Errorf("report = %+v", rep)
	}
	if rep.From != "v1.0.0" || rep.To != "1.1.0" || len(rep.Changes) != rep.Counts["ERR"]+rep.Counts["WARN"]+rep.Counts["INFO"] {
		t.Errorf("report = %+v", rep)
	}
}

func TestRunDiff_BreakingOnlyListsOnlyErrors(t *testing.T) {
	t.Parallel()

	root := diffRepo(t)
	_, stdout, _ := runCLI(t, "diff", "--from", "1.0.0", "--to", "1.1.0", "--format=json", "--breaking-only", "--root", root)
	var rep diffReport
	if err := json.Unmarshal([]byte(stdout), &rep); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if len(rep.Changes) == 0 {
		t.Fatal("no changes listed")
	}
	for _, c := range rep.Changes {
		if c.Level != "ERR" {
			t.Errorf("change %+v is not breaking", c)
		}
	}
	if rep.Counts["INFO"] < 1 {
		t.Errorf("counts should still include non-breaking changes: %+v", rep.Counts)
	}
}

func TestRunDiff_FailOnBreakingExits1WithoutPrintingError(t *testing.T) {
	t.Parallel()

	root := diffRepo(t)
	code, stdout, stderr := runCLI(t, "diff", "--from", "1.0.0", "--to", "1.1.0", "--fail-on-breaking", "--root", root)
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty (errFailed is not printed)", stderr)
	}
	if !strings.Contains(stdout, "semver guidance") {
		t.Errorf("report not printed:\n%s", stdout)
	}
}

func TestRunDiff_FailOnBreakingPassesWithoutBreakingChanges(t *testing.T) {
	t.Parallel()

	root := diffRepo(t)
	if code, _, stderr := runCLI(t, "diff", "--from", "1.0.0", "--to", "1.0.1", "--fail-on-breaking", "--root", root); code != 0 {
		t.Errorf("exit = %d, stderr = %q", code, stderr)
	}
}

func TestRunDiff_SameContentIsPatch(t *testing.T) {
	t.Parallel()

	root := diffRepo(t)
	_, stdout, _ := runCLI(t, "diff", "--from", "1.0.0", "--to", "1.0.1", "--format=json", "--root", root)
	var rep diffReport
	if err := json.Unmarshal([]byte(stdout), &rep); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if rep.Semver != "PATCH" || len(rep.Changes) != 0 || rep.Reason != "no API changes" {
		t.Errorf("report = %+v", rep)
	}
}

func TestRunDiff_Errors(t *testing.T) {
	t.Parallel()

	root := diffRepo(t)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing from", []string{"--to", "1.1.0"}, "--from and --to are required"},
		{"missing to", []string{"--from", "1.0.0"}, "--from and --to are required"},
		{"bad format", []string{"--from", "1.0.0", "--to", "1.1.0", "--format", "xml", "--root", root}, `--format must be text or json, got "xml"`},
		{"missing document", []string{"--from", "1.0.0", "--to", "7.7.7", "--root", root}, "loading"},
		{"no repository", []string{"--from", "1.0.0", "--to", "1.1.0", "--root", filepath.Join(root, "nowhere")}, "loading"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			code, _, stderr := runCLI(t, append([]string{"diff"}, tt.args...)...)
			if code != 1 || !strings.Contains(stderr, tt.want) {
				t.Errorf("exit = %d, stderr = %q, want 1 containing %q", code, stderr, tt.want)
			}
		})
	}
}

func TestSemverGuidance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		counts     map[string]int
		wantLevel  string
		wantReason string
	}{
		{"breaking", map[string]int{"ERR": 2, "WARN": 1, "INFO": 3}, "MAJOR", "2 breaking change(s)"},
		{"warnings only", map[string]int{"WARN": 1}, "MINOR", "1 possibly breaking and 0 non-breaking change(s)"},
		{"info only", map[string]int{"INFO": 4}, "MINOR", "0 possibly breaking and 4 non-breaking change(s)"},
		{"nothing", map[string]int{}, "PATCH", "no API changes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			level, reason := semverGuidance(tt.counts)
			if level != tt.wantLevel || reason != tt.wantReason {
				t.Errorf("semverGuidance = %q, %q; want %q, %q", level, reason, tt.wantLevel, tt.wantReason)
			}
		})
	}
}
