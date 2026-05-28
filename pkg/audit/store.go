package audit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)
type Store struct {
	db     *sql.DB
	driver string
}

// NewStore creates an audit store.
func NewStore(db *sql.DB, driver string) *Store {
	return &Store{db: db, driver: driver}
}

// Append inserts a new audit event under a serializable transaction.
func (s *Store) Append(ctx context.Context, e Event) (int64, error) {
	if e.Action == "" {
		return 0, errors.New("audit action is required")
	}
	occurredAt := time.Now().UTC().Format(time.RFC3339Nano)
	payload := PayloadFromEvent(occurredAt, e)

	tx, err := s.beginSerializable(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin audit tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	prevHash, err := s.latestRowHash(ctx, tx)
	if err != nil {
		return 0, err
	}
	rowHash, err := RowHash(prevHash, payload)
	if err != nil {
		return 0, err
	}

	id, err := s.insertRow(ctx, tx, prevHash, rowHash, occurredAt, e)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit audit tx: %w", err)
	}
	return id, nil
}

func (s *Store) beginSerializable(ctx context.Context) (*sql.Tx, error) {
	if s.driver == "postgres" {
		return s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	}
	return s.db.BeginTx(ctx, nil)
}

func (s *Store) latestRowHash(ctx context.Context, tx *sql.Tx) ([]byte, error) {
	var prevHash []byte
	var query string
	if s.driver == "postgres" {
		query = `SELECT row_hash FROM audit_events ORDER BY id DESC LIMIT 1 FOR UPDATE`
	} else {
		query = `SELECT row_hash FROM audit_events ORDER BY id DESC LIMIT 1`
	}
	err := tx.QueryRowContext(ctx, query).Scan(&prevHash)
	if errors.Is(err, sql.ErrNoRows) {
		return GenesisPrevHash(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read latest audit hash: %w", err)
	}
	return prevHash, nil
}

func (s *Store) insertRow(ctx context.Context, tx *sql.Tx, prevHash, rowHash []byte, occurredAt string, e Event) (int64, error) {
	if s.driver == "postgres" {
		var id int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO audit_events (
				prev_hash, row_hash, occurred_at, actor_id, action,
				resource_type, resource_id, request_id, before_hash, after_hash
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`, prevHash, rowHash, occurredAt, e.ActorID, e.Action,
			e.ResourceType, e.ResourceID, e.RequestID, e.BeforeHash, e.AfterHash).Scan(&id)
		return id, err
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO audit_events (
			prev_hash, row_hash, occurred_at, actor_id, action,
			resource_type, resource_id, request_id, before_hash, after_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, prevHash, rowHash, occurredAt, e.ActorID, e.Action,
		e.ResourceType, e.ResourceID, e.RequestID, e.BeforeHash, e.AfterHash)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListAll returns all audit rows ordered by id ascending.
func (s *Store) ListAll(ctx context.Context) ([]Row, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, prev_hash, row_hash, occurred_at, actor_id, action,
			resource_type, resource_id, request_id, before_hash, after_hash
		FROM audit_events ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(
			&r.ID, &r.PrevHash, &r.RowHash, &r.OccurredAt, &r.ActorID, &r.Action,
			&r.ResourceType, &r.ResourceID, &r.RequestID, &r.BeforeHash, &r.AfterHash,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
