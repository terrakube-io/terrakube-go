package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

// specChange is one difference between two documents, as oasdiff reports it.
type specChange struct {
	ID        string `json:"id"`
	Level     string `json:"level"` // ERR, WARN, or INFO
	Operation string `json:"operation,omitempty"`
	Path      string `json:"path,omitempty"`
	Text      string `json:"text"`
}

type diffReport struct {
	From    string         `json:"from"`
	To      string         `json:"to"`
	Counts  map[string]int `json:"counts"`
	Semver  string         `json:"semver"` // MAJOR, MINOR, or PATCH
	Reason  string         `json:"reason"`
	Changes []specChange   `json:"changes"`
}

func runDiff(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("diff", stderr)
	from := fs.String("from", "", "base API version (required), e.g. 2.34.1")
	to := fs.String("to", "", "revised API version (required)")
	format := fs.String("format", "text", "output format: text or json")
	breakingOnly := fs.Bool("breaking-only", false, "list only breaking (ERR) changes")
	failOnBreaking := fs.Bool("fail-on-breaking", false, "exit 1 if there are breaking changes")
	root := fs.String("root", "", "repository root (default: found from the working directory)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *from == "" || *to == "" {
		return errors.New("--from and --to are required")
	}
	if *format != "text" && *format != "json" {
		return fmt.Errorf("--format must be text or json, got %q", *format)
	}
	dir, err := resolveRoot(*root)
	if err != nil {
		return err
	}

	rep, err := diffSpecs(specFile(dir, *from), specFile(dir, *to))
	if err != nil {
		return err
	}
	rep.From, rep.To = *from, *to
	if *breakingOnly {
		rep.Changes = slices.DeleteFunc(rep.Changes, func(c specChange) bool { return c.Level != "ERR" })
	}

	if *format == "json" {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			return err
		}
	} else {
		printDiffReport(stdout, rep)
	}
	if *failOnBreaking && rep.Counts["ERR"] > 0 {
		return errFailed
	}
	return nil
}

// diffSpecs compares two document files with oasdiff and classifies every
// change as ERR (breaking), WARN (possibly breaking), or INFO.
func diffSpecs(fromPath, toPath string) (diffReport, error) {
	loader := openapi3.NewLoader()
	base, err := load.NewSpecInfo(loader, load.NewSource(fromPath))
	if err != nil {
		return diffReport{}, fmt.Errorf("loading %s: %w", fromPath, err)
	}
	revision, err := load.NewSpecInfo(loader, load.NewSource(toPath))
	if err != nil {
		return diffReport{}, fmt.Errorf("loading %s: %w", toPath, err)
	}
	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	if err != nil {
		return diffReport{}, fmt.Errorf("comparing documents: %w", err)
	}
	changes := checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, sources, checker.INFO)
	return buildDiffReport(changes), nil
}

func buildDiffReport(changes checker.Changes) diffReport {
	l := checker.NewDefaultLocalizer()
	rep := diffReport{Counts: map[string]int{"ERR": 0, "WARN": 0, "INFO": 0}, Changes: []specChange{}}
	for _, c := range changes {
		level := levelName(c.GetLevel())
		rep.Counts[level]++
		rep.Changes = append(rep.Changes, specChange{
			ID:        c.GetId(),
			Level:     level,
			Operation: c.GetOperation(),
			Path:      c.GetPath(),
			Text:      c.GetUncolorizedText(l),
		})
	}
	order := map[string]int{"ERR": 0, "WARN": 1, "INFO": 2}
	slices.SortStableFunc(rep.Changes, func(a, b specChange) int {
		return cmp.Or(cmp.Compare(order[a.Level], order[b.Level]), cmp.Compare(a.Path, b.Path), cmp.Compare(a.Operation, b.Operation))
	})
	rep.Semver, rep.Reason = semverGuidance(rep.Counts)
	return rep
}

func levelName(l checker.Level) string {
	switch l {
	case checker.ERR:
		return "ERR"
	case checker.WARN:
		return "WARN"
	default:
		return "INFO"
	}
}

// semverGuidance maps change levels to the release the client needs: a
// breaking API change can break callers (MAJOR), any other change adds or
// relaxes something the client may expose (MINOR), and no change needs at
// most a PATCH.
func semverGuidance(counts map[string]int) (level, reason string) {
	switch {
	case counts["ERR"] > 0:
		return "MAJOR", fmt.Sprintf("%d breaking change(s)", counts["ERR"])
	case counts["WARN"] > 0 || counts["INFO"] > 0:
		return "MINOR", fmt.Sprintf("%d possibly breaking and %d non-breaking change(s)", counts["WARN"], counts["INFO"])
	default:
		return "PATCH", "no API changes"
	}
}

func printDiffReport(w io.Writer, rep diffReport) {
	p := func(format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
	p("Terrakube API %s -> %s\n", rep.From, rep.To)
	for _, c := range rep.Changes {
		where := c.Path
		if c.Operation != "" {
			where = c.Operation + " " + c.Path
		}
		p("  %-4s %s: %s\n", c.Level, where, c.Text)
	}
	p("%d breaking (ERR), %d possibly breaking (WARN), %d non-breaking (INFO)\n", rep.Counts["ERR"], rep.Counts["WARN"], rep.Counts["INFO"])
	p("semver guidance: %s (%s)\n", rep.Semver, rep.Reason)
}
