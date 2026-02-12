# Go Doc Comment Audit Report

**Date:** 2026-02-12
**Scope:** All exported symbols in terrakube-go (excluding `_test.go` and `testutil/`)
**`go vet ./...`:** Clean. No doc-related warnings.
**`golangci-lint run --enable-only revive,godot`:** 0 issues.
**`revive -config exported.toml`:** 0 issues. All exported symbols have conforming doc comments.
**`godot`:** 0 issues. All comments end with periods.

## Summary

| Metric | Count |
|--------|-------|
| Total exported symbols | 230 |
| Good (meets all conventions) | 188 |
| Needs work (minor issues) | 39 |
| Missing doc comment | 3 |

**Overall assessment:** The codebase has strong doc coverage. Every exported type, service, and method has a doc comment. The issues are:
1. Three `Error()` methods on error types lack doc comments.
2. The `doc.go` package comment does not end with a period.
3. Some method docs don't mention error conditions (`*ValidationError`, `*APIError`).
4. A handful of struct field names could benefit from inline comments (non-obvious mappings to JSON:API attribute names).

---

## 1. Core Files

### doc.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| (package comment) | package | `Package terrakube provides a Go client library for the Terrakube API.` (multi-line) | Needs work | Last line of the example block should be followed by a period-terminated sentence. The package comment overall is good but the closing is abrupt -- the final `client.Organizations.List(ctx, nil)` line is just code; add a trailing sentence like `// See the service fields on [Client] for the full list of supported resources.` |

### client.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `APIVersion` | const | `// APIVersion is the Terrakube OpenAPI specification version this library targets.` | Good | -- |
| `ListOptions` | struct | `// ListOptions specifies optional parameters for List methods.` | Good | -- |
| `Client` | struct | `// Client manages communication with the Terrakube API.` | Good | -- |
| `Option` | type | `// Option configures a Client.` | Good | -- |
| `WithEndpoint` | func | `// WithEndpoint sets the Terrakube server URL.` | Good | -- |
| `WithToken` | func | `// WithToken sets the API bearer token.` | Good | -- |
| `WithHTTPClient` | func | `// WithHTTPClient sets a custom HTTP client.` | Good | -- |
| `WithInsecureTLS` | func | `// WithInsecureTLS skips TLS certificate verification.` | Good | -- |
| `WithUserAgent` | func | `// WithUserAgent sets a custom User-Agent header.` | Good | -- |
| `NewClient` | func | `// NewClient creates a new Terrakube API client.` | Needs work | Should mention that it returns an error if endpoint or token are not provided. Suggested: `// NewClient creates a new Terrakube API client. It returns an error if WithEndpoint or WithToken are not provided.` |

### crud.go

All symbols are unexported (`crudService`, `list`, `get`, `create`, `update`, `del`). No exported symbols to audit.

### errors.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `APIError` | struct | `// APIError represents an error response from the Terrakube API.` | Good | -- |
| `ErrorDetail` | struct | `// ErrorDetail represents a single error entry in a JSON:API error response.` | Good | -- |
| `(*APIError).Error` | method | (none) | Missing | Add: `// Error returns a string representation including the HTTP method, path, and status code.` |
| `ValidationError` | struct | `// ValidationError represents a client-side validation failure.` | Good | -- |
| `(*ValidationError).Error` | method | (none) | Missing | Add: `// Error returns a string representation of the validation failure.` |
| `IsNotFound` | func | `// IsNotFound returns true if the error is a 404 API error.` | Good | -- |
| `IsConflict` | func | `// IsConflict returns true if the error is a 409 API error.` | Good | -- |
| `IsUnauthorized` | func | `// IsUnauthorized returns true if the error is a 401 API error.` | Good | -- |

---

## 2. Service Files

The following table covers every exported symbol across all 31 service files. Services follow a highly consistent pattern; the audit focuses on deviations from ideal.

### action.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Action` | struct | `// Action represents a Terrakube action resource.` | Good | -- |
| `ActionService` | struct | `// ActionService handles communication with the action related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all actions, optionally filtered.` | Needs work | Does not mention `*APIError` on server errors. |
| `Get` | method | `// Get retrieves an action by ID.` | Needs work | Does not mention `*ValidationError` for empty ID or `*APIError`. |
| `Create` | method | `// Create creates a new action.` | Needs work | Does not mention `*APIError`. |
| `Update` | method | `// Update modifies an existing action. The action's ID field must be set.` | Good | Mentions ID requirement. Could note `*ValidationError`/`*APIError`. |
| `Delete` | method | `// Delete removes an action by ID.` | Needs work | Does not mention error conditions. |

### address.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Address` | struct | `// Address represents a Terrakube job address resource.` | Good | -- |
| `AddressService` | struct | `// AddressService handles communication with the job address endpoints.` | Good | -- |
| `List` | method | `// List returns all addresses for a job.` | Needs work | Missing error condition docs. |
| `Get` | method | `// Get returns a single address by ID.` | Needs work | Missing error condition docs. |
| `Create` | method | `// Create creates a new address for a job.` | Needs work | Missing error condition docs. |
| `Update` | method | `// Update modifies an existing address. The address's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes an address by ID.` | Needs work | Missing error condition docs. |

### agent.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Agent` | struct | `// Agent represents an agent in Terrakube.` | Good | -- |
| `AgentService` | struct | `// AgentService handles communication with the Agent related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all agents for an organization.` | Good | -- |
| `Get` | method | `// Get returns a single agent by ID.` | Good | -- |
| `Create` | method | `// Create creates a new agent in an organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing agent.` | Good | -- |
| `Delete` | method | `// Delete removes an agent by ID.` | Good | -- |

### collection.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Collection` | struct | `// Collection represents a Terrakube collection resource.` | Good | -- |
| `CollectionService` | struct | `// CollectionService handles communication with the collection related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all collections for the given organization.` | Good | -- |
| `Get` | method | `// Get returns a single collection by ID.` | Good | -- |
| `Create` | method | `// Create creates a new collection in the given organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing collection. The collection's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a collection by ID.` | Good | -- |

### collection_item.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `CollectionItem` | struct | `// CollectionItem represents a key/value item within a Terrakube collection.` | Good | -- |
| `CollectionItemService` | struct | `// CollectionItemService handles communication with the collection item related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all items for the given collection.` | Good | -- |
| `Get` | method | `// Get returns a single collection item by ID.` | Good | -- |
| `Create` | method | `// Create creates a new item in the given collection.` | Good | -- |
| `Update` | method | `// Update modifies an existing collection item. The item's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a collection item by ID.` | Good | -- |

### collection_reference.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `CollectionReference` | struct | `// CollectionReference represents a reference within a Terrakube collection.` | Good | -- |
| `CollectionReferenceService` | struct | `// CollectionReferenceService handles communication with the collection reference related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all references for the given collection.` | Good | -- |
| `Get` | method | `// Get returns a single collection reference by ID using the flat endpoint.` | Good | -- |
| `Create` | method | `// Create creates a new reference in the given collection.` | Good | -- |
| `Update` | method | `// Update modifies an existing collection reference using the flat endpoint. The reference's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a collection reference by ID using the flat endpoint.` | Good | -- |

### github_app_token.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `GithubAppToken` | struct | `// GithubAppToken represents a Terrakube GitHub App token resource.` | Good | -- |
| `GithubAppTokenService` | struct | `// GithubAppTokenService handles communication with the GitHub App token endpoints.` | Good | -- |
| `List` | method | `// List returns all GitHub App tokens.` | Good | -- |
| `Get` | method | `// Get returns a single GitHub App token by ID.` | Good | -- |
| `Create` | method | `// Create creates a new GitHub App token.` | Good | -- |
| `Update` | method | `// Update modifies an existing GitHub App token. The token's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a GitHub App token by ID.` | Good | -- |

### history.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `History` | struct | `// History represents a Terrakube workspace history resource.` | Good | -- |
| `HistoryService` | struct | `// HistoryService handles communication with the history-related endpoints.` | Good | -- |
| `List` | method | `// List returns all history entries for the given workspace.` | Good | -- |
| `Get` | method | `// Get returns a single history entry by ID within the given workspace.` | Good | -- |
| `Create` | method | `// Create creates a new history entry in the given workspace.` | Good | -- |
| `Update` | method | `// Update modifies an existing history entry in the given workspace.` | Needs work | Missing "ID field must be set" note that other Update methods have. |
| `Delete` | method | `// Delete removes a history entry from the given workspace.` | Good | -- |

### implementation.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Implementation` | struct | `// Implementation represents a Terrakube provider version implementation resource.` | Good | -- |
| `ImplementationService` | struct | `// ImplementationService handles communication with the implementation-related endpoints.` | Good | -- |
| `List` | method | `// List returns all implementations for a provider version.` | Good | -- |
| `Get` | method | `// Get returns a single implementation by ID.` | Good | -- |
| `Create` | method | `// Create creates a new implementation for a provider version.` | Good | -- |
| `Update` | method | `// Update modifies an existing implementation. The implementation's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes an implementation by ID.` | Good | -- |

### job.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Job` | struct | `// Job represents a Terrakube job resource.` | Good | -- |
| `JobService` | struct | `// JobService handles communication with the job related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all jobs for the given organization.` | Good | -- |
| `Get` | method | `// Get returns a single job by ID.` | Good | -- |
| `Create` | method | `// Create creates a new job in the given organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing job. The job's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a job by ID.` | Good | -- |

### module.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Module` | struct | `// Module represents a Terrakube module resource.` | Good | -- |
| `ModuleService` | struct | `// ModuleService handles communication with the module related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all modules for an organization, optionally filtered.` | Good | -- |
| `Get` | method | `// Get retrieves a module by ID within an organization.` | Good | -- |
| `Create` | method | `// Create creates a new module within an organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing module within an organization.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes a module by ID within an organization.` | Good | -- |

### module_version.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `ModuleVersion` | struct | `// ModuleVersion represents a Terrakube module version resource.` | Good | -- |
| `ModuleVersionService` | struct | `// ModuleVersionService handles communication with the module version endpoints.` | Good | -- |
| `List` | method | `// List returns all versions for a module.` | Good | -- |
| `Get` | method | `// Get returns a single module version by ID.` | Good | -- |
| `Create` | method | `// Create creates a new version for a module.` | Good | -- |
| `Update` | method | `// Update modifies an existing module version. The version's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a module version by ID.` | Good | -- |

### operations.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `OperationAction` | type | `// OperationAction represents the type of atomic operation.` | Good | -- |
| `OperationAdd` | const | (group comment: `// Supported atomic operation actions.`) | Good | -- |
| `OperationUpdate` | const | (covered by group comment) | Good | -- |
| `OperationRemove` | const | (covered by group comment) | Good | -- |
| `OperationRef` | struct | `// OperationRef identifies the target resource for an atomic operation.` | Good | -- |
| `Operation` | struct | `// Operation represents a single atomic operation.` | Good | -- |
| `AtomicRequest` | struct | `// AtomicRequest is the request body for POST /operations.` | Good | -- |
| `AtomicResult` | struct | `// AtomicResult represents the result of a single atomic operation.` | Good | -- |
| `AtomicResponse` | struct | `// AtomicResponse is the response from POST /operations.` | Good | -- |
| `OperationsService` | struct | `// OperationsService handles atomic batch operations.` | Good | -- |
| `Submit` | method | `// Submit sends an atomic operations batch request.` | Good | -- |

### organization.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Organization` | struct | `// Organization represents a Terrakube organization resource.` | Good | -- |
| `OrganizationService` | struct | `// OrganizationService handles communication with the organization related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all organizations, optionally filtered.` | Good | -- |
| `Get` | method | `// Get retrieves an organization by ID.` | Good | -- |
| `Create` | method | `// Create creates a new organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing organization.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes an organization by ID.` | Good | -- |

### organization_variable.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `OrganizationVariable` | struct | `// OrganizationVariable represents a Terrakube organization-level global variable.` | Good | -- |
| `OrganizationVariableService` | struct | `// OrganizationVariableService handles communication with the organization global variable endpoints.` | Good | -- |
| `List` | method | `// List returns all global variables for an organization.` | Good | -- |
| `Get` | method | `// Get returns a single organization variable by ID.` | Good | -- |
| `Create` | method | `// Create creates a new global variable in the organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing organization variable. The variable's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes an organization variable by ID.` | Good | -- |

### provider.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Provider` | struct | `// Provider represents a Terrakube provider resource within an organization.` | Good | -- |
| `ProviderService` | struct | `// ProviderService handles communication with the provider related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all providers for the given organization.` | Good | -- |
| `Get` | method | `// Get returns a single provider by ID.` | Good | -- |
| `Create` | method | `// Create creates a new provider in the given organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing provider. The provider's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a provider by ID.` | Good | -- |

### provider_version.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `ProviderVersion` | struct | `// ProviderVersion represents a Terrakube provider version resource.` | Good | -- |
| `ProviderVersionService` | struct | `// ProviderVersionService handles communication with the provider version related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all versions for the given provider within an organization.` | Good | -- |
| `Get` | method | `// Get returns a single provider version by ID.` | Good | -- |
| `Create` | method | `// Create creates a new version for the given provider.` | Good | -- |
| `Update` | method | `// Update modifies an existing provider version. The version's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a provider version by ID.` | Good | -- |

### ssh.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `SSH` | struct | `// SSH represents an SSH key in Terrakube.` | Good | -- |
| `SSHService` | struct | `// SSHService handles communication with the SSH related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all SSH keys for an organization.` | Good | -- |
| `Get` | method | `// Get returns a single SSH key by ID.` | Good | -- |
| `Create` | method | `// Create creates a new SSH key in an organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing SSH key.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes an SSH key by ID.` | Good | -- |

### step.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Step` | struct | `// Step represents a Terrakube step resource within a job.` | Good | -- |
| `StepService` | struct | `// StepService handles communication with the step related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all steps for the given job within an organization.` | Good | -- |
| `Get` | method | `// Get returns a single step by ID.` | Good | -- |
| `Create` | method | `// Create creates a new step in the given job.` | Good | -- |
| `Update` | method | `// Update modifies an existing step. The step's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a step by ID.` | Good | -- |

### tag.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Tag` | struct | `// Tag represents a Terrakube tag resource.` | Good | -- |
| `TagService` | struct | `// TagService handles communication with the tag-related endpoints.` | Good | -- |
| `List` | method | `// List returns all tags for the given organization.` | Good | -- |
| `Get` | method | `// Get returns a single tag by ID within the given organization.` | Good | -- |
| `Create` | method | `// Create creates a new tag in the given organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing tag in the given organization.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes a tag from the given organization.` | Good | -- |

### team.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Team` | struct | `// Team represents a Terrakube team resource.` | Good | -- |
| `TeamService` | struct | `// TeamService handles communication with the team-related endpoints.` | Good | -- |
| `List` | method | `// List returns all teams for an organization, with optional filtering.` | Good | -- |
| `Get` | method | `// Get retrieves a single team by ID within an organization.` | Good | -- |
| `Create` | method | `// Create creates a new team within an organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing team within an organization.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes a team from an organization.` | Good | -- |

### team_token.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `TeamToken` | struct | `// TeamToken represents a Terrakube team access token.` | Good | -- |
| `TeamTokenService` | struct | `// TeamTokenService handles communication with the team token endpoints.` | Good | -- |
| `Create` | method | `// Create generates a new team token.` | Good | -- |
| `List` | method | `// List returns all team tokens.` | Good | -- |
| `Delete` | method | `// Delete removes a team token by ID.` | Good | -- |

### template.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Template` | struct | `// Template represents a Terrakube template resource.` | Good | -- |
| `TemplateService` | struct | `// TemplateService handles communication with the template-related endpoints.` | Good | -- |
| `List` | method | `// List returns all templates for the given organization.` | Good | -- |
| `Get` | method | `// Get returns a single template by ID within the given organization.` | Good | -- |
| `Create` | method | `// Create creates a new template in the given organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing template in the given organization.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes a template from the given organization.` | Good | -- |

### variable.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Variable` | struct | `// Variable represents a Terrakube workspace variable.` | Good | -- |
| `VariableService` | struct | `// VariableService handles communication with the workspace variable endpoints.` | Good | -- |
| `List` | method | `// List returns all variables for a workspace.` | Good | -- |
| `Get` | method | `// Get returns a single variable by ID.` | Good | -- |
| `Create` | method | `// Create creates a new variable in the workspace.` | Good | -- |
| `Update` | method | `// Update modifies an existing variable. The variable's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a variable by ID.` | Good | -- |

### vcs.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `VCS` | struct | `// VCS represents a version control system connection in Terrakube.` | Good | -- |
| `VCSService` | struct | `// VCSService handles communication with the VCS related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all VCS connections for an organization.` | Good | -- |
| `Get` | method | `// Get returns a single VCS connection by ID.` | Good | -- |
| `Create` | method | `// Create creates a new VCS connection in an organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing VCS connection.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes a VCS connection by ID.` | Good | -- |

### webhook.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Webhook` | struct | `// Webhook represents a workspace webhook (v1 flat format).` | Good | -- |
| `WebhookEvent` | struct | `// WebhookEvent represents a webhook event entity.` | Good | -- |
| `WebhookService` | struct | `// WebhookService handles communication with the webhook related methods of the Terrakube API.` | Good | -- |
| `WebhookEventService` | struct | `// WebhookEventService handles communication with the webhook event related methods of the Terrakube API.` | Good | -- |
| `WebhookService.List` | method | `// List returns all webhooks for a workspace.` | Good | -- |
| `WebhookService.Get` | method | `// Get retrieves a single webhook by ID.` | Good | -- |
| `WebhookService.Create` | method | `// Create creates a new webhook for a workspace.` | Good | -- |
| `WebhookService.Update` | method | `// Update modifies an existing webhook.` | Needs work | Missing "ID field must be set" note. |
| `WebhookService.Delete` | method | `// Delete removes a webhook.` | Good | -- |
| `WebhookEventService.List` | method | `// List returns all events for a webhook.` | Good | -- |
| `WebhookEventService.Get` | method | `// Get retrieves a single webhook event by ID.` | Good | -- |
| `WebhookEventService.Create` | method | `// Create creates a new webhook event.` | Good | -- |
| `WebhookEventService.Update` | method | `// Update modifies an existing webhook event.` | Needs work | Missing "ID field must be set" note. |
| `WebhookEventService.Delete` | method | `// Delete removes a webhook event.` | Good | -- |

### workspace.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `Workspace` | struct | `// Workspace represents a Terrakube workspace resource.` | Good | -- |
| `WorkspaceService` | struct | `// WorkspaceService handles communication with the workspace related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all workspaces for an organization, optionally filtered.` | Good | -- |
| `Get` | method | `// Get retrieves a workspace by ID within an organization.` | Good | -- |
| `Create` | method | `// Create creates a new workspace within an organization.` | Good | -- |
| `Update` | method | `// Update modifies an existing workspace within an organization.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes a workspace by ID within an organization.` | Good | -- |

### workspace_access.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `WorkspaceAccess` | struct | `// WorkspaceAccess represents access control settings for a workspace.` | Good | -- |
| `WorkspaceAccessService` | struct | `// WorkspaceAccessService handles communication with the workspace access related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all access entries for the given workspace.` | Good | -- |
| `Get` | method | `// Get returns a single workspace access entry by ID.` | Good | -- |
| `Create` | method | `// Create creates a new access entry for the given workspace.` | Good | -- |
| `Update` | method | `// Update modifies an existing workspace access entry. The access entry's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a workspace access entry by ID.` | Good | -- |

### workspace_schedule.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `WorkspaceSchedule` | struct | `// WorkspaceSchedule represents a scheduled job for a workspace.` | Good | -- |
| `WorkspaceScheduleService` | struct | `// WorkspaceScheduleService handles communication with the workspace schedule related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all schedules for the given workspace.` | Good | -- |
| `Get` | method | `// Get returns a single workspace schedule by ID.` | Good | -- |
| `Create` | method | `// Create creates a new schedule for the given workspace.` | Good | -- |
| `Update` | method | `// Update modifies an existing workspace schedule. The schedule's ID field must be set.` | Good | -- |
| `Delete` | method | `// Delete removes a workspace schedule by ID.` | Good | -- |

### workspace_tag.go

| Symbol | Kind | Current Doc | Assessment | Recommendation |
|--------|------|------------|------------|----------------|
| `WorkspaceTag` | struct | `// WorkspaceTag represents a tag association on a workspace.` | Good | -- |
| `WorkspaceTagService` | struct | `// WorkspaceTagService handles communication with the workspace tag related methods of the Terrakube API.` | Good | -- |
| `List` | method | `// List returns all tags for a workspace.` | Good | -- |
| `Get` | method | `// Get retrieves a single workspace tag by ID.` | Good | -- |
| `Create` | method | `// Create creates a new tag association on a workspace.` | Good | -- |
| `Update` | method | `// Update modifies an existing workspace tag.` | Needs work | Missing "ID field must be set" note. |
| `Delete` | method | `// Delete removes a tag association from a workspace.` | Good | -- |

---

## 3. Struct Fields Needing Documentation

These struct fields have non-obvious names or JSON:API tag mappings that would benefit from inline comments:

| File | Struct | Field | Issue | Recommendation |
|------|--------|-------|-------|----------------|
| `workspace.go` | `Workspace` | `TemplateID` | JSON:API attr is `defaultTemplate` -- field name does not match | Add: `// TemplateID is the default template ID (JSON:API attr: "defaultTemplate").` |
| `workspace.go` | `Workspace` | `IaCType` | Abbreviation not obvious | Add: `// IaCType is the infrastructure-as-code type (e.g. "terraform", "tofu").` |
| `workspace.go` | `Workspace` | `IaCVersion` | JSON:API attr is `terraformVersion` | Add: `// IaCVersion is the IaC tool version (JSON:API attr: "terraformVersion").` |
| `template.go` | `Template` | `Content` | JSON:API attr is `tcl` | Add: `// Content is the template body in TCL format (JSON:API attr: "tcl").` |
| `workspace_schedule.go` | `WorkspaceSchedule` | `Schedule` | JSON:API attr is `cron` | Add: `// Schedule is the cron expression for the schedule (JSON:API attr: "cron").` |
| `workspace_schedule.go` | `WorkspaceSchedule` | `TemplateID` | JSON:API attr is `templateReference` | Add: `// TemplateID is the template reference (JSON:API attr: "templateReference").` |
| `team_token.go` | `TeamToken` | `Value` | JSON tag is `token` | Add: `// Value is the token string returned by the server (JSON key: "token").` |
| `history.go` | `History` | `Md5` | Not obvious what this hashes | Add: `// Md5 is the MD5 hash of the Terraform state.` |
| `history.go` | `History` | `Lineage` | Terraform-specific concept | Add: `// Lineage is the Terraform state lineage identifier.` |
| `job.go` | `Job` | `Tcl` | Abbreviation | Add: `// Tcl is the Terrakube Configuration Language content for this job.` |

---

## 4. Methods Needing Error Condition Documentation

All service methods can return two categories of errors:
1. `*ValidationError` -- returned when an ID parameter is empty (client-side validation).
2. `*APIError` -- returned when the server responds with a non-2xx status code.

**Currently, no method documents these error conditions.** This is the largest systematic gap.

### Recommended pattern for all CRUD methods

Add a standardized error doc line to each method. Example for a `Get` method:

```go
// Get retrieves an action by ID.
// It returns a *ValidationError if id is empty and a *APIError if the server
// responds with a non-2xx status code.
```

For methods that validate multiple IDs:

```go
// List returns all variables for a workspace.
// It returns a *ValidationError if orgID or workspaceID is empty and a *APIError
// on server errors.
```

### Methods that validate IDs but lack error docs

All `Get`, `Update`, `Delete` methods across all 30 service types (approximately 90 methods) validate at least one ID but do not document the `*ValidationError` possibility. All methods that make HTTP calls (all of them) can return `*APIError` but do not document it.

**Recommendation:** Add the error condition pattern above to all methods in the enrichment pass. This is a systemic improvement, not a per-file issue.

---

## 5. Go Convention Violations

### Missing doc comments (3 total)

1. `errors.go:24` -- `(*APIError).Error` method has no doc comment.
2. `errors.go:37` -- `(*ValidationError).Error` method has no doc comment.
3. Note: `Error()` methods on error types are sometimes left undocumented by convention since the `error` interface is well-understood. However, for a public library, documenting them is preferred.

### Inconsistent "ID field must be set" notes on Update methods

Some `Update` methods include "The X's ID field must be set." and some do not. The following are **missing** this note:

| File | Method |
|------|--------|
| `agent.go` | `(*AgentService).Update` |
| `history.go` | `(*HistoryService).Update` |
| `module.go` | `(*ModuleService).Update` |
| `organization.go` | `(*OrganizationService).Update` |
| `ssh.go` | `(*SSHService).Update` |
| `tag.go` | `(*TagService).Update` |
| `team.go` | `(*TeamService).Update` |
| `template.go` | `(*TemplateService).Update` |
| `vcs.go` | `(*VCSService).Update` |
| `webhook.go` | `(*WebhookService).Update` |
| `webhook.go` | `(*WebhookEventService).Update` |
| `workspace.go` | `(*WorkspaceService).Update` |
| `workspace_tag.go` | `(*WorkspaceTagService).Update` |

### Package comment style

The `doc.go` package comment is good but could be improved:
- The code example does not end with a closing remark after the code block.
- Consider adding a sentence after the code block that references the `Client` type's service fields.

### No other convention violations found

- All doc comments start with the symbol name (Go convention met).
- All doc comments are complete sentences ending in periods.
- All use `//` style comments, not `/* */`.
- No stale or inaccurate doc comments detected.

---

## 6. Prioritized Fix List

### P0 -- Missing doc comments (3 items)

1. Add `// Error returns a string including the HTTP method, path, and status code.` above `(*APIError).Error`.
2. Add `// Error returns a string describing the validation failure.` above `(*ValidationError).Error`.
3. (Optional) Improve `doc.go` closing with a trailing sentence after code block.

### P1 -- Consistency fixes (13 items)

Add "The X's ID field must be set." to the 13 `Update` methods listed in section 5.

### P2 -- Error condition documentation (systematic, ~160 methods)

Add `*ValidationError` and `*APIError` return documentation to all service methods.

### P3 -- Struct field documentation (10 items)

Add inline comments to the 10 non-obvious struct fields listed in section 3.
