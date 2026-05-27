package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

var (
	// ErrInvalidPasswordHash indicates a stored hash is malformed.
	ErrInvalidPasswordHash = errors.New("invalid password hash")
)

type argonParams struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

// HashPassword returns an argon2id encoded password hash.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads, b64Salt, b64Hash), nil
}

// VerifyPassword compares a password against an argon2id hash.
func VerifyPassword(encodedHash, password string) (bool, error) {
	params, salt, hash, err := decodeArgon2Hash(encodedHash)
	if err != nil {
		return false, err
	}
	other := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, params.keyLen)
	if subtle.ConstantTimeCompare(hash, other) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeArgon2Hash(encoded string) (argonParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argonParams{}, nil, nil, ErrInvalidPasswordHash
	}
	params, err := parseArgonParams(parts[3])
	if err != nil {
		return argonParams{}, nil, nil, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argonParams{}, nil, nil, fmt.Errorf("%w: decode salt: %v", ErrInvalidPasswordHash, err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argonParams{}, nil, nil, fmt.Errorf("%w: decode hash: %v", ErrInvalidPasswordHash, err)
	}
	if len(salt) == 0 || len(hash) == 0 {
		return argonParams{}, nil, nil, ErrInvalidPasswordHash
	}
	params.keyLen = uint32(len(hash)) // #nosec G115 -- key length comes from decoded hash bytes
	return params, salt, hash, nil
}

func parseArgonParams(raw string) (argonParams, error) {
	var params argonParams
	for _, part := range strings.Split(raw, ",") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			return argonParams{}, fmt.Errorf("%w: malformed params %q", ErrInvalidPasswordHash, raw)
		}
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return argonParams{}, fmt.Errorf("%w: parse param %q: %v", ErrInvalidPasswordHash, part, err)
		}
		switch k {
		case "m":
			params.memory = uint32(n)
		case "t":
			params.time = uint32(n)
		case "p":
			if n > 255 {
				return argonParams{}, fmt.Errorf("%w: threads out of range", ErrInvalidPasswordHash)
			}
			params.threads = uint8(n) // #nosec G115 -- bounded above
		default:
			return argonParams{}, fmt.Errorf("%w: unknown param %q", ErrInvalidPasswordHash, k)
		}
	}
	if params.memory == 0 || params.time == 0 || params.threads == 0 {
		return argonParams{}, fmt.Errorf("%w: incomplete params %q", ErrInvalidPasswordHash, raw)
	}
	return params, nil
}
