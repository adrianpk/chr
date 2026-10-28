package workflow

import (
	"context"
	"encoding/json"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationBudgetBinding seals the invocation-wide budget image.
type IntegrationBudgetBinding struct {
	Revision uint64                    `json:"revision"`
	Ref      workflowstate.EvidenceRef `json:"ref"`
}

// IntegrationLimits retain admission ceilings and positive grants.
type IntegrationLimits struct {
	Worker     int64 `json:"worker"`
	Review     int64 `json:"review"`
	Closing    int64 `json:"closing"`
	Merger     int64 `json:"merger"`
	Drift      int64 `json:"drift"`
	Submission int64 `json:"submission"`
}

// IntegrationCounter retains consumed actions, independently of candidates.
type IntegrationCounter struct {
	Unit                workflowstate.Ref `json:"unit"`
	Merger              int64             `json:"merger"`
	Drift               int64             `json:"drift"`
	Submission          int64             `json:"submission"`
	ExistingSubmissions int64             `json:"existing_submissions"`
}

// IntegrationUsage preserves the existing owners' consumed history.
type IntegrationUsage struct {
	Attempts         []string             `json:"attempts"`
	WorkerCount      int                  `json:"worker_count"`
	ReviewCount      int                  `json:"review_count"`
	PreparationCount int                  `json:"preparation_count"`
	ReviewRepairs    []ReviewRepairBudget `json:"review_repairs"`
	ClosingAttempts  []string             `json:"closing_attempts"`
}

// IntegrationReservation is retained before any possible dispatch.
// Unknown usage keeps Maximum charged; absence of a response is no refund.
type IntegrationReservation struct {
	AttemptInput    *workflowstate.EvidenceRef   `json:"attempt_input,omitempty"`
	ID              string                       `json:"id"`
	Kind            string                       `json:"kind"`
	Role            string                       `json:"role"`
	EntryID         string                       `json:"entry_id"`
	Unit            workflowstate.Ref            `json:"unit"`
	ClaimGeneration uint64                       `json:"claim_generation"`
	AttemptID       string                       `json:"attempt_id"`
	BaseCommit      string                       `json:"base_commit,omitempty"`
	Allowance       workflowstate.TokenAllowance `json:"allowance"`
	RuntimeDeadline time.Time                    `json:"runtime_deadline,omitempty"`
	State           string                       `json:"state"`
	ReportedInput   int64                        `json:"reported_input"`
	ReportedOutput  int64                        `json:"reported_output"`
	SettlementProof *workflowstate.EvidenceRef   `json:"settlement_proof"`
}

// IntegrationBudget is owned by the invocation, never by a queue index.
type IntegrationBudget struct {
	SchemaVersion       int                              `json:"schema_version"`
	InvocationID        string                           `json:"invocation_id"`
	Revision            uint64                           `json:"revision"`
	AdmissionTokenLimit *int64                           `json:"admission_token_limit"`
	TokenLimit          *int64                           `json:"token_limit"`
	TokensSettled       int64                            `json:"tokens_settled"`
	TokensReserved      int64                            `json:"tokens_reserved"`
	AdmissionLimits     IntegrationLimits                `json:"admission_limits"`
	LimitsSnapshot      IntegrationLimits                `json:"limits_snapshot"`
	Usage               IntegrationUsage                 `json:"existing_usage"`
	UnitCounters        []IntegrationCounter             `json:"unit_counters"`
	Grants              []workflowstate.IntegrationGrant `json:"grant_history"`
	Reservations        []IntegrationReservation         `json:"reservation_refs"`
	Extensions          map[string]json.RawMessage       `json:"extensions,omitempty"`
}

// IntegrationBudgetGuard fences budget callbacks by the owning claim.
type IntegrationBudgetGuard struct {
	InvocationID       string
	InvocationRevision uint64
	BudgetRevision     uint64
	EntryID            string
	EntryRevision      uint64
	ClaimGeneration    uint64
}

// IntegrationReserve records one idempotent action before launch.
type IntegrationReserve struct {
	IntegrationBudgetGuard
	Reservation IntegrationReservation
}

// IntegrationSettlement requires retained proof of reported usage.
type IntegrationSettlement struct {
	IntegrationBudgetGuard
	ReservationID string
	State         string
	Input, Output int64
	Proof         *workflowstate.EvidenceRef
}

// TokenSettlement proves compatible accounting for one call identity.
type TokenSettlement struct {
	SchemaVersion      int    `json:"schema_version"`
	ReservationID      string `json:"reservation_id"`
	AttemptID          string `json:"attempt_id"`
	CapabilityRevision string `json:"capability_revision"`
	WholeCall          bool   `json:"whole_call"`
	Input              int64  `json:"input"`
	Output             int64  `json:"output"`
}

// ReadWorkflowIntegrationBudget reads retained accounting without resetting it.
func (s *Store) ReadWorkflowIntegrationBudget(ctx context.Context, id string) (budget IntegrationBudget, err error) {
	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		run, owner, readErr := s.locateWorkflowInvocationPrepared(ctx, id, (*Store).requireNoWorkflowTransition)
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
func (s *Store) ReserveWorkflowIntegration(ctx context.Context, request IntegrationReserve) (budget IntegrationBudget, err error) {
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
func (s *Store) SettleWorkflowIntegration(ctx context.Context, request IntegrationSettlement) (budget IntegrationBudget, err error) {
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
