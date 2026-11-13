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

// pageInfo is the pagination metadata the server returns in meta.page.
type pageInfo struct {
	Number     int `json:"number"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

// doList executes a request and decodes a JSON:API list response into []*T.
// It also returns the response's meta.page, or nil when the server sent none.
func (c *Client) doList[T any](req *http.Request) ([]*T, *pageInfo, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are inconsequential

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, c.parseAPIError(req, resp.StatusCode, bodyBytes)
	}

	if len(bodyBytes) == 0 {
		return nil, nil, nil
	}

	items, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(bodyBytes), reflect.TypeFor[*T]())
	if err != nil {
		// A to-one relationship path (a workspace's webhook) answers with one
		// resource object instead of an array, or null when there is none.
		one := new(T)
		if jsonapi.UnmarshalPayload(bytes.NewReader(bodyBytes), one) != nil {
			return nil, nil, fmt.Errorf("decoding JSON:API list response: %w", err)
		}
		return []*T{one}, nil, nil
	}

	result := make([]*T, len(items))
	for i, item := range items {
		result[i] = item.(*T)
	}

	var envelope struct {
		Meta struct {
			Page *pageInfo `json:"page"`
		} `json:"meta"`
	}
	_ = json.Unmarshal(bodyBytes, &envelope) // the items decoded; missing or odd metadata just means no paging
	return result, envelope.Meta.Page, nil
}

// APIVersion is the Terrakube OpenAPI specification version this library targets.
const APIVersion = "2.34.1"

const (
	apiBasePath  = "/api/v1/"
	mediaType    = "application/vnd.api+json"
	jsonType     = "application/json"
	defaultAgent = "terrakube-go"
)

// ListOptions specifies optional parameters for List methods.
//
// Page and PageSize request one specific page. Leave both at zero to list the
// collection: resource types the library pages are fetched page by page until
// the server reports no more; all others, and any type for which the server
// refuses paging, are fetched in one request, so the result stops at any limit
// the server applies to that type. Links in responses are never followed.
//
// The server accepts Page and PageSize only for some resource types: workspace,
// workspace tag, action, federated, agent, organization variable, and workspace
// access (verified on Terrakube 2.34.1). For every other resource type List
// returns a *ValidationError without sending a request when either is set.
type ListOptions struct {
	Filter   string // Server-side filter expression (e.g. "name==production").
	Include  string // Comma-separated relationship names to sideload (e.g. "workspace,organization").
	Page     int    // 1-indexed page number; 0 means unset.
	PageSize int    // Items per page; 0 means unset (server default).
}

// Client manages communication with the Terrakube API.
type Client struct {
	baseURL     *url.URL
	token       string
	httpClient  *http.Client
	userAgent   string
	insecureTLS bool

	Organizations              *OrganizationService
	Workspaces                 *WorkspaceService
	Modules                    *ModuleService
	Teams                      *TeamService
	TeamTokens                 *TeamTokenService
	Variables                  *VariableService
	OrganizationVariables      *OrganizationVariableService
	Templates                  *TemplateService
	Tags                       *TagService
	VCS                        *VCSService
	SSH                        *SSHService
	Agents                     *AgentService
	Collections                *CollectionService
	CollectionItems            *CollectionItemService
	CollectionReferences       *CollectionReferenceService
	WorkspaceTags              *WorkspaceTagService
	WorkspaceAccess            *WorkspaceAccessService
	WorkspaceSchedules         *WorkspaceScheduleService
	Webhooks                   *WebhookService
	WebhookEvents              *WebhookEventService
	History                    *HistoryService
	Jobs                       *JobService
	Actions                    *ActionService
	Steps                      *StepService
	Providers                  *ProviderService
	ProviderVersions           *ProviderVersionService
	Implementations            *ImplementationService
	ModuleVersions             *ModuleVersionService
	GithubAppTokens            *GithubAppTokenService
	Addresses                  *AddressService
	Projects                   *ProjectService
	ProjectAccess              *ProjectAccessService
	Federated                  *FederatedService
	FederatedClaims            *FederatedClaimService
	Operations                 *OperationsService
	NotificationConfigurations *NotificationConfigurationService
	NotificationTriggers       *NotificationTriggerService
	PolicyAttachments          *PolicyAttachmentService
	PolicyEvaluations          *PolicyEvaluationService
	PolicyExemptions           *PolicyExemptionService
	PolicyOverrides            *PolicyOverrideService
	PolicySets                 *PolicySetService
	PolicySetParameters        *PolicySetParameterService
	RunTriggers                *RunTriggerService
}

// Option configures a Client.
type Option func(*Client) error

// WithEndpoint sets the Terrakube server URL. A path in the URL is kept as a
// prefix for every request, for a server published under a sub-path
// (https://example.com/terrakube).
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
		if httpClient == nil {
			return fmt.Errorf("http client must not be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithInsecureTLS skips TLS certificate verification. It is applied after all
// other options, so its position relative to WithHTTPClient does not matter.
// The client's transport must be an *http.Transport (or unset): NewClient
// returns an error for any other http.RoundTripper, whose TLS settings the
// library cannot change.
func WithInsecureTLS() Option {
	return func(c *Client) error {
		c.insecureTLS = true
		return nil
	}
}

// applyInsecureTLS replaces the HTTP client with a copy whose transport skips
// certificate verification. The caller's client and transport are not modified.
func (c *Client) applyInsecureTLS() error {
	rt := c.httpClient.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	base, ok := rt.(*http.Transport)
	if !ok {
		return fmt.Errorf("WithInsecureTLS needs an *http.Transport, got %T: disable verification in that transport instead", rt)
	}
	transport := base.Clone()
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{} //nolint:gosec // Struct initialized below
	}
	transport.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec // User-requested insecure mode

	hc := *c.httpClient
	hc.Transport = transport
	c.httpClient = &hc
	return nil
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) error {
		c.userAgent = ua
		return nil
	}
}

// NewClient creates a new Terrakube API client. It returns an error if WithEndpoint or WithToken are not provided.
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
	if c.insecureTLS {
		if err := c.applyInsecureTLS(); err != nil {
			return nil, err
		}
	}

	c.Organizations = &OrganizationService{client: c}
	c.Workspaces = &WorkspaceService{client: c, paged: true, filterKey: "filter[workspace]"}
	c.Modules = &ModuleService{client: c, filterKey: "filter[module]"}
	c.Teams = &TeamService{client: c, filterKey: "filter[team]"}
	c.TeamTokens = &TeamTokenService{client: c}
	c.Variables = &VariableService{client: c, filterKey: "filter[variable]"}
	c.OrganizationVariables = &OrganizationVariableService{client: c, paged: true, filterKey: "filter[globalvar]"}
	c.Templates = &TemplateService{client: c, filterKey: "filter[template]"}
	c.Tags = &TagService{client: c, filterKey: "filter[tag]"}
	c.VCS = &VCSService{client: c, filterKey: "filter[vcs]"}
	c.SSH = &SSHService{client: c, filterKey: "filter[ssh]"}
	c.Agents = &AgentService{client: c, paged: true, filterKey: "filter[agent]"}
	c.Collections = &CollectionService{client: c, filterKey: "filter[collection]"}
	c.CollectionItems = &CollectionItemService{client: c, filterKey: "filter[item]"}
	c.CollectionReferences = &CollectionReferenceService{client: c, filterKey: "filter[reference]"}
	c.WorkspaceTags = &WorkspaceTagService{client: c, paged: true, filterKey: "filter[workspacetag]"}
	c.WorkspaceAccess = &WorkspaceAccessService{client: c, paged: true, filterKey: "filter[access]"}
	c.WorkspaceSchedules = &WorkspaceScheduleService{client: c, filterKey: "filter[schedule]"}
	c.Webhooks = &WebhookService{client: c, filterKey: "filter[webhook]"}
	c.WebhookEvents = &WebhookEventService{client: c, filterKey: "filter[webhook_event]"}
	c.History = &HistoryService{client: c, filterKey: "filter[history]"}
	c.Jobs = &JobService{client: c, filterKey: "filter[job]"}
	c.Actions = &ActionService{client: c, paged: true}
	c.Steps = &StepService{client: c, filterKey: "filter[step]"}
	c.Providers = &ProviderService{client: c, filterKey: "filter[provider]"}
	c.ProviderVersions = &ProviderVersionService{client: c, filterKey: "filter[version]"}
	c.Implementations = &ImplementationService{client: c, filterKey: "filter[implementation]"}
	c.ModuleVersions = &ModuleVersionService{client: c, filterKey: "filter[module_version]"}
	c.GithubAppTokens = &GithubAppTokenService{client: c}
	c.Addresses = &AddressService{client: c, filterKey: "filter[address]"}
	c.Projects = &ProjectService{client: c, filterKey: "filter[project]"}
	c.ProjectAccess = &ProjectAccessService{client: c, filterKey: "filter[project_access]"}
	c.Federated = &FederatedService{client: c, paged: true, filterKey: "filter[federated]"}
	c.FederatedClaims = &FederatedClaimService{client: c, filterKey: "filter[federated_claim]"}
	c.Operations = &OperationsService{client: c}
	c.NotificationConfigurations = &NotificationConfigurationService{client: c, filterKey: "filter[notification_configuration]"}
	c.NotificationTriggers = &NotificationTriggerService{client: c, filterKey: "filter[notification_trigger]"}
	c.PolicyAttachments = &PolicyAttachmentService{client: c, filterKey: "filter[policy_attachment]"}
	c.PolicyEvaluations = &PolicyEvaluationService{client: c, filterKey: "filter[policy_evaluation]"}
	c.PolicyExemptions = &PolicyExemptionService{client: c, filterKey: "filter[policy_exemption]"}
	c.PolicyOverrides = &PolicyOverrideService{client: c, filterKey: "filter[policy_override]"}
	c.PolicySets = &PolicySetService{client: c, filterKey: "filter[policy_set]"}
	c.PolicySetParameters = &PolicySetParameterService{client: c, filterKey: "filter[policy_set_parameter]"}
	c.RunTriggers = &RunTriggerService{client: c, filterKey: "filter[runTrigger]"}

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

// endpointURL returns the URL of reqPath on the configured endpoint, keeping
// any path prefix the endpoint has (https://host/terrakube + /api/v1/x).
func (c *Client) endpointURL(reqPath string) *url.URL {
	u := *c.baseURL
	u.Path = path.Join("/", u.Path, reqPath)
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return &u
}

// request builds an authenticated JSON:API HTTP request.
func (c *Client) request(ctx context.Context, method, reqPath string, body any) (*http.Request, error) {
	return c.requestWithQuery(ctx, method, reqPath, nil, body)
}

// requestWithQuery builds an authenticated JSON:API HTTP request with query parameters.
func (c *Client) requestWithQuery(ctx context.Context, method, reqPath string, params url.Values, body any) (*http.Request, error) {
	u := c.endpointURL(reqPath)

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

// do executes a request and decodes a single JSON:API resource response.
// If v is nil, no response body decoding is performed (used for DELETE).
// For list responses, use the generic doList[T] method instead.
func (c *Client) do(req *http.Request, v any) (*http.Response, error) { //nolint:unparam // Response returned for future use by callers
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
		return resp, c.parseAPIError(req, resp.StatusCode, bodyBytes)
	}

	if v != nil && len(bodyBytes) > 0 {
		if err := jsonapi.UnmarshalPayload(bytes.NewReader(bodyBytes), v); err != nil {
			return resp, fmt.Errorf("decoding JSON:API response: %w", err)
		}
	}

	return resp, nil
}

// requestRaw builds an authenticated HTTP request with standard JSON content type.
// Used for non-JSON:API endpoints like TeamToken.
func (c *Client) requestRaw(ctx context.Context, method, rawPath string, body any) (*http.Request, error) {
	u := c.endpointURL(rawPath)

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
func (c *Client) doRaw(req *http.Request, v any) (*http.Response, error) { //nolint:unparam // Response returned for future use by callers
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
		return resp, c.parseAPIError(req, resp.StatusCode, bodyBytes)
	}

	if v != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, v); err != nil {
			return resp, fmt.Errorf("decoding JSON response: %w", err)
		}
	}

	return resp, nil
}

// parseAPIError constructs an APIError from an HTTP response, parsing
// structured error details from the response body when possible.
func (c *Client) parseAPIError(req *http.Request, statusCode int, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Method:     req.Method,
		Path:       req.URL.Path,
		Body:       body,
	}
	var errResp struct {
		Errors []ErrorDetail `json:"errors"`
	}
	if json.Unmarshal(body, &errResp) == nil {
		apiErr.Errors = errResp.Errors
		return apiErr
	}
	// POST /operations reports errors as an array with one such object per operation.
	var perOp []struct {
		Errors []ErrorDetail `json:"errors"`
	}
	if json.Unmarshal(body, &perOp) == nil {
		for _, e := range perOp {
			apiErr.Errors = append(apiErr.Errors, e.Errors...)
		}
	}
	return apiErr
}

// validateID checks that a resource ID is not empty.
func validateID(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "must not be empty"}
	}
	return nil
}
