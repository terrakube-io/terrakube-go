package terrakube_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/denniswebb/terrakube-go/testutil"
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
				{Data: map[string]interface{}{"type": "workspace", "id": "ws-1"}},
				{Data: map[string]interface{}{"type": "variable", "id": "var-1"}},
			},
		})
	})

	c := newTestClient(t, srv)

	resp, err := c.Operations.Submit(context.Background(), &terrakube.AtomicRequest{
		Operations: []terrakube.Operation{
			{
				Op:  terrakube.OperationAdd,
				Ref: terrakube.OperationRef{Type: "workspace"},
				Data: map[string]interface{}{
					"attributes": map[string]interface{}{"name": "ws-new"},
				},
			},
			{
				Op:  terrakube.OperationUpdate,
				Ref: terrakube.OperationRef{Type: "variable", ID: "var-1"},
				Data: map[string]interface{}{
					"attributes": map[string]interface{}{"value": "updated"},
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
