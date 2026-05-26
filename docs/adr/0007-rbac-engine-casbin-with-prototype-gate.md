# 0007. RBAC engine = Casbin, with mandatory Stage 2 UI-prototype gate

- **Status:** Accepted
- **Date:** 2026-05-25
- **Tracker entry:** [D-010](../../planning/decisions-todo.md)

## Context

PRD §6.5 calls for:

- Built-in roles (`owner`, `admin`, `developer`, `viewer`).
- Custom roles with a **permission-builder UI** (Stage 3).
- Conditions per permission: environment, stack name, secret path prefix, tags.
- Separate `describe` / `read_value` / `write` on secrets.

That is full ABAC over a hierarchical resource model, with a non-technical admin expected to author policies via UI. The engine choice and the UX are inseparable: a great engine with an unusable UI is a worse outcome than a slightly less powerful engine with a clean UI.

Candidates considered:

| Option | Strengths | Weaknesses |
|--------|-----------|------------|
| **Casbin** | Mature, widely deployed, supports ABAC via matchers, CSV-style policies | UX over Casbin matchers is hard; debugging "why was this denied?" requires reading matcher source |
| OpenFGA / Google Zanzibar-style | Excellent for hierarchical resources; explicit relationship model | Newer; relationship model is a different mental shape than role+condition; another dependency |
| Cedar (AWS) | Typed policy language; strong tooling | Newer Go binding; smaller ecosystem |
| Custom evaluator on a typed schema | Most ergonomic; we own the engine | We own the engine — long-term maintenance burden |

Casbin is the safe technical choice. The risk is not the engine — it is whether we can build a permission-builder UI on top of it that an operator can actually use. Discovering that gap in Stage 3 (after RBAC is built) is expensive. Discovering it in Stage 2 (before) is cheap.

## Decision

Use Casbin as the RBAC engine. The model is RBAC-with-domains-and-conditions; policies live in the DB (one row per assertion) and load into a Casbin enforcer at startup with hot reload on policy change.

**Mandatory gate:** Stage 2 — the auth stage — must ship a paper or static-HTML mockup of the permission-builder UI as part of the PR that introduces basic auth. The mockup walks an admin through:

1. Create a custom role `staging-deployer`.
2. Grant `swarm.stacks.deploy` scoped to `environment = "staging"`.
3. Grant `secrets.read_value` scoped to `path startsWith "/api/"`.
4. Assign the role to a user and to a group.
5. Preview "what can this user do" against a sample fixture.

If the mockup is honest and the only realistic admin interface ends up being raw Casbin CSV, that is the **trigger to revisit** the engine choice in a follow-up ADR (likely revisiting OpenFGA or a custom evaluator).

Additionally, by Stage 3 GA we ship a **policy explainer** endpoint and CLI command: given a subject + action + resource, return either `ALLOW` with the matching policy row(s) and condition evaluation trace, or `DENY` with the matched (or missing) policy row(s). This is non-optional; debugging denials by reading matchers is not acceptable for the target audience.

## Consequences

**Easier**

- Mature engine, well-known patterns, abundant examples.
- ABAC conditions modelable today with Casbin matchers.
- The UI prototype gate catches the "Casbin can model it, but no admin can configure it" failure mode before it becomes expensive.

**Harder**

- We have to build the policy-explainer tooling that other RBAC engines might give us out of the box.
- If the prototype fails the gate, Stage 3 slips by 1–2 weeks while we re-evaluate. That cost is intentional — the alternative is shipping unusable RBAC.
- Casbin matchers are powerful and easy to misuse; the `CODEOWNERS` rule on `pkg/rbac` enforces review by the maintainer on every change.

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) — `D-010`, `DEF-002`
- [planning/PRD.md §6.5, §8 Stage 2, §8 Stage 3](../../planning/PRD.md)
- [docs/security/threat-model.md §4.2](../security/threat-model.md) — authorization threats
