package state

import "time"

// WorkflowControlIntegrationGuard is the displayed integration image. Private
// claim credentials are derived by the native owner after these fields match.
type WorkflowControlIntegrationGuard struct {
	EntryID        string
	EntryRevision  uint64
	BudgetRevision uint64
	CandidateHash  string
}

// WorkflowControlMutationRequest binds one attributable control to its
// displayed invocation and any required candidate, budget or question images.
type WorkflowControlMutationRequest struct {
	Guided           *GuidedWorkflowControlProof `json:"guided,omitempty"`
	Binding          MutationBinding
	Integration      *WorkflowControlIntegrationGuard
	Grant            *WorkflowIntegrationGrant
	Answer           *WorkflowIntegrationAnswer
	Action           string
	ID               string
	ExpectedRevision uint64
	ReviewHash       string
	DecisionRevision uint64
	DecisionHash     string
	StepLimit        int `json:",omitempty"`
}

// WorkflowControlResult is the original confirmed mutation receipt. Its revision
// describes acceptance; callers observe current workflow state separately.
type WorkflowControlResult struct {
	RequestID      string `json:"request_id"`
	Action         string `json:"action"`
	InvocationID   string `json:"invocation_id"`
	ArgumentDigest string `json:"argument_digest"`
	ReceiptRef     string `json:"receipt_ref"`
	PriorRevision  uint64 `json:"prior_revision"`
	Revision       uint64 `json:"revision"`
}

// WorkflowIntegrationGrant is an append-only, attributable ceiling extension.
type WorkflowIntegrationGrant struct {
	ID            string    `json:"id"`
	Kind          string    `json:"kind"`
	Amount        int64     `json:"amount"`
	Actor         string    `json:"actor"`
	Reason        string    `json:"reason"`
	PriorRevision uint64    `json:"prior_revision"`
	RecordedAt    time.Time `json:"recorded_at"`
}

// WorkflowIntegrationAnswer is an explicit, attributable response. Resolve
// records an interpretation; defer keeps the question waiting. Neither grants
// publication, different scope, additional calls or a replacement gate.
type WorkflowIntegrationAnswer struct {
	ID             string                `json:"id"`
	DecisionID     string                `json:"decision_id"`
	Choice         string                `json:"choice"`
	Action         string                `json:"action"`
	Actor          string                `json:"actor"`
	Authority      string                `json:"authority"`
	Reason         string                `json:"reason"`
	PriorRevision  uint64                `json:"prior_revision"`
	QuestionHash   string                `json:"question_hash"`
	RecordedAt     time.Time             `json:"recorded_at"`
	OriginalCommit string                `json:"original_commit,omitempty"`
	IntegrationFix *AgenticReviewFinding `json:"integration_fix,omitempty"`
}

// GuidedWorkflowControlProof binds a native control to retained guided authority.
type GuidedWorkflowControlProof struct {
	RunID, OperationDigest string
}

// AgenticReviewFinding describes one blocking or non-blocking review note.
type AgenticReviewFinding struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}
