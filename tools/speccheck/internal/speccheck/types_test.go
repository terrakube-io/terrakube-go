package speccheck

import (
	"reflect"
	"testing"
	"time"
)

func TestGoToOpenAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		typ    reflect.Type
		want   Property
		wantOK bool
	}{
		{"string", reflect.TypeFor[string](), Property{Type: "string"}, true},
		{"pointer to string", reflect.TypeFor[*string](), Property{Type: "string"}, true},
		{"time.Time", reflect.TypeFor[time.Time](), Property{Type: "string", Format: "date-time"}, true},
		{"pointer to time.Time", reflect.TypeFor[*time.Time](), Property{Type: "string", Format: "date-time"}, true},
		{"bool", reflect.TypeFor[bool](), Property{Type: "boolean"}, true},
		{"int", reflect.TypeFor[int](), Property{Type: "integer"}, true},
		{"int32", reflect.TypeFor[int32](), Property{Type: "integer", Format: "int32"}, true},
		{"int64", reflect.TypeFor[int64](), Property{Type: "integer", Format: "int64"}, true},
		{"float32", reflect.TypeFor[float32](), Property{Type: "number", Format: "float"}, true},
		{"float64", reflect.TypeFor[float64](), Property{Type: "number", Format: "double"}, true},
		{"string slice", reflect.TypeFor[[]string](), Property{Type: "array", ItemsType: "string"}, true},
		{"map", reflect.TypeFor[map[string]any](), Property{Type: "object"}, true},
		{"map with int keys", reflect.TypeFor[map[int]string](), Property{}, false},
		{"slice of unsupported", reflect.TypeFor[[]struct{}](), Property{}, false},
		{"struct", reflect.TypeFor[struct{}](), Property{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := GoToOpenAPI(tt.typ)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("GoToOpenAPI(%v) = %+v, %v; want %+v, %v", tt.typ, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestOpenAPIToGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		attr string
		prop Property
		want string
	}{
		{"string", "name", Property{Type: "string"}, "*string"},
		{"boolean", "enabled", Property{Type: "boolean"}, "bool"},
		{"integer without format", "count", Property{Type: "integer"}, "int32"},
		{"integer int32", "count", Property{Type: "integer", Format: "int32"}, "int32"},
		{"integer int64", "size", Property{Type: "integer", Format: "int64"}, "int64"},
		{"number", "ratio", Property{Type: "number"}, "float64"},
		{"array", "tags", Property{Type: "array"}, "[]string"},
		{"date-like object", "createdDate", Property{Type: "object"}, "*time.Time"},
		{"other object", "settings", Property{Type: "object"}, "map[string]any"},
		{"unknown type", "x", Property{Type: "weird"}, "*string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := OpenAPIToGo(tt.attr, tt.prop); got != tt.want {
				t.Errorf("OpenAPIToGo(%q, %+v) = %q, want %q", tt.attr, tt.prop, got, tt.want)
			}
		})
	}
}

func TestTypesCompatible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		attr string
		typ  reflect.Type
		prop Property
		want bool
	}{
		{"date-like object vs *time.Time", "createdDate", reflect.TypeFor[*time.Time](), Property{Type: "object"}, true},
		{"date-like object vs time.Time", "createdDate", reflect.TypeFor[time.Time](), Property{Type: "object"}, true},
		{"date-like string vs *time.Time", "expiresAt", reflect.TypeFor[*time.Time](), Property{Type: "string"}, true},
		{"non-date object vs *time.Time", "settings", reflect.TypeFor[*time.Time](), Property{Type: "object"}, false},
		{"non-date string vs *time.Time", "name", reflect.TypeFor[*time.Time](), Property{Type: "string"}, false},
		{"time.Time vs integer", "createdDate", reflect.TypeFor[*time.Time](), Property{Type: "integer"}, false},
		{"date-like object vs *string", "createdDate", reflect.TypeFor[*string](), Property{Type: "object"}, false},
		{"non-date object vs *string", "settings", reflect.TypeFor[*string](), Property{Type: "object"}, false},
		{"object vs map", "settings", reflect.TypeFor[map[string]any](), Property{Type: "object"}, true},
		{"object vs bool", "settings", reflect.TypeFor[bool](), Property{Type: "object"}, false},
		{"int vs integer", "n", reflect.TypeFor[int](), Property{Type: "integer"}, true},
		{"int vs integer int32", "n", reflect.TypeFor[int](), Property{Type: "integer", Format: "int32"}, true},
		{"int32 vs integer", "n", reflect.TypeFor[int32](), Property{Type: "integer"}, true},
		{"*int32 vs integer int32", "n", reflect.TypeFor[*int32](), Property{Type: "integer", Format: "int32"}, true},
		{"int32 vs integer int64", "n", reflect.TypeFor[int32](), Property{Type: "integer", Format: "int64"}, false},
		{"int64 vs integer int32", "n", reflect.TypeFor[int64](), Property{Type: "integer", Format: "int32"}, false},
		{"string vs integer", "n", reflect.TypeFor[string](), Property{Type: "integer"}, false},
		{"float64 vs number", "n", reflect.TypeFor[float64](), Property{Type: "number"}, true},
		{"int vs number", "n", reflect.TypeFor[int](), Property{Type: "number"}, false},
		{"[]string vs array", "n", reflect.TypeFor[[]string](), Property{Type: "array"}, true},
		{"string vs array", "n", reflect.TypeFor[string](), Property{Type: "array"}, false},
		{"string vs string", "n", reflect.TypeFor[string](), Property{Type: "string"}, true},
		{"bool vs string", "n", reflect.TypeFor[bool](), Property{Type: "string"}, false},
		{"unsupported Go type", "n", reflect.TypeFor[struct{}](), Property{Type: "string"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := TypesCompatible(tt.attr, tt.typ, tt.prop); got != tt.want {
				t.Errorf("TypesCompatible(%q, %v, %+v) = %v, want %v", tt.attr, tt.typ, tt.prop, got, tt.want)
			}
		})
	}
}

func TestIsDateLike(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"createdDate":     true,
		"updatedAt":       true,
		"lastTimestamp":   true,
		"tokenExpiration": true,
		"startTime":       true,
		"name":            false,
		"dateOfBirth":     false,
		"":                false,
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := IsDateLike(name); got != want {
				t.Errorf("IsDateLike(%q) = %v, want %v", name, got, want)
			}
		})
	}
}

func TestProperty_String(t *testing.T) {
	t.Parallel()

	if got := (Property{Type: "integer", Format: "int64"}).String(); got != "integer/int64" {
		t.Errorf("String() = %q, want integer/int64", got)
	}
	if got := (Property{Type: "string"}).String(); got != "string" {
		t.Errorf("String() = %q, want string", got)
	}
}

func TestDateLikeStringAttribute(t *testing.T) {
	t.Parallel()

	// github_app_token.expiresAt: a date the document types as a string.
	p := Property{Type: "string"}
	if got := OpenAPIToGo("expiresAt", p); got != "*time.Time" {
		t.Errorf("OpenAPIToGo = %q, want *time.Time", got)
	}
	if TypesCompatible("expiresAt", reflect.TypeFor[*string](), p) {
		t.Error("*string accepted for a date-like string attribute")
	}
	if !TypesCompatible("expiresAt", reflect.TypeFor[*time.Time](), p) {
		t.Error("*time.Time rejected for a date-like string attribute")
	}
	if got := OpenAPIToGo("name", p); got != "*string" {
		t.Errorf("OpenAPIToGo(name) = %q, want *string", got)
	}
}
