package workflow

import (
	"sync"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TaskInput is the immutable, credential-free input supplied to a worker.
type TaskInput struct {
	HarnessBinding           *harnesscore.Binding                 `json:"harness_binding,omitempty"`
	OrderMap                 *OrderMapInput                       `json:"order_map,omitempty"`
	AgentCall                *AgentCall                           `json:"agent_call,omitempty"`
	Merger                   *workflowstate.MergerInput           `json:"merger,omitempty"`
	MergerHarnessRef         *workflowstate.EvidenceRef           `json:"merger_harness_ref,omitempty"`
	ValidationTimeoutSeconds int                                  `json:"validation_timeout_seconds,omitempty"`
	Preparation              *workflowstate.PreparationInput      `json:"preparation,omitempty"`
	Review                   *workflowstate.ReviewInput           `json:"review,omitempty"`
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
	Lane                     workflowstate.Lane                   `json:"lane"`
	Prerequisites            []workflowstate.InputImage           `json:"prerequisites"`
	EvidenceHashes           map[string]string                    `json:"prerequisite_evidence_hashes"`
	Worker                   []string                             `json:"worker"`
	Slice                    *workflowstate.SliceInput            `json:"slice,omitempty"`
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
