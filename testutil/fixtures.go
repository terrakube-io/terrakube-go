package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

// WriteJSONAPI writes a single JSON:API resource response.
// entity must have jsonapi struct tags with a "primary" field.
func WriteJSONAPI(t testing.TB, w http.ResponseWriter, status int, entity interface{}) {
	t.Helper()

	rv := reflect.ValueOf(entity)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()

	typeName := ""
	id := ""
	attrs := map[string]interface{}{}
	rels := map[string]interface{}{}

	for i := range rt.NumField() {
		field := rt.Field(i)
		tag := field.Tag.Get("jsonapi")
		if tag == "" {
			continue
		}

		val := rv.Field(i)
		parts := splitTag(tag)

		switch parts[0] {
		case "primary":
			if len(parts) > 1 {
				typeName = parts[1]
			}
			id = fmt.Sprintf("%v", val.Interface())
		case "attr":
			if len(parts) > 1 {
				attrVal := val.Interface()
				if val.Kind() == reflect.Ptr {
					if val.IsNil() {
						continue
					}
					attrVal = val.Elem().Interface()
				}
				attrs[parts[1]] = attrVal
			}
		case "relation":
			if len(parts) > 1 && !val.IsZero() {
				relVal := val.Interface()
				if val.Kind() == reflect.Ptr {
					if val.IsNil() {
						continue
					}
					relVal = val.Elem().Interface()
				}
				relRV := reflect.ValueOf(relVal)
				if relRV.Kind() == reflect.Ptr {
					relRV = relRV.Elem()
				}
				relRT := relRV.Type()

				relData := map[string]interface{}{}
				for j := range relRT.NumField() {
					relTag := relRT.Field(j).Tag.Get("jsonapi")
					relParts := splitTag(relTag)
					if len(relParts) > 1 && relParts[0] == "primary" {
						relData["type"] = relParts[1]
						relData["id"] = fmt.Sprintf("%v", relRV.Field(j).Interface())
						break
					}
				}
				if len(relData) > 0 {
					rels[parts[1]] = map[string]interface{}{"data": relData}
				}
			}
		}
	}

	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       typeName,
			"id":         id,
			"attributes": attrs,
		},
	}
	if len(rels) > 0 {
		resp["data"].(map[string]interface{})["relationships"] = rels
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("testutil: failed to encode JSON:API response: %v", err)
	}
}

// WriteJSONAPIList writes a JSON:API list response containing multiple resources.
func WriteJSONAPIList(t testing.TB, w http.ResponseWriter, status int, entities interface{}) {
	t.Helper()

	rv := reflect.ValueOf(entities)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		t.Fatalf("testutil: WriteJSONAPIList expects a slice, got %s", rv.Kind())
	}

	items := make([]interface{}, 0, rv.Len())
	for i := range rv.Len() {
		elem := rv.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		rt := elem.Type()

		typeName := ""
		id := ""
		attrs := map[string]interface{}{}

		for j := range rt.NumField() {
			field := rt.Field(j)
			tag := field.Tag.Get("jsonapi")
			if tag == "" {
				continue
			}

			val := elem.Field(j)
			parts := splitTag(tag)

			switch parts[0] {
			case "primary":
				if len(parts) > 1 {
					typeName = parts[1]
				}
				id = fmt.Sprintf("%v", val.Interface())
			case "attr":
				if len(parts) > 1 {
					attrVal := val.Interface()
					if val.Kind() == reflect.Ptr {
						if val.IsNil() {
							continue
						}
						attrVal = val.Elem().Interface()
					}
					attrs[parts[1]] = attrVal
				}
			}
		}

		items = append(items, map[string]interface{}{
			"type":       typeName,
			"id":         id,
			"attributes": attrs,
		})
	}

	resp := map[string]interface{}{
		"data": items,
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("testutil: failed to encode JSON:API list response: %v", err)
	}
}

// WriteError writes a JSON:API error response.
func WriteError(t testing.TB, w http.ResponseWriter, status int, detail string) {
	t.Helper()
	resp := map[string]interface{}{
		"errors": []map[string]interface{}{
			{"detail": detail, "status": fmt.Sprintf("%d", status)},
		},
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("testutil: failed to encode error response: %v", err)
	}
}

// WriteJSON writes a standard JSON response (for non-JSON:API endpoints).
func WriteJSON(t testing.TB, w http.ResponseWriter, status int, v interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("testutil: failed to encode JSON response: %v", err)
	}
}

func splitTag(tag string) []string {
	result := []string{}
	current := ""
	for _, c := range tag {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
