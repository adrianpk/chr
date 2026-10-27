package core

import (
	"context"
	"encoding/json"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationBudgetBinding seals the invocation-wide budget image.
type WorkflowIntegrationBudgetBinding struct {
	Revision uint64              `json:"revision"`
	Ref      WorkflowEvidenceRef `json:"ref"`
}

// WorkflowIntegrationLimits retain admission ceilings and positive grants.
type WorkflowIntegrationLimits struct {
	Worker     int64 `json:"worker"`
	Review     int64 `json:"review"`
	Closing    int64 `json:"closing"`
	Merger     int64 `json:"merger"`
	Drift      int64 `json:"drift"`
	Submission int64 `json:"submission"`
}

// WorkflowIntegrationCounter retains consumed actions, independently of candidates.
type WorkflowIntegrationCounter struct {
	Unit                workflowstate.Ref `json:"unit"`
	Merger              int64             `json:"merger"`
	Drift               int64             `json:"drift"`
	Submission          int64             `json:"submission"`
	ExistingSubmissions int64             `json:"existing_submissions"`
}

// WorkflowIntegrationUsage preserves the existing owners' consumed history.
type WorkflowIntegrationUsage struct {
	Attempts         []string                     `json:"attempts"`
	WorkerCount      int                          `json:"worker_count"`
	ReviewCount      int                          `json:"review_count"`
	PreparationCount int                          `json:"preparation_count"`
	ReviewRepairs    []WorkflowReviewRepairBudget `json:"review_repairs"`
	ClosingAttempts  []string                     `json:"closing_attempts"`
}

// WorkflowTokenAllowance describes whole-call accounting, including input.
// Reported output already includes reasoning; cached input remains input.
type WorkflowTokenAllowance struct {
	Maximum            int64     `json:"maximum"`
	Input              int64     `json:"input"`
	Output             int64     `json:"output"`
	CapabilityRevision string    `json:"capability_revision"`
	WholeCall          bool      `json:"whole_call"`
	Deadline           time.Time `json:"deadline"`
}

// WorkflowIntegrationReservation is retained before any possible dispatch.
// Unknown usage keeps Maximum charged; absence of a response is no refund.
type WorkflowIntegrationReservation struct {
	AttemptInput    *WorkflowEvidenceRef   `json:"attempt_input,omitempty"`
	ID              string                 `json:"id"`
	Kind            string                 `json:"kind"`
	Role            string                 `json:"role"`
	EntryID         string                 `json:"entry_id"`
	Unit            workflowstate.Ref      `json:"unit"`
	ClaimGeneration uint64                 `json:"claim_generation"`
	AttemptID       string                 `json:"attempt_id"`
	BaseCommit      string                 `json:"base_commit,omitempty"`
	Allowance       WorkflowTokenAllowance `json:"allowance"`
	RuntimeDeadline time.Time              `json:"runtime_deadline,omitempty"`
	State           string                 `json:"state"`
	ReportedInput   int64                  `json:"reported_input"`
	ReportedOutput  int64                  `json:"reported_output"`
	SettlementProof *WorkflowEvidenceRef   `json:"settlement_proof"`
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

// WorkflowIntegrationBudget is owned by the invocation, never by a queue index.
type WorkflowIntegrationBudget struct {
	SchemaVersion       int                              `json:"schema_version"`
	InvocationID        string                           `json:"invocation_id"`
	Revision            uint64                           `json:"revision"`
	AdmissionTokenLimit *int64                           `json:"admission_token_limit"`
	TokenLimit          *int64                           `json:"token_limit"`
	TokensSettled       int64                            `json:"tokens_settled"`
	TokensReserved      int64                            `json:"tokens_reserved"`
	AdmissionLimits     WorkflowIntegrationLimits        `json:"admission_limits"`
	LimitsSnapshot      WorkflowIntegrationLimits        `json:"limits_snapshot"`
	Usage               WorkflowIntegrationUsage         `json:"existing_usage"`
	UnitCounters        []WorkflowIntegrationCounter     `json:"unit_counters"`
	Grants              []WorkflowIntegrationGrant       `json:"grant_history"`
	Reservations        []WorkflowIntegrationReservation `json:"reservation_refs"`
	Extensions          map[string]json.RawMessage       `json:"extensions,omitempty"`
}

// WorkflowIntegrationBudgetGuard fences budget callbacks by the owning claim.
type WorkflowIntegrationBudgetGuard struct {
	InvocationID       string
	InvocationRevision uint64
	BudgetRevision     uint64
	EntryID            string
	EntryRevision      uint64
	ClaimGeneration    uint64
}

// WorkflowIntegrationReserve records one idempotent action before launch.
type WorkflowIntegrationReserve struct {
	WorkflowIntegrationBudgetGuard
	Reservation WorkflowIntegrationReservation
}

// WorkflowIntegrationSettlement requires retained proof of reported usage.
type WorkflowIntegrationSettlement struct {
	WorkflowIntegrationBudgetGuard
	ReservationID string
	State         string
	Input, Output int64
	Proof         *WorkflowEvidenceRef
}

// WorkflowTokenSettlement proves compatible accounting for one call identity.
type WorkflowTokenSettlement struct {
	SchemaVersion      int    `json:"schema_version"`
	ReservationID      string `json:"reservation_id"`
	AttemptID          string `json:"attempt_id"`
	CapabilityRevision string `json:"capability_revision"`
	WholeCall          bool   `json:"whole_call"`
	Input              int64  `json:"input"`
	Output             int64  `json:"output"`
}

// ReadWorkflowIntegrationBudget reads retained accounting without resetting it.
func (s *WorkflowStore) ReadWorkflowIntegrationBudget(ctx context.Context, id string) (budget WorkflowIntegrationBudget, err error) {
	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		run, owner, readErr := s.locateWorkflowInvocationPrepared(ctx, id, (*WorkflowStore).requireNoWorkflowTransition)
		if readErr != nil {
			return readErr
		}

		if recordErr := owner.checkWorkflowIntegrationRetainedRecords(ctx, run); recordErr != nil {
			return recordErr
		}

		budget, readErr = owner.readWorkflowIntegrationBudget(run)

		return readErr
	})

	return budget, err
}

// ReserveWorkflowIntegration durably charges a call or action exactly once.
func (s *WorkflowStore) ReserveWorkflowIntegration(ctx context.Context, request WorkflowIntegrationReserve) (budget WorkflowIntegrationBudget, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		budget, readErr = owner.reserveWorkflowIntegration(ctx, run, request)

		return readErr
	})

	return budget, err
}

// SettleWorkflowIntegration never releases possibly launched or unknown calls.
func (s *WorkflowStore) SettleWorkflowIntegration(ctx context.Context, request WorkflowIntegrationSettlement) (budget WorkflowIntegrationBudget, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		budget, readErr = owner.settleWorkflowIntegration(ctx, run, request)

		return readErr
	})

	return budget, err
}
