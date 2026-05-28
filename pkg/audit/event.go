package audit

// Event is an append-only audit record input.
type Event struct {
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	RequestID    string
	BeforeHash   string
	AfterHash    string
}

// Row is a stored audit event with chain metadata.
type Row struct {
	ID           int64
	PrevHash     []byte
	RowHash      []byte
	OccurredAt   string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	RequestID    string
	BeforeHash   string
	AfterHash    string
}
