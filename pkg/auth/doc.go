// Package auth implements local email/password authentication with DB-backed sessions.
// Stage 1 ships a single admin role; full RBAC is deferred to Stage 3.
// See docs/security/threat-model.md §4.1 for threat assumptions.
package auth
