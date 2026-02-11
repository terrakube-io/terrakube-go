package terrakube

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/denniswebb/terrakube-go/testutil"
)

// newCrudTestClient creates a Client for use in internal (package terrakube) tests.
func newCrudTestClient(t *testing.T, srv *testutil.Server) *Client {
	t.Helper()
	client, err := NewClient(WithEndpoint(srv.URL), WithToken("test-token"))
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}

func TestCrudService_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		opts      *ListOptions
		handler   http.HandlerFunc
		wantCount int
		wantName  string
		wantErr   bool
	}{
		{
			name: "success with results",
			opts: nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteJSONAPIList(t, w, http.StatusOK, []*Organization{
					{ID: "org-1", Name: "Alpha"},
					{ID: "org-2", Name: "Beta"},
				})
			},
			wantCount: 2,
			wantName:  "Alpha",
		},
		{
			name: "success with empty result",
			opts: nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteJSONAPIList(t, w, http.StatusOK, []*Organization{})
			},
			wantCount: 0,
		},
		{
			name: "with filter param",
			opts: &ListOptions{Filter: "name==Filtered"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				filter := r.URL.Query().Get("filter")
				if filter != "name==Filtered" {
					t.Errorf("filter = %q, want %q", filter, "name==Filtered")
				}
				testutil.WriteJSONAPIList(t, w, http.StatusOK, []*Organization{
					{ID: "org-1", Name: "Filtered"},
				})
			},
			wantCount: 1,
			wantName:  "Filtered",
		},
		{
			name: "without filter nil opts",
			opts: nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RawQuery != "" {
					t.Errorf("expected no query params, got %q", r.URL.RawQuery)
				}
				testutil.WriteJSONAPIList(t, w, http.StatusOK, []*Organization{
					{ID: "org-1", Name: "NoFilter"},
				})
			},
			wantCount: 1,
			wantName:  "NoFilter",
		},
		{
			name: "without filter empty filter string",
			opts: &ListOptions{Filter: ""},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RawQuery != "" {
					t.Errorf("expected no query params for empty filter, got %q", r.URL.RawQuery)
				}
				testutil.WriteJSONAPIList(t, w, http.StatusOK, []*Organization{
					{ID: "org-1", Name: "EmptyFilter"},
				})
			},
			wantCount: 1,
			wantName:  "EmptyFilter",
		},
		{
			name: "server error",
			opts: nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := testutil.NewServer(t)
			srv.HandleFunc("GET /api/v1/organization", tt.handler)

			client := newCrudTestClient(t, srv)
			svc := &crudService[Organization]{client: client}

			items, err := svc.list(context.Background(), client.apiPath("organization"), tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != tt.wantCount {
				t.Fatalf("got %d items, want %d", len(items), tt.wantCount)
			}
			if tt.wantName != "" && len(items) > 0 {
				if items[0].Name != tt.wantName {
					t.Errorf("Name = %q, want %q", items[0].Name, tt.wantName)
				}
			}
		})
	}
}

func TestCrudService_List_ErrorType(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusInternalServerError, "something broke")
	})

	client := newCrudTestClient(t, srv)
	svc := &crudService[Organization]{client: client}

	_, err := svc.list(context.Background(), client.apiPath("organization"), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func TestCrudService_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantID  string
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				desc := "Test org"
				testutil.WriteJSONAPI(t, w, http.StatusOK, &Organization{
					ID: "org-1", Name: "Alpha", Description: &desc, ExecutionMode: "remote",
				})
			},
			wantID: "org-1",
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteError(t, w, http.StatusNotFound, "organization not found")
			},
			wantErr: true,
			errCode: http.StatusNotFound,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteError(t, w, http.StatusInternalServerError, "internal error")
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := testutil.NewServer(t)
			srv.HandleFunc("GET /api/v1/organization/org-1", tt.handler)

			client := newCrudTestClient(t, srv)
			svc := &crudService[Organization]{client: client}

			result, err := svc.get(context.Background(), client.apiPath("organization", "org-1"))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.errCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.errCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", result.ID, tt.wantID)
			}
		})
	}
}

func TestCrudService_Get_FieldValues(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	desc := "Full details"
	icon := "icon-url"
	srv.HandleFunc("GET /api/v1/organization/org-full", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &Organization{
			ID:            "org-full",
			Name:          "FullOrg",
			Description:   &desc,
			ExecutionMode: "remote",
			Disabled:      false,
			Icon:          &icon,
		})
	})

	client := newCrudTestClient(t, srv)
	svc := &crudService[Organization]{client: client}

	result, err := svc.get(context.Background(), client.apiPath("organization", "org-full"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "FullOrg" {
		t.Errorf("Name = %q, want %q", result.Name, "FullOrg")
	}
	if result.ExecutionMode != "remote" {
		t.Errorf("ExecutionMode = %q, want %q", result.ExecutionMode, "remote")
	}
	if result.Description == nil || *result.Description != "Full details" {
		t.Errorf("Description = %v, want %q", result.Description, "Full details")
	}
	if result.Icon == nil || *result.Icon != "icon-url" {
		t.Errorf("Icon = %v, want %q", result.Icon, "icon-url")
	}
}

func TestCrudService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   *Organization
		handler http.HandlerFunc
		wantID  string
		wantErr bool
		errCode int
	}{
		{
			name:  "success",
			input: &Organization{Name: "NewOrg", ExecutionMode: "remote"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Content-Type") != mediaType {
					t.Errorf("Content-Type = %q, want %q", r.Header.Get("Content-Type"), mediaType)
				}
				testutil.WriteJSONAPI(t, w, http.StatusCreated, &Organization{
					ID: "org-new", Name: "NewOrg", ExecutionMode: "remote",
				})
			},
			wantID: "org-new",
		},
		{
			name:  "bad request",
			input: &Organization{Name: ""},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteError(t, w, http.StatusBadRequest, "name is required")
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := testutil.NewServer(t)
			srv.HandleFunc("POST /api/v1/organization", tt.handler)

			client := newCrudTestClient(t, srv)
			svc := &crudService[Organization]{client: client}

			result, err := svc.create(context.Background(), client.apiPath("organization"), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.errCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.errCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", result.ID, tt.wantID)
			}
			if result.Name != tt.input.Name {
				t.Errorf("Name = %q, want %q", result.Name, tt.input.Name)
			}
		})
	}
}

func TestCrudService_Create_ReturnedEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		desc := "Created desc"
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &Organization{
			ID:            "org-created",
			Name:          "CreatedOrg",
			Description:   &desc,
			ExecutionMode: "local",
		})
	})

	client := newCrudTestClient(t, srv)
	svc := &crudService[Organization]{client: client}

	result, err := svc.create(context.Background(), client.apiPath("organization"), &Organization{
		Name:          "CreatedOrg",
		ExecutionMode: "local",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "org-created" {
		t.Errorf("ID = %q, want %q", result.ID, "org-created")
	}
	if result.ExecutionMode != "local" {
		t.Errorf("ExecutionMode = %q, want %q", result.ExecutionMode, "local")
	}
	if result.Description == nil || *result.Description != "Created desc" {
		t.Errorf("Description = %v, want %q", result.Description, "Created desc")
	}
}

func TestCrudService_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   *Organization
		handler http.HandlerFunc
		wantErr bool
		errCode int
	}{
		{
			name:  "success",
			input: &Organization{ID: "org-1", Name: "Updated", ExecutionMode: "local"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("Method = %q, want %q", r.Method, http.MethodPatch)
				}
				testutil.WriteJSONAPI(t, w, http.StatusOK, &Organization{
					ID: "org-1", Name: "Updated", ExecutionMode: "local",
				})
			},
		},
		{
			name:  "server error",
			input: &Organization{ID: "org-1", Name: "Fail"},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteError(t, w, http.StatusInternalServerError, "update failed")
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := testutil.NewServer(t)
			srv.HandleFunc("PATCH /api/v1/organization/org-1", tt.handler)

			client := newCrudTestClient(t, srv)
			svc := &crudService[Organization]{client: client}

			result, err := svc.update(context.Background(), client.apiPath("organization", "org-1"), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.errCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.errCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != "Updated" {
				t.Errorf("Name = %q, want %q", result.Name, "Updated")
			}
			if result.ExecutionMode != "local" {
				t.Errorf("ExecutionMode = %q, want %q", result.ExecutionMode, "local")
			}
		})
	}
}

func TestCrudService_Update_ReturnedEntity(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	desc := "Updated desc"
	srv.HandleFunc("PATCH /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &Organization{
			ID:            "org-1",
			Name:          "UpdatedOrg",
			Description:   &desc,
			ExecutionMode: "remote",
		})
	})

	client := newCrudTestClient(t, srv)
	svc := &crudService[Organization]{client: client}

	result, err := svc.update(context.Background(), client.apiPath("organization", "org-1"), &Organization{
		ID:   "org-1",
		Name: "UpdatedOrg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "org-1" {
		t.Errorf("ID = %q, want %q", result.ID, "org-1")
	}
	if result.Description == nil || *result.Description != "Updated desc" {
		t.Errorf("Description = %v, want %q", result.Description, "Updated desc")
	}
}

func TestCrudService_Del(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
		errCode int
	}{
		{
			name: "success no content",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteError(t, w, http.StatusNotFound, "not found")
			},
			wantErr: true,
			errCode: http.StatusNotFound,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteError(t, w, http.StatusInternalServerError, "delete failed")
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := testutil.NewServer(t)
			srv.HandleFunc("DELETE /api/v1/organization/org-1", tt.handler)

			client := newCrudTestClient(t, srv)
			svc := &crudService[Organization]{client: client}

			err := svc.del(context.Background(), client.apiPath("organization", "org-1"))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.errCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.errCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCrudService_Del_MethodVerification(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/organization/org-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %q, want %q", r.Method, http.MethodDelete)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	client := newCrudTestClient(t, srv)
	svc := &crudService[Organization]{client: client}

	err := svc.del(context.Background(), client.apiPath("organization", "org-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
