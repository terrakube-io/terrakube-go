package terrakube

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/terrakube-io/terrakube-go/testutil"
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

	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
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
				testutil.WriteJSONAPI(t, w, http.StatusOK, &Organization{
					ID: "org-1", Name: "Alpha", Description: new("Test org"), ExecutionMode: "remote",
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
				apiErr, ok := errors.AsType[*APIError](err)
				if !ok {
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
	icon := "icon-url"
	srv.HandleFunc("GET /api/v1/organization/org-full", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &Organization{
			ID:            "org-full",
			Name:          "FullOrg",
			Description:   new("Full details"),
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
				apiErr, ok := errors.AsType[*APIError](err)
				if !ok {
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
				apiErr, ok := errors.AsType[*APIError](err)
				if !ok {
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
				apiErr, ok := errors.AsType[*APIError](err)
				if !ok {
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

func TestNewClient_FilterKeys(t *testing.T) {
	t.Parallel()

	// Expected filterKey for every crudService field on Client.
	// Empty string means the service uses the default "filter" key (top-level resources).
	// Non-empty means the service must use that exact "filter[type]" key (nested resources).
	expected := map[string]string{
		"Organizations":              "",
		"Workspaces":                 "filter[workspace]",
		"Modules":                    "filter[module]",
		"Teams":                      "filter[team]",
		"Variables":                  "filter[variable]",
		"OrganizationVariables":      "filter[globalvar]",
		"Templates":                  "filter[template]",
		"Tags":                       "filter[tag]",
		"VCS":                        "filter[vcs]",
		"SSH":                        "filter[ssh]",
		"Agents":                     "filter[agent]",
		"Collections":                "filter[collection]",
		"CollectionItems":            "filter[item]",
		"CollectionReferences":       "filter[reference]",
		"WorkspaceTags":              "filter[workspacetag]",
		"WorkspaceAccess":            "filter[access]",
		"WorkspaceSchedules":         "filter[schedule]",
		"Webhooks":                   "filter[webhook]",
		"WebhookEvents":              "filter[webhook_event]",
		"History":                    "filter[history]",
		"Jobs":                       "filter[job]",
		"Actions":                    "",
		"Steps":                      "filter[step]",
		"Providers":                  "filter[provider]",
		"ProviderVersions":           "filter[version]",
		"Implementations":            "filter[implementation]",
		"ModuleVersions":             "filter[module_version]",
		"GithubAppTokens":            "",
		"Addresses":                  "filter[address]",
		"Projects":                   "filter[project]",
		"ProjectAccess":              "filter[project_access]",
		"Federated":                  "filter[federated]",
		"FederatedClaims":            "filter[federated_claim]",
		"NotificationConfigurations": "filter[notification_configuration]",
		"NotificationTriggers":       "filter[notification_trigger]",
		"PolicyAttachments":          "filter[policy_attachment]",
		"PolicyEvaluations":          "filter[policy_evaluation]",
		"PolicyExemptions":           "filter[policy_exemption]",
		"PolicyOverrides":            "filter[policy_override]",
		"PolicySets":                 "filter[policy_set]",
		"PolicySetParameters":        "filter[policy_set_parameter]",
		"RunTriggers":                "filter[runTrigger]",
	}

	// Services that don't embed crudService (non-JSON:API).
	skipFields := map[string]bool{
		"TeamTokens": true,
		"Operations": true,
	}

	// Private fields on Client that aren't services.
	privateFields := map[string]bool{
		"baseURL":     true,
		"token":       true,
		"httpClient":  true,
		"userAgent":   true,
		"insecureTLS": true,
	}

	client, err := NewClient(WithEndpoint("https://example.com"), WithToken("test"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	v := reflect.ValueOf(client).Elem()
	typ := v.Type()

	for field, fieldValue := range v.Fields() {
		if privateFields[field.Name] || skipFields[field.Name] {
			continue
		}

		wantKey, ok := expected[field.Name]
		if !ok {
			t.Errorf("Client.%s: not in expected filterKey map — add it to TestNewClient_FilterKeys", field.Name)
			continue
		}

		svc := fieldValue.Elem()
		fk := svc.FieldByName("filterKey")
		if !fk.IsValid() {
			t.Errorf("Client.%s: no filterKey field found (missing crudService embed?)", field.Name)
			continue
		}

		if got := fk.String(); got != wantKey {
			t.Errorf("Client.%s filterKey = %q, want %q", field.Name, got, wantKey)
		}
	}

	// Detect stale entries in expected that no longer exist on Client.
	for name := range expected {
		if _, ok := typ.FieldByName(name); !ok {
			t.Errorf("expected map has %q but Client has no such field — remove it from TestNewClient_FilterKeys", name)
		}
	}
}

func TestCrudService_List_CustomFilterKey(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[workspace]"); got != "name==Test" {
			t.Errorf("filter[workspace] = %q, want %q", got, "name==Test")
		}
		if got := r.URL.Query().Get("filter"); got != "" {
			t.Errorf("unexpected plain 'filter' param = %q", got)
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*Organization{
			{ID: "org-1", Name: "Test"},
		})
	})

	client := newCrudTestClient(t, srv)
	svc := &crudService[Organization]{client: client, filterKey: "filter[workspace]"}

	items, err := svc.list(context.Background(), client.apiPath("organization"), &ListOptions{Filter: "name==Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
}

// TestCrudService_List_PageOptionsUnpagedType: page options on a resource type
// the server cannot paginate fail before any request is sent.
func TestCrudService_List_PageOptionsUnpagedType(t *testing.T) {
	t.Parallel()

	for _, opts := range []*ListOptions{{Page: 2}, {PageSize: 1}, {Page: 2, PageSize: 1}} {
		srv := newRecordedServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(listBody("organization", []string{"o1"}, "")))
		})
		items, err := srv.client(t).Organizations.List(context.Background(), opts)
		if _, ok := errors.AsType[*ValidationError](err); !ok {
			t.Errorf("opts %+v: err = %v, want *ValidationError", *opts, err)
		}
		if items != nil {
			t.Errorf("opts %+v: got %d items, want nil", *opts, len(items))
		}
		wantQueries(t, srv.recorded())
	}
}

// listBody renders a JSON:API list document. extra is appended verbatim inside
// the top-level object (for example `,"meta":…` or `,"links":…`).
func listBody(kind string, ids []string, extra string) string {
	items := make([]string, len(ids))
	for i, id := range ids {
		items[i] = fmt.Sprintf(`{"type":%q,"id":%q,"attributes":{"name":%q}}`, kind, id, id)
	}
	return `{"data":[` + strings.Join(items, ",") + `]` + extra + `}`
}

// pageMeta renders the meta.page member a Terrakube server sends with page[totals].
func pageMeta(number, limit, totalPages int) string {
	return fmt.Sprintf(`,"meta":{"page":{"number":%d,"limit":%d,"totalPages":%d,"totalRecords":%d}}`,
		number, limit, totalPages, limit*totalPages)
}

// recordedServer records the query string of every request it receives.
type recordedServer struct {
	*httptest.Server
	mu      sync.Mutex
	queries []string
}

// newRecordedServer starts a server whose handler receives the 1-based request count.
func newRecordedServer(t *testing.T, handler func(n int, w http.ResponseWriter, r *http.Request)) *recordedServer {
	t.Helper()
	rs := &recordedServer{}
	rs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		rs.mu.Lock()
		rs.queries = append(rs.queries, r.URL.Query().Encode())
		n := len(rs.queries)
		rs.mu.Unlock()
		handler(n, w, r)
	}))
	t.Cleanup(rs.Close)
	return rs
}

func (rs *recordedServer) client(t *testing.T) *Client {
	t.Helper()
	c, err := NewClient(WithEndpoint(rs.URL), WithToken("test-token"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func (rs *recordedServer) recorded() []string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return slices.Clone(rs.queries)
}

func wantQueries(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d requests %q, want %d %q", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("request %d query = %q, want %q", i+1, got[i], want[i])
		}
	}
}

func itemIDs[T any](items []*T, id func(*T) string) []string {
	ids := make([]string, len(items))
	for i, it := range items {
		ids[i] = id(it)
	}
	return ids
}

// TestCrudService_List_IgnoresLinks: links in a response body are never
// requested, for a type the library pages and for one it does not.
func TestCrudService_List_IgnoresLinks(t *testing.T) {
	var foreignHits atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		foreignHits.Add(1)
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(foreign.Close)

	for _, kind := range []string{"organization", "workspace"} {
		t.Run(kind, func(t *testing.T) {
			var srv *recordedServer
			srv = newRecordedServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				links := fmt.Sprintf(`,"links":{"next":"%s/api/v1/x?page%%5Bnumber%%5D=2","self":"%s/api/v1/x"}`, foreign.URL, srv.URL)
				_, _ = w.Write([]byte(listBody(kind, []string{"a"}, links)))
			})
			c := srv.client(t)

			var n int
			var err error
			if kind == "organization" {
				var items []*Organization
				items, err = c.Organizations.List(context.Background(), nil)
				n = len(items)
			} else {
				var items []*Workspace
				items, err = c.Workspaces.List(context.Background(), "org-1", nil)
				n = len(items)
			}
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if n != 1 {
				t.Errorf("got %d items, want 1", n)
			}
			if got := srv.recorded(); len(got) != 1 {
				t.Errorf("configured endpoint got %d requests, want 1", len(got))
			}
		})
	}

	if n := foreignHits.Load(); n != 0 {
		t.Errorf("server named in links.next got %d requests, want 0", n)
	}
}

func TestCrudService_List_AllPages(t *testing.T) {
	t.Parallel()

	const totals = "page%5Btotals%5D=&sort=id"
	pageQ := func(n, size int) string {
		return fmt.Sprintf("page%%5Bnumber%%5D=%d&page%%5Bsize%%5D=%d&%s", n, size, totals)
	}
	write := func(w http.ResponseWriter, body string) { _, _ = w.Write([]byte(body)) }

	tests := []struct {
		name        string
		opts        *ListOptions
		handler     func(n int, w http.ResponseWriter, r *http.Request)
		wantIDs     []string
		wantErr     bool
		wantQueries []string
	}{
		{
			name: "three pages with filter and include",
			opts: &ListOptions{Filter: "name==x", Include: "vcs"},
			handler: func(_ int, w http.ResponseWriter, r *http.Request) {
				pages := map[string][]string{"": {"w1", "w2"}, "2": {"w3", "w4"}, "3": {"w5"}}
				num := r.URL.Query().Get("page[number]")
				p := 1
				if num != "" {
					p, _ = strconv.Atoi(num)
				}
				write(w, listBody("workspace", pages[num], pageMeta(p, 2, 3)))
			},
			wantIDs: []string{"w1", "w2", "w3", "w4", "w5"},
			wantQueries: []string{
				"filter%5Bworkspace%5D=name%3D%3Dx&include=vcs&" + totals,
				"filter%5Bworkspace%5D=name%3D%3Dx&include=vcs&" + pageQ(2, 2),
				"filter%5Bworkspace%5D=name%3D%3Dx&include=vcs&" + pageQ(3, 2),
			},
		},
		{
			name: "one page",
			handler: func(_ int, w http.ResponseWriter, _ *http.Request) {
				write(w, listBody("workspace", []string{"w1"}, pageMeta(1, 10000, 1)))
			},
			wantIDs:     []string{"w1"},
			wantQueries: []string{totals},
		},
		{
			name: "empty collection",
			handler: func(_ int, w http.ResponseWriter, _ *http.Request) {
				write(w, listBody("workspace", nil, pageMeta(1, 500, 0)))
			},
			wantIDs:     []string{},
			wantQueries: []string{totals},
		},
		{
			name: "response without metadata",
			handler: func(_ int, w http.ResponseWriter, _ *http.Request) {
				write(w, listBody("workspace", []string{"w1"}, ""))
			},
			wantIDs:     []string{"w1"},
			wantQueries: []string{totals},
		},
		{
			name: "empty page before the reported last page",
			handler: func(n int, w http.ResponseWriter, _ *http.Request) {
				if n == 1 {
					write(w, listBody("workspace", []string{"w1"}, pageMeta(1, 1, 5)))
					return
				}
				write(w, listBody("workspace", nil, pageMeta(n, 1, 5)))
			},
			wantIDs:     []string{"w1"},
			wantQueries: []string{totals, pageQ(2, 1)},
		},
		{
			name: "page count grows while paging",
			handler: func(n int, w http.ResponseWriter, _ *http.Request) {
				total := 3
				if n == 1 {
					total = 2
				}
				write(w, listBody("workspace", []string{fmt.Sprintf("w%d", n)}, pageMeta(n, 1, total)))
			},
			wantIDs:     []string{"w1", "w2", "w3"},
			wantQueries: []string{totals, pageQ(2, 1), pageQ(3, 1)},
		},
		{
			name: "failure on a later page returns no items",
			handler: func(n int, w http.ResponseWriter, _ *http.Request) {
				if n == 1 {
					write(w, listBody("workspace", []string{"w1"}, pageMeta(1, 1, 3)))
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:     true,
			wantQueries: []string{totals, pageQ(2, 1)},
		},
		{
			name: "server refuses page parameters",
			handler: func(n int, w http.ResponseWriter, _ *http.Request) {
				if n == 1 {
					w.WriteHeader(http.StatusBadRequest)
					write(w, `{"errors":[{"detail":"Cannot paginate workspace"}]}`)
					return
				}
				write(w, listBody("workspace", []string{"w1", "w2"}, ""))
			},
			wantIDs:     []string{"w1", "w2"},
			wantQueries: []string{totals, ""},
		},
		{
			name: "genuine 400 is returned after one unpaged retry",
			opts: &ListOptions{Filter: "name=="},
			handler: func(_ int, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				write(w, `{"errors":[{"detail":"Invalid filter"}]}`)
			},
			wantErr:     true,
			wantQueries: []string{"filter%5Bworkspace%5D=name%3D%3D&" + totals, "filter%5Bworkspace%5D=name%3D%3D"},
		},
		{
			name: "one requested page",
			opts: &ListOptions{Page: 2, PageSize: 1},
			handler: func(_ int, w http.ResponseWriter, _ *http.Request) {
				write(w, listBody("workspace", []string{"w2"}, pageMeta(2, 1, 9)))
			},
			wantIDs:     []string{"w2"},
			wantQueries: []string{"page%5Bnumber%5D=2&page%5Bsize%5D=1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newRecordedServer(t, tt.handler)
			items, err := srv.client(t).Workspaces.List(context.Background(), "org-1", tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if items != nil {
					t.Errorf("got %d items with an error, want nil", len(items))
				}
			} else if got := itemIDs(items, func(w *Workspace) string { return w.ID }); !slices.Equal(got, tt.wantIDs) {
				t.Errorf("ids = %q, want %q", got, tt.wantIDs)
			}
			wantQueries(t, srv.recorded(), tt.wantQueries...)
		})
	}
}

func TestCrudService_List_CancelWhilePaging(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := newRecordedServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			_, _ = w.Write([]byte(listBody("workspace", []string{"w1"}, pageMeta(1, 1, 3))))
			return
		}
		cancel()
		// Returning would send an empty 200, which the client may prefer over the cancellation.
		<-r.Context().Done()
	})

	items, err := srv.client(t).Workspaces.List(ctx, "org-1", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if items != nil {
		t.Errorf("got %d items with an error, want nil", len(items))
	}
	if got := srv.recorded(); len(got) != 2 {
		t.Errorf("got %d requests, want 2", len(got))
	}
}

func TestCrudService_List_UnpagedType(t *testing.T) {
	t.Parallel()

	srv := newRecordedServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(listBody("organization", []string{"o1", "o2"}, "")))
	})
	items, err := srv.client(t).Organizations.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
	wantQueries(t, srv.recorded(), "")
}

// TestNewClient_PagedServices pins the set of services the library pages to
// the resource types verified against a Terrakube 2.34.1 server.
func TestNewClient_PagedServices(t *testing.T) {
	t.Parallel()

	paged := map[string]bool{
		"Workspaces": true, "WorkspaceTags": true, "Actions": true, "Federated": true,
		"Agents": true, "OrganizationVariables": true, "WorkspaceAccess": true,
	}

	c, err := NewClient(WithEndpoint("https://example.com"), WithToken("test-token"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	found := 0
	val := reflect.ValueOf(c).Elem()
	for i := range val.NumField() {
		field := val.Field(i)
		if field.Kind() != reflect.Pointer || field.IsNil() || field.Elem().Kind() != reflect.Struct {
			continue
		}
		flag := field.Elem().FieldByName("paged")
		if !flag.IsValid() {
			continue // not built on crudService (TeamTokens, Operations)
		}
		name := val.Type().Field(i).Name
		if flag.Bool() != paged[name] {
			t.Errorf("Client.%s paged = %v, want %v", name, flag.Bool(), paged[name])
		}
		if paged[name] {
			found++
		}
	}
	if found != len(paged) {
		t.Errorf("found %d paged services, want %d", found, len(paged))
	}
}

// TestCrudService_Update_NoContent: a Terrakube server answers PATCH with 204
// and no body. Update must still return the resource, read with a GET.
func TestCrudService_Update_NoContent(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var methods []string
	srv := testutil.NewServer(t)
	record := func(r *http.Request) {
		mu.Lock()
		methods = append(methods, r.Method)
		mu.Unlock()
	}
	srv.HandleFunc("PATCH /api/v1/organization/org-1", func(w http.ResponseWriter, r *http.Request) {
		record(r)
		w.WriteHeader(http.StatusNoContent)
	})
	srv.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, r *http.Request) {
		record(r)
		testutil.WriteJSONAPI(t, w, http.StatusOK, &Organization{ID: "org-1", Name: "Updated", ExecutionMode: "remote"})
	})

	client := newCrudTestClient(t, srv)
	got, err := client.Organizations.Update(context.Background(), &Organization{ID: "org-1", Name: "Updated"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.ID != "org-1" || got.Name != "Updated" || got.ExecutionMode != "remote" {
		t.Errorf("Update returned %+v, want the state the server holds", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if !slices.Equal(methods, []string{http.MethodPatch, http.MethodGet}) {
		t.Errorf("requests = %v, want PATCH then GET", methods)
	}
}

// TestCrudService_Update_NoContentThenNotFound: an update can make the resource
// unreadable (disabling an organization). The update succeeded, so that is not
// an error, and the caller gets back what it sent.
func TestCrudService_Update_NoContentThenNotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "Unknown identifier org-1 for organization")
	})

	sent := &Organization{ID: "org-1", Name: "Gone", Disabled: true}
	got, err := newCrudTestClient(t, srv).Organizations.Update(context.Background(), sent)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got != sent {
		t.Errorf("Update returned %+v, want the entity that was sent", got)
	}
}
