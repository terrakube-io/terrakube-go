// Package terrakube provides a Go client for the Terrakube API.
//
// Terrakube is an open-source platform for managing Terraform and OpenTofu
// infrastructure as code. It provides workspaces, modules, providers, team
// access control, and job execution through a JSON:API interface. This library
// covers the full Terrakube API surface as of OpenAPI version 2.27.0.
//
// # Authentication
//
// Create a [Client] with [NewClient], passing [WithEndpoint] and [WithToken]
// as required options:
//
//	client, err := terrakube.NewClient(
//		terrakube.WithEndpoint("https://terrakube.example.com"),
//		terrakube.WithToken("your-api-token"),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Additional options include [WithHTTPClient] to supply a custom http.Client,
// [WithInsecureTLS] to skip certificate verification, and [WithUserAgent] to
// set a custom User-Agent header.
//
// # Resource Hierarchy
//
// The Terrakube API organizes resources in a hierarchy rooted at organizations.
// Each service on the [Client] corresponds to one resource type. Top-level
// methods require only a context, while nested resources require parent IDs.
//
// Organization (top-level):
//   - [WorkspaceService], [ModuleService], [TeamService], [VariableService] (org-scoped),
//     [TemplateService], [TagService], [VCSService], [SSHService], [AgentService],
//     [CollectionService], [ProviderService], [JobService], [WebhookService],
//     [OrganizationVariableService]
//
// Workspace children:
//   - [WorkspaceTagService], [WorkspaceAccessService], [WorkspaceScheduleService],
//     [HistoryService], [VariableService] (workspace-scoped)
//
// Module children:
//   - [ModuleVersionService]
//
// Provider children:
//   - [ProviderVersionService] -> [ImplementationService]
//
// Collection children:
//   - [CollectionItemService], [CollectionReferenceService]
//
// Job children:
//   - [StepService], [AddressService]
//
// Team children:
//   - [TeamTokenService] (non-JSON:API, uses standard JSON)
//
// Webhook children:
//   - [WebhookEventService]
//
// Standalone resources (no organization parent):
//   - [GithubAppTokenService], [ActionService]
//
// # CRUD Operations
//
// Most services expose List, Get, Create, Update, and Delete methods that
// follow a consistent pattern. Nested resources require one or more parent IDs:
//
//	// List workspaces in an organization.
//	workspaces, err := client.Workspaces.List(ctx, orgID, nil)
//
//	// Create a workspace.
//	ws, err := client.Workspaces.Create(ctx, orgID, &terrakube.Workspace{
//		Name:   "my-workspace",
//		Source: "https://github.com/example/repo",
//		Branch: "main",
//	})
//
//	// Get a single workspace.
//	ws, err = client.Workspaces.Get(ctx, orgID, ws.ID)
//
//	// Delete a workspace.
//	err = client.Workspaces.Delete(ctx, orgID, ws.ID)
//
// # List Filtering
//
// List methods accept an optional [ListOptions] parameter for server-side
// filtering. The filter query parameter key varies by service (for example,
// "filter[workspace]" for workspaces, "filter[module]" for modules):
//
//	opts := &terrakube.ListOptions{Filter: "name==production"}
//	workspaces, err := client.Workspaces.List(ctx, orgID, opts)
//
// Pass nil for no filtering.
//
// # Error Handling
//
// Server errors are returned as [APIError], which includes the HTTP status
// code, request method and path, the raw response body, and any structured
// error details. Client-side validation failures (such as empty IDs) are
// returned as [ValidationError].
//
// Use the helper functions [IsNotFound], [IsConflict], and [IsUnauthorized]
// to check for common HTTP error conditions:
//
//	ws, err := client.Workspaces.Get(ctx, orgID, wsID)
//	if terrakube.IsNotFound(err) {
//		// Handle 404.
//	}
//
// # API Version
//
// The [APIVersion] constant reports the Terrakube OpenAPI specification version
// this library targets.
package terrakube
