package speccheck

import (
	"cmp"
	"reflect"
	"slices"
)

// IssueKind classifies a difference between a Go model and the spec.
type IssueKind string

// Issue kinds reported by ValidateModel and ValidatePaths.
const (
	// MissingInStruct: the spec has an attribute or relationship the struct lacks.
	MissingInStruct IssueKind = "missing_in_struct"
	// MissingInSchema: the struct has a field the spec does not describe.
	MissingInSchema IssueKind = "missing_in_schema"
	// TypeMismatch: the Go type cannot hold the spec's attribute type.
	TypeMismatch IssueKind = "type_mismatch"
	// RelationTargetMismatch: a relationship points to another resource type.
	RelationTargetMismatch IssueKind = "relation_target_mismatch"
	// SchemaMissing: the spec has no schema for the model's resource type.
	SchemaMissing IssueKind = "schema_missing"
	// PathMissing: the spec has no path matching the request the service sends.
	PathMissing IssueKind = "path_missing"
	// PathTypeMismatch: the spec path returns a different resource type.
	PathTypeMismatch IssueKind = "path_type_mismatch"
	// FilterKeyMismatch: the list filter key is not filter[<resource type>].
	FilterKeyMismatch IssueKind = "filter_key_mismatch"
	// KnownIssue: a difference tracked elsewhere; reported, not failing.
	KnownIssue IssueKind = "known_issue"
)

// Failing reports whether the kind means the client is out of date. Fields
// that exist only in Go (MissingInSchema) are kept for compatibility, and
// known issues are tracked elsewhere; both are reported without failing.
func (k IssueKind) Failing() bool { return k != MissingInSchema && k != KnownIssue }

// Issue is one difference between the client and the spec.
type Issue struct {
	Kind     IssueKind `json:"kind"`
	Field    string    `json:"field,omitempty"` // JSON:API attribute or relationship name
	Relation bool      `json:"relation,omitempty"`
	GoType   string    `json:"goType,omitempty"`
	SpecType string    `json:"specType,omitempty"`
	Detail   string    `json:"detail,omitempty"`
}

// ModelReport collects the issues found for one service and its model.
type ModelReport struct {
	Service      string  `json:"service"`
	Model        string  `json:"model"`
	ResourceType string  `json:"resourceType"`
	Issues       []Issue `json:"issues"`
}

// Failing reports whether the report contains an issue that fails a check.
func (r ModelReport) Failing() bool {
	return slices.ContainsFunc(r.Issues, func(i Issue) bool { return i.Kind.Failing() })
}

// ValidateModel compares a model's attributes and relationships with the
// resource schema of the same JSON:API type. A nil resource yields a single
// SchemaMissing issue.
func ValidateModel(m ModelInfo, res *Resource) ModelReport {
	rep := ModelReport{Service: m.ServiceName, Model: m.ModelType.Name(), ResourceType: m.ResourceType}
	if res == nil {
		rep.Issues = append(rep.Issues, Issue{Kind: SchemaMissing, Detail: "no schema for type " + m.ResourceType})
		return rep
	}

	goAttrs := map[string]Field{}
	for _, f := range m.Attributes {
		goAttrs[f.JSONName] = f
	}
	for name, prop := range res.Attributes {
		f, ok := goAttrs[name]
		if !ok {
			rep.Issues = append(rep.Issues, Issue{Kind: MissingInStruct, Field: name, SpecType: prop.String()})
			continue
		}
		if !TypesCompatible(name, f.Type, prop) {
			rep.Issues = append(rep.Issues, Issue{Kind: TypeMismatch, Field: name, GoType: f.Type.String(), SpecType: prop.String()})
		}
	}
	for name, f := range goAttrs {
		if _, ok := res.Attributes[name]; !ok {
			rep.Issues = append(rep.Issues, Issue{Kind: MissingInSchema, Field: name, GoType: f.Type.String()})
		}
	}

	goRels := map[string]Field{}
	for _, f := range m.Relations {
		goRels[f.JSONName] = f
	}
	for name, rel := range res.Relations {
		f, ok := goRels[name]
		if !ok {
			rep.Issues = append(rep.Issues, Issue{Kind: MissingInStruct, Field: name, Relation: true, SpecType: rel.Target})
			continue
		}
		if target, ok := relatedResourceType(f.Type); ok && rel.Target != "" && target != rel.Target {
			rep.Issues = append(rep.Issues, Issue{Kind: RelationTargetMismatch, Field: name, Relation: true, GoType: target, SpecType: rel.Target})
		}
	}
	for name, f := range goRels {
		if _, ok := res.Relations[name]; !ok {
			rep.Issues = append(rep.Issues, Issue{Kind: MissingInSchema, Field: name, Relation: true, GoType: f.Type.String()})
		}
	}

	sortIssues(rep.Issues)
	return rep
}

// relatedResourceType returns the JSON:API type of the struct a relationship
// field points to.
func relatedResourceType(t reflect.Type) (string, bool) {
	target, _, ok := RelationTarget(t)
	if !ok {
		return "", false
	}
	typ, _, _, err := ParseModel(target)
	if err != nil {
		return "", false
	}
	return typ, true
}

func sortIssues(issues []Issue) {
	slices.SortFunc(issues, func(a, b Issue) int {
		return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Field, b.Field))
	})
}

// ExpectedFilterKey is the list filter query key for a JSON:API type.
func ExpectedFilterKey(resourceType string) string { return "filter[" + resourceType + "]" }

// ValidateFilterKey checks that a service filters lists with
// filter[<resource type>]. An empty key means the generic "filter" parameter
// and is accepted.
func ValidateFilterKey(m ModelInfo) []Issue {
	if m.FilterKey == "" || m.FilterKey == ExpectedFilterKey(m.ResourceType) {
		return nil
	}
	return []Issue{{Kind: FilterKeyMismatch, GoType: m.FilterKey, SpecType: ExpectedFilterKey(m.ResourceType)}}
}
