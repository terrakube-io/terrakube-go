package main

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/tools/speccheck/internal/speccheck"
)

// copyLibrary copies the library's non-test Go files and crud_test.go from
// the repository root into dst.
func copyLibrary(t *testing.T, root, dst string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		base := filepath.Base(f)
		if strings.HasSuffix(base, "_test.go") && base != "crud_test.go" {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dst, base), string(data))
	}
}

func readDir(t *testing.T, dir string) map[string]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		out[filepath.Base(f)] = string(data)
	}
	return out
}

func mustContain(t *testing.T, file, content string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(content, want) {
			t.Errorf("%s does not contain %q", file, want)
		}
	}
}

const gadgetSpec = `{"openapi":"3.0.1","info":{"title":"t","version":"1"},
"paths":{
 "/organization/{organizationId}/gadget":{"get":{"responses":{"200":{"description":"ok","content":{"application/vnd.api+json":{"schema":{"type":"object","properties":{"data":{"type":"array","items":{"$ref":"#/components/schemas/gadget"}}}}}}}}}},
 "/widget":{"get":{"responses":{"200":{"description":"ok","content":{"application/vnd.api+json":{"schema":{"type":"object","properties":{"data":{"type":"array","items":{"$ref":"#/components/schemas/widget"}}}}}}}}}}
},
"components":{"schemas":{
 "gadget":{"type":"object","properties":{"attributes":{"type":"object","properties":{
   "updatedDate":{"type":"object"},"name":{"type":"string"},"enabled":{"type":"boolean"},"retries":{"type":"integer","format":"int32"}}}}},
 "widget":{"type":"object","properties":{"attributes":{"type":"object","properties":{"label":{"type":"string"}}}}}
}}}`

func loadSpec(t *testing.T, data string) *speccheck.Spec {
	t.Helper()
	spec, err := speccheck.LoadData([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func plannedResource(t *testing.T, typ string) genResource {
	t.Helper()
	g, skip := planResource(loadSpec(t, gadgetSpec), typ)
	if skip != "" {
		t.Fatalf("planResource(%s) skipped: %s", typ, skip)
	}
	return g
}

// doctoredSpec returns the vendored document with mutate applied to its schemas.
func doctoredSpec(t *testing.T, root string, mutate func(schemas map[string]any)) string {
	t.Helper()
	data, err := os.ReadFile(specFile(root, terrakube.APIVersion))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	mutate(doc["components"].(map[string]any)["schemas"].(map[string]any))
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func orgAttributes(schemas map[string]any) map[string]any {
	return schemas["organization"].(map[string]any)["properties"].(map[string]any)["attributes"].(map[string]any)["properties"].(map[string]any)
}

func TestPlanSync_RealDocumentIsInSync(t *testing.T) {
	t.Parallel()

	root := realRoot(t)
	spec, err := speccheck.LoadFile(specFile(root, terrakube.APIVersion))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := planSync(spec, terrakube.APIVersion)
	if err != nil {
		t.Fatalf("planSync: %v", err)
	}
	if err := dropMarkedDeprecations(root, &plan); err != nil {
		t.Fatalf("dropMarkedDeprecations: %v", err)
	}
	if !plan.empty() {
		t.Errorf("plan is not empty: %+v", plan)
	}
}

func TestPlanSync_DetectsNewAndRemovedAttributes(t *testing.T) {
	t.Parallel()

	root := realRoot(t)
	data := doctoredSpec(t, root, func(schemas map[string]any) {
		attrs := orgAttributes(schemas)
		delete(attrs, "icon")
		attrs["quota"] = map[string]any{"type": "integer", "format": "int32"}
	})
	plan, err := planSync(loadSpec(t, data), "9.9.9")
	if err != nil {
		t.Fatalf("planSync: %v", err)
	}
	if plan.empty() || plan.Version != "9.9.9" || plan.OldVersion != terrakube.APIVersion {
		t.Errorf("plan = %+v", plan)
	}
	wantField := newField{Model: "Organization", GoName: "Quota", GoType: "*int32", JSONName: "quota"}
	if !slices.Contains(plan.NewFields, wantField) {
		t.Errorf("NewFields = %+v, want to contain %+v", plan.NewFields, wantField)
	}
	wantDep := deprecation{Model: "Organization", GoName: "Icon", JSONName: "icon"}
	if !slices.Contains(plan.Deprecations, wantDep) {
		t.Errorf("Deprecations = %+v, want to contain %+v", plan.Deprecations, wantDep)
	}
}

func TestPlanSync_PlansResourcesTheClientLacks(t *testing.T) {
	t.Parallel()

	plan, err := planSync(loadSpec(t, tinySpec), "9.9.9")
	if err != nil {
		t.Fatalf("planSync: %v", err)
	}
	if len(plan.NewResources) != 1 || plan.NewResources[0].Name != "Gadget" {
		t.Errorf("NewResources = %+v, want Gadget", plan.NewResources)
	}
}

func TestPlanResource(t *testing.T) {
	t.Parallel()

	t.Run("organization-scoped resource", func(t *testing.T) {
		t.Parallel()
		g := plannedResource(t, "gadget")
		if g.Name != "Gadget" || g.Field != "Gadgets" || g.Human != "gadget" || g.HumanPlural != "gadgets" ||
			g.FilterKey != "filter[gadget]" || g.Segment != "gadget" || g.IDVar != "gadgetID" || g.IDValue != "gadget-1" {
			t.Errorf("resource = %+v", g)
		}
		if len(g.Parents) != 1 || g.Parents[0] != (genParent{Segment: "organization", Var: "orgID", Label: "organizationID", Value: "org-1"}) {
			t.Errorf("Parents = %+v", g.Parents)
		}
		if got := g.pathTemplate(); got != "/organization/{orgID}/gadget" {
			t.Errorf("pathTemplate = %q", got)
		}
		names := make([]string, 0, len(g.Attrs))
		for _, a := range g.Attrs {
			names = append(names, a.JSONName+":"+a.GoType)
		}
		want := []string{"enabled:bool", "name:*string", "retries:*int32", "updatedDate:*time.Time"}
		if !slices.Equal(names, want) {
			t.Errorf("Attrs = %v, want %v (audit attributes last)", names, want)
		}
		if !g.Methods.all() || len(g.methodNames()) != 5 {
			t.Errorf("Methods = %+v", g.Methods)
		}
	})

	t.Run("root resource has no parents", func(t *testing.T) {
		t.Parallel()
		g := plannedResource(t, "widget")
		if len(g.Parents) != 0 || g.pathTemplate() != "/widget" {
			t.Errorf("resource = %+v", g)
		}
	})

	t.Run("type without a collection path is skipped", func(t *testing.T) {
		t.Parallel()
		_, skip := planResource(loadSpec(t, gadgetSpec), "ghost")
		if !strings.Contains(skip, "no collection path") {
			t.Errorf("skip = %q", skip)
		}
	})

	t.Run("rule path missing from the document is skipped", func(t *testing.T) {
		t.Parallel()
		_, skip := planResource(loadSpec(t, gadgetSpec), "policy_set")
		if !strings.Contains(skip, "does not resolve to policy_set") {
			t.Errorf("skip = %q", skip)
		}
	})
}

func TestApplyPlan_AppliesChangesAndIsIdempotent(t *testing.T) {
	t.Parallel()

	repo := realRoot(t)
	dir := t.TempDir()
	copyLibrary(t, repo, dir)

	plan := syncPlan{
		Version:    "9.9.9",
		OldVersion: terrakube.APIVersion,
		Retypes:    []retype{{Model: "Tag", Old: "tag", New: "tag2"}},
		FilterKeys: []filterKeyFix{{Service: "Workspaces", Old: "filter[workspace]", New: "filter[workspace2]"}},
		NewFields: []newField{
			{Model: "Organization", GoName: "Quota", GoType: "*int32", JSONName: "quota"},
			{Model: "Organization", GoName: "ExpiresAt", GoType: "*time.Time", JSONName: "expiresAt"},
			{Model: "Organization", GoName: "Gadgets", GoType: "[]*Workspace", JSONName: "gadget", Relation: true},
		},
		Deprecations: []deprecation{{Model: "Organization", GoName: "Icon", JSONName: "icon"}},
		NewResources: []genResource{plannedResource(t, "gadget"), plannedResource(t, "widget")},
	}

	written, err := applyPlan(dir, plan)
	if err != nil {
		t.Fatalf("applyPlan: %v", err)
	}
	for _, name := range []string{"client.go", "organization.go", "tag.go", "interfaces.go", "crud_test.go", "gadget.go", "gadget_test.go", "widget.go", "widget_test.go"} {
		if !slices.Contains(written, filepath.Join(dir, name)) {
			t.Errorf("%s not reported as written: %v", name, written)
		}
	}

	files := readDir(t, dir)
	mustContain(t, "client.go", files["client.go"],
		`const APIVersion = "9.9.9"`,
		`c.Gadgets = &GadgetService{client: c, filterKey: "filter[gadget]"}`,
		`c.Widgets = &WidgetService{client: c, filterKey: "filter[widget]"}`,
		`filterKey: "filter[workspace2]"`)
	if !regexp.MustCompile(`Gadgets\s+\*GadgetService`).MatchString(files["client.go"]) {
		t.Error("client.go has no Gadgets field")
	}
	mustContain(t, "organization.go", files["organization.go"],
		"Quota", `jsonapi:"attr,quota"`, "ExpiresAt", "*time.Time", `jsonapi:"attr,expiresAt,rfc3339"`, `"time"`, `jsonapi:"relation,gadget,omitempty"`,
		"DEPRECATED (absent from the Terrakube API 9.9.9 OpenAPI document; kept for compatibility).")
	mustContain(t, "tag.go", files["tag.go"], `primary,tag2`)
	mustContain(t, "interfaces.go", files["interfaces.go"], "type GadgetLister interface", "GadgetCRUD", "_ WidgetCRUD")
	mustContain(t, "crud_test.go", files["crud_test.go"], `"filter[workspace2]"`, `"Gadgets":`, `"Widgets":`)
	mustContain(t, "gadget.go", files["gadget.go"], "type Gadget struct", `jsonapi:"primary,gadget"`, `jsonapi:"attr,updatedDate,rfc3339"`, `"time"`,
		"a field left at its zero value overwrites the value on the server.", "func (s *GadgetService) Delete(")
	mustContain(t, "gadget_test.go", files["gadget_test.go"], "func TestGadgetService_Create(", "TestGadget_BooleanFieldsSerialization")

	fset := token.NewFileSet()
	for name, src := range files {
		if _, err := parser.ParseFile(fset, name, src, parser.AllErrors); err != nil {
			t.Errorf("%s does not parse: %v", name, err)
		}
	}

	again, err := applyPlan(dir, plan)
	if err != nil {
		t.Fatalf("second applyPlan: %v", err)
	}
	if len(again) != 0 {
		t.Errorf("second apply wrote %v, want nothing", again)
	}
	after := readDir(t, dir)
	for name, src := range files {
		if after[name] != src {
			t.Errorf("%s changed on the second apply", name)
		}
	}
	if n := strings.Count(after["organization.go"], "DEPRECATED"); n != 1 {
		t.Errorf("organization.go has %d DEPRECATED markers, want 1", n)
	}
}

const miniClient = `package terrakube

const APIVersion = "1.0.0"
`

const miniThing = `package terrakube

// Thing is a thing.
type Thing struct {
	ID     string  ` + "`" + `jsonapi:"primary,thing"` + "`" + `
	Legacy *string ` + "`" + `jsonapi:"attr,legacy"` + "`" + ` // keeps its own comment
	Gone   *string ` + "`" + `jsonapi:"attr,gone"` + "`" + `
	// Marked is already marked.
	// DEPRECATED (absent from the Terrakube API 0.9 OpenAPI document; kept for compatibility).
	Marked *string ` + "`" + `jsonapi:"attr,marked"` + "`" + `
	Tail   *string ` + "`" + `jsonapi:"attr,tail"` + "`" + ` // DEPRECATED already
}
`

func miniLibrary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "client.go"), miniClient)
	writeFile(t, filepath.Join(dir, "thing.go"), miniThing)
	return dir
}

func TestApplyPlan_DeprecationMarkers(t *testing.T) {
	t.Parallel()

	dir := miniLibrary(t)
	plan := syncPlan{Version: "1.0.0", OldVersion: "1.0.0", Deprecations: []deprecation{
		{Model: "Thing", GoName: "Legacy", JSONName: "legacy"},
		{Model: "Thing", GoName: "Gone", JSONName: "gone"},
		{Model: "Thing", GoName: "Marked", JSONName: "marked"},
		{Model: "Thing", GoName: "Tail", JSONName: "tail"},
	}}
	if _, err := applyPlan(dir, plan); err != nil {
		t.Fatalf("applyPlan: %v", err)
	}
	src := readDir(t, dir)["thing.go"]
	if n := strings.Count(src, "DEPRECATED"); n != 4 { // two new markers, two existing
		t.Errorf("got %d DEPRECATED markers, want 4:\n%s", n, src)
	}
	// A field with its own trailing comment gets the marker on a line above.
	if !strings.Contains(src, "kept for compatibility).\n\tLegacy") {
		t.Errorf("Legacy marker not placed above the field:\n%s", src)
	}
	// A field without a comment gets a trailing one.
	if !regexp.MustCompile("`jsonapi:\"attr,gone\"`\\s+// DEPRECATED").MatchString(src) {
		t.Errorf("Gone marker not trailing:\n%s", src)
	}
}

func TestDropMarkedDeprecations(t *testing.T) {
	t.Parallel()

	dir := miniLibrary(t)
	plan := syncPlan{Deprecations: []deprecation{
		{Model: "Thing", GoName: "Gone", JSONName: "gone"},
		{Model: "Thing", GoName: "Marked", JSONName: "marked"},
		{Model: "Thing", GoName: "Tail", JSONName: "tail"},
	}}
	if err := dropMarkedDeprecations(dir, &plan); err != nil {
		t.Fatalf("dropMarkedDeprecations: %v", err)
	}
	if len(plan.Deprecations) != 1 || plan.Deprecations[0].JSONName != "gone" {
		t.Errorf("Deprecations = %+v, want only gone", plan.Deprecations)
	}

	empty := syncPlan{}
	if err := dropMarkedDeprecations(filepath.Join(dir, "nowhere"), &empty); err != nil {
		t.Errorf("empty plan: %v", err)
	}
	bad := syncPlan{Deprecations: []deprecation{{Model: "Missing", JSONName: "x"}}}
	if err := dropMarkedDeprecations(dir, &bad); err == nil {
		t.Error("unknown model returned nil error")
	}
}

func TestApplyPlan_Errors(t *testing.T) {
	t.Parallel()

	t.Run("unknown model", func(t *testing.T) {
		t.Parallel()
		_, err := applyPlan(miniLibrary(t), syncPlan{Version: "1", Retypes: []retype{{Model: "Nope", Old: "a", New: "b"}}})
		if err == nil || !strings.Contains(err.Error(), "type Nope not found") {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("no client.go", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "thing.go"), miniThing)
		if _, err := applyPlan(dir, syncPlan{Version: "1"}); err == nil || !strings.Contains(err.Error(), "client.go not found") {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("no APIVersion constant", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "client.go"), "package terrakube\n")
		if _, err := applyPlan(dir, syncPlan{Version: "1"}); err == nil || !strings.Contains(err.Error(), "APIVersion constant not found") {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("unparsable source", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "client.go"), "not go")
		if _, err := applyPlan(dir, syncPlan{Version: "1"}); err == nil {
			t.Error("applyPlan returned nil error")
		}
	})
}

func TestRunSync_AppliesDocumentChangesToLibrary(t *testing.T) {
	t.Parallel()

	repo := realRoot(t)
	root := writeRepo(t)
	copyLibrary(t, repo, root)
	writeFile(t, specFile(root, "9.9.9"), doctoredSpec(t, repo, func(schemas map[string]any) {
		orgAttributes(schemas)["quota"] = map[string]any{"type": "integer", "format": "int32"}
	}))

	var stdout, stderr strings.Builder
	if err := runSync([]string{"--to", "v9.9.9", "--root", root, "--dry-run"}, &stdout, &stderr); err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if strings.Contains(stdout.String(), "wrote ") || strings.Contains(readDir(t, root)["organization.go"], "Quota") {
		t.Errorf("dry run wrote files:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "APIVersion 2.34.1 -> 9.9.9") || !strings.Contains(stdout.String(), "fields   Organization: 1 new") {
		t.Errorf("dry run output:\n%s", stdout.String())
	}

	stdout.Reset()
	if err := runSync([]string{"--to", "9.9.9", "--root", root, "--verbose"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSync: %v", err)
	}
	mustContain(t, "stdout", stdout.String(), "attr Quota *int32 (quota)", "wrote "+filepath.Join(root, "organization.go"), "wrote "+filepath.Join(root, "client.go"), "review the changes")
	files := readDir(t, root)
	mustContain(t, "organization.go", files["organization.go"], `jsonapi:"attr,quota"`)
	mustContain(t, "client.go", files["client.go"], `const APIVersion = "9.9.9"`)

	// Running it again with the same document writes nothing new.
	stdout.Reset()
	if err := runSync([]string{"--to", "9.9.9", "--root", root}, &stdout, &stderr); err != nil {
		t.Fatalf("second runSync: %v", err)
	}
	if strings.Contains(stdout.String(), "wrote ") {
		t.Errorf("second run wrote files:\n%s", stdout.String())
	}
	if after := readDir(t, root); after["organization.go"] != files["organization.go"] || after["client.go"] != files["client.go"] {
		t.Error("second run changed files")
	}
}

func TestRunSync_InSyncAndErrors(t *testing.T) {
	t.Parallel()

	repo := realRoot(t)
	root := writeRepo(t)
	copyLibrary(t, repo, root)
	data, err := os.ReadFile(specFile(repo, terrakube.APIVersion))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, specFile(root, terrakube.APIVersion), string(data))

	code, stdout, stderr := runCLI(t, "sync", "--to", terrakube.APIVersion, "--root", root)
	if code != 0 || !strings.Contains(stdout, "client is in sync with Terrakube API "+terrakube.APIVersion) {
		t.Errorf("exit = %d, stdout = %q, stderr = %q", code, stdout, stderr)
	}
	if code, _, stderr := runCLI(t, "sync"); code != 1 || !strings.Contains(stderr, "--to is required") {
		t.Errorf("missing --to: exit = %d, stderr = %q", code, stderr)
	}
	if code, _, stderr := runCLI(t, "sync", "--to", "0.0.1", "--root", root); code != 1 || !strings.Contains(stderr, "loading OpenAPI document") {
		t.Errorf("missing document: exit = %d, stderr = %q", code, stderr)
	}
}

func TestPrintPlan(t *testing.T) {
	t.Parallel()

	g := plannedResource(t, "gadget")
	plan := syncPlan{
		Version: "2", OldVersion: "1",
		Retypes:      []retype{{Model: "Tag", Old: "a", New: "b"}},
		FilterKeys:   []filterKeyFix{{Service: "Tags", Old: "x", New: "y"}},
		NewFields:    []newField{{Model: "Tag", GoName: "N", GoType: "bool", JSONName: "n"}, {Model: "Tag", GoName: "Rel", GoType: "*Tag", JSONName: "rel", Relation: true}},
		Deprecations: []deprecation{{Model: "Tag", GoName: "Old", JSONName: "old"}},
		NewResources: []genResource{g},
		Skipped:      []string{"zeta", "alpha"},
	}

	var quiet, verbose strings.Builder
	printPlan(&quiet, plan, false)
	printPlan(&verbose, plan, true)
	mustContain(t, "quiet", quiet.String(), "APIVersion 1 -> 2", "retype   Tag: JSON:API type a -> b", "filter   Tags: x -> y",
		"resource Gadget (gadget): Client.Gadgets, path /organization/{orgID}/gadget, methods List, Get, Create, Update, Delete",
		"fields   Tag: 2 new", "deprecate 1 fields not in the document", "skipped 2 items")
	mustContain(t, "verbose", verbose.String(), "attr N bool (n)", "rel Rel *Tag (rel)", "deprecate Tag.Old (old)", "skip     alpha\nskip     zeta")

	var empty strings.Builder
	printPlan(&empty, syncPlan{Version: "1", OldVersion: "1"}, false)
	if empty.String() != "client is in sync with Terrakube API 1\n" {
		t.Errorf("empty plan output = %q", empty.String())
	}
}

func TestSnakeCase(t *testing.T) {
	t.Parallel()

	tests := map[string]string{"PolicySet": "policy_set", "Workspace": "workspace", "VCS": "vcs", "ModuleVersion": "module_version"}
	for in, want := range tests {
		if got := snakeCase(in); got != want {
			t.Errorf("snakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGoFieldName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"triggeredByJobId": "TriggeredByJobID",
		"storageUri":       "StorageURI",
		"migratedV2":       "MigratedV2",
		"policy_set":       "PolicySet",
		"name":             "Name",
		"vcs":              "VCS",
		"apiUrl":           "APIURL",
	}
	for in, want := range tests {
		if got := goFieldName(in); got != want {
			t.Errorf("goFieldName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLowerFirst(t *testing.T) {
	t.Parallel()

	tests := map[string]string{"PolicySet": "policySet", "ID": "id", "URLPath": "urlPath", "name": "name", "X": "x", "": ""}
	for in, want := range tests {
		if got := lowerFirst(in); got != want {
			t.Errorf("lowerFirst(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPlural(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"policy set": "policy sets", "entry": "entries", "address": "addresses", "box": "boxes",
		"match": "matches", "dash": "dashes", "key": "keys", "y": "ys",
	}
	for in, want := range tests {
		if got := plural(in); got != want {
			t.Errorf("plural(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHumanize(t *testing.T) {
	t.Parallel()

	tests := map[string]string{"PolicySet": "policy set", "Workspace": "workspace", "WorkspaceURL": "workspace URL", "VCS": "VCS"}
	for in, want := range tests {
		if got := humanize(in); got != want {
			t.Errorf("humanize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestJoinOr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a or b"},
		{[]string{"a", "b", "c"}, "a, b, or c"},
	}
	for _, tt := range tests {
		if got := joinOr(tt.in); got != tt.want {
			t.Errorf("joinOr(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTemplates_RenderGofumptableGo(t *testing.T) {
	t.Parallel()

	attrs := []genField{
		{GoName: "Name", GoType: "*string", JSONName: "name"},
		{GoName: "Enabled", GoType: "bool", JSONName: "enabled"},
		{GoName: "Retries", GoType: "*int32", JSONName: "retries"},
	}
	rels := []genField{{GoName: "Workspace", GoType: "*Workspace", JSONName: "workspace"}}
	base := genResource{
		Type: "policy_set", Name: "PolicySet", Field: "PolicySets", Human: "policy set", HumanPlural: "policy sets",
		FilterKey: "filter[policy_set]", Segment: "policySet", IDVar: "policySetID", IDValue: "policyset-1", Attrs: attrs, Rels: rels,
	}
	org := genParent{Segment: "organization", Var: "orgID", Label: "organizationID", Value: "org-1"}
	job := genParent{Segment: "job", Var: "jobID", Label: "jobID", Value: "job-1"}
	all := methodSet{List: true, Get: true, Create: true, Update: true, Delete: true}

	tests := []struct {
		name    string
		parents []genParent
		methods methodSet
		attrs   []genField
	}{
		{"no parents", nil, all, attrs},
		{"one parent", []genParent{org}, all, attrs},
		{"two parents", []genParent{org, job}, all, attrs},
		{"read only", []genParent{org}, methodSet{List: true, Get: true}, attrs},
		{"no delete", []genParent{org}, methodSet{List: true, Get: true, Create: true, Update: true}, attrs},
		{"no boolean or string attributes", nil, all, []genField{{GoName: "Retries", GoType: "*int32", JSONName: "retries"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := base
			g.Parents, g.Methods, g.Attrs = tt.parents, tt.methods, tt.attrs
			for _, tmpl := range []struct{ name, text, prefix string }{
				{"resource.go", resourceTemplate, ""},
				{"resource_test.go", testTemplate, ""},
				{"interfaces.go", interfacesTemplate, "package terrakube\n\nimport \"context\"\n\n"},
			} {
				src, err := render(tmpl.name, tmpl.text, g)
				if err != nil {
					t.Fatalf("render %s: %v", tmpl.name, err)
				}
				out, err := formatGo(tmpl.name, append([]byte(tmpl.prefix), src...))
				if err != nil {
					t.Fatalf("formatGo %s: %v\n%s", tmpl.name, err, src)
				}
				if tmpl.name == "resource.go" {
					mustContain(t, tmpl.name, string(out), "type PolicySet struct", "type PolicySetService struct")
				}
			}
		})
	}
}

func TestRender_BadTemplate(t *testing.T) {
	t.Parallel()

	if _, err := render("bad", "{{.Missing", nil); err == nil {
		t.Error("render accepted a broken template")
	}
	if _, err := render("bad", "{{.Missing}}", genResource{}); err == nil {
		t.Error("render accepted a missing field")
	}
}

func TestGenResource_Helpers(t *testing.T) {
	t.Parallel()

	g := plannedResource(t, "gadget")
	if got := g.ParentParams(); got != "orgID string, " {
		t.Errorf("ParentParams = %q", got)
	}
	if got := g.IDParams(); got != "orgID, gadgetID string" {
		t.Errorf("IDParams = %q", got)
	}
	if got := g.TestPath(true); got != "/api/v1/organization/org-1/gadget/gadget-1" {
		t.Errorf("TestPath = %q", got)
	}
	if got := g.Scope(); got != " for an organization" {
		t.Errorf("Scope = %q", got)
	}
	if got := g.EmptyParentArgs(0); got != `"", ` {
		t.Errorf("EmptyParentArgs = %q", got)
	}
	if got := g.InterfaceNames(); !slices.Equal(got, []string{"GadgetCRUD"}) {
		t.Errorf("InterfaceNames = %v", got)
	}
	if got := g.StringAttr(); got == nil || got.JSONName != "name" {
		t.Errorf("StringAttr = %+v", got)
	}
	if got := g.BoolAttrs(); len(got) != 1 || got[0].JSONName != "enabled" {
		t.Errorf("BoolAttrs = %+v", got)
	}

	w := plannedResource(t, "widget")
	if w.ParentParams() != "" || w.Scope() != "" || w.ParentNames() != "" {
		t.Errorf("widget helpers = %q %q %q", w.ParentParams(), w.Scope(), w.ParentNames())
	}
	w.Methods = methodSet{List: true, Get: true}
	if got := w.InterfaceNames(); !slices.Equal(got, []string{"WidgetLister", "WidgetGetter"}) {
		t.Errorf("InterfaceNames = %v", got)
	}
}

// A to-one relationship gets no List (policy_evaluation.override is @OneToOne).
func TestPlanResource_ToOneHasNoList(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "..", "openapi", "v2.34.1.json")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("vendored document not available: %v", err)
	}
	spec, err := speccheck.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	g, skip := planResource(spec, "policy_override")
	if skip != "" {
		t.Fatalf("skip = %q", skip)
	}
	if g.Methods.List || !g.Methods.Get || !g.Methods.Create || !g.Methods.Update || g.Methods.Delete {
		t.Errorf("methods = %+v, want Get, Create, Update only", g.Methods)
	}
	set, _ := planResource(spec, "policy_set")
	if !set.Methods.all() {
		t.Errorf("policy_set methods = %+v, want all", set.Methods)
	}
}
