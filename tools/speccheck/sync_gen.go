package main

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"unicode"

	"github.com/terrakube-io/terrakube-go/tools/speccheck/internal/speccheck"
)

// resourceRule overrides how sync generates a new resource. Elide documents
// several paths per resource and does not say which operations callers may
// use, so the choice is recorded here for the resources it cannot infer.
type resourceRule struct {
	path     string // collection path template without the API base
	readOnly bool   // List and Get only
	noDelete bool
}

var resourceRules = map[string]resourceRule{
	// Created and deleted by the server ("user is a super service").
	"policy_evaluation": {path: "organization/{organizationId}/job/{jobId}/policyEvaluation", readOnly: true},
	// Created by teams, deleted only by the server.
	"policy_override":      {path: "organization/{organizationId}/job/{jobId}/policyEvaluation/{policy_evaluationId}/override", noDelete: true},
	"policy_set":           {path: "organization/{organizationId}/policySet"},
	"policy_set_parameter": {path: "organization/{organizationId}/policySet/{policy_setId}/parameters"},
	"policy_attachment":    {path: "organization/{organizationId}/policySet/{policy_setId}/attachments"},
	"policy_exemption":     {path: "organization/{organizationId}/policyExemption"},
	// runTriggers lists the triggers whose destination is the workspace.
	"runTrigger": {path: "organization/{organizationId}/workspace/{workspaceId}/runTriggers"},
}

// relationToMany records the cardinality of relationships the document does
// not reveal (no relationship endpoint), keyed "type.relationship". Values
// come from the JPA annotations of the Terrakube 2.34.1 entities.
var relationToMany = map[string]bool{
	"policy_attachment.policySet":    false, // @ManyToOne
	"policy_attachment.workspace":    false, // @ManyToOne
	"policy_attachment.project":      false, // @ManyToOne
	"policy_attachment.tag":          false, // @ManyToOne
	"policy_evaluation.job":          false, // @ManyToOne
	"policy_evaluation.step":         false, // @ManyToOne
	"policy_override.evaluation":     false, // @OneToOne
	"policy_set_parameter.policySet": false, // @ManyToOne
	"step.policyEvaluation":          true,  // inverse of policy_evaluation.step
}

// genResource is the template data for a generated resource.
type genResource struct {
	Type        string // JSON:API type, e.g. policy_set
	Name        string // Go struct name, e.g. PolicySet
	Field       string // Client field, e.g. PolicySets
	Human       string // e.g. "policy set"
	HumanPlural string // e.g. "policy sets"
	FilterKey   string
	Parents     []genParent
	Segment     string // collection segment, e.g. policySet
	IDVar       string // e.g. policySetID
	IDValue     string // test fixture ID
	Attrs       []genField
	Rels        []genField
	Methods     methodSet
	Version     string
}

type genParent struct {
	Segment string // path segment, e.g. organization
	Var     string // Go parameter, e.g. orgID
	Label   string // ValidationError field, e.g. organizationID
	Value   string // test fixture ID, e.g. org-1
}

type genField struct {
	GoName   string
	GoType   string
	JSONName string
}

type methodSet struct {
	List, Get, Create, Update, Delete bool
}

func (m methodSet) all() bool { return m.List && m.Get && m.Create && m.Update && m.Delete }

func (g genResource) methodNames() []string {
	var names []string
	for _, n := range []struct {
		name string
		on   bool
	}{{"List", g.Methods.List}, {"Get", g.Methods.Get}, {"Create", g.Methods.Create}, {"Update", g.Methods.Update}, {"Delete", g.Methods.Delete}} {
		if n.on {
			names = append(names, n.name)
		}
	}
	return names
}

func (g genResource) pathTemplate() string {
	var segs []string
	for _, p := range g.Parents {
		segs = append(segs, p.Segment, "{"+p.Var+"}")
	}
	return "/" + strings.Join(append(segs, g.Segment), "/")
}

// planResource works out how to generate a resource type no service covers.
// skip is set when sync cannot generate it.
func planResource(spec *speccheck.Spec, typ string) (genResource, string) {
	if !safeName.MatchString(typ) {
		return genResource{}, fmt.Sprintf("resource type %q is not a plain identifier", typ)
	}
	rule, hasRule := resourceRules[typ]
	path := strings.Trim(rule.path, "/")
	if !hasRule {
		path = defaultCollectionPath(spec, typ)
	}
	if path == "" {
		return genResource{}, fmt.Sprintf("resource %s: no collection path; add a rule to resourceRules", typ)
	}
	segs := strings.Split(path, "/")
	if len(segs)%2 != 1 {
		return genResource{}, fmt.Sprintf("resource %s: path %s does not end in a collection", typ, path)
	}

	g := genResource{Type: typ, Name: goTypeName(typ), FilterKey: speccheck.ExpectedFilterKey(typ), Segment: segs[len(segs)-1]}
	g.Field = plural(g.Name)
	g.Human = humanize(g.Name)
	g.HumanPlural = plural(g.Human)
	g.IDVar = lowerFirst(g.Name) + "ID"
	g.IDValue = strings.ToLower(g.Name) + "-1"

	concrete := make([]string, len(segs))
	for i, seg := range segs {
		if i%2 == 0 && !safeName.MatchString(seg) {
			return genResource{}, fmt.Sprintf("resource %s: path segment %q is not a plain identifier", typ, seg)
		}
		if i%2 == 1 {
			concrete[i] = fmt.Sprintf("id%d", i/2)
			continue
		}
		concrete[i] = seg
		if i == len(segs)-1 {
			break
		}
		p := genParent{Segment: seg}
		if seg == "organization" {
			p.Var, p.Label, p.Value = "orgID", "organizationID", "org-1"
		} else {
			p.Var = lowerFirst(goFieldName(seg)) + "ID"
			p.Label = p.Var
			p.Value = lowerFirst(goFieldName(seg)) + "-1"
		}
		g.Parents = append(g.Parents, p)
	}
	got, collection, _, ok := spec.ResolvePath("/" + strings.Join(concrete, "/"))
	if !ok || got != typ {
		return genResource{}, fmt.Sprintf("resource %s: path /%s does not resolve to %s in the document", typ, path, typ)
	}

	res, _ := spec.Resource(typ)
	for _, name := range sortedKeys(res.Attributes) {
		if !safeName.MatchString(name) {
			return genResource{}, fmt.Sprintf("resource %s: attribute name %q is not a plain identifier", typ, name)
		}
		g.Attrs = append(g.Attrs, genField{GoName: goFieldName(name), GoType: fieldGoType(name, res.Attributes[name]), JSONName: name})
	}
	slices.SortStableFunc(g.Attrs, func(a, b genField) int { return auditRank(a.JSONName) - auditRank(b.JSONName) })

	// A to-one relationship returns one resource, not a list: no List.
	g.Methods = methodSet{List: collection, Get: true, Create: !rule.readOnly, Update: !rule.readOnly, Delete: !rule.readOnly && !rule.noDelete}
	return g, ""
}

// defaultCollectionPath picks an organization-scoped collection path if the
// document has one, otherwise the shortest.
func defaultCollectionPath(spec *speccheck.Spec, typ string) string {
	paths := spec.CollectionPaths(typ)
	for _, p := range paths {
		if strings.HasPrefix(p, "/organization/{organizationId}/") {
			return strings.Trim(p, "/")
		}
	}
	if len(paths) > 0 {
		return strings.Trim(paths[0], "/")
	}
	return ""
}

// auditRank puts the audit attributes after the others, as in the hand-written models.
func auditRank(name string) int {
	return slices.Index([]string{"createdBy", "createdDate", "updatedBy", "updatedDate"}, name) + 1
}

// initialisms Go spells in capitals in identifiers.
var initialisms = map[string]string{
	"id": "ID", "url": "URL", "uri": "URI", "api": "API", "json": "JSON", "http": "HTTP",
	"https": "HTTPS", "sql": "SQL", "ssh": "SSH", "vcs": "VCS", "ip": "IP", "tcl": "TCL",
	"html": "HTML", "xml": "XML", "uuid": "UUID", "pr": "PR", "ttl": "TTL",
}

var wordBoundary = regexp.MustCompile(`[A-Z]?[a-z0-9]+|[A-Z]+(?:[0-9]+)?`)

func words(s string) []string {
	var out []string
	for part := range strings.SplitSeq(s, "_") {
		out = append(out, wordBoundary.FindAllString(part, -1)...)
	}
	return out
}

// goFieldName turns a JSON:API name into an exported Go identifier with Go's
// initialisms, e.g. triggeredByJobId -> TriggeredByJobID.
func goFieldName(name string) string {
	var b strings.Builder
	for _, w := range words(name) {
		if up, ok := initialisms[strings.ToLower(w)]; ok {
			b.WriteString(up)
			continue
		}
		r := []rune(w)
		b.WriteString(string(unicode.ToUpper(r[0])) + string(r[1:]))
	}
	return b.String()
}

// goTypeName turns a JSON:API type into a Go struct name, e.g. policy_set -> PolicySet.
func goTypeName(typ string) string { return goFieldName(typ) }

func lowerFirst(s string) string {
	r := []rune(s)
	// Lower a leading initialism as a whole: ID -> id, URLPath -> urlPath.
	n := 0
	for n < len(r) && unicode.IsUpper(r[n]) {
		n++
	}
	switch {
	case n == 0:
		return s
	case n == 1 || n == len(r):
		return strings.ToLower(string(r[:n])) + string(r[n:])
	default:
		return strings.ToLower(string(r[:n-1])) + string(r[n-1:])
	}
}

func humanize(name string) string {
	ws := words(name)
	for i, w := range ws {
		if _, ok := initialisms[strings.ToLower(w)]; !ok {
			ws[i] = strings.ToLower(w)
		}
	}
	return strings.Join(ws, " ")
}

func plural(s string) string {
	switch {
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "sh"):
		return s + "es"
	case strings.HasSuffix(s, "y") && len(s) > 1 && !strings.ContainsRune("aeiou", rune(s[len(s)-2])):
		return s[:len(s)-1] + "ies"
	default:
		return s + "s"
	}
}

// Template helpers.

func (g genResource) ParentParams() string {
	if len(g.Parents) == 0 {
		return ""
	}
	vars := make([]string, len(g.Parents))
	for i, p := range g.Parents {
		vars[i] = p.Var
	}
	return strings.Join(vars, ", ") + " string, "
}

// IDParams is the parent parameters followed by the item ID parameter,
// e.g. "orgID, policySetID string".
func (g genResource) IDParams() string {
	return strings.Join(append(g.parentVars(), g.IDVar), ", ") + " string"
}

func (g genResource) ParentArgs() string {
	var b strings.Builder
	for _, p := range g.Parents {
		fmt.Fprintf(&b, "%q, ", p.Value)
	}
	return b.String()
}

func (g genResource) PathArgs(item string) string {
	var parts []string
	for _, p := range g.Parents {
		parts = append(parts, fmt.Sprintf("%q", p.Segment), p.Var)
	}
	parts = append(parts, fmt.Sprintf("%q", g.Segment))
	if item != "" {
		parts = append(parts, item)
	}
	return strings.Join(parts, ", ")
}

func (g genResource) TestPath(withItem bool) string {
	var segs []string
	for _, p := range g.Parents {
		segs = append(segs, p.Segment, p.Value)
	}
	segs = append(segs, g.Segment)
	if withItem {
		segs = append(segs, g.IDValue)
	}
	return "/api/v1/" + strings.Join(segs, "/")
}

func (g genResource) ParentNames() string { return joinOr(g.parentVars()) }

func (g genResource) WithID() string { return joinOr(append(g.parentVars(), g.IDVar)) }

func (g genResource) WithEntityID() string { return joinOr(append(g.parentVars(), "the ID")) }

func (g genResource) parentVars() []string {
	vars := make([]string, len(g.Parents))
	for i, p := range g.Parents {
		vars[i] = p.Var
	}
	return vars
}

func joinOr(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
	}
}

func (g genResource) Scope() string {
	if len(g.Parents) == 0 {
		return ""
	}
	return " for " + article(humanize(goFieldName(g.Parents[len(g.Parents)-1].Segment)))
}

func article(s string) string {
	if s != "" && strings.ContainsRune("aeiou", rune(s[0])) {
		return "an " + s
	}
	return "a " + s
}

// StringAttr is the first *string attribute other than the audit ones, used
// for fixture values in tests.
func (g genResource) StringAttr() *genField {
	for _, a := range g.Attrs {
		if a.GoType == "*string" && auditRank(a.JSONName) == 0 {
			return &a
		}
	}
	return nil
}

func (g genResource) BoolAttrs() []genField {
	var out []genField
	for _, a := range g.Attrs {
		if a.GoType == "bool" {
			out = append(out, a)
		}
	}
	return out
}

func (g genResource) InterfaceNames() []string {
	if g.Methods.all() {
		return []string{g.Name + "CRUD"}
	}
	var names []string
	for _, m := range g.methodNames() {
		names = append(names, g.Name+methodInterface[m])
	}
	return names
}

var methodInterface = map[string]string{"List": "Lister", "Get": "Getter", "Create": "Creator", "Update": "Updater", "Delete": "Deleter"}

func upperFirst(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}

func render(name, text string, data any) ([]byte, error) {
	t, err := template.New(name).Funcs(template.FuncMap{"upperFirst": upperFirst}).Parse(text)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

const resourceTemplate = `package terrakube

import "context"

// {{.Name}} represents a Terrakube {{.Human}} resource.
type {{.Name}} struct {
	ID string ` + "`" + `jsonapi:"primary,{{.Type}}"` + "`" + `
{{- range .Attrs}}
	{{.GoName}} {{.GoType}} ` + "`" + `jsonapi:"attr,{{.JSONName}}"` + "`" + `
{{- end}}
{{- range .Rels}}
	{{.GoName}} {{.GoType}} ` + "`" + `jsonapi:"relation,{{.JSONName}},omitempty"` + "`" + `
{{- end}}
}

// {{.Name}}Service handles communication with the {{.Human}} related
// methods of the Terrakube API.
type {{.Name}}Service struct {
	crudService[{{.Name}}]
}
{{- if .Methods.List}}

// List returns all {{.HumanPlural}}{{.Scope}}.
{{- if .Parents}}
// It returns a *ValidationError if {{.ParentNames}} is empty and a *APIError on server errors.
{{- else}}
// It returns a *APIError on server errors.
{{- end}}
func (s *{{.Name}}Service) List(ctx context.Context, {{.ParentParams}}opts *ListOptions) ([]*{{.Name}}, error) {
{{- range .Parents}}
	if err := validateID("{{.Label}}", {{.Var}}); err != nil {
		return nil, err
	}
{{- end}}

	path := s.client.apiPath({{.PathArgs ""}})
	return s.list(ctx, path, opts)
}
{{- end}}
{{- if .Methods.Get}}

// Get retrieves a single {{.Human}} by ID.
// It returns a *ValidationError if {{.WithID}} is empty and a *APIError on server errors.
func (s *{{.Name}}Service) Get(ctx context.Context, {{.IDParams}}) (*{{.Name}}, error) {
{{- range .Parents}}
	if err := validateID("{{.Label}}", {{.Var}}); err != nil {
		return nil, err
	}
{{- end}}
	if err := validateID("{{.IDVar}}", {{.IDVar}}); err != nil {
		return nil, err
	}

	path := s.client.apiPath({{.PathArgs .IDVar}})
	return s.get(ctx, path)
}
{{- end}}
{{- if .Methods.Create}}

// Create creates a new {{.Human}}.
{{- if .Parents}}
// It returns a *ValidationError if {{.ParentNames}} is empty and a *APIError on server errors.
{{- else}}
// It returns a *APIError on server errors.
{{- end}}
func (s *{{.Name}}Service) Create(ctx context.Context, {{.ParentParams}}entity *{{.Name}}) (*{{.Name}}, error) {
{{- range .Parents}}
	if err := validateID("{{.Label}}", {{.Var}}); err != nil {
		return nil, err
	}
{{- end}}

	path := s.client.apiPath({{.PathArgs ""}})
	return s.create(ctx, path, entity)
}
{{- end}}
{{- if .Methods.Update}}

// Update modifies an existing {{.Human}}. The {{.Human}}'s ID field must be set.
// It returns a *ValidationError if {{.WithEntityID}} is empty and a *APIError on server errors.
func (s *{{.Name}}Service) Update(ctx context.Context, {{.ParentParams}}entity *{{.Name}}) (*{{.Name}}, error) {
{{- range .Parents}}
	if err := validateID("{{.Label}}", {{.Var}}); err != nil {
		return nil, err
	}
{{- end}}
	if err := validateID("{{.IDVar}}", entity.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath({{.PathArgs "entity.ID"}})
	return s.update(ctx, path, entity)
}
{{- end}}
{{- if .Methods.Delete}}

// Delete removes a {{.Human}}.
// It returns a *ValidationError if {{.WithID}} is empty and a *APIError on server errors.
func (s *{{.Name}}Service) Delete(ctx context.Context, {{.IDParams}}) error {
{{- range .Parents}}
	if err := validateID("{{.Label}}", {{.Var}}); err != nil {
		return err
	}
{{- end}}
	if err := validateID("{{.IDVar}}", {{.IDVar}}); err != nil {
		return err
	}

	path := s.client.apiPath({{.PathArgs .IDVar}})
	return s.del(ctx, path)
}
{{- end}}
`

const interfacesTemplate = `
{{- if .Methods.List}}
// {{.Name}}Lister lists {{.HumanPlural}}.
type {{.Name}}Lister interface {
	List(ctx context.Context, {{.ParentParams}}opts *ListOptions) ([]*{{.Name}}, error)
}
{{end}}
{{- if .Methods.Get}}
// {{.Name}}Getter retrieves a single {{.Human}}.
type {{.Name}}Getter interface {
	Get(ctx context.Context, {{.IDParams}}) (*{{.Name}}, error)
}
{{end}}
{{- if .Methods.Create}}
// {{.Name}}Creator creates {{.HumanPlural}}.
type {{.Name}}Creator interface {
	Create(ctx context.Context, {{.ParentParams}}entity *{{.Name}}) (*{{.Name}}, error)
}
{{end}}
{{- if .Methods.Update}}
// {{.Name}}Updater updates {{.HumanPlural}}.
type {{.Name}}Updater interface {
	Update(ctx context.Context, {{.ParentParams}}entity *{{.Name}}) (*{{.Name}}, error)
}
{{end}}
{{- if .Methods.Delete}}
// {{.Name}}Deleter deletes {{.HumanPlural}}.
type {{.Name}}Deleter interface {
	Delete(ctx context.Context, {{.IDParams}}) error
}
{{end}}
{{- if .Methods.All}}
// {{.Name}}CRUD combines all {{.Human}} operations.
type {{.Name}}CRUD interface {
	{{.Name}}Lister
	{{.Name}}Getter
	{{.Name}}Creator
	{{.Name}}Updater
	{{.Name}}Deleter
}
{{end}}`

// All is the template-visible form of all.
func (m methodSet) All() bool { return m.all() }

const testTemplate = `package terrakube_test

import (
	"context"
	"errors"
{{- if .BoolAttrs}}
	"io"
{{- end}}
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func newTest{{.Name}}() *terrakube.{{.Name}} {
	return &terrakube.{{.Name}}{
		ID: "{{.IDValue}}",
{{- with .StringAttr}}
		{{.GoName}}: new("test-{{.JSONName}}"),
{{- end}}
	}
}
{{- if .Methods.List}}

func Test{{.Name}}Service_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET {{.TestPath false}}", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.{{.Name}}{newTest{{.Name}}()})
	})

	client := newTestClient(t, srv)

	items, err := client.{{.Field}}.List(context.Background(), {{.ParentArgs}}nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ID != "{{.IDValue}}" {
		t.Errorf("ID = %q, want %q", items[0].ID, "{{.IDValue}}")
	}
}

func Test{{.Name}}Service_List_WithFilter(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET {{.TestPath false}}", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("{{.FilterKey}}"); got != "id==x" {
			t.Errorf("{{.FilterKey}} = %q, want %q", got, "id==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.{{.Name}}{})
	})

	client := newTestClient(t, srv)

	if _, err := client.{{.Field}}.List(context.Background(), {{.ParentArgs}}&terrakube.ListOptions{Filter: "id==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test{{.Name}}Service_List_ServerError(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET {{.TestPath false}}", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
	})

	client := newTestClient(t, srv)

	_, err := client.{{.Field}}.List(context.Background(), {{.ParentArgs}}nil)
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}
{{- range $i, $p := .Parents}}

func Test{{$.Name}}Service_List_Empty{{upperFirst $p.Label}}(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.{{$.Field}}.List(context.Background(), {{$.EmptyParentArgs $i}}nil)
	assertValidationError(t, err, "{{$p.Label}}")
}
{{- end}}
{{- end}}
{{- if .Methods.Get}}

func Test{{.Name}}Service_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET {{.TestPath true}}", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTest{{.Name}}())
	})

	client := newTestClient(t, srv)

	item, err := client.{{.Field}}.Get(context.Background(), {{.ParentArgs}}"{{.IDValue}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "{{.IDValue}}" {
		t.Errorf("ID = %q, want %q", item.ID, "{{.IDValue}}")
	}
{{- with .StringAttr}}
	if item.{{.GoName}} == nil || *item.{{.GoName}} != "test-{{.JSONName}}" {
		t.Errorf("{{.GoName}} = %v, want %q", item.{{.GoName}}, "test-{{.JSONName}}")
	}
{{- end}}
}

func Test{{.Name}}Service_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET {{.TestPath true}}", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	client := newTestClient(t, srv)

	_, err := client.{{.Field}}.Get(context.Background(), {{.ParentArgs}}"{{.IDValue}}")
	if !terrakube.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func Test{{.Name}}Service_Get_EmptyIDs(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")
{{range $i, $p := .Parents}}
	_, err{{$i}} := client.{{$.Field}}.Get(context.Background(), {{$.EmptyParentArgs $i}}"{{$.IDValue}}")
	assertValidationError(t, err{{$i}}, "{{$p.Label}}")
{{end}}
	_, err := client.{{.Field}}.Get(context.Background(), {{.ParentArgs}}"")
	assertValidationError(t, err, "{{.IDVar}}")
}
{{- end}}
{{- if .Methods.Create}}

func Test{{.Name}}Service_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST {{.TestPath false}}", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/vnd.api+json")
		}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTest{{.Name}}())
	})

	client := newTestClient(t, srv)

	item, err := client.{{.Field}}.Create(context.Background(), {{.ParentArgs}}&terrakube.{{.Name}}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "{{.IDValue}}" {
		t.Errorf("ID = %q, want %q", item.ID, "{{.IDValue}}")
	}
}

func Test{{.Name}}Service_Create_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST {{.TestPath false}}", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusUnprocessableEntity, "invalid")
	})

	client := newTestClient(t, srv)

	_, err := client.{{.Field}}.Create(context.Background(), {{.ParentArgs}}&terrakube.{{.Name}}{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}
{{- range $i, $p := .Parents}}

func Test{{$.Name}}Service_Create_Empty{{upperFirst $p.Label}}(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.{{$.Field}}.Create(context.Background(), {{$.EmptyParentArgs $i}}&terrakube.{{$.Name}}{})
	assertValidationError(t, err, "{{$p.Label}}")
}
{{- end}}
{{- if .BoolAttrs}}

func Test{{.Name}}_BooleanFieldsSerialization(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST {{.TestPath false}}", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}
{{- range .BoolAttrs}}
		assertJSONBoolField(t, body, "{{.JSONName}}", false)
{{- end}}
		testutil.WriteJSONAPI(t, w, http.StatusCreated, newTest{{.Name}}())
	})

	client := newTestClient(t, srv)

	if _, err := client.{{.Field}}.Create(context.Background(), {{.ParentArgs}}&terrakube.{{.Name}}{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
{{- end}}
{{- end}}
{{- if .Methods.Update}}

func Test{{.Name}}Service_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH {{.TestPath true}}", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, newTest{{.Name}}())
	})

	client := newTestClient(t, srv)

	item, err := client.{{.Field}}.Update(context.Background(), {{.ParentArgs}}newTest{{.Name}}())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "{{.IDValue}}" {
		t.Errorf("ID = %q, want %q", item.ID, "{{.IDValue}}")
	}
}

func Test{{.Name}}Service_Update_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.{{.Field}}.Update(context.Background(), {{.ParentArgs}}&terrakube.{{.Name}}{})
	assertValidationError(t, err, "{{.IDVar}}")
}
{{- range $i, $p := .Parents}}

func Test{{$.Name}}Service_Update_Empty{{upperFirst $p.Label}}(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	_, err := client.{{$.Field}}.Update(context.Background(), {{$.EmptyParentArgs $i}}newTest{{$.Name}}())
	assertValidationError(t, err, "{{$p.Label}}")
}
{{- end}}
{{- end}}
{{- if .Methods.Delete}}

func Test{{.Name}}Service_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE {{.TestPath true}}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)

	if err := client.{{.Field}}.Delete(context.Background(), {{.ParentArgs}}"{{.IDValue}}"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test{{.Name}}Service_Delete_EmptyID(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.{{.Field}}.Delete(context.Background(), {{.ParentArgs}}"")
	assertValidationError(t, err, "{{.IDVar}}")
}
{{- range $i, $p := .Parents}}

func Test{{$.Name}}Service_Delete_Empty{{upperFirst $p.Label}}(t *testing.T) {
	t.Parallel()

	client := newTestClientFromURL(t, "https://example.com")

	err := client.{{$.Field}}.Delete(context.Background(), {{$.EmptyParentArgs $i}}"{{$.IDValue}}")
	assertValidationError(t, err, "{{$p.Label}}")
}
{{- end}}
{{- end}}
`

// EmptyParentArgs returns the parent arguments with parent i empty.
func (g genResource) EmptyParentArgs(i int) string {
	var b strings.Builder
	for j, p := range g.Parents {
		if j == i {
			b.WriteString(`"", `)
			continue
		}
		fmt.Fprintf(&b, "%q, ", p.Value)
	}
	return b.String()
}
