package main

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/tools/speccheck/internal/speccheck"
)

func runSync(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("sync", stderr)
	to := fs.String("to", "", "API version to sync the client to (required); openapi/v{TO}.json must exist")
	root := fs.String("root", "", "repository root (default: found from the working directory)")
	dryRun := fs.Bool("dry-run", false, "print the planned changes without writing files")
	verbose := fs.Bool("verbose", false, "print every planned change and every skipped item")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *to == "" {
		return fmt.Errorf("--to is required")
	}
	version := strings.TrimPrefix(*to, "v")
	if !versionPattern.MatchString(version) {
		return fmt.Errorf("--to %q is not of the form X.Y.Z", *to)
	}
	dir, err := resolveRoot(*root)
	if err != nil {
		return err
	}
	spec, err := speccheck.LoadFile(specFile(dir, version))
	if err != nil {
		return err
	}

	plan, err := planSync(spec, version)
	if err != nil {
		return err
	}
	if err := dropMarkedDeprecations(dir, &plan); err != nil {
		return err
	}
	printPlan(stdout, plan, *verbose)
	if *dryRun || plan.empty() {
		return nil
	}
	changed, err := applyPlan(dir, plan)
	if err != nil {
		return err
	}
	for _, f := range changed {
		_, _ = fmt.Fprintf(stdout, "wrote %s\n", f)
	}
	_, _ = fmt.Fprintln(stdout, `review the changes, add the new resources to the README table, then run "mise run check"`)
	return nil
}

// syncPlan is everything sync would change in the client's source.
type syncPlan struct {
	Version      string
	OldVersion   string
	Retypes      []retype
	FilterKeys   []filterKeyFix
	NewFields    []newField
	Deprecations []deprecation
	NewResources []genResource
	Skipped      []string
}

func (p syncPlan) empty() bool {
	return p.Version == p.OldVersion && len(p.Retypes) == 0 && len(p.FilterKeys) == 0 &&
		len(p.NewFields) == 0 && len(p.Deprecations) == 0 && len(p.NewResources) == 0
}

// retype changes a model's JSON:API type, e.g. ModuleVersion version -> module_version.
type retype struct {
	Model    string
	Old, New string
}

// filterKeyFix changes the filter key a service is built with in NewClient.
type filterKeyFix struct {
	Service  string
	Old, New string
}

// newField adds an attribute or relationship to an existing model struct.
type newField struct {
	Model    string
	GoName   string
	GoType   string
	JSONName string
	Relation bool
}

// deprecation marks a model field the document no longer describes.
type deprecation struct {
	Model    string
	GoName   string
	JSONName string
}

// planSync compares the compiled client with spec and works out the source
// changes. Type fixes come first so that fields are reconciled against the
// right schema.
func planSync(spec *speccheck.Spec, version string) (syncPlan, error) {
	rec := &speccheck.Recorder{}
	client, err := terrakube.NewClient(
		terrakube.WithEndpoint("https://speccheck.invalid"),
		terrakube.WithToken("speccheck"),
		terrakube.WithHTTPClient(&http.Client{Transport: rec}),
	)
	if err != nil {
		return syncPlan{}, err
	}
	models, err := speccheck.DiscoverModels(client)
	if err != nil {
		return syncPlan{}, err
	}
	plan := syncPlan{Version: version, OldVersion: terrakube.APIVersion}

	// 1. Correct JSON:API types from the type the service's own path returns.
	typeOf := map[string]string{} // model name -> resource type after fixes
	for i, m := range models {
		paths, err := speccheck.CapturePaths(client, m.ServiceName, rec)
		if err != nil {
			return syncPlan{}, err
		}
		probe := paths.Collection
		if probe == "" {
			probe = paths.Item
		}
		typ, _, _, ok := spec.ResolvePath(probe)
		if ok && typ != m.ResourceType && safeName.MatchString(typ) {
			if _, known := spec.Resource(typ); known {
				if typeOf[m.ModelType.Name()] == "" {
					plan.Retypes = append(plan.Retypes, retype{Model: m.ModelType.Name(), Old: m.ResourceType, New: typ})
				}
				models[i].ResourceType = typ
			}
		}
		typeOf[m.ModelType.Name()] = models[i].ResourceType
	}

	// 2. Filter keys follow the (corrected) type.
	for _, m := range models {
		if m.FilterKey != "" && m.FilterKey != speccheck.ExpectedFilterKey(m.ResourceType) {
			plan.FilterKeys = append(plan.FilterKeys, filterKeyFix{Service: m.ServiceName, Old: m.FilterKey, New: speccheck.ExpectedFilterKey(m.ResourceType)})
		}
	}

	// 3. New resources: document types no service covers.
	goNameOf := map[string]string{} // resource type -> Go struct name
	for _, m := range models {
		goNameOf[m.ResourceType] = m.ModelType.Name()
	}
	for _, typ := range spec.ResourceTypes() {
		if goNameOf[typ] != "" || len(spec.CollectionPaths(typ)) == 0 {
			continue
		}
		g, skip := planResource(spec, typ)
		if skip != "" {
			plan.Skipped = append(plan.Skipped, skip)
			continue
		}
		plan.NewResources = append(plan.NewResources, g)
		goNameOf[typ] = g.Name
	}
	// Relationships of new resources can only point to known types.
	for i := range plan.NewResources {
		g := &plan.NewResources[i]
		res, _ := spec.Resource(g.Type)
		for _, name := range sortedKeys(res.Relations) {
			f, skip := relationField(spec, g.Type, g.Name, name, res.Relations[name], goNameOf)
			if skip != "" {
				plan.Skipped = append(plan.Skipped, skip)
				continue
			}
			g.Rels = append(g.Rels, f)
		}
	}

	// 4. Reconcile fields of existing models (each struct once).
	seen := map[string]bool{}
	for _, m := range models {
		name := m.ModelType.Name()
		if seen[name] {
			continue
		}
		seen[name] = true
		res, ok := spec.Resource(m.ResourceType)
		if !ok {
			continue
		}
		taken := map[string]bool{}
		for f := range m.ModelType.Fields() {
			taken[f.Name] = true
		}
		goAttrs := map[string]speccheck.Field{}
		for _, f := range m.Attributes {
			goAttrs[f.JSONName] = f
		}
		goRels := map[string]bool{}
		for _, f := range m.Relations {
			goRels[f.JSONName] = true
		}
		for _, attr := range sortedKeys(res.Attributes) {
			if _, ok := goAttrs[attr]; ok {
				continue
			}
			if !safeName.MatchString(attr) {
				plan.Skipped = append(plan.Skipped, fmt.Sprintf("%s: attribute name %q is not a plain identifier", name, attr))
				continue
			}
			goName := goFieldName(attr)
			if taken[goName] {
				plan.Skipped = append(plan.Skipped, fmt.Sprintf("%s.%s: attribute %s would collide with an existing field", name, goName, attr))
				continue
			}
			taken[goName] = true
			plan.NewFields = append(plan.NewFields, newField{Model: name, GoName: goName, GoType: fieldGoType(attr, res.Attributes[attr]), JSONName: attr})
		}
		for _, rel := range sortedKeys(res.Relations) {
			if goRels[rel] {
				continue
			}
			f, skip := relationField(spec, m.ResourceType, name, rel, res.Relations[rel], goNameOf)
			if skip != "" {
				plan.Skipped = append(plan.Skipped, skip)
				continue
			}
			if taken[f.GoName] {
				plan.Skipped = append(plan.Skipped, fmt.Sprintf("%s.%s: relationship %s would collide with an existing field", name, f.GoName, rel))
				continue
			}
			taken[f.GoName] = true
			plan.NewFields = append(plan.NewFields, newField{Model: name, GoName: f.GoName, GoType: f.GoType, JSONName: rel, Relation: true})
		}
		for _, f := range m.Attributes {
			if _, ok := res.Attributes[f.JSONName]; !ok {
				plan.Deprecations = append(plan.Deprecations, deprecation{Model: name, GoName: f.GoName, JSONName: f.JSONName})
			}
		}
	}
	return plan, nil
}

// safeName matches names from the document that sync may copy into Go
// source: struct tags, string literals, and comments. Anything else is
// skipped, so a hostile document cannot inject code into generated files.
var safeName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// relationField builds the struct field for relationship rel of owner, or
// explains why it is skipped.
func relationField(spec *speccheck.Spec, ownerType, ownerName, rel string, r speccheck.Relation, goNameOf map[string]string) (genField, string) {
	if !safeName.MatchString(rel) {
		return genField{}, fmt.Sprintf("%s: relationship name %q is not a plain identifier", ownerName, rel)
	}
	target := goNameOf[r.Target]
	if target == "" {
		return genField{}, fmt.Sprintf("%s: relationship %s points to %s, which has no Go model", ownerName, rel, r.Target)
	}
	toMany, ok := spec.RelationIsToMany(ownerType, rel)
	if !ok {
		toMany, ok = relationToMany[ownerType+"."+rel]
	}
	if !ok {
		return genField{}, fmt.Sprintf("%s: relationship %s: the document does not say whether it is to-one or to-many; add it to relationToMany", ownerName, rel)
	}
	goType := "*" + target
	if toMany {
		goType = "[]*" + target
	}
	return genField{GoName: goFieldName(rel), GoType: goType, JSONName: rel}, ""
}

// fieldGoType returns the Go type for a new attribute. New integers are
// pointers, so that a zero value is distinguishable from "not returned".
func fieldGoType(name string, p speccheck.Property) string {
	t := speccheck.OpenAPIToGo(name, p)
	if t == "int32" || t == "int64" {
		return "*" + t
	}
	return t
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func printPlan(w io.Writer, p syncPlan, verbose bool) {
	pr := func(format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
	if p.empty() {
		pr("client is in sync with Terrakube API %s\n", p.Version)
	}
	if p.Version != p.OldVersion {
		pr("APIVersion %s -> %s\n", p.OldVersion, p.Version)
	}
	for _, r := range p.Retypes {
		pr("retype   %s: JSON:API type %s -> %s\n", r.Model, r.Old, r.New)
	}
	for _, f := range p.FilterKeys {
		pr("filter   %s: %s -> %s\n", f.Service, f.Old, f.New)
	}
	for _, g := range p.NewResources {
		pr("resource %s (%s): Client.%s, path %s, methods %s\n", g.Name, g.Type, g.Field, g.pathTemplate(), strings.Join(g.methodNames(), ", "))
	}
	byModel := map[string][]string{}
	for _, f := range p.NewFields {
		kind := "attr"
		if f.Relation {
			kind = "rel"
		}
		byModel[f.Model] = append(byModel[f.Model], fmt.Sprintf("%s %s %s (%s)", kind, f.GoName, f.GoType, f.JSONName))
	}
	for _, model := range sortedKeys(byModel) {
		fields := byModel[model]
		if verbose {
			pr("fields   %s:\n           %s\n", model, strings.Join(fields, "\n           "))
		} else {
			pr("fields   %s: %d new\n", model, len(fields))
		}
	}
	for _, d := range p.Deprecations {
		if verbose {
			pr("deprecate %s.%s (%s): not in the document\n", d.Model, d.GoName, d.JSONName)
		}
	}
	if !verbose && len(p.Deprecations) > 0 {
		pr("deprecate %d fields not in the document (use --verbose to list them)\n", len(p.Deprecations))
	}
	skipped := slices.Clone(p.Skipped)
	slices.SortFunc(skipped, cmp.Compare)
	for _, s := range skipped {
		if verbose {
			pr("skip     %s\n", s)
		}
	}
	if !verbose && len(skipped) > 0 {
		pr("skipped %d items (use --verbose to list them)\n", len(skipped))
	}
}
