package state

import (
	"time"
)

// Blocker identifies one prerequisite without adding it to the extent.
type Blocker struct {
	Dependency      Ref
	Reason          string
	NextActionOwner string
}

// DependencyTarget binds a resumed independent cursor to canonical
// descendant history and the immutable prerequisite receipts observed there.
type DependencyTarget struct {
	Schema       string            `json:"schema"`
	InvocationID string            `json:"invocation_id"`
	Cursor       int               `json:"cursor"`
	TicketKey    string            `json:"ticket_key"`
	Previous     PreviewTarget     `json:"previous_target"`
	Target       PreviewTarget     `json:"target"`
	Branch       string            `json:"branch"`
	Blocker      Blocker           `json:"blocker"`
	Receipts     map[string]string `json:"receipt_hashes"`
	ConfirmedAt  time.Time         `json:"confirmed_at"`
	EvidenceRef  string            `json:"evidence_ref,omitempty"`
	EvidenceHash string            `json:"evidence_hash,omitempty"`
}

// InputImage freezes an owning set contract separately from its tasks.
type InputImage struct {
	Path       string `json:"path"`
	Revision   uint64 `json:"revision"`
	SourceHash string `json:"source_hash"`
	UseCase    string `json:"use_case,omitempty"`
}

// LooseQueue retains admission order independently of live pending order.
// Excluded members remain admitted facts; they are never counted as delivered.
type LooseQueue struct {
	Admitted []string          `json:"admitted"`
	Cursor   int               `json:"cursor"`
	Eligible int               `json:"eligible"`
	Excluded map[string]string `json:"excluded"`
}

// Member retains admission facts and the latest accepted owned image.
type Member struct {
	Admission          PreviewMember `json:"admission"`
	ExpectedRevision   uint64        `json:"expected_revision"`
	ExpectedSourceHash string        `json:"expected_source_hash"`
}

// PreviewMember freezes the identity, revision and contract of one task.
type PreviewMember struct {
	Key          string
	Path         string
	Revision     uint64
	SourceHash   string
	ContractHash string
	SetID        string
	SliceID      string
	TaskOrder    int
	Status       string
	Phase        string
	Readiness    string
	Blocker      *Blocker
}
