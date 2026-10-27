package core

import (
	"context"
	"math"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) workflowGrantBudget(ctx context.Context, run workflowInvocationFile, guard WorkflowIntegrationBudgetGuard) (WorkflowIntegrationEntry, WorkflowIntegrationBudget, error) {
	if run.Integration != nil {
		return s.currentWorkflowIntegrationBudget(ctx, run, guard)
	}

	var entry WorkflowIntegrationEntry
	if guard.EntryID != "" || guard.EntryRevision != 0 || guard.ClaimGeneration != 0 {
		return entry, WorkflowIntegrationBudget{}, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowIntegrationRetainedRecords(ctx, run); err != nil {
		return entry, WorkflowIntegrationBudget{}, err
	}

	if run.IntegrationBudget == nil {
		budget, err := s.workflowGrantBudgetAdmission(ctx, run, guard)

		return entry, budget, err
	}

	budget, err := s.readWorkflowIntegrationBudget(run)

	return entry, budget, err
}

func guardWorkflowGrant(run workflowInvocationFile, entry WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, guard WorkflowIntegrationBudgetGuard) error {
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

func (s *WorkflowStore) workflowGrantBudgetAdmission(ctx context.Context, run workflowInvocationFile, guard WorkflowIntegrationBudgetGuard) (WorkflowIntegrationBudget, error) {
	if guard.BudgetRevision != 0 || run.Policy.AgentTokenBudget != 0 {
		return WorkflowIntegrationBudget{}, workflowstate.ErrConflict
	}

	limits := workflowIntegrationAdmissionLimits(run.Policy)
	budget := WorkflowIntegrationBudget{SchemaVersion: 1, InvocationID: run.ID, Revision: 1, AdmissionLimits: limits, LimitsSnapshot: limits}

	usage, err := s.workflowIntegrationUsage(ctx, run)
	if err != nil {
		return budget, err
	}

	budget.Usage = usage
	budget.UnitCounters, err = s.workflowIntegrationLegacySubmissions(ctx, run)

	return budget, err
}

func boundedWorkflowIntegrationGrant(limits WorkflowIntegrationLimits) error {
	// Explicit grants may extend automatic policy ceilings, not persistence bounds
	// or the existing finite role owners' absolute admission capacities.
	if limits.Worker > workflowstate.WorkflowRecordLimit || limits.Review > workflowstate.WorkflowPageLimit || limits.Closing > workflowstate.WorkflowPageLimit || limits.Merger > workflowstate.WorkflowPageLimit || limits.Drift > workflowstate.WorkflowPageLimit || limits.Submission > workflowstate.WorkflowPageLimit {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func (s *WorkflowStore) workflowGrantedLimit(run workflowInvocationFile, role string) (int64, error) {
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

func (s *WorkflowStore) checkWorkflowGrantInputs(ctx context.Context, run workflowInvocationFile) error {
	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return err
	}

	if run.StopIntent == "cancel" {
		return ErrWorkflowStop
	}

	return nil
}
