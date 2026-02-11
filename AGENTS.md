# Terrakube Go Client Library

## Architecture

Single flat `terrakube` package at the repo root. No sub-packages except `testutil/`.

### Three layers

1. **Client** (`client.go`): HTTP transport, authentication, JSON:API serialization
2. **Services** (`*.go`): One file per resource type, each containing entity struct + service struct + CRUD methods
3. **Test infrastructure** (`testutil/`): httptest server wrapper and fixture helpers

### Key design decisions

- `google/jsonapi` for JSON:API serialization (uses reflection on concrete structs)
- Functional options pattern for `NewClient`
- NO `omitempty` on boolean struct tags (fixes dropped `false` values)
- `*string` for optional string fields
- `context.Context` on all service methods
- HTTP status code checking on all responses (non-2xx returns `*APIError`)
- ID validation before URL path construction

### Non-JSON:API endpoints

TeamToken uses `/access-token/v1/teams` with standard JSON. Uses `requestRaw`/`doRaw` methods instead of `request`/`do`.

## Testing

- Tests use `package terrakube_test` (external test package)
- `testutil/` MUST NOT import the main `terrakube` package (avoids import cycles)
- Each `_test.go` covers: CRUD success, error responses, empty ID validation, boolean serialization, auth headers

## Known patterns to follow

- Service methods: `List`, `Get`, `Create`, `Update`, `Delete`
- Nested resources take parent IDs as parameters
- `ListOptions.Filter` for query string filtering
- `*ValidationError` for client-side validation failures
- `*APIError` for server-side errors

## Commit format

Conventional commits: `type(scope): description`

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
