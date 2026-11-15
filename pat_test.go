package terrakube_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

// springNotFound is Spring's default error body, as the server sends it for a
// route it does not have and for a ResponseStatusException without a message.
const springNotFound = `{"timestamp":"2026-11-15T10:00:00.000+00:00","status":404,"error":"Not Found","path":"/x"}`

// fakeJWT builds an unsigned token whose payload is claims.
func fakeJWT(claims string) string {
	enc := base64.RawURLEncoding
	return enc.EncodeToString([]byte(`{"alg":"HS256"}`)) + "." + enc.EncodeToString([]byte(claims)) + ".c2ln"
}

func TestPATService_Create(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)
	token := fakeJWT(`{"iss":"Terrakube","jti":"5b8c7d1e-0000-4000-8000-000000000001","name":"Jane (Token)"}`)

	srv.HandleFunc("POST /pat/v1", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", auth)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decoding body %s: %v", body, err)
		}
		if len(got) != 2 || got["days"] != float64(30) || got["description"] != "ci" {
			t.Errorf("body = %s, want {days: 30, description: ci}", body)
		}
		testutil.WriteJSON(t, w, http.StatusCreated, map[string]string{"token": token})
	})

	got, err := newTestClient(t, srv).PATs.Create(context.Background(), 30, "ci")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Token != token {
		t.Errorf("Token = %q, want %q", got.Token, token)
	}
	if got.ID != "5b8c7d1e-0000-4000-8000-000000000001" {
		t.Errorf("ID = %q, want the token's jti", got.ID)
	}
}

func TestPATService_Create_NoExpiry(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("POST /pat/v1", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if days, ok := got["days"]; !ok || days != float64(0) {
			t.Errorf("body = %s, want days 0 sent explicitly", body)
		}
		testutil.WriteJSON(t, w, http.StatusCreated, map[string]string{"token": "not-a-jwt"})
	})

	got, err := newTestClient(t, srv).PATs.Create(context.Background(), 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Token != "not-a-jwt" || got.ID != "" {
		t.Errorf("got %+v, want the token and no ID", got)
	}
}

func TestPATService_Create_TokenID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, token, want string
	}{
		{"jti", fakeJWT(`{"jti":"abc"}`), "abc"},
		{"no jti", fakeJWT(`{"sub":"x"}`), ""},
		{"payload not base64", "a.%%%.c", ""},
		{"payload not JSON", "a." + base64.RawURLEncoding.EncodeToString([]byte("nope")) + ".c", ""},
		{"jti not a string", fakeJWT(`{"jti":7}`), ""},
		{"two parts", "a.b", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := testutil.NewServer(t)
			srv.HandleFunc("POST /pat/v1", func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteJSON(t, w, http.StatusCreated, map[string]string{"token": tt.token})
			})
			got, err := newTestClient(t, srv).PATs.Create(context.Background(), 1, "x")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.want {
				t.Errorf("ID = %q, want %q", got.ID, tt.want)
			}
		})
	}
}

func TestPATService_Create_Errors(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)
	status := map[string]int{"forbidden": http.StatusForbidden, "too-long": http.StatusBadRequest, "unauth": http.StatusUnauthorized}

	srv.HandleFunc("POST /pat/v1", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Description string `json:"description"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		switch in.Description {
		case "empty":
			// The server answers 201 with an empty token when signing fails.
			testutil.WriteJSON(t, w, http.StatusCreated, map[string]string{"token": ""})
		case "garbage":
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, "<html>")
		default:
			w.WriteHeader(status[in.Description])
		}
	})
	client := newTestClient(t, srv)
	ctx := context.Background()

	_, err := client.PATs.Create(ctx, -1, "x")
	ve, ok := errors.AsType[*terrakube.ValidationError](err)
	if !ok || ve.Field != "days" {
		t.Errorf("negative days: err = %v, want a *ValidationError for days", err)
	}

	if _, err := client.PATs.Create(ctx, 1, "forbidden"); !terrakube.IsForbidden(err) {
		t.Errorf("forbidden: err = %v, want 403", err)
	}
	if _, err := client.PATs.Create(ctx, 1, "unauth"); !terrakube.IsUnauthorized(err) {
		t.Errorf("unauth: err = %v, want 401", err)
	}
	apiErr, ok := errors.AsType[*terrakube.APIError](func() error { _, err := client.PATs.Create(ctx, 1, "too-long"); return err }())
	if !ok || apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("too-long: err = %v, want 400", apiErr)
	}
	if _, err := client.PATs.Create(ctx, 1, "empty"); err == nil {
		t.Error("empty token: err = nil, want an error")
	}
	if _, err := client.PATs.Create(ctx, 1, "garbage"); err == nil {
		t.Error("garbage body: err = nil, want an error")
	}
}

func TestPATService_List(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)

	srv.HandleFunc("GET /pat/v1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted) // what the server answers
		_, _ = io.WriteString(w, `[
			{"id":"p-1","days":30,"deleted":false,"description":"ci","source":"API",
			 "lastUsedAt":"2026-11-14T08:30:00.000+00:00",
			 "createdDate":"2026-11-01T10:00:00.000+00:00","createdBy":"jane@example.com",
			 "updatedDate":"2026-11-01T10:00:00.000+00:00","updatedBy":"jane@example.com"},
			{"id":"p-2","days":0,"deleted":false,"description":null,"source":"CLI_LOGIN",
			 "lastUsedAt":null,"createdDate":null,"createdBy":null,"updatedDate":null,"updatedBy":null}
		]`)
	})

	pats, err := newTestClient(t, srv).PATs.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pats) != 2 {
		t.Fatalf("got %d PATs, want 2", len(pats))
	}
	p := pats[0]
	created := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	if p.ID != "p-1" || p.Days != 30 || p.Deleted || p.Description != "ci" || p.Source != "API" {
		t.Errorf("pats[0] = %+v", p)
	}
	if p.CreatedDate == nil || !p.CreatedDate.Equal(created) || p.UpdatedDate == nil || !p.UpdatedDate.Equal(created) {
		t.Errorf("pats[0] dates = %v, %v, want %v", p.CreatedDate, p.UpdatedDate, created)
	}
	if p.LastUsedAt == nil || !p.LastUsedAt.Equal(time.Date(2026, 11, 14, 8, 30, 0, 0, time.UTC)) {
		t.Errorf("pats[0].LastUsedAt = %v", p.LastUsedAt)
	}
	if p.CreatedBy == nil || *p.CreatedBy != "jane@example.com" || p.UpdatedBy == nil || *p.UpdatedBy != "jane@example.com" {
		t.Errorf("pats[0] authors = %v, %v", p.CreatedBy, p.UpdatedBy)
	}
	q := pats[1]
	if q.Days != 0 || q.Source != "CLI_LOGIN" || q.Description != "" || q.LastUsedAt != nil || q.CreatedDate != nil || q.CreatedBy != nil {
		t.Errorf("pats[1] = %+v", q)
	}
}

func TestPATService_List_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		body   string
		check  func(error) bool
	}{
		{"federated identity", http.StatusForbidden, "", terrakube.IsForbidden},
		{"unauthenticated", http.StatusUnauthorized, "", terrakube.IsUnauthorized},
		{"old server", http.StatusNotFound, springNotFound, func(err error) bool { return errors.Is(err, terrakube.ErrUnsupported) }},
		{"bad body", http.StatusAccepted, "{", func(err error) bool { return err != nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := testutil.NewServer(t)
			srv.HandleFunc("GET /pat/v1", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			})
			_, err := newTestClient(t, srv).PATs.List(context.Background())
			if !tt.check(err) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestPATService_Delete(t *testing.T) {
	t.Parallel()
	srv := testutil.NewServer(t)
	called := false

	srv.HandleFunc("DELETE /pat/v1/p-1", func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	})
	srv.HandleFunc("DELETE /pat/v1/missing", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, springNotFound)
	})
	srv.HandleFunc("DELETE /pat/v1/other", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	client := newTestClient(t, srv)
	ctx := context.Background()

	if err := client.PATs.Delete(ctx, "p-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("DELETE was not sent")
	}
	err := client.PATs.Delete(ctx, "missing")
	if !terrakube.IsNotFound(err) || errors.Is(err, terrakube.ErrUnsupported) {
		t.Errorf("missing: err = %v, want a 404 that is not ErrUnsupported", err)
	}
	if err := client.PATs.Delete(ctx, "other"); !terrakube.IsForbidden(err) {
		t.Errorf("other user's token: err = %v, want 403", err)
	}
	assertValidationError(t, client.PATs.Delete(ctx, ""), "id")
	if _, ok := errors.AsType[*terrakube.ValidationError](client.PATs.Delete(ctx, "a/b")); !ok {
		t.Error(`id "a/b": want a *ValidationError`)
	}
}
