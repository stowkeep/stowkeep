package audit

import (
	"context"
	"database/sql"
	"fmt"
)

// VerifyResult describes chain verification outcome.
type VerifyResult struct {
	OK             bool
	BreakAtEventID int64
	Detail         string
}

// VerifyChain checks hash chain integrity for all rows.
func VerifyChain(ctx context.Context, store *Store) (VerifyResult, error) {
	rows, err := store.ListAll(ctx)
	if err != nil {
		return VerifyResult{}, err
	}
	if len(rows) == 0 {
		return VerifyResult{OK: true}, nil
	}

	prev := GenesisPrevHash()
	for _, row := range rows {
		if !EqualHash(row.PrevHash, prev) {
			return VerifyResult{
				OK:             false,
				BreakAtEventID: row.ID,
				Detail:         fmt.Sprintf("prev_hash mismatch at event %d", row.ID),
			}, nil
		}
		occurredAt := row.OccurredAt
		payload := PayloadFromEvent(occurredAt, Event{
			ActorID:      row.ActorID,
			Action:       row.Action,
			ResourceType: row.ResourceType,
			ResourceID:   row.ResourceID,
			RequestID:    row.RequestID,
			BeforeHash:   row.BeforeHash,
			AfterHash:    row.AfterHash,
		})
		expected, err := RowHash(row.PrevHash, payload)
		if err != nil {
			return VerifyResult{}, err
		}
		if !EqualHash(row.RowHash, expected) {
			return VerifyResult{
				OK:             false,
				BreakAtEventID: row.ID,
				Detail:         fmt.Sprintf("row_hash mismatch at event %d", row.ID),
			}, nil
		}
		prev = row.RowHash
	}
	return VerifyResult{OK: true}, nil
}

// RecordIntegrityBreak writes an audit_integrity_events row.
func RecordIntegrityBreak(ctx context.Context, db *sql.DB, driver string, breakAt int64, detail string) error {
	if driver == "postgres" {
		_, err := db.ExecContext(ctx, `
			INSERT INTO audit_integrity_events (break_at_event_id, detail) VALUES ($1, $2)
		`, breakAt, detail)
		return err
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO audit_integrity_events (break_at_event_id, detail) VALUES (?, ?)
	`, breakAt, detail)
	return err
}

// StartVerifier runs chain verification in the background on startup.
func StartVerifier(ctx context.Context, store *Store, onBreak func(VerifyResult)) {
	go func() {
		res, err := VerifyChain(ctx, store)
		if err != nil || res.OK {
			return
		}
		if onBreak != nil {
			onBreak(res)
		}
	}()
}
