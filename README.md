# terrakube-go

Go client library for the [Terrakube](https://terrakube.io) API.

## Installation

```bash
go get github.com/denniswebb/terrakube-go
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    terrakube "github.com/denniswebb/terrakube-go"
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

## Supported Resources

| Resource | Service Field | Scope |
|----------|--------------|-------|
| Organization | `Organizations` | Top-level |
| Workspace | `Workspaces` | Organization |
| Module | `Modules` | Organization |
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

Requires Go 1.24+ and [mise](https://mise.jdx.dev/).

```bash
mise install          # install tools
mise run test         # run tests
mise run lint         # run linter
mise run check        # vet + lint + test
mise run coverage     # test with coverage report
```

## License

Apache 2.0 - see [LICENSE](LICENSE).
