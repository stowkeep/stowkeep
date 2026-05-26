# Development Guide

How to set up a local Stowkeep development environment.

> **Status:** Stage 0 scaffold — commands below describe the target workflow; some targets are added as implementation lands.

---

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.23+ | Backend |
| Node.js | 22 LTS | Frontend |
| Docker | 24+ | Engine API + optional Swarm |
| Make | any | Task runner |
| golangci-lint | latest | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

Optional:

- PostgreSQL 16 — when testing production DB path (`make dev-postgres`)
- Redis 7 — Stage 5+ GitOps workers

---

## Clone and bootstrap

```bash
git clone https://github.com/stowkeep/stowkeep.git
cd stowkeep

# Install frontend dependencies (once web/ exists)
cd web && npm ci && cd ..

# Copy example env
cp .env.example .env
```

---

## Database modes for development

### SQLite (default — fastest)

No external services. Database file at `./.data/dev.db`.

```bash
make dev
# or explicitly:
STOWKEEP_DATABASE_DRIVER=sqlite \
STOWKEEP_DATABASE_PATH=./.data/dev.db \
make dev
```

### PostgreSQL

```bash
docker compose -f docker-compose.dev.yml up -d postgres
make dev-postgres
```

See [database.md](./database.md) for production vs embedded guidance.

---

## Common commands

| Command | Description |
|---------|-------------|
| `make dev` | API + web hot reload (SQLite) |
| `make dev-postgres` | API + web with PostgreSQL |
| `make test` | All unit tests (Go + web) |
| `make test-integration` | Integration tests (Docker + DB) |
| `make lint` | golangci-lint + eslint |
| `make migrate-up` | Apply DB migrations |
| `make migrate-down` | Roll back one migration |
| `make build` | Production binary + web bundle |
| `make docker-build` | Build container image locally |

---

## Project layout

```
stowkeep/
├── api/              # HTTP server entrypoint
├── pkg/              # Reusable Go packages
├── web/              # React frontend
├── migrations/       # goose SQL migrations (shared + dialect-specific)
├── docs/             # User and contributor documentation
├── planning/         # PRD and engineering plans
├── deploy/           # Reference Compose / Swarm stacks
├── openapi/          # API specification
└── .github/          # CI, issue templates
```

---

## Running tests

```bash
# Go unit tests (fast)
go test ./... -short -race

# Go integration tests (requires Docker)
go test ./... -tags=integration -race

# SQLite migration smoke
make test-migrations-sqlite

# PostgreSQL migration smoke (testcontainers)
make test-migrations-postgres

# Frontend
cd web && npm run test

# E2E (nightly / manual)
cd web && npm run test:e2e
```

Both database backends are tested in CI. If your change touches migrations or SQL queries, run both locally.

---

## Docker socket access

Swarm features require access to Docker:

```bash
# Linux: add your user to docker group, or run API with socket mount
export STOWKEEP_DOCKER_HOST=unix:///var/run/docker.sock

# macOS Docker Desktop: same default socket path
```

Initialize Swarm for stack deploy tests:

```bash
docker swarm init
# after tests:
docker swarm leave --force
```

---

## API development

- OpenAPI spec: `openapi/openapi.yaml`
- Regenerate TS client: `make generate-api`
- Health check: `curl localhost:8080/healthz`

---

## Frontend development

```bash
cd web
npm run dev    # Vite dev server, proxies /api to backend
npm run lint
npm run typecheck
```

---

## Debugging

- **Backend logs:** `STOWKEEP_LOG_LEVEL=debug STOWKEEP_LOG_FORMAT=text make dev`
- **Production-like logs locally:** `STOWKEEP_LOG_FORMAT=json make dev`
- **Database:** SQLite → `sqlite3 .data/dev.db`; Postgres → `psql $DATABASE_URL`
- **Feature flags:** `STOWKEEP_FEATURES=gitops,previews`

---

## Before opening a PR

1. `make lint && make test`
2. Update [CHANGELOG.md](../CHANGELOG.md) if user-visible
3. Add godoc/TSDoc for new exported APIs
4. Read [CONTRIBUTING.md](../CONTRIBUTING.md)

---

## Getting help

- [GitHub Discussions](https://github.com/stowkeep/stowkeep/discussions) — questions
- [GitHub Issues](https://github.com/stowkeep/stowkeep/issues) — bugs and features

Replace `stowkeep` with the actual GitHub organization when the repo is published.
