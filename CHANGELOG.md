# Changelog

Notable changes to this library. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) as described in the [compatibility policy](README.md#compatibility).

Releases up to v0.5.0 are described in the [GitHub release notes](https://github.com/terrakube-io/terrakube-go/releases).

## [Unreleased]

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

[Unreleased]: https://github.com/terrakube-io/terrakube-go/compare/v0.5.0...HEAD
