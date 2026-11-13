package speccheck

import (
	"context"
	"reflect"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
)

// crudService mimics the generic base the library services embed.
type crudService[T any] struct{ filterKey string }

type fakeOwner struct {
	ID string `jsonapi:"primary,owner"`
}

type fakeModel struct {
	ID    string       `jsonapi:"primary,fake"`
	Name  *string      `jsonapi:"attr,name"`
	Owner *fakeOwner   `jsonapi:"relation,owner,omitempty"`
	Kids  []*fakeOwner `jsonapi:"relation,kids,omitempty"`
	Skip  string       // no jsonapi tag
	Bad   string       `jsonapi:"attr"` // malformed tag, ignored
}

type fakeService struct{ crudService[fakeModel] } //nolint:unused // embed is read by reflection

func (*fakeService) Get(_ context.Context, _ string) (*fakeModel, error) { return nil, nil }

type fakeFiltered struct{ crudService[fakeModel] }

func (*fakeFiltered) Get(_ context.Context, _ string) (*fakeModel, error) { return nil, nil }

type fakePlain struct{}

type fakeNoGet struct{ crudService[fakeModel] } //nolint:unused // embed is read by reflection

type fakeBadGet struct{ crudService[fakeModel] } //nolint:unused // embed is read by reflection

func (*fakeBadGet) Get(_ context.Context, _ string) (string, error) { return "", nil }

type noPrimary struct {
	Name string `jsonapi:"attr,name"`
}

type fakeNoPrimary struct{ crudService[noPrimary] } //nolint:unused // embed is read by reflection

func (*fakeNoPrimary) Get(_ context.Context, _ string) (*noPrimary, error) { return nil, nil }

func TestDiscoverModels_FakeClient(t *testing.T) {
	t.Parallel()

	client := &struct {
		Items    *fakeService
		Filtered *fakeFiltered
		Plain    *fakePlain
		Nil      *fakeService
		Other    *int
		Value    string
		hidden   *fakeService
	}{
		Items:    &fakeService{},
		Filtered: &fakeFiltered{crudService[fakeModel]{filterKey: "filter[fake]"}},
		Plain:    &fakePlain{},
		Other:    new(int),
		hidden:   &fakeService{},
	}
	_ = client.hidden

	got, err := DiscoverModels(client)
	if err != nil {
		t.Fatalf("DiscoverModels: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d models, want 2: %+v", len(got), got)
	}
	if got[0].ServiceName != "Items" || got[1].ServiceName != "Filtered" {
		t.Errorf("services = %s, %s; want Items, Filtered", got[0].ServiceName, got[1].ServiceName)
	}
	m := got[1]
	if m.ResourceType != "fake" {
		t.Errorf("ResourceType = %q, want fake", m.ResourceType)
	}
	if m.FilterKey != "filter[fake]" {
		t.Errorf("FilterKey = %q, want filter[fake]", m.FilterKey)
	}
	if got[0].FilterKey != "" {
		t.Errorf("FilterKey = %q, want empty", got[0].FilterKey)
	}
	if m.ModelType != reflect.TypeFor[fakeModel]() {
		t.Errorf("ModelType = %v, want fakeModel", m.ModelType)
	}
	if len(m.Attributes) != 1 || m.Attributes[0].JSONName != "name" || m.Attributes[0].GoName != "Name" {
		t.Errorf("Attributes = %+v, want one field name", m.Attributes)
	}
	if len(m.Relations) != 2 || m.Relations[0].JSONName != "owner" || m.Relations[1].JSONName != "kids" {
		t.Errorf("Relations = %+v, want owner, kids", m.Relations)
	}
}

func TestDiscoverModels_Errors(t *testing.T) {
	t.Parallel()

	var nilClient *struct{ Items *fakeService }
	tests := []struct {
		name   string
		client any
		want   string
	}{
		{"non-pointer", struct{ Items *fakeService }{}, "non-nil pointer"},
		{"nil pointer", nilClient, "non-nil pointer"},
		{"pointer to non-struct", new(int), "non-nil pointer"},
		{"service without Get", &struct{ Svc *fakeNoGet }{&fakeNoGet{}}, "no Get method"},
		{"Get not returning a struct pointer", &struct{ Svc *fakeBadGet }{&fakeBadGet{}}, "does not return a pointer to a struct"},
		{"model without primary tag", &struct{ Svc *fakeNoPrimary }{&fakeNoPrimary{}}, "no jsonapi primary tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := DiscoverModels(tt.client)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestDiscoverModels_RealClient(t *testing.T) {
	t.Parallel()

	client, err := terrakube.NewClient(terrakube.WithEndpoint("https://example.invalid"), terrakube.WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	models, err := DiscoverModels(client)
	if err != nil {
		t.Fatalf("DiscoverModels: %v", err)
	}
	if len(models) != 42 {
		t.Errorf("got %d services, want 42", len(models))
	}
	names := map[string]ModelInfo{}
	for _, m := range models {
		names[m.ServiceName] = m
	}
	for _, absent := range []string{"TeamTokens", "Operations"} {
		if _, ok := names[absent]; ok {
			t.Errorf("service %s discovered, want skipped", absent)
		}
	}
	mv, ok := names["ModuleVersions"]
	if !ok {
		t.Fatal("ModuleVersions not discovered")
	}
	if mv.ResourceType != "module_version" {
		t.Errorf("ModuleVersions ResourceType = %q, want module_version", mv.ResourceType)
	}
	if mv.FilterKey != "filter[module_version]" {
		t.Errorf("ModuleVersions FilterKey = %q, want filter[module_version]", mv.FilterKey)
	}
}

func TestRelationTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		typ        reflect.Type
		wantTarget reflect.Type
		wantMany   bool
		wantOK     bool
	}{
		{"pointer to struct", reflect.TypeFor[*fakeOwner](), reflect.TypeFor[fakeOwner](), false, true},
		{"slice of pointers", reflect.TypeFor[[]*fakeOwner](), reflect.TypeFor[fakeOwner](), true, true},
		{"string", reflect.TypeFor[string](), nil, false, false},
		{"pointer to string", reflect.TypeFor[*string](), nil, false, false},
		{"slice of structs", reflect.TypeFor[[]fakeOwner](), nil, false, false},
		{"slice of pointers to string", reflect.TypeFor[[]*string](), nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			target, many, ok := RelationTarget(tt.typ)
			if target != tt.wantTarget || many != tt.wantMany || ok != tt.wantOK {
				t.Errorf("RelationTarget(%v) = %v, %v, %v; want %v, %v, %v", tt.typ, target, many, ok, tt.wantTarget, tt.wantMany, tt.wantOK)
			}
		})
	}
}

func TestParseModel_NotAStruct(t *testing.T) {
	t.Parallel()

	if _, _, _, err := ParseModel(reflect.TypeFor[string]()); err == nil {
		t.Fatal("ParseModel(string) returned nil error")
	}
}
