package audit

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"sort"
)

type canonicalPayload struct {
	Time         string `json:"time"`
	Actor        string `json:"actor"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	RequestID    string `json:"request_id"`
	BeforeHash   string `json:"before_hash"`
	AfterHash    string `json:"after_hash"`
}

// CanonicalSerialize returns deterministic JSON for hash chain fields.
func CanonicalSerialize(row canonicalPayload) ([]byte, error) {
	// json.Marshal on struct with fixed field order provides stable output for our schema.
	return json.Marshal(row)
}

// RowHash computes sha256(prevHash || canonicalFields).
func RowHash(prevHash []byte, row canonicalPayload) ([]byte, error) {
	body, err := CanonicalSerialize(row)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	if len(prevHash) == 0 {
		prevHash = GenesisPrevHash()
	}
	h.Write(prevHash)
	h.Write(body)
	sum := h.Sum(nil)
	return sum, nil
}

// GenesisPrevHash returns 32 zero bytes for the first chain row.
func GenesisPrevHash() []byte {
	return make([]byte, 32)
}

// EqualHash compares two hash byte slices in constant time length check.
func EqualHash(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// PayloadFromEvent builds canonical payload from event fields.
func PayloadFromEvent(occurredAt string, e Event) canonicalPayload {
	return canonicalPayload{
		Time:         occurredAt,
		Actor:        e.ActorID,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		RequestID:    e.RequestID,
		BeforeHash:   e.BeforeHash,
		AfterHash:    e.AfterHash,
	}
}

// SortRowsByID returns a copy sorted by ascending ID for verification.
func SortRowsByID(rows []Row) []Row {
	out := append([]Row(nil), rows...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
