package speccheck

import (
	"reflect"
	"slices"
	"testing"
)

type valTarget struct {
	ID string `jsonapi:"primary,target"`
}

type valOther struct {
	ID string `jsonapi:"primary,other"`
}

type valModel struct {
	ID string `jsonapi:"primary,thing"`
}

func testModelInfo() ModelInfo {
	return ModelInfo{
		ServiceName:  "Things",
		ModelType:    reflect.TypeFor[valModel](),
		ResourceType: "thing",
		Attributes: []Field{
			{GoName: "Name", JSONName: "name", Type: reflect.TypeFor[*string]()},
			{GoName: "Count", JSONName: "count", Type: reflect.TypeFor[string]()},
			{GoName: "Legacy", JSONName: "legacy", Type: reflect.TypeFor[*string]()},
		},
		Relations: []Field{
			{GoName: "Owner", JSONName: "owner", Type: reflect.TypeFor[*valTarget]()},
			{GoName: "Others", JSONName: "others", Type: reflect.TypeFor[[]*valOther]()},
			{GoName: "Old", JSONName: "old", Type: reflect.TypeFor[*valTarget]()},
		},
	}
}

func TestValidateModel_ReportsEveryKind(t *testing.T) {
	t.Parallel()

	res := &Resource{
		Type: "thing",
		Attributes: map[string]Property{
			"name":   {Type: "string"},
			"count":  {Type: "integer", Format: "int32"}, // Go has string: TypeMismatch
			"extra":  {Type: "boolean"},                  // MissingInStruct
			"active": {Type: "boolean"},                  // MissingInStruct
		},
		Relations: map[string]Relation{
			"owner":  {Target: "target"},
			"others": {Target: "target"}, // Go points to other: RelationTargetMismatch
			"added":  {Target: "target"}, // MissingInStruct
		},
	}
	rep := ValidateModel(testModelInfo(), res)

	if rep.Service != "Things" || rep.Model != "valModel" || rep.ResourceType != "thing" {
		t.Errorf("header = %+v", rep)
	}
	want := []Issue{
		{Kind: MissingInSchema, Field: "legacy", GoType: "*string"},
		{Kind: MissingInSchema, Field: "old", Relation: true, GoType: "*speccheck.valTarget"},
		{Kind: MissingInStruct, Field: "active", SpecType: "boolean"},
		{Kind: MissingInStruct, Field: "added", Relation: true, SpecType: "target"},
		{Kind: MissingInStruct, Field: "extra", SpecType: "boolean"},
		{Kind: RelationTargetMismatch, Field: "others", Relation: true, GoType: "other", SpecType: "target"},
		{Kind: TypeMismatch, Field: "count", GoType: "string", SpecType: "integer/int32"},
	}
	if !slices.Equal(rep.Issues, want) {
		t.Errorf("issues =\n%+v\nwant\n%+v", rep.Issues, want)
	}
	if !rep.Failing() {
		t.Error("Failing() = false, want true")
	}
}

func TestValidateModel_NilResourceReportsSchemaMissing(t *testing.T) {
	t.Parallel()

	rep := ValidateModel(testModelInfo(), nil)
	if len(rep.Issues) != 1 || rep.Issues[0].Kind != SchemaMissing {
		t.Fatalf("issues = %+v, want one SchemaMissing", rep.Issues)
	}
	if !rep.Failing() {
		t.Error("Failing() = false, want true")
	}
}

func TestValidateModel_MatchingSchemaHasNoIssues(t *testing.T) {
	t.Parallel()

	m := ModelInfo{
		ModelType:    reflect.TypeFor[valModel](),
		ResourceType: "thing",
		Attributes:   []Field{{GoName: "Name", JSONName: "name", Type: reflect.TypeFor[*string]()}},
		Relations:    []Field{{GoName: "Owner", JSONName: "owner", Type: reflect.TypeFor[*valTarget]()}},
	}
	res := &Resource{
		Attributes: map[string]Property{"name": {Type: "string"}},
		Relations:  map[string]Relation{"owner": {Target: "target"}},
	}
	rep := ValidateModel(m, res)
	if len(rep.Issues) != 0 || rep.Failing() {
		t.Errorf("issues = %+v, want none", rep.Issues)
	}
}

func TestValidateModel_RelationWithUnknownTargetIsAccepted(t *testing.T) {
	t.Parallel()

	m := ModelInfo{
		ModelType:    reflect.TypeFor[valModel](),
		ResourceType: "thing",
		Relations:    []Field{{GoName: "Owner", JSONName: "owner", Type: reflect.TypeFor[*valTarget]()}},
	}
	res := &Resource{Relations: map[string]Relation{"owner": {}}}
	if rep := ValidateModel(m, res); len(rep.Issues) != 0 {
		t.Errorf("issues = %+v, want none", rep.Issues)
	}
}

func TestValidateModel_RelationToModelWithoutPrimaryTagIsAccepted(t *testing.T) {
	t.Parallel()

	m := ModelInfo{
		ModelType:    reflect.TypeFor[valModel](),
		ResourceType: "thing",
		Relations:    []Field{{GoName: "Owner", JSONName: "owner", Type: reflect.TypeFor[*noPrimary]()}},
	}
	res := &Resource{Relations: map[string]Relation{"owner": {Target: "target"}}}
	if rep := ValidateModel(m, res); len(rep.Issues) != 0 {
		t.Errorf("issues = %+v, want none", rep.Issues)
	}
}

func TestIssueKind_Failing(t *testing.T) {
	t.Parallel()

	tests := map[IssueKind]bool{
		MissingInStruct:        true,
		MissingInSchema:        false,
		TypeMismatch:           true,
		RelationTargetMismatch: true,
		SchemaMissing:          true,
		PathMissing:            true,
		PathTypeMismatch:       true,
		FilterKeyMismatch:      true,
		KnownIssue:             false,
	}
	for kind, want := range tests {
		t.Run(string(kind), func(t *testing.T) {
			t.Parallel()
			if got := kind.Failing(); got != want {
				t.Errorf("%s.Failing() = %v, want %v", kind, got, want)
			}
		})
	}
}

func TestModelReport_Failing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		issues []Issue
		want   bool
	}{
		{"no issues", nil, false},
		{"only non-failing", []Issue{{Kind: MissingInSchema}, {Kind: KnownIssue}}, false},
		{"one failing", []Issue{{Kind: MissingInSchema}, {Kind: TypeMismatch}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := (ModelReport{Issues: tt.issues}).Failing(); got != tt.want {
				t.Errorf("Failing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFilterKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"empty key is accepted", "", false},
		{"exact key is accepted", "filter[thing]", false},
		{"wrong key is reported", "filter[other]", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issues := ValidateFilterKey(ModelInfo{ResourceType: "thing", FilterKey: tt.key})
			if (len(issues) > 0) != tt.wantErr {
				t.Fatalf("issues = %+v, wantErr %v", issues, tt.wantErr)
			}
			if tt.wantErr {
				want := Issue{Kind: FilterKeyMismatch, GoType: "filter[other]", SpecType: "filter[thing]"}
				if issues[0] != want {
					t.Errorf("issue = %+v, want %+v", issues[0], want)
				}
			}
		})
	}
}

func TestExpectedFilterKey(t *testing.T) {
	t.Parallel()

	if got := ExpectedFilterKey("module_version"); got != "filter[module_version]" {
		t.Errorf("ExpectedFilterKey = %q", got)
	}
}
