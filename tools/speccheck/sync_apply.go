package main

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	gofumpt "mvdan.cc/gofumpt/format"
)

// edit replaces src[start:end] with text.
type edit struct {
	start, end int
	text       string
}

// sourceFile is a Go file of the library loaded for editing.
type sourceFile struct {
	path  string
	src   []byte
	fset  *token.FileSet
	file  *ast.File
	edits []edit

	timeAdded bool // importTime already queued its edit
}

func (f *sourceFile) offset(p token.Pos) int { return f.fset.Position(p).Offset }

func (f *sourceFile) lineStart(off int) int {
	return bytes.LastIndexByte(f.src[:off], '\n') + 1
}

func (f *sourceFile) lineEnd(off int) int {
	if i := bytes.IndexByte(f.src[off:], '\n'); i >= 0 {
		return off + i
	}
	return len(f.src)
}

func (f *sourceFile) insert(off int, text string) {
	f.edits = append(f.edits, edit{start: off, end: off, text: text})
}

// importTime makes sure the file imports "time".
func (f *sourceFile) importTime() {
	if f.timeAdded {
		return
	}
	f.timeAdded = true
	for _, imp := range f.file.Imports {
		if imp.Path.Value == `"time"` && imp.Name == nil {
			return
		}
	}
	for _, decl := range f.file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.IMPORT {
			continue
		}
		if gd.Lparen.IsValid() {
			f.insert(f.lineStart(f.offset(gd.Rparen)), "\t\"time\"\n")
		} else {
			f.insert(f.offset(gd.End()), "\nimport \"time\"")
		}
		return
	}
	f.insert(f.offset(f.file.Name.End()), "\n\nimport \"time\"")
}

// result applies the edits and formats the file with gofumpt.
func (f *sourceFile) result() ([]byte, error) {
	out := slices.Clone(f.src)
	// Apply from the end so offsets stay valid. Edits at the same offset are
	// applied last-added first, so that their texts end up in added order.
	edits := slices.Clone(f.edits)
	slices.Reverse(edits)
	slices.SortStableFunc(edits, func(a, b edit) int { return cmp.Compare(b.start, a.start) })
	for _, e := range edits {
		out = slices.Concat(out[:e.start], []byte(e.text), out[e.end:])
	}
	return formatGo(f.path, out)
}

func formatGo(path string, src []byte) ([]byte, error) {
	out, err := gofumpt.Source(src, gofumpt.Options{LangVersion: "go1.27", ModulePath: libraryModule})
	if err != nil {
		return nil, fmt.Errorf("formatting %s: %w", filepath.Base(path), err)
	}
	return out, nil
}

type library struct {
	dir   string
	files map[string]*sourceFile // by base name
}

func loadLibrary(dir string) (*library, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}
	lib := &library{dir: dir, files: map[string]*sourceFile{}}
	for _, path := range matches {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		lib.files[filepath.Base(path)] = &sourceFile{path: path, src: src, fset: fset, file: file}
	}
	return lib, nil
}

// findStruct returns the file and struct type declaring the named type,
// looking only at non-test files.
func (l *library) findStruct(name string) (*sourceFile, *ast.StructType, error) {
	for base, f := range l.files {
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		for _, decl := range f.file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts := spec.(*ast.TypeSpec)
				if ts.Name.Name != name {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return nil, nil, fmt.Errorf("%s is not a struct", name)
				}
				return f, st, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("type %s not found in %s", name, l.dir)
}

func (l *library) file(base string) (*sourceFile, error) {
	f, ok := l.files[base]
	if !ok {
		return nil, fmt.Errorf("%s not found in %s", base, l.dir)
	}
	return f, nil
}

// jsonapiTag returns the kind and name of a field's jsonapi tag.
func jsonapiTag(field *ast.Field) (kind, name string) {
	if field.Tag == nil {
		return "", ""
	}
	raw, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return "", ""
	}
	tag, ok := reflect.StructTag(raw).Lookup("jsonapi")
	if !ok {
		return "", ""
	}
	parts := strings.Split(tag, ",")
	if len(parts) < 2 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// applyPlan writes the planned changes into the library at dir and returns
// the files it wrote. Every step checks the current source first, so running
// it again with the same plan changes nothing.
func applyPlan(dir string, p syncPlan) ([]string, error) {
	lib, err := loadLibrary(dir)
	if err != nil {
		return nil, err
	}
	touched := map[*sourceFile]bool{}
	mark := func(f *sourceFile) { touched[f] = true }

	for _, r := range p.Retypes {
		f, st, err := lib.findStruct(r.Model)
		if err != nil {
			return nil, err
		}
		for _, field := range st.Fields.List {
			if kind, name := jsonapiTag(field); kind == "primary" && name == r.Old {
				start, end := f.offset(field.Tag.Pos()), f.offset(field.Tag.End())
				lit := strings.Replace(string(f.src[start:end]), "primary,"+r.Old, "primary,"+r.New, 1)
				f.edits = append(f.edits, edit{start: start, end: end, text: lit})
				mark(f)
			}
		}
	}

	if err := applyFieldChanges(lib, p, mark); err != nil {
		return nil, err
	}

	client, err := lib.file("client.go")
	if err != nil {
		return nil, err
	}
	if err := applyClientChanges(client, p); err != nil {
		return nil, err
	}
	mark(client)

	if len(p.NewResources) > 0 {
		ifaces, err := lib.file("interfaces.go")
		if err != nil {
			return nil, err
		}
		if err := applyInterfaces(ifaces, p.NewResources); err != nil {
			return nil, err
		}
		mark(ifaces)
	}
	if crudTest, ok := lib.files["crud_test.go"]; ok {
		applyFilterKeyTest(crudTest, p)
		mark(crudTest)
	}

	var written []string
	for f := range touched {
		out, err := f.result()
		if err != nil {
			return nil, err
		}
		if bytes.Equal(out, f.src) {
			continue
		}
		if err := os.WriteFile(f.path, out, 0o600); err != nil {
			return nil, err
		}
		written = append(written, f.path)
	}

	for _, g := range p.NewResources {
		base := snakeCase(g.Name)
		for _, gen := range []struct{ name, tmpl string }{{base + ".go", resourceTemplate}, {base + "_test.go", testTemplate}} {
			path := filepath.Join(dir, gen.name)
			if _, err := os.Stat(path); err == nil {
				continue // never overwrite hand-edited files
			} else if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			src, err := render(gen.name, gen.tmpl, g)
			if err != nil {
				return nil, err
			}
			out, err := formatGo(path, src)
			if err != nil {
				return nil, fmt.Errorf("%w\n%s", err, src)
			}
			if err := os.WriteFile(path, out, 0o600); err != nil {
				return nil, err
			}
			written = append(written, path)
		}
	}
	slices.Sort(written)
	return written, nil
}

func applyFieldChanges(lib *library, p syncPlan, mark func(*sourceFile)) error {
	byModel := map[string][]newField{}
	for _, f := range p.NewFields {
		byModel[f.Model] = append(byModel[f.Model], f)
	}
	deprecated := map[string][]deprecation{}
	for _, d := range p.Deprecations {
		deprecated[d.Model] = append(deprecated[d.Model], d)
	}

	for _, model := range slices.Sorted(mapsKeys(byModel, deprecated)) {
		f, st, err := lib.findStruct(model)
		if err != nil {
			return err
		}
		existing := map[string]bool{}
		lastAttr := -1
		for i, field := range st.Fields.List {
			kind, name := jsonapiTag(field)
			if kind == "attr" || kind == "relation" {
				existing[kind+":"+name] = true
			}
			if kind == "primary" || kind == "attr" {
				lastAttr = i
			}
		}

		// Markers first: they share an insertion point with new attributes
		// added after the last attribute, and must stay on the marked field.
		for _, d := range deprecated[model] {
			for _, field := range st.Fields.List {
				if kind, name := jsonapiTag(field); kind != "attr" || name != d.JSONName {
					continue
				}
				if hasMarker(st, d.JSONName) {
					break
				}
				// Go's "Deprecated:" form, which gopls, staticcheck and pkg.go.dev
				// understand. It has to be a paragraph of the field's doc comment:
				// a trailing comment is not read as documentation. Elide hides
				// some fields the server still serves, so review each marker.
				marker := fmt.Sprintf("\t// Deprecated: absent from the Terrakube API %s OpenAPI document; kept for compatibility.\n", p.Version)
				if field.Doc != nil {
					marker = "\t//\n" + marker
				}
				f.insert(f.lineStart(f.offset(field.Pos())), marker)
				mark(f)
			}
		}

		var attrs, rels strings.Builder
		needTime := false
		for _, nf := range byModel[model] {
			if nf.Relation {
				if !existing["relation:"+nf.JSONName] {
					fmt.Fprintf(&rels, "\t%s %s `jsonapi:\"relation,%s,omitempty\"`\n", nf.GoName, nf.GoType, nf.JSONName)
				}
				continue
			}
			if !existing["attr:"+nf.JSONName] {
				fmt.Fprintf(&attrs, "\n\t%s %s `jsonapi:\"%s\"`", nf.GoName, nf.GoType, attrTag(nf.GoType, nf.JSONName))
				needTime = needTime || nf.GoType == "*time.Time"
			}
		}
		if attrs.Len() > 0 {
			if lastAttr < 0 {
				return fmt.Errorf("%s has no primary or attribute field to insert after", model)
			}
			f.insert(f.lineEnd(f.offset(st.Fields.List[lastAttr].End())), attrs.String())
			if needTime {
				f.importTime()
			}
			mark(f)
		}
		if rels.Len() > 0 {
			f.insert(f.lineStart(f.offset(st.Fields.Closing)), rels.String())
			mark(f)
		}

	}
	return nil
}

func mapsKeys[A, B any](a map[string]A, b map[string]B) func(func(string) bool) {
	return func(yield func(string) bool) {
		seen := map[string]bool{}
		for k := range a {
			seen[k] = true
		}
		for k := range b {
			seen[k] = true
		}
		for k := range seen {
			if !yield(k) {
				return
			}
		}
	}
}

var apiVersionLine = regexp.MustCompile(`(const APIVersion = )"[^"]*"`)

func applyClientChanges(f *sourceFile, p syncPlan) error {
	src := string(f.src)
	if loc := apiVersionLine.FindStringSubmatchIndex(src); loc != nil {
		f.edits = append(f.edits, edit{start: loc[0], end: loc[1], text: fmt.Sprintf("const APIVersion = %q", p.Version)})
	} else {
		return errors.New("client.go: APIVersion constant not found")
	}

	for _, fk := range p.FilterKeys {
		re := regexp.MustCompile(`c\.` + regexp.QuoteMeta(fk.Service) + ` = &\w+\{[^}\n]*filterKey: "(` + regexp.QuoteMeta(fk.Old) + `)"`)
		loc := re.FindStringSubmatchIndex(src)
		if loc == nil {
			continue
		}
		f.edits = append(f.edits, edit{start: loc[2], end: loc[3], text: fk.New})
	}

	if len(p.NewResources) == 0 {
		return nil
	}
	clientStruct, err := structClosing(f, "Client")
	if err != nil {
		return err
	}
	ret := strings.Index(src, "\n\n\treturn c, nil\n}")
	if ret < 0 {
		return errors.New("client.go: end of NewClient not found")
	}
	var fields, inits strings.Builder
	for _, g := range p.NewResources {
		if regexp.MustCompile(`(?m)^\t` + g.Field + `\s+\*`).MatchString(src) {
			continue
		}
		fmt.Fprintf(&fields, "\t%s *%sService\n", g.Field, g.Name)
		fmt.Fprintf(&inits, "\n\tc.%s = &%sService{client: c, filterKey: %q}", g.Field, g.Name, g.FilterKey)
	}
	f.insert(f.lineStart(clientStruct), fields.String())
	f.insert(ret, inits.String())
	return nil
}

// structClosing returns the offset of the closing brace of a struct type.
func structClosing(f *sourceFile, name string) (int, error) {
	for _, decl := range f.file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts := spec.(*ast.TypeSpec)
			if st, ok := ts.Type.(*ast.StructType); ok && ts.Name.Name == name {
				return f.offset(st.Fields.Closing), nil
			}
		}
	}
	return 0, fmt.Errorf("%s: struct %s not found", filepath.Base(f.path), name)
}

const assertionsMarker = "// Compile-time interface satisfaction checks."

func applyInterfaces(f *sourceFile, resources []genResource) error {
	src := string(f.src)
	marker := strings.Index(src, assertionsMarker)
	if marker < 0 {
		return fmt.Errorf("interfaces.go: %q not found", assertionsMarker)
	}
	closing := strings.Index(src[marker:], "\n)")
	if closing < 0 {
		return errors.New("interfaces.go: end of the assertion block not found")
	}
	var ifaces, asserts strings.Builder
	for _, g := range resources {
		if strings.Contains(src, "type "+g.Name+"Lister interface") {
			continue
		}
		block, err := render("interfaces", interfacesTemplate, g)
		if err != nil {
			return err
		}
		ifaces.Write(block)
		ifaces.WriteString("\n")
		for _, name := range g.InterfaceNames() {
			fmt.Fprintf(&asserts, "\n\t_ %s = (*%sService)(nil)", name, g.Name)
		}
	}
	f.insert(marker, ifaces.String())
	f.insert(marker+closing, asserts.String())
	return nil
}

var filterKeyMap = regexp.MustCompile(`expected := map\[string\]string\{\n`)

// applyFilterKeyTest keeps TestNewClient_FilterKeys in step: new services get
// an entry and fixed filter keys their new value.
func applyFilterKeyTest(f *sourceFile, p syncPlan) {
	src := string(f.src)
	for _, fk := range p.FilterKeys {
		re := regexp.MustCompile(`("` + regexp.QuoteMeta(fk.Service) + `":\s*)"` + regexp.QuoteMeta(fk.Old) + `"`)
		if loc := re.FindStringSubmatchIndex(src); loc != nil {
			f.edits = append(f.edits, edit{start: loc[3], end: loc[1], text: strconv.Quote(fk.New)})
		}
	}
	loc := filterKeyMap.FindStringIndex(src)
	if loc == nil || len(p.NewResources) == 0 {
		return
	}
	end := strings.Index(src[loc[1]:], "\n\t}")
	if end < 0 {
		return
	}
	var lines strings.Builder
	for _, g := range p.NewResources {
		if strings.Contains(src[loc[1]:loc[1]+end], strconv.Quote(g.Field)+":") {
			continue
		}
		fmt.Fprintf(&lines, "\n\t\t%q: %q,", g.Field, g.FilterKey)
	}
	f.insert(loc[1]+end, lines.String())
}

// snakeCase turns a Go name into a file name stem, e.g. PolicySet -> policy_set.
func snakeCase(name string) string {
	ws := words(name)
	for i, w := range ws {
		ws[i] = strings.ToLower(w)
	}
	return strings.Join(ws, "_")
}

// dropMarkedDeprecations removes from the plan the fields that already carry
// a deprecation marker, which reflection on the compiled client cannot see.
func dropMarkedDeprecations(dir string, p *syncPlan) error {
	if len(p.Deprecations) == 0 {
		return nil
	}
	lib, err := loadLibrary(dir)
	if err != nil {
		return err
	}
	var keep []deprecation
	for _, d := range p.Deprecations {
		_, st, err := lib.findStruct(d.Model)
		if err != nil {
			return err
		}
		if !hasMarker(st, d.JSONName) {
			keep = append(keep, d)
		}
	}
	p.Deprecations = keep
	return nil
}

func hasMarker(st *ast.StructType, jsonName string) bool {
	for _, field := range st.Fields.List {
		if kind, name := jsonapiTag(field); kind == "attr" && name == jsonName {
			// "DEPRECATED" is the form this tool wrote before it used Go's.
			marked := func(c *ast.CommentGroup) bool {
				return c != nil && (strings.Contains(c.Text(), "Deprecated:") || strings.Contains(c.Text(), "DEPRECATED"))
			}
			return marked(field.Doc) || marked(field.Comment)
		}
	}
	return false
}
