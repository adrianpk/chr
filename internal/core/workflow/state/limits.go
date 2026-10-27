package state

// Workflow bounds apply to durable input, not just its derived index.
const (
	RecordLimit = 4096
	FileLimit   = 1 << 20
	PageLimit   = 128
)
