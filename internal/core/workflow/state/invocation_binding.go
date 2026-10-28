package state

import (
	"time"
)

// CandidateRepairBinding retains actionable combined-result evidence.
// It cannot authorize publication or rewind accepted source task execution.
type CandidateRepairBinding struct {
	Unit            Ref         `json:"unit"`
	EntryID         string      `json:"entry_id"`
	ClaimGeneration uint64      `json:"claim_generation"`
	CandidateHash   string      `json:"candidate_hash"`
	FindingRef      EvidenceRef `json:"finding_ref"`
	RecordRef       EvidenceRef `json:"record_ref"`
}

// ContinuationCheckpoint binds a bounded call to frozen request identity.
// Prior images remain immutable; ParentRef/Hash retain the previous checkpoint.
type ContinuationCheckpoint struct {
	Schema          string    `json:"schema"`
	InvocationID    string    `json:"invocation_id"`
	RequestHash     string    `json:"request_hash"`
	Sequence        int       `json:"sequence"`
	SourceRevision  uint64    `json:"source_revision"`
	Cursor          int       `json:"cursor"`
	ConfirmedCursor int       `json:"confirmed_cursor"`
	MemberCount     int       `json:"member_count"`
	AttemptCount    int       `json:"attempt_count"`
	Steps           int       `json:"steps"`
	Reason          string    `json:"reason"`
	ParentRef       string    `json:"parent_ref,omitempty"`
	ParentHash      string    `json:"parent_hash,omitempty"`
	RecordedAt      time.Time `json:"recorded_at"`
	EvidenceRef     string    `json:"evidence_ref,omitempty"`
	EvidenceHash    string    `json:"evidence_hash,omitempty"`
}

// IntegrationBudgetBinding seals the invocation-wide budget image.
type IntegrationBudgetBinding struct {
	Revision uint64      `json:"revision"`
	Ref      EvidenceRef `json:"ref"`
}

// InvocationClosure retains a confirmed set and its refreshed execution
// base without changing earlier unit delivery images or admission identity.
type InvocationClosure struct {
	SetID   string          `json:"set_id"`
	Cursor  int             `json:"cursor"`
	Receipt ClosingEvidence `json:"receipt"`
	Target  PreviewTarget   `json:"target"`
}

// PreparationBinding retains bounded preparation across candidate changes.
type PreparationBinding struct {
	Round      int    `json:"round"`
	Unit       Ref    `json:"unit"`
	State      string `json:"state"`
	SourceHash string `json:"source_hash"`
	ResultRef  string `json:"result_ref"`
	ResultHash string `json:"result_hash"`
}

// PullRequestBinding retains creation intent across crash and candidate repair.
type PullRequestBinding struct {
	Schema           string             `json:"schema"`
	ID               string             `json:"id"`
	Unit             Ref                `json:"unit"`
	Request          PullRequestRequest `json:"request"`
	TargetCommit     string             `json:"target_commit"`
	CandidateID      string             `json:"candidate_id"`
	CandidateRef     string             `json:"candidate_ref"`
	CandidateHash    string             `json:"candidate_hash"`
	SourceCommit     string             `json:"source_commit"`
	SourceTree       string             `json:"source_tree"`
	URL              string             `json:"url"`
	State            string             `json:"state"`
	Reason           string             `json:"reason"`
	ObservationCount int                `json:"observation_count"`
	ObservationRef   string             `json:"observation_ref"`
	ObservationHash  string             `json:"observation_hash"`
}

// ReviewRepairBudget survives candidate replacement and continuation.
type ReviewRepairBudget struct {
	Unit     Ref `json:"unit"`
	Consumed int `json:"consumed"`
}

// SubmissionBinding projects the repository journal into an invocation.
// The common-directory journal remains authoritative after interrupted updates.
type SubmissionBinding struct {
	JournalKey       string          `json:"journal_key"`
	JournalHash      string          `json:"journal_hash"`
	TargetKey        string          `json:"target_key"`
	ID               string          `json:"id"`
	Unit             Ref             `json:"unit"`
	Attempt          int             `json:"attempt"`
	State            SubmissionState `json:"state"`
	ExternalDelivery bool            `json:"external_delivery,omitempty"`
}

// UnitDelivery retains a delivered unit before its active bindings move.
// EvidenceRef binds the exact frontier, target and receipt to an immutable image.
type UnitDelivery struct {
	Schema            string              `json:"schema"`
	InvocationID      string              `json:"invocation_id"`
	Unit              Ref                 `json:"unit"`
	SetID             string              `json:"set_id,omitempty"`
	Cursor            int                 `json:"cursor"`
	Target            PreviewTarget       `json:"target"`
	NextTarget        PreviewTarget       `json:"next_target"`
	Lane              Lane                `json:"lane"`
	CandidateRef      string              `json:"candidate_ref"`
	CandidateHash     string              `json:"candidate_hash"`
	PullRequest       *PullRequestBinding `json:"pull_request_binding"`
	DeliveryProofRef  string              `json:"delivery_proof_ref"`
	DeliveryProofHash string              `json:"delivery_proof_hash"`
	DeliveryRef       string              `json:"delivery_ref"`
	DeliveryHash      string              `json:"delivery_hash"`
	Preparation       *PreparationBinding `json:"preparation_binding,omitempty"`
	Review            *ReviewBinding      `json:"review_binding,omitempty"`
	Submission        *SubmissionBinding  `json:"submission_binding,omitempty"`
	Proof             IntegrationProof    `json:"proof"`
	EvidenceRef       string              `json:"evidence_ref,omitempty"`
	EvidenceHash      string              `json:"evidence_hash,omitempty"`
}
