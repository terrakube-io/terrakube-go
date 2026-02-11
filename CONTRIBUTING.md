# Contributing

## Development Setup

1. Install [Go 1.24+](https://go.dev/dl/)
2. Install [mise](https://mise.jdx.dev/) for tool management
3. Clone the repo and install dependencies:

```bash
git clone https://github.com/denniswebb/terrakube-go.git
cd terrakube-go
mise install
go mod download
```

## Code Standards

- All exported types and functions must have doc comments
- No `omitempty` on boolean `jsonapi` struct tags (booleans must serialize `false`)
- Use `*string` for optional string fields
- All service methods take `context.Context` as first parameter
- Validate non-empty IDs before building URL paths
- Use `google/jsonapi` for JSON:API serialization, `encoding/json` only for non-JSON:API endpoints

## Testing

Every service file has a corresponding `_test.go` covering:

- Successful CRUD operations
- Error responses (404, 422, 500)
- Empty ID validation
- Boolean serialization (false not dropped)
- Auth header presence

Run tests:

```bash
mise run test       # all tests
mise run coverage   # with coverage report
mise run check      # vet + lint + test
```

## Commit Format

Use [conventional commits](https://www.conventionalcommits.org/):

```
type(scope): description

feat(workspace): add schedule support
fix(client): check HTTP status codes in do()
test(team): add boolean serialization tests
docs(readme): add webhook examples
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

## Pull Requests

1. Create a feature branch from `main`
2. Make changes with tests
3. Ensure `mise run check` passes
4. Open a PR against `main`
