# Code Standards

Coding conventions and documentation expectations for Stowkeep contributors.

---

## General principles

1. **Readability over cleverness** — Future maintainers include people who have never met you.
2. **Minimal scope** — PRs do one thing well.
3. **Match existing patterns** — Consistency beats personal preference.
4. **Explain the non-obvious** — Security, Swarm quirks, and concurrency deserve comments.
5. **Don't comment the obvious** — `// increment counter` above `counter++` adds noise.

---

## Go

### Formatting and lint

- `gofmt` / `goimports` — enforced by CI
- `golangci-lint` with project config (`.golangci.yml`) — no disabling rules without issue discussion
- Errors wrapped with context: `fmt.Errorf("deploy stack %q: %w", name, err)`

### Naming

| Item | Convention |
|------|------------|
| Packages | Short, lowercase, no underscores (`secrets`, not `secret_mgmt`) |
| Files | `snake_case.go` for multi-word |
| Interfaces | `-er` suffix when natural (`Store`, `Reconciler`) |
| Test files | `*_test.go` |

### Godoc

Required on all exported identifiers:

```go
// Store persists encrypted secrets and their version history.
type Store interface {
    // Get returns the secret metadata and decrypted value when the caller
    // has read_value permission. Returns ErrNotFound when the path does not exist.
    Get(ctx context.Context, env, path, name string) (*Secret, error)
}
```

Package comment in `doc.go`:

```go
// Package gitops implements pull-based reconciliation from Git repositories
// to Docker Swarm stacks. See planning/PRD.md for product requirements.
package gitops
```

### Tests

- Table-driven tests for permission matrices and parsers
- Use `t.Parallel()` where safe
- Integration tests: build tag `//go:build integration`
- Test names: `TestStore_Get_DeniesWithoutReadValue`

### Security-sensitive code

- Never log secret values, tokens, or session cookies
- Use `crypto/rand` for keys
- Document threat assumptions in package doc for `secrets`, `auth`, `rbac`

### Logging (Go)

Use the shared logger from `pkg/observability/log` (stdlib `slog` wrapper). Full guide: [logging.md](./logging.md).

```go
// Prefer context-aware logging so request_id propagates.
log.InfoContext(ctx, "stack deploy completed",
    slog.String("component", "swarm"),
    slog.String("stack", name),
    slog.Int("duration_ms", elapsed),
)

log.ErrorContext(ctx, "gitops sync failed",
    slog.String("component", "gitops"),
    slog.String("error", err.Error()),
)
```

Rules:

- Always set `component` (`http`, `gitops`, `swarm`, `secrets`, `auth`)
- Use `Info` for normal ops, `Warn` for recoverable issues, `Error` for failures
- Never use `fmt.Sprintf` to embed dynamic secrets in `msg`
- Log errors at the boundary where they are handled, not every wrap layer
- Do not log HTTP bodies, `Authorization` headers, or env vars from compose files

---

## TypeScript / React

### Formatting and lint

- Prettier + ESLint (flat config)
- Strict TypeScript (`strict: true`)

### Naming

| Item | Convention |
|------|------------|
| Components | `PascalCase.tsx` |
| Hooks | `useSomething.ts` |
| Utilities | `camelCase.ts` |
| Constants | `SCREAMING_SNAKE` or `camelCase` for module-level |

### TSDoc

```typescript
/**
 * Fetches stack list for the active Swarm endpoint.
 * Refetches on window focus when {@link refetchOnFocus} is enabled.
 */
export function useStacks(options?: UseStacksOptions) { ... }
```

### Components

- Prefer function components
- Colocate tests: `Component.test.tsx` next to `Component.tsx`
- Extract hooks when logic exceeds ~30 lines in a component

---

## SQL migrations

```sql
-- 003_audit_events: append-only audit log for mutating API actions.
-- +goose Up
CREATE TABLE audit_events ( ... );
```

- Reversible when feasible (`+goose Down`)
- Never store plaintext secrets in schema defaults or seed data

---

## API design

- REST paths: `/api/v1/...`, kebab-case resources
- JSON field names: `camelCase`
- Errors: consistent envelope `{ "error": { "code": "...", "message": "..." } }`
- Document in OpenAPI before implementing (or in same PR)

---

## Comments quality checklist

Before opening a PR, ask:

- [ ] Would a new contributor understand **why** this code exists?
- [ ] Are exported APIs documented without restating the name?
- [ ] Are security constraints visible at the package boundary?
- [ ] Did I remove debug comments and dead code?

See also [open-source-standards.md](../planning/open-source-standards.md).
