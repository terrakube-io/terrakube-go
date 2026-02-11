package terrakube

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/google/jsonapi"
)

const (
	apiBasePath  = "/api/v1/"
	mediaType    = "application/vnd.api+json"
	jsonType     = "application/json"
	defaultAgent = "terrakube-go"
)

// ListOptions specifies optional parameters for List methods.
type ListOptions struct {
	Filter string
}

// Client manages communication with the Terrakube API.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
	userAgent  string

	Organizations         *OrganizationService
	Workspaces            *WorkspaceService
	Modules               *ModuleService
	Teams                 *TeamService
	TeamTokens            *TeamTokenService
	Variables             *VariableService
	OrganizationVariables *OrganizationVariableService
	Templates             *TemplateService
	Tags                  *TagService
	VCS                   *VCSService
	SSH                   *SSHService
	Agents                *AgentService
	Collections           *CollectionService
	CollectionItems       *CollectionItemService
	CollectionReferences  *CollectionReferenceService
	WorkspaceTags         *WorkspaceTagService
	WorkspaceAccess       *WorkspaceAccessService
	WorkspaceSchedules    *WorkspaceScheduleService
	Webhooks              *WebhookService
	WebhookEvents         *WebhookEventService
	History               *HistoryService
	Jobs                  *JobService
}

// Option configures a Client.
type Option func(*Client) error

// WithEndpoint sets the Terrakube server URL.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) error {
		if endpoint == "" {
			return fmt.Errorf("endpoint must not be empty")
		}
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "https://" + endpoint
		}
		u, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("invalid endpoint URL: %w", err)
		}
		c.baseURL = u
		return nil
	}
}

// WithToken sets the API bearer token.
func WithToken(token string) Option {
	return func(c *Client) error {
		if token == "" {
			return fmt.Errorf("token must not be empty")
		}
		c.token = token
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}

// WithInsecureTLS skips TLS certificate verification.
func WithInsecureTLS() Option {
	return func(c *Client) error {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // User-requested insecure mode
		c.httpClient = &http.Client{Transport: transport}
		return nil
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) error {
		c.userAgent = ua
		return nil
	}
}

// NewClient creates a new Terrakube API client.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
		userAgent:  defaultAgent,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.baseURL == nil {
		return nil, fmt.Errorf("endpoint is required: use WithEndpoint()")
	}
	if c.token == "" {
		return nil, fmt.Errorf("token is required: use WithToken()")
	}

	c.Organizations = &OrganizationService{client: c}
	c.Workspaces = &WorkspaceService{client: c}
	c.Modules = &ModuleService{client: c}
	c.Teams = &TeamService{client: c}
	c.TeamTokens = &TeamTokenService{client: c}
	c.Variables = &VariableService{client: c}
	c.OrganizationVariables = &OrganizationVariableService{client: c}
	c.Templates = &TemplateService{client: c}
	c.Tags = &TagService{client: c}
	c.VCS = &VCSService{client: c}
	c.SSH = &SSHService{client: c}
	c.Agents = &AgentService{client: c}
	c.Collections = &CollectionService{client: c}
	c.CollectionItems = &CollectionItemService{client: c}
	c.CollectionReferences = &CollectionReferenceService{client: c}
	c.WorkspaceTags = &WorkspaceTagService{client: c}
	c.WorkspaceAccess = &WorkspaceAccessService{client: c}
	c.WorkspaceSchedules = &WorkspaceScheduleService{client: c}
	c.Webhooks = &WebhookService{client: c}
	c.WebhookEvents = &WebhookEventService{client: c}
	c.History = &HistoryService{client: c}
	c.Jobs = &JobService{client: c}

	return c, nil
}

// apiPath constructs a full API path by joining segments under apiBasePath.
func (c *Client) apiPath(segments ...string) string {
	p := apiBasePath
	if len(segments) > 0 {
		p = path.Join(apiBasePath, path.Join(segments...))
	}
	return p
}

// request builds an authenticated JSON:API HTTP request.
func (c *Client) request(ctx context.Context, method, reqPath string, body interface{}) (*http.Request, error) {
	return c.requestWithQuery(ctx, method, reqPath, nil, body)
}

// requestWithQuery builds an authenticated JSON:API HTTP request with query parameters.
func (c *Client) requestWithQuery(ctx context.Context, method, reqPath string, params url.Values, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: reqPath}
	u := c.baseURL.ResolveReference(rel)

	if params != nil {
		u.RawQuery = params.Encode()
	}

	var buf io.Reader
	if body != nil {
		var b bytes.Buffer
		if err := jsonapi.MarshalPayload(&b, body); err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		buf = &b
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", mediaType)
	}
	req.Header.Set("Accept", mediaType)

	return req, nil
}

// do executes a request and decodes the JSON:API response.
// If v is nil, no response body decoding is performed (used for DELETE).
// If v is a pointer to a slice, jsonapi.UnmarshalManyPayload is used.
// Otherwise jsonapi.UnmarshalPayload is used.
func (c *Client) do(_ context.Context, req *http.Request, v interface{}) (*http.Response, error) { //nolint:unparam // Response returned for future use by callers
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are inconsequential

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Method:     req.Method,
			Path:       req.URL.Path,
			Body:       bodyBytes,
		}
		var errResp struct {
			Errors []ErrorDetail `json:"errors"`
		}
		if json.Unmarshal(bodyBytes, &errResp) == nil {
			apiErr.Errors = errResp.Errors
		}
		return resp, apiErr
	}

	if v != nil && len(bodyBytes) > 0 {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Slice {
			items, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(bodyBytes), rv.Elem().Type().Elem())
			if err != nil {
				return resp, fmt.Errorf("decoding JSON:API list response: %w", err)
			}
			slice := reflect.MakeSlice(rv.Elem().Type(), len(items), len(items))
			for i, item := range items {
				slice.Index(i).Set(reflect.ValueOf(item))
			}
			rv.Elem().Set(slice)
		} else {
			if err := jsonapi.UnmarshalPayload(bytes.NewReader(bodyBytes), v); err != nil {
				return resp, fmt.Errorf("decoding JSON:API response: %w", err)
			}
		}
	}

	return resp, nil
}

// requestRaw builds an authenticated HTTP request with standard JSON content type.
// Used for non-JSON:API endpoints like TeamToken.
func (c *Client) requestRaw(ctx context.Context, method, rawPath string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: rawPath}
	u := c.baseURL.ResolveReference(rel)

	var buf io.Reader
	if body != nil {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(body); err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		buf = &b
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", jsonType)
	}
	req.Header.Set("Accept", jsonType)

	return req, nil
}

// doRaw executes a request and decodes the response using encoding/json.
// Used for non-JSON:API endpoints.
func (c *Client) doRaw(_ context.Context, req *http.Request, v interface{}) (*http.Response, error) { //nolint:unparam // Response returned for future use by callers
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are inconsequential

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, &APIError{
			StatusCode: resp.StatusCode,
			Method:     req.Method,
			Path:       req.URL.Path,
			Body:       bodyBytes,
		}
	}

	if v != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, v); err != nil {
			return resp, fmt.Errorf("decoding JSON response: %w", err)
		}
	}

	return resp, nil
}

// validateID checks that a resource ID is not empty.
func validateID(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "must not be empty"}
	}
	return nil
}
