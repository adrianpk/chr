package core

import (
	"encoding/json"
	"time"
)

// WorkflowSchema identifies explicitly executable workflow records.
const WorkflowSchema = "chr-workflow/v1"

// Workflow bounds apply to durable input, not just its derived index.
const (
	WorkflowRecordLimit = 4096
	WorkflowFileLimit   = 1 << 20
	WorkflowPageLimit   = 128
)

// WorkflowIdentity is a scoped durable identity independent of a checkout path.
type WorkflowIdentity struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Namespace string    `json:"namespace"`
	Revision  uint64    `json:"revision"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WorkflowRef is a ticket-key or slice-ID reference in the containing scope.
type WorkflowRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// WorkflowTask extends a native ticket without replacing its display fields.
type WorkflowTask struct {
	Schema          string                `json:"workflow_schema"`
	Key             string                `json:"key"`
	Status          string                `json:"status"`
	UseCase         string                `json:"use_case,omitempty"`
	SetID           string                `json:"slice_set_id,omitempty"`
	SliceID         string                `json:"slice_id,omitempty"`
	Order           int                   `json:"task_order,omitempty"`
	DependsOn       []WorkflowRef         `json:"depends_on,omitempty"`
	Phase           string                `json:"task_phase"`
	ContractHash    string                `json:"execution_contract_hash"`
	AttemptID       string                `json:"active_attempt_id,omitempty"`
	ClaimGeneration uint64                `json:"claim_generation,omitempty"`
	Evidence        []string              `json:"task_evidence,omitempty"`
	DeliveryRef     string                `json:"slice_delivery_ref,omitempty"`
	Delivery        *WorkflowTaskDelivery `json:"slice_delivery,omitempty"`
}

// WorkflowTaskDelivery projects the common receipt's immutable delivery facts.
// The set-owned PR and shared receipt remain authoritative.
type WorkflowTaskDelivery struct {
	PullRequest  string    `json:"pull_request"`
	Branch       string    `json:"branch"`
	Method       string    `json:"method"`
	ResultCommit string    `json:"result_commit"`
	ConfirmedAt  time.Time `json:"confirmed_at"`
}

// WorkflowValidation is an ordered argv gate with one shared deadline.
type WorkflowValidation struct {
	Required       bool       `json:"required"`
	Commands       [][]string `json:"commands,omitempty"`
	TimeoutSeconds int        `json:"timeout_seconds"`
	WaiverReason   string     `json:"waiver_reason,omitempty"`
}

// WorkflowSlice is a set-owned descriptor; tickets alone own task membership.
type WorkflowSlice struct {
	ID            string        `json:"id"`
	Order         int           `json:"slice_order"`
	Title         string        `json:"title"`
	ContractRef   string        `json:"contract_ref"`
	ContractHash  string        `json:"contract_hash"`
	TaskInputHash string        `json:"task_input_hash"`
	DependsOn     []WorkflowRef `json:"depends_on,omitempty"`
	Phase         string        `json:"phase"`
	Branch        string        `json:"branch,omitempty"`
	CandidateRef  string        `json:"candidate_ref"`
	PullRequest   string        `json:"pull_request,omitempty"`
	DeliveryRef   string        `json:"delivery_ref,omitempty"`
}

// WorkflowSet owns ordered slices and its closing contract.
type WorkflowSet struct {
	Closing           *WorkflowClosingBinding `json:"closing,omitempty"`
	Amendments        []WorkflowAmendmentRef  `json:"amendments,omitempty"`
	Schema            string                  `json:"workflow_schema"`
	UseCase           string                  `json:"use_case,omitempty"`
	UseCaseOrder      int                     `json:"use_case_order,omitempty"`
	Status            string                  `json:"status"`
	PlanRef           string                  `json:"plan_ref"`
	PlanHash          string                  `json:"plan_hash"`
	Slices            []WorkflowSlice         `json:"slices"`
	ClosingValidation WorkflowValidation      `json:"closing_validation"`
	ClosureRef        string                  `json:"closure_ref,omitempty"`
	ClosureHistory    []string                `json:"closure_history,omitempty"`
}

// WorkflowReceipt retains immutable integration facts, separate from projections.
type WorkflowReceipt struct {
	Native          *WorkflowNativeReceipt `json:"native,omitempty"`
	Schema          string                 `json:"workflow_schema"`
	Unit            WorkflowRef            `json:"unit"`
	CandidateID     string                 `json:"candidate_id"`
	SubmissionID    string                 `json:"submission_id"`
	Provider        string                 `json:"provider"`
	Repository      string                 `json:"repository"`
	PullRequest     string                 `json:"pull_request"`
	Method          string                 `json:"method"`
	SourceCommit    string                 `json:"source_commit"`
	BaseCommit      string                 `json:"base_commit"`
	PreparedTree    string                 `json:"prepared_tree"`
	HistoryProof    string                 `json:"history_proof"`
	ResultCommit    string                 `json:"result_commit"`
	VerifiedTree    string                 `json:"verified_tree"`
	ObservedTarget  string                 `json:"observed_target"`
	VerificationRef string                 `json:"verification_ref"`
	ValidationRefs  []string               `json:"validation_refs"`
	ConfirmedAt     time.Time              `json:"confirmed_at"`
}

// WorkflowRecord separates envelope identity, native content and workflow domain.
// Path is relative to the namespace root. Metadata and Domain retain unknown fields.
type WorkflowRecord struct {
	WorkflowIdentity
	Kind       string
	Path       string
	Body       []byte
	Task       *WorkflowTask
	Set        *WorkflowSet
	Receipt    *WorkflowReceipt
	Metadata   map[string]json.RawMessage
	Domain     map[string]json.RawMessage
	SourceHash string
}

// WorkflowSnapshot is one validated namespace, including independent tickets.
type WorkflowSnapshot struct {
	Records []WorkflowRecord
}
