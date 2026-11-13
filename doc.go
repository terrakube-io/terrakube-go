// Package terrakube provides a Go client for the Terrakube API.
//
// Terrakube is an open-source platform for managing Terraform and OpenTofu
// infrastructure as code. It provides workspaces, modules, providers, team
// access control, and job execution through a JSON:API interface. This library
// targets one Terrakube API version, reported by [APIVersion].
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
//     [CollectionService], [ProviderService], [JobService],
//     [OrganizationVariableService], [ProjectService],
//     [NotificationConfigurationService], [PolicySetService],
//     [PolicyExemptionService]
//
// Workspace children:
//   - [WorkspaceTagService], [WorkspaceAccessService], [WorkspaceScheduleService],
//     [HistoryService], [VariableService] (workspace-scoped), [WebhookService],
//     [RunTriggerService]
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
//   - [StepService], [AddressService], [PolicyEvaluationService] (read-only)
//     -> [PolicyOverrideService]
//
// Project children:
//   - [ProjectAccessService]
//
// Policy set children:
//   - [PolicySetParameterService], [PolicyAttachmentService]
//
// Notification configuration children:
//   - [NotificationTriggerService]
//
// Federated children:
//   - [FederatedClaimService]
//
// Team children:
//   - [TeamTokenService] (non-JSON:API, uses standard JSON)
//
// Webhook children:
//   - [WebhookEventService]
//
// Standalone resources (no organization parent):
//   - [GithubAppTokenService], [ActionService], [FederatedService]
//
// [OperationsService] submits JSON:API atomic operations: several changes in
// one request that the server applies together or not at all.
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
// # Updating
//
// Update sends a JSON:API PATCH with every attribute of the struct, not only
// the ones that changed. A field left at its zero value therefore overwrites
// the value on the server: an empty string for a string, false for a bool.
// Fetch the resource, change it, and pass it back:
//
//	ws, err := client.Workspaces.Get(ctx, orgID, wsID)
//	if err != nil {
//		return err
//	}
//	ws.Description = new("managed by automation")
//	ws, err = client.Workspaces.Update(ctx, orgID, ws)
//
// Relationships are different: a nil or empty relationship is not sent and is
// left unchanged.
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
// A List call with no page options returns the whole collection where the
// server can page it and whatever one request returns otherwise; see
// [ListOptions] for the details and for requesting a single page.
//
// # Error Handling
//
// Server errors are returned as [APIError], which includes the HTTP status
// code, request method and path, the raw response body, and any structured
// error details. Client-side validation failures (such as empty IDs) are
// returned as [ValidationError].
//
// Use the helper functions [IsNotFound], [IsConflict], [IsUnauthorized], and
// [IsForbidden] to check for common HTTP error conditions:
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
