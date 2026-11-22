# Changelog

Notable changes to this library. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) as described in the [compatibility policy](README.md#compatibility).

Releases up to v0.5.0 are described in the [GitHub release notes](https://github.com/terrakube-io/terrakube-go/releases).

## [Unreleased]

### Breaking changes

- **`Team.DestroyJob`, `Team.ManagePolicies`, `ProjectAccess.DestroyJob` and `WorkspaceAccess.DestroyJob` are now `*bool`** (tag `omitempty`). Older Terrakube servers leave these flags out, and as a plain `bool` a missing flag was read as `false`, so callers could not tell "not returned" from "false". A missing flag is now `nil`, and a `nil` flag is left out of requests instead of being sent as `false`, so `Update` leaves it unchanged on the server. This change to the exported API ships in a minor release on purpose: the old type could not report the server's real value, and the Terraform provider is the only user of these fields we know of. To migrate, write `DestroyJob: new(true)` instead of `DestroyJob: true`, and check for `nil` before dereferencing a value you read.

## [1.4.0] - 2026-11-15

### Added

- **Methods that address a resource by its own ID** on the server's top-level paths, for callers that do not know its organization or workspace (for example a Terraform resource imported by ID). Each method that takes an ID returns a `*ValidationError` when it is empty.
  - `RunTriggers.CreateFieldsTopLevel`, `GetByID`, `UpdateFieldsByID` and `DeleteByID` on `/runTrigger[/{id}]`.
  - `Webhooks.GetByID`, `WebhookEvents.GetByID` and `Workspaces.GetByID` on `/webhook/{id}`, `/webhook_event/{id}` and `/workspace/{id}`. A webhook's `Workspace` and a workspace's `Organization` relationship then give the parent IDs.
  - `PolicySets`, `PolicyExemptions` and `NotificationConfigurations`: `GetByID`, `UpdateFieldsByID` and `DeleteByID` on `/policy_set/{id}`, `/policy_exemption/{id}` and `/notification_configuration/{id}`. The last also serves workspace-level configurations.
  - `PolicySetParameters` and `PolicyAttachments`: `ListByPolicySet`, `GetByPolicySet`, `CreateFieldsByPolicySet`, `UpdateFieldsByPolicySet` and `DeleteByPolicySet` on `/policy_set/{id}/parameters` and `/policy_set/{id}/attachments`.
  - `NotificationTriggers.ListByConfiguration`, `CreateFieldsByConfiguration` and `DeleteByConfiguration` on `/notification_configuration/{id}/triggers`.
  - Each method has its own interface, next to the resource's other interfaces: `<Resource>ByIDGetter`, `<Resource>ByIDFieldsUpdater` and `<Resource>ByIDDeleter`, `RunTriggerTopLevelFieldsCreator`, `PolicySetParameterByPolicySetLister` (and `Getter`, `FieldsCreator`, `FieldsUpdater`, `Deleter`), the same for `PolicyAttachment`, and `NotificationTriggerByConfigurationLister`, `FieldsCreator` and `Deleter`.

## [1.3.0] - 2026-11-15

### Added

- **`PATs` service** for the caller's personal access tokens (`/pat/v1`). `Create(ctx, days, description)` returns a `PATCreated` with the token, which the server returns only once, and its `ID`, read from the token's `jti` claim. `days` 0 creates a token that never expires (the server refuses it when it has a maximum lifetime); a negative `days` returns a `*ValidationError`. `List` returns the tokens that are not deleted, expired ones included, and `Delete` revokes one. The interfaces are `PATLister`, `PATCreator`, `PATDeleter` and `PATCRUD`.
- **`Access` service** for what the caller may do. `CurrentTeams` returns the groups the server sees for the caller. `Permissions` and `WorkspacePermissions` return the caller's effective `Permissions` in an organization or a workspace, as the server computes them from the caller's teams and the workspace and project access entries; an organization the caller has no team in gives no permissions rather than an error. The interfaces are `CurrentTeamsLister`, `PermissionsGetter` and `WorkspacePermissionsGetter`.

## [1.2.0] - 2026-11-15

### Added

- **`StepLogs` service** for job step logs. `Get` returns a step's log (the output so far while the step runs); `StepLogOptions.Offset` reads from a byte offset on with a `Range` request, and `StepLog.Incomplete` reports the server's `X-Terrakube-Output-Incomplete` mark. `Follow` reads the log's Server-Sent Events as they are written and calls a function for each output event and each gap (output the server could not replay; read it with `Get`). It returns nil when the job is complete and reconnects with `Last-Event-ID` after the server rotates the stream, a dropped connection or a 429/5xx answer, without delivering an event twice; it gives up after `StepLogFollowOptions.MaxFailures` failed attempts in a row. The interfaces are `StepLogGetter` and `StepLogFollower`.
- `APIError.RetryAfter` holds the wait a `Retry-After` header asked for, such as on a 503.
- `ErrUnsupported` also matches a missing route on `/tfoutput/v1/organization/{org}/job/{job}/step/{step}`. An unknown job or step there is a plain 404 and does not match.

## [1.1.0] - 2026-11-15

### Added

- **Sparse writes.** Every service with `Create` and `Update` gains `CreateFields` and `UpdateFields` (with the same parent IDs), which take a `Patch` and send only the attributes and relationships it names. `CreateFields` lets the server apply its defaults to everything else, and `UpdateFields` rejects an empty `Patch` with a `*ValidationError`. `NotificationConfigurations.CreateForWorkspaceFields` is the sparse form of `CreateForWorkspace`. Each method has a matching `<Resource>FieldsCreator` or `<Resource>FieldsUpdater` interface; the existing `<Resource>CRUD` interfaces are unchanged.
  - A to-one `Relationship` sets `ID`; a nil `ID` clears the link (`{"data": null}`).
  - A to-many `Relationship` sets `Many` and lists every target in `IDs`, which replaces the set; an empty list clears it (`{"data": []}`).
- **`State` service** for workspace state: `Current` streams the current raw state, `Download` streams the raw (`StateRaw`) or `terraform show -json` (`StateJSON`) file of a history entry, `Rollback` makes a history entry current, and `StateName` returns a history entry's state file name. An empty response returns `ErrStateNotFound`. `Download` reads a history entry whose `output` has the API's state path (`/tfstate/v1/organization/…/state/{name}.json`) from the configured endpoint, whatever host the server or executor wrote into it (often an in-cluster address).
- **`Outputs` service.** `List` reads a workspace's outputs, with their real types and `sensitive` flag, from the JSON state of the newest history entry that names a state file, falling back to the current raw state. A workspace without history has no outputs.
- `ErrForeignHost`: a URL taken from server data that is neither on the configured endpoint nor an API state path is refused before any request, so the token is never sent to another host.
- `ErrUnsupported`, matched with `errors.Is`, for an error that shows the server lacks a resource type (JSON:API 404 `Unknown collection`) or one of the newer raw endpoints (404 or 405 for a missing route, not for a missing entity). Known limit: on `/tfoutput/v1/…/stream`, Terrakube 2.34.1 answers an unknown job or step with Spring's default error body, which looks like a missing route, so that error also matches `ErrUnsupported`.
- `ListOptions.Sort` sets the `sort` query parameter (for example `-createdDate`). When a collection is listed page by page, `id` is added as a tie-breaker unless the order already uses it. `ListOptions` gaining a field breaks only unkeyed struct literals (`ListOptions{"x", "", 0, 0}`), which the Go 1 compatibility guidelines do not cover; use field names.

### Changed

- The client no longer follows a redirect off the configured endpoint (another scheme, host, port or path prefix), so a redirect cannot carry the token elsewhere; the request fails with an error that wraps `ErrForeignHost`. An `http` endpoint may still be redirected to `https` on the same host and path. A `CheckRedirect` set on your `http.Client` still applies to redirects that stay on the endpoint, and your client is not modified.
- IDs passed to service methods must be a single path segment: an ID containing `/`, or equal to `.` or `..`, returns a `*ValidationError` instead of sending a request to another path.

## [1.0.0] - 2026-11-14

Targets Terrakube API 2.34.1.

### Breaking changes

- **Go 1.27 is required.** v0.5.0 is the last release for Go 1.24 to 1.26.
- **Date attributes are `*time.Time`.** `CreatedDate`, `UpdatedDate`, `LastJobDate`, `ExpiresAt`, `ApprovedAt`, `OverriddenAt` and the job dispatch timestamps were `*string` (`string` on `WebhookEvent`). Replace string parsing with the time value; a nil pointer still means "not set".
- **Attributes the API does not expose are removed.** The server never returned them and rejected some of them on create (`Job.Command`):
  - `Job.Command`, `Job.Tcl`, `WorkspaceSchedule.Tcl`
  - `Address.Type`, `Agent.Type`, `Step.LogStatus`
  - `VCS.RefreshToken`, `VCS.TokenExpiration`
  - `Webhook.Path`, `Webhook.Branch`, `Webhook.TemplateID`, `Webhook.Event` (these live on `WebhookEvent`)
  - `CreatedBy`, `CreatedDate`, `UpdatedBy`, `UpdatedDate` on `Variable`, `Step` and `ModuleVersion`
- **`TeamTokens.List` returns `[]*TeamToken`** instead of `[]TeamToken`, like every other `List`.
- **`ListOptions.Page` and `PageSize` return a `*ValidationError`** on resource types the server cannot paginate, without sending a request. Before, the server answered 400 `Cannot paginate`. They work for workspaces, workspace tags, actions, federated credentials, agents, organization variables and workspace access.
- **`WithInsecureTLS` returns an error from `NewClient`** when the HTTP client's transport is not an `*http.Transport`. Before, it silently replaced that transport with a default one. Disable verification inside your own transport instead.

### Added

- Terrakube API 2.34.1: policy sets, policy set parameters, policy attachments, policy exemptions, policy evaluations, policy overrides and run triggers, plus the new attributes and relationships of existing resources.
- `tools/speccheck`, a separate module that compares the client with the vendored OpenAPI document and updates it for a new Terrakube version.
- `WithEndpoint` accepts a URL with a path prefix (`https://example.com/terrakube`) and keeps it for every request.

### Changed

- `List` without page options returns the whole collection for the resource types the server can paginate, ordered by id. It used to stop at the server's limit (500 or 10,000 records, depending on the type) without an error.
- `List` no longer follows `links` found in responses. Every request goes to the configured endpoint.
- `WithInsecureTLS` is applied after all other options, so its position relative to `WithHTTPClient` does not matter, and the `http.Client` you pass in is never modified.

### Fixed

- `Operations.Submit` sends the JSON:API atomic operations media type; the server answered 415 to every call. An operation with only `Href` no longer carries an empty `ref`, which the server rejects. Error details of failed operations reach `APIError.Errors`.
- `Webhooks.List` works for a workspace that has a webhook. It failed to decode the response and now returns zero or one element.
- `Modules.List` no longer fails when a module has versions.
- `WithHTTPClient(nil)` returns an error instead of causing a nil dereference on the first request.
- Every `Update` returns the updated resource. The server answers an update with 204 and no body, so the methods returned an empty struct; the state is now read back from the server.
- `Jobs.Create` works. The server rejects a create request that mentions an attribute only it may set, even as `null`, and the library sent all of them. Those `Job` attributes, and `Webhook.SyncError`, are now sent only when set.

### Documentation

- `Update` sends every attribute of the struct, so start from the result of `Get`. This is now stated on every `Update` method, in the package documentation and in the README.

[Unreleased]: https://github.com/terrakube-io/terrakube-go/compare/v1.4.0...HEAD
[1.4.0]: https://github.com/terrakube-io/terrakube-go/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/terrakube-io/terrakube-go/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/terrakube-io/terrakube-go/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/terrakube-io/terrakube-go/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/terrakube-io/terrakube-go/compare/v0.5.0...v1.0.0
