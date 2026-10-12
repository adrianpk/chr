package core

import (
	"context"
	"slices"
)

func (s *WorkflowStore) commitWorkflowProcessStart(ctx context.Context, run *workflowInvocationFile, prepared workflowPreparedTask) error {
	if prepared.input.Merger == nil {
		return s.commitWorkflowState(run)
	}

	budget, err := s.readWorkflowIntegrationBudget(*run)
	if err != nil {
		return err
	}

	input := prepared.input.Merger

	index := slices.IndexFunc(budget.Reservations, func(r WorkflowIntegrationReservation) bool {
		return r.ID == input.BudgetReservationID && r.AttemptID == prepared.input.AttemptID && r.EntryID == input.EntryID && r.ClaimGeneration == input.ClaimGeneration
	})
	if index < 0 || budget.Reservations[index].State != "reserved" {
		return ErrWorkflowConflict
	}

	// The birth identity and launch reservation reach disk before the supervisor
	// receives its permit. A lost response cannot refund an uncertain dispatch.
	budget.Reservations[index].State = "launched"

	file, err := s.workflowIntegrationBudgetChange(ctx, run, &budget)
	if err != nil {
		return err
	}

	return s.commitWorkflowPullRequestFiles(run, []workflowFileChange{file}, workflowFileChange{})
}
