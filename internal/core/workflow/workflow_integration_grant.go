package workflow

import (
	"context"
	"math"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowGrantBudget(ctx context.Context, run workflowInvocationFile, guard IntegrationBudgetGuard) (IntegrationEntry, IntegrationBudget, error) {
	if run.Integration != nil {
		return s.currentWorkflowIntegrationBudget(ctx, run, guard)
	}

	var entry IntegrationEntry
	if guard.EntryID != "" || guard.EntryRevision != 0 || guard.ClaimGeneration != 0 {
		return entry, IntegrationBudget{}, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowIntegrationRetainedRecords(ctx, run); err != nil {
		return entry, IntegrationBudget{}, err
	}

	if run.IntegrationBudget == nil {
		budget, err := s.workflowGrantBudgetAdmission(ctx, run, guard)

		return entry, budget, err
	}

	budget, err := s.readWorkflowIntegrationBudget(run)

	return entry, budget, err
}

func guardWorkflowGrant(run workflowInvocationFile, entry IntegrationEntry, budget IntegrationBudget, guard IntegrationBudgetGuard) error {
	if run.Integration != nil {
		return guardWorkflowIntegrationBudget(run, entry, budget, guard)
	}

	revision := budget.Revision
	if run.IntegrationBudget == nil {
		revision = 0
	}

	if guard.InvocationID != run.ID || guard.InvocationRevision != run.Revision || guard.BudgetRevision != revision || run.Revision == math.MaxUint64 || budget.Revision == math.MaxUint64 {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) workflowGrantBudgetAdmission(ctx context.Context, run workflowInvocationFile, guard IntegrationBudgetGuard) (IntegrationBudget, error) {
	if guard.BudgetRevision != 0 || run.Policy.AgentTokenBudget != 0 {
		return IntegrationBudget{}, workflowstate.ErrConflict
	}

	limits := workflowIntegrationAdmissionLimits(run.Policy)
	budget := IntegrationBudget{SchemaVersion: 1, InvocationID: run.ID, Revision: 1, AdmissionLimits: limits, LimitsSnapshot: limits}

	usage, err := s.workflowIntegrationUsage(ctx, run)
	if err != nil {
		return budget, err
	}

	budget.Usage = usage
	budget.UnitCounters, err = s.workflowIntegrationLegacySubmissions(ctx, run)

	return budget, err
}

func boundedWorkflowIntegrationGrant(limits IntegrationLimits) error {
	// Explicit grants may extend automatic policy ceilings, not persistence bounds
	// or the existing finite role owners' absolute admission capacities.
	if limits.Worker > workflowstate.RecordLimit || limits.Review > workflowstate.PageLimit || limits.Closing > workflowstate.PageLimit || limits.Merger > workflowstate.PageLimit || limits.Drift > workflowstate.PageLimit || limits.Submission > workflowstate.PageLimit {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func (s *Store) workflowGrantedLimit(run workflowInvocationFile, role string) (int64, error) {
	limits := workflowIntegrationAdmissionLimits(run.Policy)
	if run.IntegrationBudget != nil {
		budget, err := s.readWorkflowIntegrationBudget(run)
		if err != nil {
			return 0, err
		}

		limits = budget.LimitsSnapshot
	}

	switch role {
	case "executor":
		return limits.Worker, nil
	default:
		return 0, workflowstate.ErrConflict
	}
}

func (s *Store) checkWorkflowGrantInputs(ctx context.Context, run workflowInvocationFile) error {
	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return err
	}

	if run.StopIntent == "cancel" {
		return ErrWorkflowStop
	}

	return nil
}
