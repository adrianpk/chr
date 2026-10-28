package state

import (
	"time"
)

// Lane retains one slice or independent ticket's cumulative Git checkout.
type Lane struct {
	Unit       Ref    `json:"unit"`
	Branch     string `json:"branch"`
	Worktree   string `json:"worktree"`
	BaseCommit string `json:"base_commit"`
	HeadCommit string `json:"head_commit"`
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
