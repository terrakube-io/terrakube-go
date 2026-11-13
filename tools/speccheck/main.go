// Command speccheck keeps the Terrakube Go client in step with the Terrakube
// OpenAPI document.
//
// Usage:
//
//	speccheck fetch [flags]   download an OpenAPI document into openapi/
//	speccheck diff  [flags]   compare two vendored documents, with semver guidance
//	speccheck check [flags]   compare the client's models and paths with a document
//	speccheck sync  [flags]   update the client's source to match a document
//
// Run "speccheck <command> -h" for the flags of a command.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// libraryModule is the module path of the client library.
const libraryModule = "github.com/terrakube-io/terrakube-go"

// errFailed signals that a command ran but its result is a failure (for
// example breaking changes with --fail-on-breaking). main exits 1 without
// printing it again.
var errFailed = errors.New("check failed")

type command struct {
	name    string
	summary string
	run     func(args []string, stdout, stderr io.Writer) error
}

func commands() []command {
	return []command{
		{"fetch", "download an OpenAPI document into openapi/", runFetch},
		{"diff", "compare two vendored documents, with semver guidance", runDiff},
		{"check", "compare the client's models and paths with a document", runCheck},
		{"sync", "update the client's source to match a document", runSync},
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		usage(stderr)
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	for _, c := range commands() {
		if c.name != args[0] {
			continue
		}
		err := c.run(args[1:], stdout, stderr)
		switch {
		case err == nil:
			return 0
		case errors.Is(err, flag.ErrHelp):
			return 0
		case errors.Is(err, errFailed):
			return 1
		default:
			_, _ = fmt.Fprintf(stderr, "speccheck %s: %v\n", c.name, err)
			return 1
		}
	}
	_, _ = fmt.Fprintf(stderr, "speccheck: unknown command %q\n\n", args[0])
	usage(stderr)
	return 2
}

func usage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: speccheck <command> [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Commands:")
	for _, c := range commands() {
		_, _ = fmt.Fprintf(w, "  %-6s %s\n", c.name, c.summary)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, `Run "speccheck <command> -h" for the flags of a command.`)
}

// newFlagSet returns a flag set that reports errors instead of exiting.
func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("speccheck "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

// findRepoRoot walks up from dir to the directory whose go.mod declares the
// client library module.
func findRepoRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && modulePath(string(data)) == libraryModule {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod for %s found above the current directory; use --root", libraryModule)
		}
		dir = parent
	}
}

func modulePath(gomod string) string {
	for line := range strings.Lines(gomod) {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// resolveRoot returns root if set, otherwise the repository root above the
// working directory.
func resolveRoot(root string) (string, error) {
	if root != "" {
		return filepath.Abs(root)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return findRepoRoot(wd)
}

// specFile is the vendored document for a version, e.g. openapi/v2.34.1.json.
func specFile(root, version string) string {
	return filepath.Join(root, "openapi", "v"+strings.TrimPrefix(version, "v")+".json")
}
