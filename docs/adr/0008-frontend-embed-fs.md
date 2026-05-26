# 0008. Frontend embedded via `embed.FS`

- **Status:** Accepted
- **Date:** 2026-05-25
- **Tracker entry:** [D-026](../../planning/decisions-todo.md)

## Context

Stowkeep pitches a single-binary, single-container, "one `docker run`" deploy story for homelab users. The frontend is a React+Vite SPA built into a `web/dist/` directory. The question is how the SPA gets served in production.

Two real options:

- **`embed.FS`** — bake the built `web/dist/*` files into the Go binary at compile time and serve them from the same Go HTTP server that hosts the API.
- **nginx sidecar** — ship a separate nginx container that serves the static files; users run both containers (or one container with two processes — anti-pattern) and put a reverse proxy in front.

For our shape:

| Concern | `embed.FS` | nginx sidecar |
|---------|-----------|---------------|
| Containers per install | 1 | 2 (or 1 + anti-pattern) |
| Quick-start `docker run` | One command | Compose file or two commands |
| Release atomicity | FE/API versions cannot drift | Possible to ship mismatched versions |
| Asset serving perf | Fine for SPA-sized payloads at our scale | Faster at the static layer (invisible at our load) |
| FE-only deploys | Not supported (Go rebuild required) | Supported (not useful for self-hosted) |
| Operator mental model | "It's one binary" | "It's a Go service and an nginx" |
| Comparable products | Portainer, Grafana, Vault, ArgoCD, Gitea, Caddy | Few precedents in this category |

The only scenarios where the sidecar wins materially are CDN-scale FE delivery (not relevant), independent FE-version rollouts (not relevant for self-hosted), or "we already run nginx for something" — and in that case users already have a reverse proxy in front and the upstream container shape is irrelevant.

## Decision

Embed via `//go:embed web/dist/*` in a small `pkg/web` package. The Go HTTP server mounts a `http.FileServer(http.FS(distFS))` on a non-API prefix and serves `index.html` for SPA fallback routes. Cache headers are set by a small middleware: `Cache-Control: public, max-age=31536000, immutable` for hashed asset paths (Vite produces `*-[hash].js|css`), `Cache-Control: no-cache` for `index.html`.

The Dockerfile is a three-stage build:

```
Stage 1: node:22-alpine          # npm ci && npm run build → web/dist/
Stage 2: golang:1.23-alpine      # COPY --from=Stage1 web/dist /src/web/dist; go build
Stage 3: gcr.io/distroless/static # COPY --from=Stage2 /out/stowkeep
```

Local dev is unchanged: `make dev` runs Vite with HMR on one port and the Go server on another; Vite proxies `/api` to the Go process. The embed is only relevant at production build time.

## Consequences

**Easier**

- One binary, one image, one port. Atomic FE/API releases — version skew is structurally impossible.
- Matches the "single container quick-start" pitch in the README.
- Standard pattern in our reference set (Portainer, Grafana, Vault, ArgoCD, Gitea, Caddy).
- No nginx config to maintain; no second image to scan, sign, or update.

**Harder**

- Every UI-only change requires a Go rebuild in CI. Transparent — both Node and Go stages are already required for any release.
- Binary grows by the SPA bundle size (~3–10 MB). Negligible at our scale.
- Cannot ship FE-only patches without a backend release. Not a real constraint for a self-hosted product where users pull a new image regardless.
- Cache-header and SPA-fallback middleware is now our responsibility (well-trodden, ~30 lines).

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) — `D-026`
- [planning/tech-stack.md — Frontend](../../planning/tech-stack.md)
- [planning/PRD.md §8 Stage 0, §12](../../planning/PRD.md)
