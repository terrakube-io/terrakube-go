package terrakube_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"
)

// The optional permission flags (destroyJob, managePolicies) are *bool: an older
// server leaves them out, so a missing flag must decode as nil rather than false,
// and a nil flag must be left out of requests instead of being sent as null.

type optionalFlagCase struct {
	name     string
	typ      string
	attr     string
	newModel func() any
	field    func(model any) *bool
	set      func(model any, v *bool)
}

var optionalFlagCases = []optionalFlagCase{
	{
		name: "Team.DestroyJob", typ: "team", attr: "destroyJob",
		newModel: func() any { return new(terrakube.Team) },
		field:    func(m any) *bool { return m.(*terrakube.Team).DestroyJob },
		set:      func(m any, v *bool) { m.(*terrakube.Team).DestroyJob = v },
	},
	{
		name: "Team.ManagePolicies", typ: "team", attr: "managePolicies",
		newModel: func() any { return new(terrakube.Team) },
		field:    func(m any) *bool { return m.(*terrakube.Team).ManagePolicies },
		set:      func(m any, v *bool) { m.(*terrakube.Team).ManagePolicies = v },
	},
	{
		name: "ProjectAccess.DestroyJob", typ: "project_access", attr: "destroyJob",
		newModel: func() any { return new(terrakube.ProjectAccess) },
		field:    func(m any) *bool { return m.(*terrakube.ProjectAccess).DestroyJob },
		set:      func(m any, v *bool) { m.(*terrakube.ProjectAccess).DestroyJob = v },
	},
	{
		name: "WorkspaceAccess.DestroyJob", typ: "access", attr: "destroyJob",
		newModel: func() any { return new(terrakube.WorkspaceAccess) },
		field:    func(m any) *bool { return m.(*terrakube.WorkspaceAccess).DestroyJob },
		set:      func(m any, v *bool) { m.(*terrakube.WorkspaceAccess).DestroyJob = v },
	},
}

func TestOptionalPermissionFlags_Decode(t *testing.T) {
	t.Parallel()

	for _, tc := range optionalFlagCases {
		for _, sent := range []struct {
			name string
			attr string // JSON fragment for the attribute, empty when the server omits it
			want *bool
		}{
			{"omitted", "", nil},
			{"null", `,"` + tc.attr + `":null`, nil},
			{"false", `,"` + tc.attr + `":false`, new(false)},
			{"true", `,"` + tc.attr + `":true`, new(true)},
		} {
			t.Run(tc.name+"/"+sent.name, func(t *testing.T) {
				t.Parallel()

				body := `{"data":{"type":"` + tc.typ + `","id":"id-1","attributes":{"name":"n"` + sent.attr + `}}}`
				model := tc.newModel()
				if err := jsonapi.UnmarshalPayload(strings.NewReader(body), model); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}

				if got := tc.field(model); !reflect.DeepEqual(got, sent.want) {
					t.Errorf("%s = %s, want %s", tc.attr, fmtBoolPtr(got), fmtBoolPtr(sent.want))
				}
			})
		}
	}
}

func TestOptionalPermissionFlags_Encode(t *testing.T) {
	t.Parallel()

	for _, tc := range optionalFlagCases {
		for _, value := range []*bool{nil, new(false), new(true)} {
			t.Run(tc.name+"/"+fmtBoolPtr(value), func(t *testing.T) {
				t.Parallel()

				model := tc.newModel()
				tc.set(model, value)

				var buf bytes.Buffer
				if err := jsonapi.MarshalPayload(&buf, model); err != nil {
					t.Fatalf("marshal: %v", err)
				}
				var payload struct {
					Data struct {
						Attributes map[string]any `json:"attributes"`
					} `json:"data"`
				}
				if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
					t.Fatalf("decode payload: %v", err)
				}

				got, present := payload.Data.Attributes[tc.attr]
				switch {
				case value == nil && present:
					t.Errorf("nil %s was sent as %v, want it left out", tc.attr, got)
				case value != nil && !present:
					t.Errorf("%s = %v was dropped from the payload", tc.attr, *value)
				case value != nil && got != *value:
					t.Errorf("%s = %v, want %v", tc.attr, got, *value)
				}
			})
		}
	}
}

func fmtBoolPtr(b *bool) string {
	if b == nil {
		return "nil"
	}
	if *b {
		return "true"
	}
	return "false"
}
