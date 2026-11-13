# terrakube-go

[![Go Reference](https://pkg.go.dev/badge/github.com/terrakube-io/terrakube-go.svg)](https://pkg.go.dev/github.com/terrakube-io/terrakube-go)

Go client library for the [Terrakube](https://terrakube.io) API.

## Installation

```bash
go get github.com/terrakube-io/terrakube-go
```

Requires Go 1.27 or later.

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    terrakube "github.com/terrakube-io/terrakube-go"
)

func main() {
    client, err := terrakube.NewClient(
        terrakube.WithEndpoint("https://terrakube.example.com"),
        terrakube.WithToken("your-api-token"),
    )
    if err != nil {
        log.Fatal(err)
    }

    orgs, err := client.Organizations.List(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, org := range orgs {
        fmt.Printf("Organization: %s (%s)\n", org.Name, org.ID)
    }
}
```

### Listing

`List` with no page options lists the collection and never follows links found in responses. Workspaces, workspace tags, actions, federated, agents, organization variables, and workspace access are fetched page by page until the server reports no more. Every other resource type, and any type for which the server refuses paging, is fetched in one request, so the result stops at any limit the server applies to it.

`ListOptions.Page` and `PageSize` return one page. On Terrakube 2.34.1 the server accepts them for those seven types and answers 400 `Cannot paginate` for most others.

## Supported Resources

| Resource | Service Field | Scope |
|----------|--------------|-------|
| Organization | `Organizations` | Top-level |
| Workspace | `Workspaces` | Organization |
| Module | `Modules` | Organization |
| Module Version | `ModuleVersions` | Module |
| Team | `Teams` | Organization |
| Team Token | `TeamTokens` | Organization |
| Template | `Templates` | Organization |
| Tag | `Tags` | Organization |
| VCS | `VCS` | Organization |
| SSH | `SSH` | Organization |
| Agent | `Agents` | Organization |
| Collection | `Collections` | Organization |
| Collection Item | `CollectionItems` | Collection |
| Collection Reference | `CollectionReferences` | Collection |
| Variable | `Variables` | Workspace |
| Organization Variable | `OrganizationVariables` | Organization |
| Workspace Tag | `WorkspaceTags` | Workspace |
| Workspace Access | `WorkspaceAccess` | Workspace |
| Workspace Schedule | `WorkspaceSchedules` | Workspace |
| Webhook | `Webhooks` | Workspace |
| Webhook Event | `WebhookEvents` | Webhook |
| History | `History` | Workspace |
| Job | `Jobs` | Organization |
| Action | `Actions` | Top-level |
| Step | `Steps` | Job |
| Address | `Addresses` | Job |
| Project | `Projects` | Organization |
| Project Access | `ProjectAccess` | Project |
| Federated | `Federated` | Top-level |
| Federated Claim | `FederatedClaims` | Federated |
| Provider | `Providers` | Organization |
| Provider Version | `ProviderVersions` | Provider |
| Implementation | `Implementations` | Provider Version |
| GitHub App Token | `GithubAppTokens` | Top-level |
| Notification Configuration | `NotificationConfigurations` | Organization |
| Notification Trigger | `NotificationTriggers` | Notification Configuration |
| Policy Set | `PolicySets` | Organization |
| Policy Set Parameter | `PolicySetParameters` | Policy Set |
| Policy Attachment | `PolicyAttachments` | Policy Set |
| Policy Exemption | `PolicyExemptions` | Organization |
| Policy Evaluation | `PolicyEvaluations` | Job (read-only) |
| Policy Override | `PolicyOverrides` | Policy Evaluation |
| Run Trigger | `RunTriggers` | Workspace |

## API Version

The `APIVersion` constant tracks which Terrakube OpenAPI specification version this library targets.

```go
fmt.Println(terrakube.APIVersion) // "2.34.1"
```

## Configuration

| Option | Description | Required |
|--------|-------------|----------|
| `WithEndpoint(url)` | Terrakube server URL | Yes |
| `WithToken(token)` | API bearer token | Yes |
| `WithHTTPClient(client)` | Custom `*http.Client` | No |
| `WithInsecureTLS()` | Skip TLS verification | No |
| `WithUserAgent(ua)` | Custom User-Agent header | No |

## Error Handling

```go
org, err := client.Organizations.Get(ctx, "org-id")
if err != nil {
    if terrakube.IsNotFound(err) {
        // handle 404
    }
    if terrakube.IsUnauthorized(err) {
        // handle 401
    }
    // handle other errors
}
```

## Development

Requires Go 1.27+ and [mise](https://mise.jdx.dev/).

```bash
mise install          # install tools
mise run test         # run tests
mise run lint         # run linter
mise run check        # vet + lint + test
mise run coverage     # test with coverage report
```

## OpenAPI Spec Tracking

The library supports one Terrakube API version at a time, the one in the `APIVersion` constant. Its OpenAPI document is vendored in [`openapi/`](openapi/). The `speccheck` tool in [`tools/speccheck`](tools/speccheck) (a separate Go module, so its dependencies never reach the library) keeps the client and the document in step:

```bash
mise run spec-check                                  # compare models, paths, and filter keys with openapi/v{APIVersion}.json
mise run spec-check -- --verbose                     # also list fields that exist only in Go
mise run spec-fetch -- --server=https://terrakube-api.example.com --version=2.35.0   # token from $TERRAKUBE_TOKEN
mise run spec-diff -- --from=2.34.1 --to=2.35.0     # what changed, with semver guidance
mise run spec-sync -- --to=2.35.0 --dry-run         # what sync would change
mise run spec-sync -- --to=2.35.0                   # update the source
```

To move to a new Terrakube release:

1. Fetch its document. Terrakube no longer publishes current documents on GitHub, so fetch it from a running server of that version (`GET /doc`, bearer token from `$TERRAKUBE_TOKEN`). `--repo`/`--ref` fetch from GitHub when an upstream file exists.
2. Run `spec-diff` against the current version to see what changed.
3. Run `spec-sync --to=<version>`. It updates `APIVersion`, fixes resource types and filter keys, adds new attributes and relationships, marks fields the document no longer lists with a `// DEPRECATED (...)` comment (it never deletes them), and generates new resources with their service, interfaces, and tests.
4. Review the generated code, add new resources to the table above, and run `mise run check` and `mise run spec-check -- --fail-on-missing`.

CI runs the tool's tests and fails when the client no longer matches the vendored document.

## Documentation

Full API reference is available on [pkg.go.dev](https://pkg.go.dev/github.com/terrakube-io/terrakube-go). This README covers quickstart and configuration basics.

## License

Apache 2.0 - see [LICENSE](LICENSE).
