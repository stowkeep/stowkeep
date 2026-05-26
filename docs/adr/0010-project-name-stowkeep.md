# 0010. Project name: Stowkeep

- **Status:** Accepted
- **Date:** 2026-05-25
- **Decision ID:** D-031 (closes OPEN-001; carries forward OPEN-006 for namespace reservation)

## Context

The project went through its planning phase under the working title "Swarm Operator". That name has two structural problems that compound over time:

1. **Namespace collision with Kubernetes.** The "operator" pattern in cloud-native software is strongly associated with Kubernetes (Operator Framework, OperatorHub, controller-runtime). Calling a Docker Swarm control plane a "Swarm Operator" sets the wrong mental model on first contact and harms discoverability on every search surface.
2. **Generic + already-taken.** "Swarm Operator" appears in multiple unrelated GitHub repos and is not registrable as a clean handle on the surfaces we care about (GitHub org, Docker Hub org, npm package, `.dev`/`.app` domains).

A rename pre-`v0.1.0`, before any tag, container push, or HN/Reddit post, is cheap. After that, it's painful. So we picked a name now.

We deliberately set a high bar for the selected name:

1. **All four key technical namespaces clean:** GitHub org, Docker Hub org, npm package, top-level domain (`.dev` strongly preferred — HTTPS-enforced and OSS-infra-friendly).
2. **No active software project competing on the bare name** that would create user confusion.
3. **No trademark conflict in software classes** (Nice Classes 9 and 42). Conflicts in unrelated classes (e.g. apparel, sports equipment) are acceptable.
4. **No phonetic collision** with widely-used dev tools (this is what eliminated "Sailwind" — the Tailwind echo is permanent).
5. **Etymologically fits the product** — a control plane that holds compose stacks, secrets (encrypted), GitOps state, audit logs, and backups.

## Decision

The project is named **Stowkeep**.

### Canonical naming surfaces

| Attribute | Value |
|-----------|-------|
| Display name | Stowkeep |
| Repository | `stowkeep` |
| Go module | `github.com/stowkeep/stowkeep` |
| Container image | `ghcr.io/stowkeep/stowkeep` |
| Config manifest | `.stowkeep.yml` |
| CLI binary | `stowkeep` |
| Env prefix | `STOWKEEP_*` |
| Default SQLite filename | `stowkeep.db` |
| Default Postgres database | `stowkeep` |
| Default Docker volume | `stowkeep-data` |
| Security contact | `security@stowkeep.dev` (pending OPEN-006) |
| Conduct contact | `conduct@stowkeep.dev` (pending OPEN-006) |

### Why "Stowkeep" passed the bar

| Criterion | Evidence (verified 2026-05-25) |
|-----------|----|
| GitHub org | `github.com/stowkeep` returned **404** — available |
| Docker Hub | `hub.docker.com/u/stowkeep` returned **404** — available |
| npm package | `npmjs.com/package/stowkeep` returned **404** — available |
| Software-namespace search | Zero direct hits. Nearest neighbors (`gnu-stow`, `pystow`) are unrelated dotfile-symlink managers in a clearly different problem domain. |
| Brand / trademark | No conflict in software classes. Nearby trademarks (`Stow Stick™` — golf, Class 28; `Sailwind®` — apparel, Class 25) are in unrelated Nice classes. |
| Phonetic clarity | Two syllables. Spells how it sounds. No tech-namespace homophone (contrast: "Sailwind" vs Tailwind). |
| Fit for the product | "Stow" = "to put something in a place where it can be kept safely" (literal dictionary definition). "Keep" = both "maintain" and (medieval) "the fortified inner tower of a castle." "Stowkeep" = the keeper of what is stowed away. Reads honestly for a control plane that holds stacks, secrets, audit logs, and backups. |

### Names eliminated during the search (briefly recorded for audit)

- **Skipper** — saturated; multiple OSS projects and a Kubernetes-adjacent routing tool.
- **Drydock** — direct competition (existing container-tooling projects + a security scanner with the same name).
- **Hopper** — Kafka Connect tooling and data-pipeline projects already claim it.
- **Pilothouse** — clean enough, but a mouthful spoken aloud (eliminated for ergonomics).
- **Sailwind** — fatal phonetic collision with Tailwind CSS; `pglevy/sailwind` is literally a React component library built on Tailwind. Also: `sailwind.ai` is a funded AI-orchestration startup, and the Steam game floods unqualified search results.
- **Hivekeep** — direct phonetic mirror of the active `keephive` project on GitHub/PyPI. Brutal spoken-word confusion.
- **Roostward** — viable, but no clear advantage over Stowkeep and a slightly worse fit ("ward" implies guard, not custody).

## Consequences

### Positive

- Single, defensible name across every technical surface that matters in 2026.
- Etymologically aligned with the product's job — "the thing that keeps your stowed configuration, secrets, and state safe" — so the marketing pitch writes itself.
- No spoken-word friction. People will not have to spell it; they will not confuse it with another tool.
- The naming decision is now auditable: anyone asking "why this name?" gets a one-screen answer.

### Negative

- A rename touches every doc, env var (`SWARM_OPERATOR_*` → `STOWKEEP_*`), and example. This ADR is being landed alongside that sweep, and the cost compounds with every week it is deferred — so we are paying it now.
- "Stowkeep" is not a real English word. New users need a one-sentence pitch on first encounter ("the keeper of what is stowed away — Docker Swarm control plane"). The README and the GitHub repo description must carry that sentence prominently.
- Coined words have a small SEO cold-start cost vs. real words. This is acceptable given the upside of zero competing meaning.

### Reservation outcome (updated 2026-05-25)

| Surface | Status |
|---|---|
| `stowkeep.dev` domain | ✅ Registered |
| GitHub organization | ✅ Created |
| Docker Hub organization | ✅ Created |
| npm package `stowkeep` | ⏸️ **Deferred per D-033.** Go-first project, frontend embedded via `embed.FS` (ADR 0008), no JS package planned. Revisit only if a `@stowkeep/*` JS SDK is ever scoped or public visibility rises and squat risk grows. |
| `contact@stowkeep.dev` inbox | ✅ Live; serves as single security + conduct contact pre-`v0.1.0` (D-032) |
| Dedicated `security@` / `conduct@` inboxes | ⏸️ Deferred post-`v0.1.0` |
| Defensive `stowkeep.app` / `.com` | ⏸️ Optional |

With the first three surfaces held and a live contact inbox, the name is considered **fully locked** as of 2026-05-25. The deferred items above are conveniences or future polish, not blockers.

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) D-031 (this decision), OPEN-006 (reservation follow-up).
- Selection criteria, evidence, and rejected names captured in the chat transcript dated 2026-05-25.
