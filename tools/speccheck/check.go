package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/tools/speccheck/internal/speccheck"
)

// knownIssue is a difference tracked elsewhere. It is still listed, with the
// reason, but does not fail a check.
type knownIssue struct {
	service string
	kind    speccheck.IssueKind
	path    string // concrete request path the issue is about
	reason  string
}

var knownIssues = []knownIssue{
	// workspace.webhook is to-one; Webhooks.List decodes a collection.
	{"Webhooks", speccheck.PathTypeMismatch, "/organization/id0/workspace/id1/webhook", "a workspace has one webhook; List returns zero or one element"},
}

func matchKnownIssue(service string, i speccheck.Issue) (string, bool) {
	for _, k := range knownIssues {
		if k.service == service && k.kind == i.Kind && strings.HasPrefix(i.Detail, k.path+" ") {
			return k.reason, true
		}
	}
	return "", false
}

// checkReport is the result of comparing the client with one document.
type checkReport struct {
	Version          string                  `json:"version"`
	Models           []speccheck.ModelReport `json:"models"`
	MissingResources []string                `json:"missingResources"`
}

// failing reports whether anything in the report means the client is out of date.
func (r checkReport) failing() bool {
	if len(r.MissingResources) > 0 {
		return true
	}
	return slices.ContainsFunc(r.Models, speccheck.ModelReport.Failing)
}

func (r checkReport) counts() (failing, info int) {
	for _, m := range r.Models {
		for _, i := range m.Issues {
			if i.Kind.Failing() {
				failing++
			} else {
				info++
			}
		}
	}
	return failing + len(r.MissingResources), info
}

func runCheck(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("check", stderr)
	version := fs.String("version", terrakube.APIVersion, "API version of the vendored document to check against")
	root := fs.String("root", "", "repository root (default: found from the working directory)")
	verbose := fs.Bool("verbose", false, "list every issue, including fields kept only in Go")
	asJSON := fs.Bool("json", false, "print the report as JSON")
	failOnMissing := fs.Bool("fail-on-missing", false, "exit 1 if the client is missing anything the document describes")
	if err := fs.Parse(args); err != nil {
		return err
	}

	dir, err := resolveRoot(*root)
	if err != nil {
		return err
	}
	spec, err := speccheck.LoadFile(specFile(dir, *version))
	if err != nil {
		return err
	}
	rep, err := buildCheckReport(spec, *version)
	if err != nil {
		return err
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			return err
		}
	} else {
		printCheckReport(stdout, rep, *verbose)
	}
	if *failOnMissing && rep.failing() {
		return errFailed
	}
	return nil
}

// buildCheckReport validates every CRUD service of the client against spec.
func buildCheckReport(spec *speccheck.Spec, version string) (checkReport, error) {
	rec := &speccheck.Recorder{}
	client, err := terrakube.NewClient(
		terrakube.WithEndpoint("https://speccheck.invalid"),
		terrakube.WithToken("speccheck"),
		terrakube.WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		return checkReport{}, err
	}
	models, err := speccheck.DiscoverModels(client)
	if err != nil {
		return checkReport{}, err
	}

	rep := checkReport{Version: version, Models: []speccheck.ModelReport{}, MissingResources: []string{}}
	covered := map[string]bool{}
	for _, m := range models {
		res, _ := spec.Resource(m.ResourceType)
		mr := speccheck.ValidateModel(m, res)
		paths, err := speccheck.CapturePaths(client, m.ServiceName, rec)
		if err != nil {
			return checkReport{}, err
		}
		mr.Issues = append(mr.Issues, speccheck.ValidatePaths(spec, m, paths)...)
		mr.Issues = append(mr.Issues, speccheck.ValidateFilterKey(m)...)
		for i, issue := range mr.Issues {
			if reason, ok := matchKnownIssue(m.ServiceName, issue); ok {
				mr.Issues[i].Kind = speccheck.KnownIssue
				mr.Issues[i].Detail = strings.TrimSpace(issue.Detail + " — " + reason)
			}
		}
		rep.Models = append(rep.Models, mr)
		covered[m.ResourceType] = true
	}
	for _, typ := range spec.ResourceTypes() {
		if !covered[typ] && len(spec.CollectionPaths(typ)) > 0 {
			rep.MissingResources = append(rep.MissingResources, typ)
		}
	}
	return rep, nil
}

func printCheckReport(w io.Writer, rep checkReport, verbose bool) {
	failing, info := rep.counts()
	p := func(format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }

	p("Terrakube API %s: %d services checked\n", rep.Version, len(rep.Models))
	for _, m := range rep.Models {
		var lines []string
		for _, i := range m.Issues {
			if !verbose && !i.Kind.Failing() {
				continue
			}
			lines = append(lines, "    "+describeIssue(i))
		}
		if len(lines) == 0 {
			if verbose {
				p("  ok    %s (%s)\n", m.Service, m.ResourceType)
			}
			continue
		}
		status := "FAIL"
		if !m.Failing() {
			status = "info"
		}
		p("  %-5s %s (%s)\n%s\n", status, m.Service, m.ResourceType, strings.Join(lines, "\n"))
	}
	for _, typ := range rep.MissingResources {
		p("  FAIL  resource %s has no service in the client\n", typ)
	}
	p("%d failing, %d informational", failing, info)
	if info > 0 && !verbose {
		p(" (use --verbose to list them)")
	}
	p("\n")
}

func describeIssue(i speccheck.Issue) string {
	what := "attribute " + i.Field
	if i.Relation {
		what = "relationship " + i.Field
	}
	switch i.Kind {
	case speccheck.MissingInStruct:
		return fmt.Sprintf("missing in Go: %s (%s)", what, i.SpecType)
	case speccheck.MissingInSchema:
		return fmt.Sprintf("not in spec (kept): %s (%s)", what, i.GoType)
	case speccheck.TypeMismatch:
		return fmt.Sprintf("type mismatch: %s is %s in Go, %s in spec", what, i.GoType, i.SpecType)
	case speccheck.RelationTargetMismatch:
		return fmt.Sprintf("target mismatch: %s points to %s in Go, %s in spec", what, i.GoType, i.SpecType)
	case speccheck.SchemaMissing:
		return "no schema in spec: " + i.Detail
	case speccheck.PathMissing:
		return "path not in spec: " + i.Detail
	case speccheck.PathTypeMismatch:
		return fmt.Sprintf("path %s returns %s in spec, model is %s", i.Detail, i.SpecType, i.GoType)
	case speccheck.KnownIssue:
		return "known: " + i.Detail
	case speccheck.FilterKeyMismatch:
		return fmt.Sprintf("filter key is %s, spec uses %s", i.GoType, i.SpecType)
	}
	return string(i.Kind)
}
