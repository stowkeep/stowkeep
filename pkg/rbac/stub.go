// Package rbac provides authorization hooks; Stage 2 ships a stub for Stage 3 Casbin.
package rbac

import (
	"context"

	"github.com/stowkeep/stowkeep/pkg/auth"
)

// Authorizer decides whether an action is permitted.
type Authorizer interface {
	Allow(ctx context.Context, user *auth.User, action, resourceType, resourceID string) bool
}

// AdminOnly allows only users with role "admin".
type AdminOnly struct{}

// Allow returns true when the user has the admin role.
func (AdminOnly) Allow(_ context.Context, user *auth.User, _, _, _ string) bool {
	return user != nil && user.Role == "admin"
}
