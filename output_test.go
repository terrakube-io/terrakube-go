package terrakube_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

const historyPath = "/api/v1/organization/org-1/workspace/ws-1/history"

// serveHistory answers the history list with entries whose Output is under base.
func serveHistory(t *testing.T, srv *testutil.Server, entries []*terrakube.History) {
	t.Helper()
	srv.HandleFunc("GET "+historyPath, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("sort"); got != "-createdDate" {
			t.Errorf("sort = %q, want %q", got, "-createdDate")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, entries)
	})
}

func at(day int) *time.Time {
	d := time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC)
	return &d
}

func TestOutputService_List_JSONState(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	// The server's order is not trusted: the newest entry is picked by createdDate.
	serveHistory(t, srv, []*terrakube.History{
		{ID: "h-old", Output: srv.URL + stateDir + "old.json", CreatedDate: at(1)},
		{ID: "h-new", Output: srv.URL + stateDir + "new.json", CreatedDate: at(3)},
		{ID: "h-mid", Output: srv.URL + stateDir + "mid.json", CreatedDate: at(2)},
	})
	srv.HandleFunc("GET "+stateDir+"new.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"format_version":"1.0","values":{"outputs":{
			"vpc":{"value":{"id":"vpc-1","subnets":["a","b"],"count":12345678901234567890},"type":["object",{"id":"string"}],"sensitive":false},
			"password":{"value":"s3cret","type":"string","sensitive":true},
			"az":{"value":"eu-1","type":"string"}
		}}}`)
	})
	srv.HandleFunc("GET "+stateDir+"terraform.tfstate", func(http.ResponseWriter, *http.Request) {
		t.Error("raw state read although the JSON state has outputs")
	})

	outputs, err := newTestClient(t, srv).Outputs.List(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 3 {
		t.Fatalf("got %d outputs, want 3", len(outputs))
	}
	names := []string{outputs[0].Name, outputs[1].Name, outputs[2].Name}
	if !reflect.DeepEqual(names, []string{"az", "password", "vpc"}) {
		t.Errorf("names = %v, want sorted", names)
	}
	if !outputs[1].Sensitive || outputs[1].Value != "s3cret" {
		t.Errorf("password = %+v, want sensitive with its value", outputs[1])
	}
	if outputs[0].Sensitive {
		t.Errorf("az.Sensitive = true, want false")
	}
	vpc, ok := outputs[2].Value.(map[string]any)
	if !ok {
		t.Fatalf("vpc value is %T, want map", outputs[2].Value)
	}
	if vpc["id"] != "vpc-1" || !reflect.DeepEqual(vpc["subnets"], []any{"a", "b"}) {
		t.Errorf("vpc value = %v", vpc)
	}
	if n, ok := vpc["count"].(json.Number); !ok || n.String() != "12345678901234567890" {
		t.Errorf("count = %#v, want the exact json.Number", vpc["count"])
	}
	if !reflect.DeepEqual(outputs[2].Type, []any{"object", map[string]any{"id": "string"}}) {
		t.Errorf("vpc type = %#v", outputs[2].Type)
	}
}

func TestOutputService_List_FallbackToRawState(t *testing.T) {
	t.Parallel()

	for name, jsonState := range map[string]string{
		"no outputs":  `{"format_version":"1.0","values":{"root_module":{}}}`,
		"empty state": ``,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			srv := testutil.NewServer(t)
			serveHistory(t, srv, []*terrakube.History{{ID: "h-1", Output: srv.URL + stateDir + "abc.json", CreatedDate: at(1)}})
			srv.HandleFunc("GET "+stateDir+"abc.json", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, jsonState)
			})
			srv.HandleFunc("GET "+stateDir+"terraform.tfstate", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"version":4,"outputs":{"token":{"value":"t","type":"string","sensitive":true},"n":{"value":3,"type":"number"}}}`)
			})

			outputs, err := newTestClient(t, srv).Outputs.List(context.Background(), "org-1", "ws-1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(outputs) != 2 || outputs[0].Name != "n" || outputs[1].Name != "token" {
				t.Fatalf("outputs = %+v", outputs)
			}
			if outputs[0].Value != json.Number("3") || outputs[0].Type != "number" || outputs[0].Sensitive {
				t.Errorf("n = %+v", outputs[0])
			}
			if !outputs[1].Sensitive {
				t.Error("token.Sensitive = false, want true")
			}
		})
	}
}

func TestOutputService_List_NoHistory(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	serveHistory(t, srv, []*terrakube.History{})

	outputs, err := newTestClient(t, srv).Outputs.List(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outputs == nil || len(outputs) != 0 {
		t.Errorf("outputs = %#v, want an empty, non-nil slice", outputs)
	}
}

func TestOutputService_List_NoState(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	serveHistory(t, srv, []*terrakube.History{{ID: "h-1", Output: srv.URL + stateDir + "abc.json"}})
	srv.HandleFunc("GET "+stateDir+"abc.json", func(http.ResponseWriter, *http.Request) {})
	srv.HandleFunc("GET "+stateDir+"terraform.tfstate", func(http.ResponseWriter, *http.Request) {})

	outputs, err := newTestClient(t, srv).Outputs.List(context.Background(), "org-1", "ws-1")
	if err != nil || len(outputs) != 0 {
		t.Errorf("outputs = %v, err = %v; want none and no error", outputs, err)
	}
}

func TestOutputService_List_Errors(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	client := newTestClient(t, srv)

	_, err := client.Outputs.List(context.Background(), "", "ws-1")
	assertValidationError(t, err, "organizationID")
	_, err = client.Outputs.List(context.Background(), "org-1", "")
	assertValidationError(t, err, "workspaceID")

	bad := testutil.NewServer(t)
	serveHistory(t, bad, []*terrakube.History{{ID: "h-1", Output: bad.URL + stateDir + "abc.json"}})
	bad.HandleFunc("GET "+stateDir+"abc.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "not json")
	})
	if _, err := newTestClient(t, bad).Outputs.List(context.Background(), "org-1", "ws-1"); err == nil {
		t.Error("invalid state: err = nil")
	}

	denied := testutil.NewServer(t)
	denied.HandleFunc("GET "+historyPath, func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusForbidden, "denied")
	})
	if _, err := newTestClient(t, denied).Outputs.List(context.Background(), "org-1", "ws-1"); !terrakube.IsForbidden(err) {
		t.Errorf("history 403: err = %v", err)
	}
}

func TestOutput_JSONKeys(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(&terrakube.Output{Name: "n", Value: 1, Type: "number", Sensitive: false})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"name":"n","value":1,"type":"number","sensitive":false}`; got != want {
		t.Errorf("json = %s, want %s", got, want)
	}
}

func TestOutputService_List_StateOutsideAPI(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	// A state URL that is neither on the endpoint nor an API state path: nothing
	// is sent there, and the outputs come from the raw state read through the API.
	serveHistory(t, srv, []*terrakube.History{{ID: "h-1", Output: "https://bucket.example.com/states/abc.json", CreatedDate: at(1)}})
	srv.HandleFunc("GET "+stateDir+"terraform.tfstate", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"outputs":{"a":{"value":"1","type":"string"}}}`)
	})

	outputs, err := newTestClient(t, srv).Outputs.List(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 1 || outputs[0].Name != "a" {
		t.Errorf("outputs = %+v, want the raw state's", outputs)
	}
}

func TestOutputService_List_InClusterStateURL(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	serveHistory(t, srv, []*terrakube.History{{ID: "h-1", Output: "http://terrakube-api:8080" + stateDir + "abc.json", CreatedDate: at(1)}})
	srv.HandleFunc("GET "+stateDir+"abc.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"values":{"outputs":{"b":{"value":true,"type":"bool"}}}}`)
	})

	outputs, err := newTestClient(t, srv).Outputs.List(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 1 || outputs[0].Name != "b" || outputs[0].Value != true {
		t.Errorf("outputs = %+v, want the JSON state's", outputs)
	}
}

func TestOutputService_List_SkipsEntriesWithoutOutput(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	serveHistory(t, srv, []*terrakube.History{
		{ID: "h-old", Output: srv.URL + stateDir + "old.json", CreatedDate: at(1)},
		{ID: "h-new", Output: "", CreatedDate: at(5)}, // saved before its state file
	})
	srv.HandleFunc("GET "+stateDir+"old.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"values":{"outputs":{"old":{"value":1,"type":"number"}}}}`)
	})

	outputs, err := newTestClient(t, srv).Outputs.List(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 1 || outputs[0].Name != "old" {
		t.Errorf("outputs = %+v, want the newest entry with a state file", outputs)
	}

	// With no entry naming a state file, the raw state is read.
	only := testutil.NewServer(t)
	serveHistory(t, only, []*terrakube.History{{ID: "h-1", Output: "", CreatedDate: at(1)}})
	only.HandleFunc("GET "+stateDir+"terraform.tfstate", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"outputs":{"raw":{"value":"r","type":"string"}}}`)
	})
	outputs, err = newTestClient(t, only).Outputs.List(context.Background(), "org-1", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 1 || outputs[0].Name != "raw" {
		t.Errorf("outputs = %+v, want the raw state's", outputs)
	}
}
