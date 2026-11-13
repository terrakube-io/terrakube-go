package terrakube_test

import (
	"context"
	"encoding/json"
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
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
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

// TestOperation_ZeroRefIsSerialized pins that "ref" is always sent, even for a
// zero OperationRef: encoding/json ignores omitempty on struct-typed fields.
func TestOperation_ZeroRefIsSerialized(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(terrakube.Operation{Op: terrakube.OperationAdd, Href: "/x"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ref, ok := got["ref"].(map[string]any)
	if !ok {
		t.Fatalf("ref missing or not an object in %s", b)
	}
	if typ, ok := ref["type"]; !ok || typ != "" {
		t.Errorf("ref = %v, want {\"type\":\"\"}", ref)
	}
}
