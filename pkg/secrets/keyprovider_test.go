package secrets_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/secrets"
)

func testMasterKey() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestEnvKeyNotConfigured(t *testing.T) {
	_, err := secrets.NewEnvKey("")
	if err != secrets.ErrMasterKeyNotConfigured {
		t.Fatalf("NewEnvKey(\"\") = %v, want ErrMasterKeyNotConfigured", err)
	}
}

func TestEnvKeyWrapUnwrapRoundtrip(t *testing.T) {
	provider, err := secrets.NewEnvKey(testMasterKey())
	if err != nil {
		t.Fatalf("NewEnvKey: %v", err)
	}
	dek := []byte("data-encryption-key-bytes!")
	wrapped, keyID, err := provider.Wrap(context.Background(), dek)
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	if keyID != provider.ActiveKeyID() {
		t.Fatalf("keyID = %q, want %q", keyID, provider.ActiveKeyID())
	}
	unwrapped, err := provider.Unwrap(context.Background(), wrapped, keyID)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if string(unwrapped) != string(dek) {
		t.Fatalf("unwrap mismatch: got %q want %q", unwrapped, dek)
	}
}

func TestStubKMSProviderReturnsError(t *testing.T) {
	var stub secrets.StubKMSProvider
	_, _, err := stub.Wrap(context.Background(), []byte("dek"))
	if err == nil {
		t.Fatal("expected KMS stub Wrap error")
	}
	_, err = stub.Unwrap(context.Background(), []byte("wrapped"), stub.ActiveKeyID())
	if err == nil {
		t.Fatal("expected KMS stub Unwrap error")
	}
}

func TestEnvKeyInvalidBase64(t *testing.T) {
	_, err := secrets.NewEnvKey("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestEnvKeyWrongLength(t *testing.T) {
	_, err := secrets.NewEnvKey(base64.StdEncoding.EncodeToString([]byte("short")))
	if err == nil {
		t.Fatal("expected error for wrong key length")
	}
}
