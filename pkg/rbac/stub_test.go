package rbac_test

import (
	"context"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/auth"
	"github.com/stowkeep/stowkeep/pkg/rbac"
)

func TestAdminOnlyAllow(t *testing.T) {
	authz := rbac.AdminOnly{}
	ctx := context.Background()
	if !authz.Allow(ctx, &auth.User{Role: "admin"}, "swarm.stacks.deploy", "stack", "web") {
		t.Fatal("admin should be allowed")
	}
	if authz.Allow(ctx, &auth.User{Role: "viewer"}, "swarm.stacks.deploy", "stack", "web") {
		t.Fatal("viewer should be denied")
	}
}
