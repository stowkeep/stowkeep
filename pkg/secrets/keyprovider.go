// Package secrets provides envelope encryption interfaces for Stowkeep secret storage.
// Secret plaintext must never be logged or written to audit payloads.
package secrets

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// MasterKeyProvider wraps and unwraps data encryption keys using a master encryption key.
type MasterKeyProvider interface {
	Wrap(ctx context.Context, dek []byte) (wrapped []byte, keyID string, err error)
	Unwrap(ctx context.Context, wrapped []byte, keyID string) (dek []byte, err error)
	ActiveKeyID() string
}

// KMSProvider is a stub interface for future cloud KMS implementations (Stage 7+).
type KMSProvider interface {
	MasterKeyProvider
}

var (
	// ErrMasterKeyNotConfigured is returned when STOWKEEP_MASTER_KEY is unset.
	ErrMasterKeyNotConfigured = errors.New("STOWKEEP_MASTER_KEY is not configured")
	// ErrInvalidMasterKey is returned when the master key cannot be decoded.
	ErrInvalidMasterKey = errors.New("STOWKEEP_MASTER_KEY must be base64-encoded 32 bytes")
)

// EnvKey implements MasterKeyProvider using STOWKEEP_MASTER_KEY from the environment.
type EnvKey struct {
	key   []byte
	keyID string
}

func envKeyID(key []byte) string {
	sum := sha256.Sum256(key)
	return "env:" + hex.EncodeToString(sum[:8])
}

// NewEnvKey creates an EnvKey from a base64-encoded 32-byte master key.
func NewEnvKey(masterKeyB64 string) (*EnvKey, error) {
	if masterKeyB64 == "" {
		return nil, ErrMasterKeyNotConfigured
	}
	raw, err := base64.StdEncoding.DecodeString(masterKeyB64)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidMasterKey, err)
	}
	if len(raw) != 32 {
		return nil, ErrInvalidMasterKey
	}
	return &EnvKey{key: raw, keyID: envKeyID(raw)}, nil
}

// ActiveKeyID returns the key identifier for env-based MEK.
func (e *EnvKey) ActiveKeyID() string {
	return e.keyID
}

// Wrap stores the DEK using XOR with the MEK (Stage 0 stub — real AES-GCM in Stage 4).
func (e *EnvKey) Wrap(_ context.Context, dek []byte) ([]byte, string, error) {
	if len(dek) == 0 {
		return nil, "", errors.New("dek must not be empty")
	}
	out := make([]byte, len(dek))
	for i := range dek {
		out[i] = dek[i] ^ e.key[i%len(e.key)]
	}
	return out, e.keyID, nil
}

// Unwrap recovers the DEK using the MEK identified by keyID.
func (e *EnvKey) Unwrap(_ context.Context, wrapped []byte, keyID string) ([]byte, error) {
	if keyID != e.keyID {
		return nil, fmt.Errorf("unsupported key_id %q", keyID)
	}
	if len(wrapped) == 0 {
		return nil, errors.New("wrapped dek must not be empty")
	}
	out := make([]byte, len(wrapped))
	for i := range wrapped {
		out[i] = wrapped[i] ^ e.key[i%len(e.key)]
	}
	return out, nil
}

// StubKMSProvider documents the future KMS integration point without an implementation.
type StubKMSProvider struct{}

// ActiveKeyID returns a placeholder key id for the stub.
func (StubKMSProvider) ActiveKeyID() string { return "kms:stub" }

// Wrap returns an error indicating KMS is not implemented.
func (StubKMSProvider) Wrap(context.Context, []byte) ([]byte, string, error) {
	return nil, "", errors.New("KMSProvider is not implemented in Stage 0")
}

// Unwrap returns an error indicating KMS is not implemented.
func (StubKMSProvider) Unwrap(context.Context, []byte, string) ([]byte, error) {
	return nil, errors.New("KMSProvider is not implemented in Stage 0")
}
