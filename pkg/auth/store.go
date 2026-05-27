package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrBootstrapComplete is returned when bootstrap is attempted but users already exist.
	ErrBootstrapComplete = errors.New("bootstrap already completed")
	// ErrInvalidCredentials is returned for failed login attempts.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrSessionNotFound is returned when a session token is unknown or expired.
	ErrSessionNotFound = errors.New("session not found")
	// ErrInvalidEmail is returned when bootstrap email is missing.
	ErrInvalidEmail = errors.New("email is required")
	// ErrInvalidPassword is returned when bootstrap password is too short.
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
)

// User is a local account.
type User struct {
	ID        string
	Email     string
	Role      string
	CreatedAt time.Time
}

// Store persists users and sessions.
type Store struct {
	db     *sql.DB
	driver string
}

// NewStore returns a Store backed by db.
func NewStore(db *sql.DB, driver string) *Store {
	return &Store{db: db, driver: driver}
}

// NeedsBootstrap reports whether no users exist yet.
func (s *Store) NeedsBootstrap(ctx context.Context) (bool, error) {
	count, err := s.UserCount(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// UserCount returns the number of registered users.
func (s *Store) UserCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// CreateBootstrapAdmin creates the first admin user when the database is empty.
func (s *Store) CreateBootstrapAdmin(ctx context.Context, email, password string) (*User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	q := s.q(`
		INSERT INTO users (id, email, password_hash, role, created_at)
		SELECT ?, ?, ?, 'admin', ?
		WHERE NOT EXISTS (SELECT 1 FROM users)
	`)
	res, err := s.db.ExecContext(ctx, q, id, email, hash, formatDBTime(now))
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrBootstrapComplete
	}
	return &User{ID: id, Email: email, Role: "admin", CreatedAt: now}, nil
}

// GetUserByEmail returns a user by email or nil when not found.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, string, error) {
	email = normalizeEmail(email)
	var u User
	var hash string
	var createdAt any
	q := s.q(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = ?`)
	err := s.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Email, &hash, &u.Role, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("get user by email: %w", err)
	}
	u.CreatedAt, err = scanTime(createdAt)
	if err != nil {
		return nil, "", err
	}
	return &u, hash, nil
}

// CreateSession stores a new session for userID.
func (s *Store) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	id := uuid.NewString()
	now := time.Now().UTC()
	q := s.q(`INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`)
	_, err := s.db.ExecContext(ctx, q, id, userID, tokenHash, formatDBTime(expiresAt.UTC()), formatDBTime(now))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// GetUserBySessionToken returns the user for a valid, unexpired session token hash.
func (s *Store) GetUserBySessionToken(ctx context.Context, tokenHash string) (*User, error) {
	var u User
	var createdAt any
	var expiresAt any
	q := s.q(`
		SELECT u.id, u.email, u.role, u.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ?
	`)
	err := s.db.QueryRowContext(ctx, q, tokenHash).Scan(&u.ID, &u.Email, &u.Role, &createdAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session user: %w", err)
	}
	exp, err := scanTime(expiresAt)
	if err != nil {
		return nil, err
	}
	if time.Now().UTC().After(exp) {
		_ = s.DeleteSession(ctx, tokenHash)
		return nil, ErrSessionNotFound
	}
	u.CreatedAt, err = scanTime(createdAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// DeleteSession removes a session by token hash.
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	q := s.q(`DELETE FROM sessions WHERE token_hash = ?`)
	_, err := s.db.ExecContext(ctx, q, tokenHash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// Authenticate verifies email/password and returns the user.
func (s *Store) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, hash, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	ok, err := VerifyPassword(hash, password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Store) q(query string) string {
	if s.driver != "postgres" {
		return query
	}
	var b strings.Builder
	n := 1
	for _, r := range query {
		if r == '?' {
			fmt.Fprintf(&b, "$%d", n)
			n++
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func scanTime(value any) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v.UTC(), nil
	case string:
		return parseTimeString(v)
	case []byte:
		return parseTimeString(string(v))
	default:
		return time.Time{}, fmt.Errorf("unsupported time value type %T", value)
	}
}

func parseTimeString(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t, err = time.Parse(time.RFC3339, value)
	}
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", value)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", value, err)
	}
	return t.UTC(), nil
}

func formatDBTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
