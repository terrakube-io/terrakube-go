---
title: Introduction
---

# Terrakube Go Client

`github.com/terrakube-io/terrakube-go` is the Go client library for the
[Terrakube](https://terrakube.io) API — the open-source platform for managing
Terraform and OpenTofu at scale (workspaces, VCS connections, modules, teams,
variables and more).

This site is a browsable, source-mapped reference generated from the package's
own godoc. Every symbol links back to the exact line of Go source on `main`.

## Layout at a glance

The library is a **single flat `terrakube` package** at the module root. There
are no sub-packages except `testutil/` (test helpers). It is organised in four
layers:

1. **Client** (`client.go`) — HTTP transport, authentication, and JSON:API
   serialization. Constructed with the functional-options pattern via
   `NewClient`.
2. **Generic CRUD base** (`crud.go`) — `crudService[T]` provides `list`, `get`,
   `create`, `update`, `del` once, so the ~30 resource services stay thin.
3. **Resource services** (`*.go`, one file per resource) — each pairs an entity
   struct with a service struct that embeds `crudService[T]` and exposes
   `List` / `Get` / `Create` / `Update` / `Delete`.
4. **Test infrastructure** (`testutil/`) — an `httptest` server wrapper and
   fixture helpers.

## Conventions worth knowing before you read the reference

These are the non-obvious rules the code follows — knowing them makes the API
reference read faster:

- **JSON:API everywhere except TeamToken.** Resources serialize through
  `google/jsonapi`. The one exception is `TeamToken`, which uses the
  `/access-token/v1/teams` endpoint with standard JSON via the `requestRaw` /
  `doRaw` path and does **not** use `crudService[T]`.
- **Every service method takes `context.Context` first.** No exceptions.
- **Optional string fields are `*string`.** A `nil` pointer means "unset"
  rather than "empty string".
- **Booleans always serialize, even `false`.** The jsonapi struct tags
  deliberately omit `omitempty` on booleans, so a `false` is sent on the wire
  instead of being dropped.
- **Non-2xx responses return `*APIError`, not a generic error.** Type-assert to
  `*APIError` to read the status code and API-provided message.
- **IDs are validated before the URL path is built.** An empty ID is rejected
  client-side rather than producing a malformed request.

## Regenerating this reference

```bash
# Needs the Go toolchain — snapshots the module's godoc:
npx sourcey godoc -m . -o godoc.json
# No Go required — builds the static site from the committed snapshot:
npx sourcey build
```

The snapshot (`godoc.json`) is committed, so the site rebuilds in CI without a
Go step.
