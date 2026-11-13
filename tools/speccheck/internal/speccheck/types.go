package speccheck

import (
	"reflect"
	"regexp"
)

// dateLike matches attribute names that Elide renders as "object" because the
// Java field is a date or timestamp. The client models these as strings.
var dateLike = regexp.MustCompile(`(Date|At|Timestamp|Expiration|Time)$`)

// IsDateLike reports whether an attribute name denotes a date or timestamp.
func IsDateLike(name string) bool { return dateLike.MatchString(name) }

// GoToOpenAPI maps a Go field type to the OpenAPI property it corresponds to.
// Pointers map to their element type; nullability is not part of the result
// because Elide documents do not mark properties as nullable.
func GoToOpenAPI(t reflect.Type) (Property, bool) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return Property{Type: "string"}, true
	case reflect.Bool:
		return Property{Type: "boolean"}, true
	case reflect.Int32:
		return Property{Type: "integer", Format: "int32"}, true
	case reflect.Int64:
		return Property{Type: "integer", Format: "int64"}, true
	case reflect.Int:
		return Property{Type: "integer"}, true
	case reflect.Float32:
		return Property{Type: "number", Format: "float"}, true
	case reflect.Float64:
		return Property{Type: "number", Format: "double"}, true
	case reflect.Slice:
		item, ok := GoToOpenAPI(t.Elem())
		if !ok {
			return Property{}, false
		}
		return Property{Type: "array", ItemsType: item.Type}, true
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return Property{Type: "object"}, true
		}
	}
	return Property{}, false
}

// OpenAPIToGo returns the Go type expression the client uses for an
// attribute: strings are optional (*string), booleans are plain bool so that
// false is always sent, dates rendered as "object" are *string, other objects
// are map[string]any, and arrays are []string.
func OpenAPIToGo(name string, p Property) string {
	switch p.Type {
	case "string":
		return "*string"
	case "boolean":
		return "bool"
	case "integer":
		switch p.Format {
		case "int64":
			return "int64"
		default:
			return "int32"
		}
	case "number":
		return "float64"
	case "array":
		return "[]string"
	case "object":
		if IsDateLike(name) {
			return "*string"
		}
		return "map[string]any"
	}
	return "*string"
}

// TypesCompatible reports whether a Go field type can hold the attribute
// described by p. It accepts the client's conventions: a string for a date
// rendered as "object", any integer kind for an integer without a matching
// format, and []string for arrays whose items Elide leaves untyped.
func TypesCompatible(name string, goType reflect.Type, p Property) bool {
	got, ok := GoToOpenAPI(goType)
	if !ok {
		return false
	}
	switch p.Type {
	case "object":
		if got.Type == "string" {
			return IsDateLike(name)
		}
		return got.Type == "object"
	case "integer":
		if got.Type != "integer" {
			return false
		}
		return got.Format == "" || p.Format == "" || got.Format == p.Format
	case "number":
		return got.Type == "number"
	case "array":
		return got.Type == "array"
	default:
		return got.Type == p.Type
	}
}
