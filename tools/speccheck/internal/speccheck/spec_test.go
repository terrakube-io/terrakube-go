package speccheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// fixtureSpec builds a small OpenAPI 3.0.1 document shaped like Elide output.
//
//	organization -- workspace (to-many) -- job (to-many) -- step (to-many, no relationship endpoint)
//	workspace -- organization (to-one)
func fixtureSpec(t *testing.T) *Spec {
	t.Helper()

	ref := func(typ string) map[string]any { return map[string]any{"$ref": "#/components/schemas/" + typ} }
	get := func(data any) map[string]any {
		return map[string]any{"get": map[string]any{"responses": map[string]any{"200": map[string]any{
			"description": "ok",
			"content": map[string]any{"application/vnd.api+json": map[string]any{"schema": map[string]any{
				"type": "object", "properties": map[string]any{"data": data},
			}}},
		}}}}
	}
	coll := func(typ string) map[string]any {
		return get(map[string]any{"type": "array", "items": ref(typ)})
	}
	item := func(typ string) map[string]any { return get(ref(typ)) }
	relArr := get(map[string]any{"type": "array", "items": map[string]any{"type": "object"}})
	relObj := get(map[string]any{"type": "object"})
	rel := func(typ string) map[string]any {
		return map[string]any{"type": "object", "properties": map[string]any{"data": map[string]any{
			"type": "array",
			"items": map[string]any{"type": "object", "properties": map[string]any{
				"id":   map[string]any{"type": "string"},
				"type": map[string]any{"type": "string", "enum": []string{typ}},
			}},
		}}}
	}
	schema := func(attrs, rels map[string]any) map[string]any {
		return map[string]any{"type": "object", "properties": map[string]any{
			"id":            map[string]any{"type": "string"},
			"attributes":    map[string]any{"type": "object", "properties": attrs},
			"relationships": map[string]any{"type": "object", "properties": rels},
		}}
	}
	str := map[string]any{"type": "string"}

	doc := map[string]any{
		"openapi": "3.0.1",
		"info":    map[string]any{"title": "fixture", "version": "1"},
		"paths": map[string]any{
			"/organization":                            coll("organization"),
			"/organization/{organizationId}":           item("organization"),
			"/single":                                  item("organization"),
			"/organization/default":                    item("organization"),
			"/organization/{organizationId}/workspace": coll("workspace"),
			"/organization/{organizationId}/relationships/workspace": relArr,
			"/workspace":               coll("workspace"),
			"/workspace/{workspaceId}": item("workspace"),
			"/workspace/{workspaceId}/relationships/organization": relObj,
			"/workspace/{workspaceId}/relationships/job":          relArr,
			"/job":                   coll("job"),
			"/job/{jobId}":           item("job"),
			"/job/{jobId}/step":      coll("step"),
			"/job/{jobId}/workspace": item("workspace"),
			"/nodata":                map[string]any{"get": map[string]any{"responses": map[string]any{"200": map[string]any{"description": "ok"}}}},
			"/workspace/{workspaceId}/relationships/broken":                        map[string]any{"get": map[string]any{"responses": map[string]any{"204": map[string]any{"description": "none"}}}},
			"/organization/{organizationId}/workspace/{workspaceId}/relationships": map[string]any{"put": map[string]any{}},
		},
		"components": map[string]any{"schemas": map[string]any{
			"organization": schema(
				map[string]any{"name": str, "disabled": map[string]any{"type": "boolean"}, "count": map[string]any{"type": "integer", "format": "int32"}},
				map[string]any{"workspace": rel("workspace")}),
			"workspace": schema(
				map[string]any{"name": str, "createdDate": map[string]any{"type": "object"}, "tags": map[string]any{"type": "array", "items": str}},
				map[string]any{"organization": rel("organization"), "job": rel("job")}),
			"job":   schema(map[string]any{"status": str}, map[string]any{"step": rel("step"), "workspace": rel("workspace"), "ghost": map[string]any{}}),
			"step":  schema(map[string]any{"number": map[string]any{"type": "integer", "format": "int64"}}, map[string]any{}),
			"plain": map[string]any{"type": "object", "properties": map[string]any{"id": str}},
			"relsOnly": map[string]any{"type": "object", "properties": map[string]any{
				"relationships": map[string]any{"type": "object", "properties": map[string]any{}},
			}},
		}},
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	spec, err := LoadData(data)
	if err != nil {
		t.Fatalf("LoadData: %v", err)
	}
	return spec
}

func TestLoadData_InvalidDocument(t *testing.T) {
	t.Parallel()

	if _, err := LoadData([]byte("{not json")); err == nil {
		t.Fatal("LoadData returned nil error for invalid input")
	}
}

func TestLoadFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "spec.json")
	if err := os.WriteFile(path, []byte(`{"openapi":"3.0.1","info":{"title":"t","version":"1"},"paths":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if spec.Doc() == nil || spec.Doc().OpenAPI != "3.0.1" {
		t.Errorf("Doc() = %+v", spec.Doc())
	}
}

func TestLoadFile_Missing(t *testing.T) {
	t.Parallel()

	if _, err := LoadFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("LoadFile returned nil error for a missing file")
	}
}

func TestSpec_EmptyDocument(t *testing.T) {
	t.Parallel()

	spec, err := LoadData([]byte(`{"openapi":"3.0.1","info":{"title":"t","version":"1"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.ResourceTypes(); got != nil {
		t.Errorf("ResourceTypes = %v, want nil", got)
	}
	if _, ok := spec.Resource("x"); ok {
		t.Error("Resource found in an empty document")
	}
	if got := spec.Paths(); got != nil {
		t.Errorf("Paths = %v, want nil", got)
	}
	if got := spec.Operations("/x"); got != nil {
		t.Errorf("Operations = %v, want nil", got)
	}
	if _, _, ok := spec.PathResource("/x"); ok {
		t.Error("PathResource found in an empty document")
	}
}

func TestSpec_ResourceTypes(t *testing.T) {
	t.Parallel()

	got := fixtureSpec(t).ResourceTypes()
	want := []string{"job", "organization", "relsOnly", "step", "workspace"}
	if !slices.Equal(got, want) {
		t.Errorf("ResourceTypes = %v, want %v", got, want)
	}
}

func TestSpec_Resource(t *testing.T) {
	t.Parallel()

	spec := fixtureSpec(t)
	res, ok := spec.Resource("workspace")
	if !ok {
		t.Fatal("Resource(workspace) not found")
	}
	if res.Type != "workspace" {
		t.Errorf("Type = %q", res.Type)
	}
	if got := res.Attributes["createdDate"]; got.Type != "object" {
		t.Errorf("createdDate = %+v, want object", got)
	}
	if got := res.Attributes["tags"]; got.Type != "array" || got.ItemsType != "string" {
		t.Errorf("tags = %+v, want array of string", got)
	}
	if got := res.Relations["organization"].Target; got != "organization" {
		t.Errorf("organization target = %q", got)
	}
	if got := res.Relations["job"].Target; got != "job" {
		t.Errorf("job target = %q", got)
	}
	if got := spec.mustResource(t, "organization").Attributes["count"]; got != (Property{Type: "integer", Format: "int32"}) {
		t.Errorf("count = %+v", got)
	}
}

func (s *Spec) mustResource(t *testing.T, typ string) *Resource {
	t.Helper()
	res, ok := s.Resource(typ)
	if !ok {
		t.Fatalf("Resource(%s) not found", typ)
	}
	return res
}

func TestSpec_Resource_RelationWithoutDataHasNoTarget(t *testing.T) {
	t.Parallel()

	res := fixtureSpec(t).mustResource(t, "job")
	if got, ok := res.Relations["ghost"]; !ok || got.Target != "" {
		t.Errorf("ghost = %+v, %v; want present with empty target", got, ok)
	}
}

func TestSpec_Resource_Missing(t *testing.T) {
	t.Parallel()

	if _, ok := fixtureSpec(t).Resource("nope"); ok {
		t.Error("Resource(nope) = true, want false")
	}
}

func TestSpec_RelationTarget_ToOneData(t *testing.T) {
	t.Parallel()

	spec, err := LoadData([]byte(`{"openapi":"3.0.1","info":{"title":"t","version":"1"},"paths":{},"components":{"schemas":{
		"a":{"type":"object","properties":{"relationships":{"type":"object","properties":{
			"b":{"type":"object","properties":{"data":{"type":"object","properties":{"type":{"type":"string","enum":["bee"]}}}}}}}}}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.mustResource(t, "a").Relations["b"].Target; got != "bee" {
		t.Errorf("target = %q, want bee", got)
	}
}

func TestSpec_Paths(t *testing.T) {
	t.Parallel()

	got := fixtureSpec(t).Paths()
	if !slices.IsSorted(got) {
		t.Errorf("Paths not sorted: %v", got)
	}
	if !slices.Contains(got, "/organization/{organizationId}/workspace") {
		t.Errorf("Paths missing nested collection: %v", got)
	}
}

func TestSpec_Operations(t *testing.T) {
	t.Parallel()

	spec := fixtureSpec(t)
	if got := spec.Operations("/organization"); !slices.Equal(got, []string{"GET"}) {
		t.Errorf("Operations(/organization) = %v, want [GET]", got)
	}
	if got := spec.Operations("/organization/{organizationId}/workspace/{workspaceId}/relationships"); !slices.Equal(got, []string{"PUT"}) {
		t.Errorf("Operations(put path) = %v, want [PUT]", got)
	}
	if got := spec.Operations("/unknown"); got != nil {
		t.Errorf("Operations(/unknown) = %v, want nil", got)
	}
}

func TestSpec_PathResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path           string
		wantType       string
		wantCollection bool
		wantOK         bool
	}{
		{"/organization", "organization", true, true},
		{"/organization/{organizationId}", "organization", false, true},
		{"/organization/{organizationId}/workspace", "workspace", true, true},
		{"/unknown", "", false, false},
		{"/nodata", "", false, false},
		{"/workspace/{workspaceId}/relationships/broken", "", false, false},
		{"/organization/{organizationId}/workspace/{workspaceId}/relationships", "", false, false}, // no GET
	}
	spec := fixtureSpec(t)
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			typ, coll, ok := spec.PathResource(tt.path)
			if typ != tt.wantType || coll != tt.wantCollection || ok != tt.wantOK {
				t.Errorf("PathResource(%s) = %q, %v, %v; want %q, %v, %v", tt.path, typ, coll, ok, tt.wantType, tt.wantCollection, tt.wantOK)
			}
		})
	}
}

func TestSpec_MatchPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		concrete string
		want     string
		wantOK   bool
	}{
		{"template segment matches any value", "/organization/x0", "/organization/{organizationId}", true},
		{"prefers more literal segments", "/organization/default", "/organization/default", true},
		{"nested collection", "/organization/x0/workspace", "/organization/{organizationId}/workspace", true},
		{"surrounding slashes ignored", "organization/x0/", "/organization/{organizationId}", true},
		{"literal mismatch", "/organization/x0/nothing", "", false},
		{"length mismatch", "/a/b/c/d/e/f", "", false},
	}
	spec := fixtureSpec(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := spec.MatchPath(tt.concrete)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("MatchPath(%q) = %q, %v; want %q, %v", tt.concrete, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestSpec_ResolvePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		concrete       string
		wantType       string
		wantCollection bool
		wantVia        string
		wantOK         bool
	}{
		{"documented collection", "/organization/id0/workspace", "workspace", true, "path", true},
		{"documented item", "/organization/id0", "organization", false, "path", true},
		{"undocumented to-many chain", "/organization/id0/workspace/id1/job", "job", true, "relationships", true},
		{"undocumented to-one relation", "/workspace/id0/organization", "organization", false, "relationships", true},
		{"deep chain ending in an item", "/organization/id0/workspace/id1/job/id2", "job", false, "relationships", true},
		{"relation known only by fallback path", "/organization/id0/workspace/id1/job/id2/step", "step", true, "relationships", true},
		{"unknown relation", "/organization/id0/nothing", "", false, "", false},
		{"relation without target", "/job/id0/ghost", "", false, "", false},
		{"unknown root", "/nothing", "", false, "", false},
		{"non-collection root", "/single/id0/workspace", "", false, "", false},
		// job.workspace has no relationship endpoint, so its cardinality is unknown and assumed to-many.
		{"unknown cardinality is assumed to-many", "/workspace/id0/job/id1/workspace", "workspace", true, "relationships", true},
	}
	spec := fixtureSpec(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			typ, coll, via, ok := spec.ResolvePath(tt.concrete)
			if typ != tt.wantType || coll != tt.wantCollection || via != tt.wantVia || ok != tt.wantOK {
				t.Errorf("ResolvePath(%q) = %q, %v, %q, %v; want %q, %v, %q, %v",
					tt.concrete, typ, coll, via, ok, tt.wantType, tt.wantCollection, tt.wantVia, tt.wantOK)
			}
		})
	}
}

func TestSpec_ResolvePath_RelationWithoutDocumentedCardinalityIsCollection(t *testing.T) {
	t.Parallel()

	// job.ghost has no data and so no target; step.* has no relations at all.
	typ, coll, _, ok := fixtureSpec(t).ResolvePath("/job/id0/step/id1/anything")
	if ok {
		t.Errorf("ResolvePath = %q, %v, true; want not found", typ, coll)
	}
}

func TestSpec_CollectionPaths(t *testing.T) {
	t.Parallel()

	spec := fixtureSpec(t)
	got := spec.CollectionPaths("workspace")
	want := []string{"/workspace", "/organization/{organizationId}/workspace"}
	if !slices.Equal(got, want) {
		t.Errorf("CollectionPaths(workspace) = %v, want %v", got, want)
	}
	if got := spec.CollectionPaths("step"); !slices.Equal(got, []string{"/job/{jobId}/step"}) {
		t.Errorf("CollectionPaths(step) = %v", got)
	}
	if got := spec.CollectionPaths("nothing"); got != nil {
		t.Errorf("CollectionPaths(nothing) = %v, want nil", got)
	}
}

func TestSpec_RelationIsToMany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		owner    string
		rel      string
		wantMany bool
		wantOK   bool
	}{
		{"relationship endpoint returning array", "organization", "workspace", true, true},
		{"relationship endpoint returning object", "workspace", "organization", false, true},
		// Nested GETs are arrays for every Elide relationship, so a path without a
		// relationship endpoint says nothing about cardinality.
		{"nested path without relationship endpoint (collection)", "job", "step", false, false},
		{"nested path without relationship endpoint (item)", "job", "workspace", false, false},
		{"unknown relation", "job", "nothing", false, false},
		{"relationship endpoint of another owner", "step", "workspace", false, false},
		{"relationship endpoint without data", "workspace", "broken", false, false},
	}
	spec := fixtureSpec(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			many, ok := spec.RelationIsToMany(tt.owner, tt.rel)
			if many != tt.wantMany || ok != tt.wantOK {
				t.Errorf("RelationIsToMany(%s, %s) = %v, %v; want %v, %v", tt.owner, tt.rel, many, ok, tt.wantMany, tt.wantOK)
			}
		})
	}
}
