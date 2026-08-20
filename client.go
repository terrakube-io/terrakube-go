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

// doList executes a request and decodes a JSON:API list response into []*T.
// It avoids runtime reflection by leveraging Go generics.
func doList[T any](c *Client, req *http.Request) ([]*T, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are inconsequential

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.parseAPIError(req, resp.StatusCode, bodyBytes)
	}

	if len(bodyBytes) == 0 {
		return nil, nil
	}

	items, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(bodyBytes), reflect.TypeOf((*T)(nil)))
	if err != nil {
		return nil, fmt.Errorf("decoding JSON:API list response: %w", err)
	}

	result := make([]*T, len(items))
	for i, item := range items {
		result[i] = item.(*T)
	}
	return result, nil
}

// APIVersion is the Terrakube OpenAPI specification version this library targets.
const APIVersion = "2.33.0"

const (
	apiBasePath  = "/api/v1/"
	mediaType    = "application/vnd.api+json"
	jsonType     = "application/json"
	defaultAgent = "terrakube-go"
)

// ListOptions specifies optional parameters for List methods.
type ListOptions struct {
	Filter   string // Server-side filter expression (e.g. "name==production").
	Include  string // Comma-separated relationship names to sideload (e.g. "workspace,organization").
	Page     int    // 1-indexed page number; 0 means unset (auto-paginate all pages).
	PageSize int    // Items per page; 0 means unset (server default).
}

// Client manages communication with the Terrakube API.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
	userAgent  string

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
// It modifies the TLS configuration of the existing transport rather than
// replacing the entire http.Client, so it composes safely with WithHTTPClient.
func WithInsecureTLS() Option {
	return func(c *Client) error {
		var transport *http.Transport
		if t, ok := c.httpClient.Transport.(*http.Transport); ok && t != nil {
			transport = t.Clone()
		} else {
			transport = http.DefaultTransport.(*http.Transport).Clone()
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{} //nolint:gosec // Struct initialized below
		}
		transport.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec // User-requested insecure mode
		c.httpClient = &http.Client{
			Transport:     transport,
			Timeout:       c.httpClient.Timeout,
			CheckRedirect: c.httpClient.CheckRedirect,
			Jar:           c.httpClient.Jar,
		}
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

	c.Organizations = &OrganizationService{crudService[Organization]{client: c}}
	c.Workspaces = &WorkspaceService{crudService[Workspace]{client: c, filterKey: "filter[workspace]"}}
	c.Modules = &ModuleService{crudService[Module]{client: c, filterKey: "filter[module]"}}
	c.Teams = &TeamService{crudService[Team]{client: c, filterKey: "filter[team]"}}
	c.TeamTokens = &TeamTokenService{client: c}
	c.Variables = &VariableService{crudService[Variable]{client: c, filterKey: "filter[variable]"}}
	c.OrganizationVariables = &OrganizationVariableService{crudService[OrganizationVariable]{client: c, filterKey: "filter[globalvar]"}}
	c.Templates = &TemplateService{crudService[Template]{client: c, filterKey: "filter[template]"}}
	c.Tags = &TagService{crudService[Tag]{client: c, filterKey: "filter[tag]"}}
	c.VCS = &VCSService{crudService[VCS]{client: c, filterKey: "filter[vcs]"}}
	c.SSH = &SSHService{crudService[SSH]{client: c, filterKey: "filter[ssh]"}}
	c.Agents = &AgentService{crudService[Agent]{client: c, filterKey: "filter[agent]"}}
	c.Collections = &CollectionService{crudService[Collection]{client: c, filterKey: "filter[collection]"}}
	c.CollectionItems = &CollectionItemService{crudService[CollectionItem]{client: c, filterKey: "filter[item]"}}
	c.CollectionReferences = &CollectionReferenceService{crudService[CollectionReference]{client: c, filterKey: "filter[reference]"}}
	c.WorkspaceTags = &WorkspaceTagService{crudService[WorkspaceTag]{client: c, filterKey: "filter[workspacetag]"}}
	c.WorkspaceAccess = &WorkspaceAccessService{crudService[WorkspaceAccess]{client: c, filterKey: "filter[access]"}}
	c.WorkspaceSchedules = &WorkspaceScheduleService{crudService[WorkspaceSchedule]{client: c, filterKey: "filter[schedule]"}}
	c.Webhooks = &WebhookService{crudService[Webhook]{client: c, filterKey: "filter[webhook]"}}
	c.WebhookEvents = &WebhookEventService{crudService[WebhookEvent]{client: c, filterKey: "filter[webhook_event]"}}
	c.History = &HistoryService{crudService[History]{client: c, filterKey: "filter[history]"}}
	c.Jobs = &JobService{crudService[Job]{client: c, filterKey: "filter[job]"}}
	c.Actions = &ActionService{crudService[Action]{client: c}}
	c.Steps = &StepService{crudService[Step]{client: c, filterKey: "filter[step]"}}
	c.Providers = &ProviderService{crudService[Provider]{client: c, filterKey: "filter[provider]"}}
	c.ProviderVersions = &ProviderVersionService{crudService[ProviderVersion]{client: c, filterKey: "filter[version]"}}
	c.Implementations = &ImplementationService{crudService[Implementation]{client: c, filterKey: "filter[implementation]"}}
	c.ModuleVersions = &ModuleVersionService{crudService[ModuleVersion]{client: c, filterKey: "filter[version]"}}
	c.GithubAppTokens = &GithubAppTokenService{crudService[GithubAppToken]{client: c}}
	c.Addresses = &AddressService{crudService[Address]{client: c, filterKey: "filter[address]"}}
	c.Projects = &ProjectService{crudService[Project]{client: c, filterKey: "filter[project]"}}
	c.ProjectAccess = &ProjectAccessService{crudService[ProjectAccess]{client: c, filterKey: "filter[projectAccess]"}}
	c.Federated = &FederatedService{crudService[Federated]{client: c, filterKey: "filter[federated]"}}
	c.FederatedClaims = &FederatedClaimService{crudService[FederatedClaim]{client: c, filterKey: "filter[claims]"}}
	c.Operations = &OperationsService{client: c}
	c.NotificationConfigurations = &NotificationConfigurationService{crudService[NotificationConfiguration]{client: c, filterKey: "filter[notification_configuration]"}}
	c.NotificationTriggers = &NotificationTriggerService{crudService[NotificationTrigger]{client: c, filterKey: "filter[notification_trigger]"}}

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

// do executes a request and decodes a single JSON:API resource response.
// If v is nil, no response body decoding is performed (used for DELETE).
// For list responses, use the generic doList[T] function instead.
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) { //nolint:unparam // Response returned for future use by callers
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
func (c *Client) doRaw(req *http.Request, v interface{}) (*http.Response, error) { //nolint:unparam // Response returned for future use by callers
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
