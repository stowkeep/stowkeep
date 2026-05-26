# Logging Guide

How Stowkeep handles logs in development and production.

> **TL;DR:** The app writes **structured JSON to stdout**. Docker/Swarm captures it. You point your log stack (Loki, Grafana Cloud, Datadog, etc.) at container output. Default level in production is `info`. Secret values are never logged.

---

## Why this approach

Stowkeep runs as a **container on Swarm** (or locally). The standard pattern for containers is:

1. App prints logs to **stdout** and **stderr**
2. Docker logging driver collects them
3. Your platform ships, stores, and searches them (optional)

We do **not** write log files inside the container. Ephemeral filesystems and log rotation inside the app are anti-patterns for this deployment model ([12-factor logs](https://12factor.net/logs)).

---

## Stack choice

| Component | Choice | Notes |
|-----------|--------|-------|
| **Logger** | Go [`log/slog`](https://pkg.go.dev/log/slog) (stdlib) | Structured, JSON in prod, no extra dependency |
| **Format (prod)** | JSON lines | One JSON object per line — easy for Loki/ELK/Datadog |
| **Format (dev)** | Human-readable text | Easier to read in terminal during `make dev` |
| **Output** | stdout only | Swarm/Docker handles collection |
| **Correlation** | `request_id` + optional OpenTelemetry `trace_id` | Tie HTTP requests and background jobs together |
| **Metrics** | Prometheus (separate) | Counts/latency — not duplicated in logs |
| **Audit trail** | PostgreSQL/SQLite `audit_events` table | Security-sensitive actions — **not** a substitute for audit DB |

We use **slog** instead of Zap or logrus to keep dependencies minimal. It is production-ready in Go 1.21+ and supports JSON natively.

---

## Log levels

| Level | When to use | Production default |
|-------|-------------|-------------------|
| **debug** | Verbose internals (Docker API payloads summary, SQL query names) | Off |
| **info** | Normal operations (startup, sync completed, stack deployed, login success) | On |
| **warn** | Recoverable problems (retry, deprecated config, slow request) | On |
| **error** | Failures needing attention (sync failed, DB error, Docker unreachable) | On |

Configure via environment:

```bash
STOWKEEP_LOG_LEVEL=info    # production default
STOWKEEP_LOG_FORMAT=json   # json | text
```

Development defaults: `level=debug`, `format=text`.

---

## Standard log fields

Every log line includes a consistent schema where applicable:

| Field | Example | Purpose |
|-------|---------|---------|
| `time` | `2026-05-25T12:00:00.000Z` | RFC3339 timestamp (JSON handler) |
| `level` | `INFO` | Severity |
| `msg` | `stack deploy completed` | Human-readable summary |
| `service` | `stowkeep` | Constant service name |
| `version` | `v0.1.0` | Build version / git SHA |
| `component` | `gitops`, `http`, `secrets` | Subsystem |
| `request_id` | `7f3c2a1b-...` | HTTP request correlation |
| `trace_id` | `abc123...` | OpenTelemetry trace (when enabled) |
| `user_id` | `usr_01...` | Actor (ID only — avoid email in logs) |
| `duration_ms` | `42` | For timed operations |
| `stack` | `web-api` | Domain context |
| `error` | `connection refused` | Error message string on failures |

**Avoid:** secret values, passwords, tokens, session cookies, Authorization headers, full compose env blocks, raw webhook secrets.

---

## What gets logged (by area)

### HTTP API

Middleware logs **once per request** at `info` (or `warn` if slow/error):

```json
{
  "time": "2026-05-25T12:00:01Z",
  "level": "INFO",
  "msg": "request completed",
  "component": "http",
  "method": "POST",
  "path": "/api/v1/stacks",
  "status": 201,
  "duration_ms": 87,
  "request_id": "7f3c2a1b-4e5d-6f7a-8b9c-0d1e2f3a4b5c",
  "user_id": "usr_01H..."
}
```

- **Never** log request/response bodies by default.
- Paths may include IDs; redact query strings containing `token`, `key`, `secret`.

### GitOps reconciler

| Event | Level | Example msg |
|-------|-------|-------------|
| Sync started | info | `gitops sync started` + `app`, `commit` |
| Sync succeeded | info | `gitops sync completed` + `duration_ms`, `stack` |
| Sync failed | error | `gitops sync failed` + `error`, no secret env dump |
| Poll tick | debug | `gitops poll check` |

### Docker / Swarm

| Event | Level |
|-------|-------|
| Connection established | info |
| Deploy / remove stack | info |
| Docker API error | error |
| Full API request/response | debug only |

### Auth

| Event | Level | Notes |
|-------|-------|-------|
| Login success | info | `user_id` only |
| Login failure | warn | Log reason category (`invalid_password`), not the password |
| Logout | debug | |
| Permission denied | warn | `user_id`, `action`, `resource` — for security monitoring |

### Secrets

| Event | Level | Notes |
|-------|-------|-------|
| Secret created/updated/deleted | info | path + name metadata only |
| Secret value read | **audit DB only** | Never info-level log of values |
| Encryption failure | error | No ciphertext in log |

---

## Audit log vs application logs

| | Application logs (stdout) | Audit log (database) |
|---|---------------------------|----------------------|
| **Purpose** | Debugging, ops, aggregation | Compliance, security review |
| **Retention** | Your log platform (30–90 days typical) | Long-lived, queryable in UI |
| **Secret values** | Never | Never |
| **Who read a secret** | Not logged here | `audit_events` row |

Both are required for a complete picture. Do not rely on stdout alone for compliance.

---

## Configuration reference

| Variable | Default | Description |
|----------|---------|-------------|
| `STOWKEEP_LOG_LEVEL` | `info` (prod), `debug` (dev) | `debug`, `info`, `warn`, `error` |
| `STOWKEEP_LOG_FORMAT` | `json` (prod), `text` (dev) | Output format |
| `STOWKEEP_LOG_ADD_SOURCE` | `false` | Include `source` file:line (dev debugging) |

Future (Stage 7):

| Variable | Description |
|----------|-------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector for traces (+ log correlation) |

---

## Production deployment

### Docker / Swarm

Logs go to stdout. Configure Swarm logging driver if you need remote shipping:

```yaml
services:
  stowkeep:
    image: ghcr.io/stowkeep/stowkeep:latest
    environment:
      STOWKEEP_LOG_LEVEL: info
      STOWKEEP_LOG_FORMAT: json
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

`json-file` with rotation prevents a single node disk from filling if you don't ship logs elsewhere.

### Viewing logs locally

```bash
# Follow container logs
docker service logs -f stowkeep_stowkeep

# Pretty-print JSON lines (requires jq)
docker logs stowkeep 2>&1 | jq -r '.'
```

### Shipping to a log platform (optional)

Stowkeep does **not** embed a log forwarder. Common patterns:

| Platform | Typical setup |
|----------|---------------|
| **Grafana Loki** | Promtail / Grafana Alloy on nodes, scrape Docker logs |
| **Elasticsearch (ELK)** | Filebeat → Elasticsearch |
| **Datadog** | Datadog Agent with Docker integration |
| **CloudWatch** | `awslogs` driver on ECS; agent on EC2 |

Search by `service=stowkeep`, filter `level=error`, group by `component=gitops`.

---

## Development

```bash
# Human-readable, verbose
STOWKEEP_LOG_LEVEL=debug STOWKEEP_LOG_FORMAT=text make dev

# Simulate production logging locally
STOWKEEP_LOG_LEVEL=info STOWKEEP_LOG_FORMAT=json make dev
```

---

## Implementation conventions (Go)

Package: `pkg/observability/log` (or `internal/log`)

```go
// From an HTTP handler — logger on context
log.InfoContext(ctx, "stack deploy completed",
    slog.String("component", "swarm"),
    slog.String("stack", name),
    slog.Int("duration_ms", elapsed),
)

// Errors — always attach context, never secret values
log.ErrorContext(ctx, "gitops sync failed",
    slog.String("component", "gitops"),
    slog.String("app", appName),
    slog.String("error", err.Error()),
)
```

Rules:

1. Use `log.InfoContext(ctx, ...)` so `request_id` propagates from context.
2. Prefer structured attributes over `fmt.Sprintf` in messages.
3. Wrap errors with `%w`; log the error string at the handling boundary.
4. Use `component` on every log from a subsystem.

See [code-standards.md](./code-standards.md#logging-go).

---

## Testing

CI includes tests that:

- Secret values never appear in log output during CRUD operations
- HTTP middleware does not log Authorization headers
- JSON format produces valid JSON lines

See [planning/testing-strategy.md](../planning/testing-strategy.md).

---

## Troubleshooting cheat sheet

| Symptom | Check |
|---------|-------|
| No logs | Process running? `docker logs <container>` |
| Too noisy | Set `STOWKEEP_LOG_LEVEL=info` or `warn` |
| Can't correlate requests | Search logs for same `request_id` |
| Sync failures | Filter `component=gitops level=error` |
| Auth issues | Filter `component=auth` |
| Missing history | App logs rotate quickly — use log platform retention or audit DB |

---

## Related docs

- [development-guide.md](./development-guide.md) — local setup
- [planning/tech-stack.md](../planning/tech-stack.md) — observability stack
- [docs/adr/0002-structured-logging-slog.md](./adr/0002-structured-logging-slog.md) — ADR
