package app

import (
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationBudgetStatus preserves ceilings and consumed accounting
// while excluding reservation inputs, claim generations and harness identities.
type WorkflowIntegrationBudgetStatus struct {
	Revision        uint64                         `json:"revision"`
	AdmissionLimit  *int64                         `json:"admission_token_limit,omitempty"`
	TokenLimit      *int64                         `json:"token_limit,omitempty"`
	TokensSettled   int64                          `json:"tokens_settled"`
	TokensReserved  int64                          `json:"tokens_reserved"`
	TokensAvailable *int64                         `json:"tokens_available,omitempty"`
	AdmissionLimits workflowcore.IntegrationLimits `json:"admission_limits"`
	Limits          workflowcore.IntegrationLimits `json:"limits"`
	Consumed        workflowcore.IntegrationLimits `json:"consumed"`
	Remaining       workflowcore.IntegrationLimits `json:"remaining"`
}

func workflowIntegrationBudgetStatus(observed *workflowcore.IntegrationInspection, unit workflowstate.Ref) WorkflowIntegrationBudgetStatus {
	budget, usage := observed.Budget, observed.Usage
	consumed := workflowcore.IntegrationLimits{Worker: int64(usage.WorkerCount), Review: int64(usage.ReviewCount), Closing: int64(len(usage.ClosingAttempts))}

	for _, counter := range budget.UnitCounters {
		if counter.Unit == unit {
			consumed.Merger, consumed.Drift, consumed.Submission = counter.Merger, counter.Drift, counter.Submission
		}
	}

	limits := budget.LimitsSnapshot
	remaining := workflowcore.IntegrationLimits{Worker: max(0, limits.Worker-consumed.Worker), Review: max(0, limits.Review-consumed.Review), Closing: max(0, limits.Closing-consumed.Closing), Merger: max(0, limits.Merger-consumed.Merger), Drift: max(0, limits.Drift-consumed.Drift), Submission: max(0, limits.Submission-consumed.Submission)}

	result := WorkflowIntegrationBudgetStatus{Revision: budget.Revision, AdmissionLimit: budget.AdmissionTokenLimit, TokenLimit: budget.TokenLimit, TokensSettled: budget.TokensSettled, TokensReserved: budget.TokensReserved, AdmissionLimits: budget.AdmissionLimits, Limits: limits, Consumed: consumed, Remaining: remaining}
	if budget.TokenLimit != nil {
		available := max(0, *budget.TokenLimit-budget.TokensSettled-budget.TokensReserved)
		result.TokensAvailable = &available
	}

	return result
}
