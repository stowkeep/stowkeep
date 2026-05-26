# 0002. Structured logging with log/slog

- **Status:** Accepted
- **Date:** 2026-05-25

## Context

Stowkeep needs production logging that works in Docker/Swarm without extra infrastructure, supports log aggregation platforms, and never leaks secrets. Contributors should have one clear, low-friction logging API.

Options considered: Uber Zap (fast, popular), Zerolog (zero-allocation JSON), Go stdlib `log/slog` (structured, JSON handler, Go 1.21+).

## Decision

Use **Go stdlib `log/slog`** as the sole application logger:

- **Production:** JSON lines to stdout (`STOWKEEP_LOG_FORMAT=json`)
- **Development:** Human-readable text (`format=text`, `level=debug`)
- **Correlation:** `request_id` on every HTTP request via context; OpenTelemetry `trace_id` when tracing is enabled (Stage 7)
- **No file logging** inside the container
- **Audit-sensitive events** go to the `audit_events` database table, not stdout

HTTP middleware logs method, path, status, duration — never bodies or auth headers.

## Consequences

**Easier**

- Zero logging dependencies; smaller supply chain
- Standard pattern for container platforms (stdout → Docker → Loki/ELK/etc.)
- New contributors learn one API documented in Go official docs

**Harder**

- slog is slightly less ergonomic than Zap for very high-throughput scenarios (not a concern at expected scale)
- Log shipping requires external agent/driver — we document patterns but don't bundle Promtail/Alloy

## References

- [docs/logging.md](../logging.md)
- [12-factor logs](https://12factor.net/logs)
