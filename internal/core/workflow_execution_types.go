package core

import (
	"sync"
	"time"
)

// WorkflowLane retains one slice or independent ticket's cumulative Git checkout.
type WorkflowLane struct {
	Unit       WorkflowRef `json:"unit"`
	Branch     string      `json:"branch"`
	Worktree   string      `json:"worktree"`
	BaseCommit string      `json:"base_commit"`
	HeadCommit string      `json:"head_commit"`
}

// WorkflowTaskInput is the immutable, credential-free input supplied to a worker.
type WorkflowTaskInput struct {
	Schema         string                `json:"schema"`
	InvocationID   string                `json:"invocation_id"`
	AttemptID      string                `json:"attempt_id"`
	Generation     uint64                `json:"generation"`
	TaskKey        string                `json:"ticket_key"`
	InputRevision  uint64                `json:"input_revision"`
	SourceHash     string                `json:"source_hash"`
	ContractHash   string                `json:"execution_contract_hash"`
	ScopeHash      string                `json:"scope_hash"`
	Contract       GeneratedWorkflowTask `json:"execution_contract"`
	Body           string                `json:"body"`
	Target         WorkflowPreviewTarget `json:"target"`
	Lane           WorkflowLane          `json:"lane"`
	Prerequisites  []WorkflowInputImage  `json:"prerequisites"`
	EvidenceHashes map[string]string     `json:"prerequisite_evidence_hashes"`
	Worker         []string              `json:"worker"`
}

// WorkflowCommandResult records host-observed command output with explicit bounds.
type WorkflowCommandResult struct {
	Command   []string `json:"command"`
	ExitCode  int      `json:"exit_code"`
	Stdout    string   `json:"stdout"`
	Stderr    string   `json:"stderr"`
	Truncated bool     `json:"truncated"`
}

// WorkflowTaskResult binds task validation to the frozen input and actual Git tree.
type WorkflowTaskResult struct {
	WorkflowTaskValidation
	Outcome            string                  `json:"outcome"`
	IntegrationPending bool                    `json:"integration_pending"`
	AttemptID          string                  `json:"attempt_id"`
	Generation         uint64                  `json:"generation"`
	InputRef           string                  `json:"input_ref"`
	InputHash          string                  `json:"input_hash"`
	Worker             WorkflowCommandResult   `json:"worker_result"`
	Validation         []WorkflowCommandResult `json:"validation"`
	WaiverReason       string                  `json:"waiver_reason,omitempty"`
	Failure            string                  `json:"failure,omitempty"`
	FinishedAt         time.Time               `json:"finished_at"`
}

type workflowClaimExecution struct {
	sync.Mutex
	running  bool
	groupPID int
}

type workflowPreparedTask struct {
	input      WorkflowTaskInput
	inputRef   string
	inputHash  string
	data       []byte
	leaseUntil time.Time
}
