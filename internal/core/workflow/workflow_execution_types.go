package workflow

import (
	"sync"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Lane retains one slice or independent ticket's cumulative Git checkout.
type Lane struct {
	Unit       workflowstate.Ref `json:"unit"`
	Branch     string            `json:"branch"`
	Worktree   string            `json:"worktree"`
	BaseCommit string            `json:"base_commit"`
	HeadCommit string            `json:"head_commit"`
}

// TaskInput is the immutable, credential-free input supplied to a worker.
type TaskInput struct {
	HarnessBinding           *harnesscore.Binding                 `json:"harness_binding,omitempty"`
	OrderMap                 *OrderMapInput                       `json:"order_map,omitempty"`
	AgentCall                *AgentCall                           `json:"agent_call,omitempty"`
	Merger                   *MergerInput                         `json:"merger,omitempty"`
	MergerHarnessRef         *workflowstate.EvidenceRef           `json:"merger_harness_ref,omitempty"`
	ValidationTimeoutSeconds int                                  `json:"validation_timeout_seconds,omitempty"`
	Preparation              *PreparationInput                    `json:"preparation,omitempty"`
	Review                   *ReviewInput                         `json:"review,omitempty"`
	ReviewHash               string                               `json:"review_hash,omitempty"`
	Instructions             string                               `json:"instructions,omitempty"`
	Schema                   string                               `json:"schema"`
	ExecutionMode            string                               `json:"execution_mode"`
	InvocationID             string                               `json:"invocation_id"`
	AttemptID                string                               `json:"attempt_id"`
	Generation               uint64                               `json:"generation"`
	TaskKey                  string                               `json:"ticket_key"`
	InputRevision            uint64                               `json:"input_revision"`
	SourceHash               string                               `json:"source_hash"`
	ContractHash             string                               `json:"execution_contract_hash"`
	ScopeHash                string                               `json:"scope_hash"`
	Contract                 generationcore.GeneratedWorkflowTask `json:"execution_contract"`
	Body                     string                               `json:"body"`
	Target                   workflowstate.PreviewTarget          `json:"target"`
	Lane                     Lane                                 `json:"lane"`
	Prerequisites            []InputImage                         `json:"prerequisites"`
	EvidenceHashes           map[string]string                    `json:"prerequisite_evidence_hashes"`
	Worker                   []string                             `json:"worker"`
	Slice                    *SliceInput                          `json:"slice,omitempty"`
}

// CommandResult records host-observed command output with explicit bounds.
type CommandResult struct {
	Command   []string `json:"command"`
	ExitCode  int      `json:"exit_code"`
	Stdout    string   `json:"stdout"`
	Stderr    string   `json:"stderr"`
	Truncated bool     `json:"truncated"`
}

// TaskResult binds task validation to the frozen input and actual Git tree.
type TaskResult struct {
	TaskValidation
	Outcome            string          `json:"outcome"`
	ExecutionMode      string          `json:"execution_mode"`
	IntegrationPending bool            `json:"integration_pending"`
	AttemptID          string          `json:"attempt_id"`
	Generation         uint64          `json:"generation"`
	InputRef           string          `json:"input_ref"`
	InputHash          string          `json:"input_hash"`
	Worker             CommandResult   `json:"worker_result"`
	Validation         []CommandResult `json:"validation"`
	WaiverReason       string          `json:"waiver_reason,omitempty"`
	Failure            string          `json:"failure,omitempty"`
	FinishedAt         time.Time       `json:"finished_at"`
}

type workflowClaimExecution struct {
	sync.Mutex
	running  bool
	groupPID int
}

type workflowPreparedTask struct {
	mergerProcedure bool
	input           TaskInput
	inputRef        string
	inputHash       string
	data            []byte
	leaseUntil      time.Time
	environment     []string
	directory       string
}
