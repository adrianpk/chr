package state

import (
	"encoding/json"
	"time"
)

// WorkflowSchema identifies explicitly executable workflow records.
const WorkflowSchema = "chr-workflow/v1"

// Task extends a native ticket without replacing its display fields.
type Task struct {
	Schema          string        `json:"workflow_schema"`
	Key             string        `json:"key"`
	Status          string        `json:"status"`
	UseCase         string        `json:"use_case,omitempty"`
	SetID           string        `json:"slice_set_id,omitempty"`
	SliceID         string        `json:"slice_id,omitempty"`
	Order           int           `json:"task_order,omitempty"`
	DependsOn       []Ref         `json:"depends_on,omitempty"`
	Phase           string        `json:"task_phase"`
	ContractHash    string        `json:"execution_contract_hash"`
	AttemptID       string        `json:"active_attempt_id,omitempty"`
	ClaimGeneration uint64        `json:"claim_generation,omitempty"`
	Evidence        []string      `json:"task_evidence,omitempty"`
	DeliveryRef     string        `json:"slice_delivery_ref,omitempty"`
	Delivery        *TaskDelivery `json:"slice_delivery,omitempty"`
}

// TaskDelivery projects the common receipt's immutable delivery facts.
// The set-owned PR and shared receipt remain authoritative.
type TaskDelivery struct {
	PullRequest  string    `json:"pull_request"`
	Branch       string    `json:"branch"`
	Method       string    `json:"method"`
	ResultCommit string    `json:"result_commit"`
	ConfirmedAt  time.Time `json:"confirmed_at"`
}

// Receipt retains immutable integration facts, separate from projections.
type Receipt struct {
	Native          *NativeReceipt `json:"native,omitempty"`
	Schema          string         `json:"workflow_schema"`
	Unit            Ref            `json:"unit"`
	CandidateID     string         `json:"candidate_id"`
	SubmissionID    string         `json:"submission_id"`
	Provider        string         `json:"provider"`
	Repository      string         `json:"repository"`
	PullRequest     string         `json:"pull_request"`
	Method          string         `json:"method"`
	SourceCommit    string         `json:"source_commit"`
	BaseCommit      string         `json:"base_commit"`
	PreparedTree    string         `json:"prepared_tree"`
	HistoryProof    string         `json:"history_proof"`
	ResultCommit    string         `json:"result_commit"`
	VerifiedTree    string         `json:"verified_tree"`
	ObservedTarget  string         `json:"observed_target"`
	VerificationRef string         `json:"verification_ref"`
	ValidationRefs  []string       `json:"validation_refs"`
	ConfirmedAt     time.Time      `json:"confirmed_at"`
}

// Record separates envelope identity, native content and workflow domain.
// Path is relative to the namespace root. Metadata and Domain retain unknown fields.
type Record struct {
	Identity
	Kind       string
	Path       string
	Body       []byte
	Task       *Task
	Set        *WorkflowSet
	Receipt    *Receipt
	Metadata   map[string]json.RawMessage
	Domain     map[string]json.RawMessage
	SourceHash string
}

// Snapshot is one validated namespace, including independent tickets.
type Snapshot struct {
	Records []Record
}

// SliceDeliveryProof retains verified external delivery before finalization.
type SliceDeliveryProof struct {
	Schema           string              `json:"schema"`
	InvocationID     string              `json:"invocation_id"`
	Unit             Ref                 `json:"unit"`
	CandidateID      string              `json:"candidate_id"`
	CandidateRef     string              `json:"candidate_ref"`
	CandidateHash    string              `json:"candidate_hash"`
	SubmissionID     string              `json:"submission_id"`
	PullRequest      PullRequestSnapshot `json:"pull_request"`
	Git              IntegrationProof    `json:"git"`
	ConfirmedAt      time.Time           `json:"confirmed_at"`
	NativeReceiptRef *EvidenceRef        `json:"native_receipt_ref,omitempty"`
}
