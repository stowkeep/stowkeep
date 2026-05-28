package compose_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/compose"
)

func TestValidateValidStack(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "compose", "valid-stack.yml"))
	if err != nil {
		t.Fatal(err)
	}
	res := compose.Validate(context.Background(), content, "web")
	if !res.Valid {
		t.Fatalf("expected valid, got errors: %+v", res.Errors)
	}
	if res.Hash == "" {
		t.Fatal("expected content hash")
	}
}

func TestValidateInvalidSyntax(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "compose", "invalid-syntax.yml"))
	if err != nil {
		t.Fatal(err)
	}
	res := compose.Validate(context.Background(), content, "web")
	if res.Valid {
		t.Fatal("expected invalid compose")
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestValidateInvalidSchema(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "compose", "invalid-schema.yml"))
	if err != nil {
		t.Fatal(err)
	}
	res := compose.Validate(context.Background(), content, "web")
	if res.Valid {
		t.Fatal("expected invalid compose")
	}
}

func TestValidateStackName(t *testing.T) {
	tests := []struct {
		name    string
		stack   string
		wantErr bool
	}{
		{"valid", "web", false},
		{"empty", "", true},
		{"uppercase", "Web", true},
		{"too long", string(make([]byte, 64)), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := compose.Validate(context.Background(), []byte("services:\n  a:\n    image: nginx:alpine\n"), tc.stack)
			if tc.wantErr && res.Valid {
				t.Fatal("expected error")
			}
			if !tc.wantErr && !res.Valid {
				t.Fatalf("unexpected errors: %+v", res.Errors)
			}
		})
	}
}

func TestValidateOversize(t *testing.T) {
	content := make([]byte, compose.MaxFileSize+1)
	res := compose.Validate(context.Background(), content, "web")
	if res.Valid {
		t.Fatal("expected oversize rejection")
	}
}

func TestContentHashStable(t *testing.T) {
	a := compose.ContentHash([]byte("hello"))
	b := compose.ContentHash([]byte("hello"))
	if a != b {
		t.Fatalf("hash not stable: %s vs %s", a, b)
	}
}
