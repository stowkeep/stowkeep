package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/argon2"
)

func TestRebindPostgresQuery(t *testing.T) {
	store := &Store{driver: "postgres"}
	got := store.q(`SELECT * FROM users WHERE email = ? AND id = ?`)
	want := `SELECT * FROM users WHERE email = $1 AND id = $2`
	if got != want {
		t.Fatalf("q() = %q, want %q", got, want)
	}
}

func TestScanTimeVariants(t *testing.T) {
	cases := []struct {
		name  string
		value any
	}{
		{"time", time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)},
		{"rfc3339", "2026-05-26T12:00:00Z"},
		{"bytes", []byte("2026-05-26T12:00:00Z")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := scanTime(tc.value)
			if err != nil {
				t.Fatalf("scanTime: %v", err)
			}
			if got.IsZero() {
				t.Fatal("expected non-zero time")
			}
		})
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	_, err := VerifyPassword("not-a-hash", "pw")
	if err != ErrInvalidPasswordHash {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyPasswordUsesStoredParams(t *testing.T) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("rand: %v", err)
	}
	const (
		customMemory  = uint32(32 * 1024)
		customTime    = uint32(2)
		customThreads = uint8(2)
	)
	hash := argon2.IDKey([]byte("password123"), salt, customTime, customMemory, customThreads, argonKeyLen)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		customMemory, customTime, customThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	ok, err := VerifyPassword(encoded, "password123")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify with stored params")
	}
}

func TestOptionalAuthWithoutSession(t *testing.T) {
	nextCalled := false
	handler := OptionalAuth(&Store{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		if _, ok := UserFromContext(r.Context()); ok {
			t.Fatal("unexpected user in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !nextCalled || rec.Code != http.StatusOK {
		t.Fatal("expected next handler to run")
	}
}

func TestRequestClientKey(t *testing.T) {
	if got := requestClientKey("192.168.1.1:4321"); got != "192.168.1.1" {
		t.Fatalf("got %q", got)
	}
	if got := requestClientKey("no-port"); got != "no-port" {
		t.Fatalf("got %q", got)
	}
}

func TestNewHandlerDefaultTTL(t *testing.T) {
	h := NewHandler(&Store{}, HandlerConfig{})
	if h.cfg.SessionIdleTTL != 24*time.Hour {
		t.Fatalf("ttl = %v", h.cfg.SessionIdleTTL)
	}
}

func TestParseTimeGoStringFormat(t *testing.T) {
	_, err := parseTimeString("2026-05-26 17:01:52.5524 +0000 UTC")
	if err != nil {
		t.Fatalf("parseTimeString: %v", err)
	}
}

func TestScanTimeUnsupportedType(t *testing.T) {
	_, err := scanTime(123)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeArgon2InvalidBase64(t *testing.T) {
	_, _, _, err := decodeArgon2Hash("$argon2id$v=19$m=65536,t=3,p=4$!!!$!!!")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeArgon2EmptySaltOrHash(t *testing.T) {
	_, _, _, err := decodeArgon2Hash("$argon2id$v=19$m=65536,t=3,p=4$$")
	if err != ErrInvalidPasswordHash {
		t.Fatalf("err = %v", err)
	}
}

func TestDecodeArgon2WrongParts(t *testing.T) {
	_, _, _, err := decodeArgon2Hash("$argon2id$v=19$m=65536")
	if err != ErrInvalidPasswordHash {
		t.Fatalf("err = %v", err)
	}
}

func TestParseTimeInvalid(t *testing.T) {
	_, err := parseTimeString("not-a-time")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSQLiteQueryPassthrough(t *testing.T) {
	store := &Store{driver: "sqlite"}
	if got := store.q("SELECT ?"); got != "SELECT ?" {
		t.Fatalf("got %q", got)
	}
}
