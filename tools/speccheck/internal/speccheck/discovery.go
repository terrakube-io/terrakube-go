package speccheck

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ModelInfo describes one CRUD service of the client and the model it serves.
type ModelInfo struct {
	ServiceName  string       // Client field name, e.g. "Workspaces"
	ServiceType  reflect.Type // e.g. *terrakube.WorkspaceService
	ModelType    reflect.Type // e.g. terrakube.Workspace (a struct type)
	ResourceType string       // JSON:API type from the primary tag, e.g. "workspace"
	FilterKey    string       // list filter query key; empty means "filter"
	Attributes   []Field
	Relations    []Field
}

// Field is one jsonapi-tagged struct field.
type Field struct {
	GoName   string       // struct field name
	JSONName string       // name in the jsonapi tag
	Type     reflect.Type // Go type of the field
	Options  []string     // tag options after the name (omitempty, rfc3339)
}

// crudEmbedName is the name of the generic CRUD base embedded by services.
const crudEmbedName = "crudService"

// DiscoverModels inspects a *terrakube.Client (or any struct pointer laid out
// the same way) and returns one ModelInfo per service that embeds the generic
// CRUD base. Services without it, such as TeamTokens and Operations, are
// skipped. The model type is taken from the result of the service's Get
// method, so the client does not have to be connected to anything.
func DiscoverModels(client any) ([]ModelInfo, error) {
	v := reflect.ValueOf(client)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return nil, errors.New("speccheck: DiscoverModels needs a non-nil pointer to a client struct")
	}
	v = v.Elem()

	var models []ModelInfo
	for field, value := range v.Fields() {
		if !field.IsExported() || value.Kind() != reflect.Pointer || value.IsNil() {
			continue
		}
		svc := value.Elem()
		if svc.Kind() != reflect.Struct {
			continue
		}
		base, ok := embeddedCRUD(svc)
		if !ok {
			continue
		}
		info, err := modelOfService(field.Name, value.Type())
		if err != nil {
			return nil, err
		}
		if fk := base.FieldByName("filterKey"); fk.IsValid() && fk.Kind() == reflect.String {
			info.FilterKey = fk.String()
		}
		models = append(models, info)
	}
	return models, nil
}

func embeddedCRUD(svc reflect.Value) (reflect.Value, bool) {
	for field, value := range svc.Fields() {
		if field.Anonymous && strings.HasPrefix(field.Name, crudEmbedName) {
			return value, true
		}
	}
	return reflect.Value{}, false
}

func modelOfService(name string, svcType reflect.Type) (ModelInfo, error) {
	get, ok := svcType.MethodByName("Get")
	if !ok {
		return ModelInfo{}, fmt.Errorf("speccheck: service %s has no Get method", name)
	}
	ft := get.Type
	if ft.NumOut() == 0 || ft.Out(0).Kind() != reflect.Pointer || ft.Out(0).Elem().Kind() != reflect.Struct {
		return ModelInfo{}, fmt.Errorf("speccheck: %s.Get does not return a pointer to a struct", name)
	}
	model := ft.Out(0).Elem()
	resourceType, attrs, rels, err := ParseModel(model)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("speccheck: service %s: %w", name, err)
	}
	return ModelInfo{
		ServiceName:  name,
		ServiceType:  svcType,
		ModelType:    model,
		ResourceType: resourceType,
		Attributes:   attrs,
		Relations:    rels,
	}, nil
}

// ParseModel reads the jsonapi tags of a model struct and returns its
// JSON:API type, attributes, and relationships in declaration order.
func ParseModel(t reflect.Type) (resourceType string, attrs, rels []Field, err error) {
	if t.Kind() != reflect.Struct {
		return "", nil, nil, fmt.Errorf("%s is not a struct", t)
	}
	for f := range t.Fields() {
		tag, ok := f.Tag.Lookup("jsonapi")
		if !ok {
			continue
		}
		parts := strings.Split(tag, ",")
		if len(parts) < 2 {
			continue
		}
		field := Field{GoName: f.Name, JSONName: parts[1], Type: f.Type, Options: parts[2:]}
		switch parts[0] {
		case "primary":
			resourceType = parts[1]
		case "attr":
			attrs = append(attrs, field)
		case "relation":
			rels = append(rels, field)
		}
	}
	if resourceType == "" {
		return "", nil, nil, fmt.Errorf("%s has no jsonapi primary tag", t)
	}
	return resourceType, attrs, rels, nil
}

// RelationTarget returns the struct type a relationship field points to and
// whether it is to-many ([]*T) or to-one (*T).
func RelationTarget(t reflect.Type) (target reflect.Type, toMany bool, ok bool) {
	switch {
	case t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct:
		return t.Elem(), false, true
	case t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Pointer && t.Elem().Elem().Kind() == reflect.Struct:
		return t.Elem().Elem(), true, true
	}
	return nil, false, false
}
