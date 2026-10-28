package state

import (
	"time"
)

// ClosingMember freezes each actual ticket, including work excluded by
// an invocation's default membership policy.
type ClosingMember struct {
	Identity
	Key        string `json:"key"`
	Path       string `json:"path"`
	SourceHash string `json:"source_hash"`
	SliceID    string `json:"slice_id"`
	Order      int    `json:"task_order"`
}

// ClosingSlice binds ordered descriptors to actual integration evidence.
type ClosingSlice struct {
	ID           string           `json:"id"`
	Order        int              `json:"slice_order"`
	Receipt      ClosingEvidence  `json:"receipt"`
	Verification ClosingEvidence  `json:"verification"`
	Proof        IntegrationProof `json:"proof"`
}

// ClosingInput is the immutable member/plan/contract/target snapshot.
type ClosingInput struct {
	ValidationTimeoutSeconds int               `json:"validation_timeout_seconds,omitempty"`
	Schema                   string            `json:"schema"`
	SetID                    string            `json:"set_id"`
	SetRevision              uint64            `json:"set_revision"`
	SetSourceHash            string            `json:"set_source_hash"`
	ProjectID                string            `json:"project_id"`
	Namespace                string            `json:"namespace"`
	Plan                     ClosingEvidence   `json:"plan"`
	ContractHash             string            `json:"closing_contract_hash"`
	Validation               Validation        `json:"closing_validation"`
	Members                  []ClosingMember   `json:"members"`
	Slices                   []ClosingSlice    `json:"slices"`
	Target                   PreviewTarget     `json:"target"`
	Branch                   string            `json:"branch"`
	Tree                     string            `json:"tree"`
	EvidenceHashes           map[string]string `json:"evidence_hashes"`
}

// ClosingReceipt retains exact inputs and host results, including an
// explicit waiver. Only a successful receipt can become the current closure.
type ClosingReceipt struct {
	ClosingInput
	Input       ClosingEvidence `json:"input"`
	Validation  []CommandResult `json:"validation"`
	Outcome     string          `json:"outcome"`
	Waiver      string          `json:"waiver_reason,omitempty"`
	ConfirmedAt time.Time       `json:"confirmed_at"`
}
