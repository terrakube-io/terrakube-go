package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRepo creates a temporary directory that looks like the library root
// (a go.mod declaring libraryModule) and returns it.
func writeRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module "+libraryModule+"\n\ngo 1.27\n")
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil { //nolint:gosec // G703: test helper, path is under t.TempDir
		t.Fatal(err)
	}
}

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRun_NoArgsPrintsUsageAndExits2(t *testing.T) {
	t.Parallel()

	code, stdout, stderr := runCLI(t)
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	for _, want := range []string{"Usage: speccheck", "fetch", "diff", "check", "sync"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("usage missing %q:\n%s", want, stderr)
		}
	}
}

func TestRun_HelpExits0(t *testing.T) {
	t.Parallel()

	for _, arg := range []string{"help", "-h", "--help"} {
		t.Run(arg, func(t *testing.T) {
			t.Parallel()
			code, _, stderr := runCLI(t, arg)
			if code != 0 {
				t.Errorf("exit = %d, want 0", code)
			}
			if !strings.Contains(stderr, "Usage: speccheck") {
				t.Errorf("no usage in stderr: %q", stderr)
			}
		})
	}
}

func TestRun_UnknownCommandExits2(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI(t, "bogus")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, `unknown command "bogus"`) || !strings.Contains(stderr, "Usage:") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestRun_CommandHelpExits0(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI(t, "check", "-h")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stderr, "-fail-on-missing") {
		t.Errorf("flag help missing in stderr: %q", stderr)
	}
}

func TestRun_CommandErrorIsPrintedWithCommandName(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI(t, "diff")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "speccheck diff: --from and --to are required") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestFindRepoRoot(t *testing.T) {
	t.Parallel()

	root := writeRepo(t)
	nested := filepath.Join(root, "tools", "speccheck")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	// A nested module with another name must be skipped.
	writeFile(t, filepath.Join(nested, "go.mod"), "module example.com/other\n")

	got, err := findRepoRoot(nested)
	if err != nil {
		t.Fatalf("findRepoRoot: %v", err)
	}
	want, _ := filepath.EvalSymlinks(root)
	gotResolved, _ := filepath.EvalSymlinks(got)
	if gotResolved != want {
		t.Errorf("findRepoRoot = %q, want %q", got, root)
	}
}

func TestFindRepoRoot_NotFound(t *testing.T) {
	t.Parallel()

	if _, err := findRepoRoot(t.TempDir()); err == nil || !strings.Contains(err.Error(), "--root") {
		t.Errorf("err = %v, want one mentioning --root", err)
	}
}

func TestResolveRoot(t *testing.T) {
	t.Parallel()

	root := writeRepo(t)
	got, err := resolveRoot(root)
	if err != nil || got != root {
		t.Errorf("resolveRoot(%q) = %q, %v", root, got, err)
	}
}

func TestResolveRoot_FromWorkingDirectory(t *testing.T) {
	t.Parallel()

	// Tests run in tools/speccheck, below the repository root.
	got, err := resolveRoot("")
	if err != nil {
		t.Fatalf("resolveRoot: %v", err)
	}
	if _, err := os.Stat(filepath.Join(got, "client.go")); err != nil {
		t.Errorf("resolved root %q has no client.go: %v", got, err)
	}
}

func TestModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "module example.com/m\n\ngo 1.27\n", "example.com/m"},
		{"leading comment and spaces", "// c\n  module   example.com/m  \n", "example.com/m"},
		{"missing", "go 1.27\n", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := modulePath(tt.in); got != tt.want {
				t.Errorf("modulePath = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSpecFile(t *testing.T) {
	t.Parallel()

	want := filepath.Join("root", "openapi", "v2.34.1.json")
	for _, version := range []string{"2.34.1", "v2.34.1"} {
		if got := specFile("root", version); got != want {
			t.Errorf("specFile(%q) = %q, want %q", version, got, want)
		}
	}
}
