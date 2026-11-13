package terrakube_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestOperationsService_Submit(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/operations", func(w http.ResponseWriter, r *http.Request) {
		const atomic = `application/vnd.api+json; ext="https://jsonapi.org/ext/atomic"`
		if ct := r.Header.Get("Content-Type"); ct != atomic {
			t.Errorf("Content-Type = %q, want %q", ct, atomic)
		}
		if got := r.Header.Get("Accept"); got != atomic {
			t.Errorf("Accept = %q, want %q", got, atomic)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var req terrakube.AtomicRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}
		if len(req.Operations) != 2 {
			t.Fatalf("got %d operations, want 2", len(req.Operations))
		}
		if req.Operations[0].Op != terrakube.OperationAdd {
			t.Errorf("operations[0].Op = %q, want %q", req.Operations[0].Op, terrakube.OperationAdd)
		}
		if req.Operations[1].Op != terrakube.OperationUpdate {
			t.Errorf("operations[1].Op = %q, want %q", req.Operations[1].Op, terrakube.OperationUpdate)
		}

		testutil.WriteJSON(t, w, http.StatusOK, &terrakube.AtomicResponse{
			Results: []terrakube.AtomicResult{
				{Data: map[string]any{"type": "workspace", "id": "ws-1"}},
				{Data: map[string]any{"type": "variable", "id": "var-1"}},
			},
		})
	})

	c := newTestClient(t, srv)

	resp, err := c.Operations.Submit(context.Background(), &terrakube.AtomicRequest{
		Operations: []terrakube.Operation{
			{
				Op:  terrakube.OperationAdd,
				Ref: terrakube.OperationRef{Type: "workspace"},
				Data: map[string]any{
					"attributes": map[string]any{"name": "ws-new"},
				},
			},
			{
				Op:  terrakube.OperationUpdate,
				Ref: terrakube.OperationRef{Type: "variable", ID: "var-1"},
				Data: map[string]any{
					"attributes": map[string]any{"value": "updated"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 2 {
		t.Fatalf("got %d results, want 2", len(resp.Results))
	}
	if resp.Results[0].Data["id"] != "ws-1" {
		t.Errorf("results[0].Data[id] = %v, want %q", resp.Results[0].Data["id"], "ws-1")
	}
	if resp.Results[1].Data["id"] != "var-1" {
		t.Errorf("results[1].Data[id] = %v, want %q", resp.Results[1].Data["id"], "var-1")
	}
}

func TestOperationsService_Submit_EmptyOperations(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/operations", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSON(t, w, http.StatusOK, &terrakube.AtomicResponse{
			Results: []terrakube.AtomicResult{},
		})
	})

	c := newTestClient(t, srv)

	resp, err := c.Operations.Submit(context.Background(), &terrakube.AtomicRequest{
		Operations: []terrakube.Operation{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 0 {
		t.Fatalf("got %d results, want 0", len(resp.Results))
	}
}

func TestOperationsService_Submit_ServerError(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /api/v1/operations", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := newTestClient(t, srv)

	_, err := c.Operations.Submit(context.Background(), &terrakube.AtomicRequest{
		Operations: []terrakube.Operation{
			{
				Op:  terrakube.OperationRemove,
				Ref: terrakube.OperationRef{Type: "workspace", ID: "ws-1"},
			},
		},
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// TestOperation_ZeroRefIsOmitted: the server rejects an operation that has both
// "ref" and "href", so an unset Ref must not be serialized.
func TestOperation_ZeroRefIsOmitted(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(terrakube.Operation{Op: terrakube.OperationRemove, Href: "/organization/o1"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if want := `{"op":"remove","href":"/organization/o1"}`; string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}

	b, err = json.Marshal(terrakube.Operation{Op: terrakube.OperationRemove, Ref: terrakube.OperationRef{Type: "organization", ID: "o1"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if want := `{"op":"remove","ref":{"type":"organization","id":"o1"}}`; string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}

// TestOperations_Submit_ErrorArray: the server reports operation errors as an
// array of error objects; their details must reach APIError.
func TestOperations_Submit_ErrorArray(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/operations", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`[{"errors":[{"detail":"Unknown identifier x for organization","status":"404"}]}]`))
	})

	client := newTestClient(t, srv)
	_, err := client.Operations.Submit(context.Background(), &terrakube.AtomicRequest{})
	apiErr, ok := errors.AsType[*terrakube.APIError](err)
	if !ok {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if len(apiErr.Errors) != 1 || apiErr.Errors[0].Detail != "Unknown identifier x for organization" {
		t.Errorf("Errors = %+v, want the server's detail", apiErr.Errors)
	}
}
