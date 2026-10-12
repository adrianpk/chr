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
	if index < 0 {
		return ErrWorkflowConflict
	}

	if prepared.mergerProcedure {
		return s.commitWorkflowMergerProcedureStart(run, prepared, budget.Reservations[index])
	}

	if budget.Reservations[index].State != "reserved" {
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

func (s *WorkflowStore) commitWorkflowMergerProcedureStart(run *workflowInvocationFile, prepared workflowPreparedTask, reservation WorkflowIntegrationReservation) error {
	if reservation.State != "launched" {
		return ErrWorkflowConflict
	}

	processes := run.Attempts[len(run.Attempts)-1].Processes
	if len(processes) < 2 {
		return ErrWorkflowConflict
	}

	command := processes[len(processes)-1].Command

	allowed := slices.ContainsFunc(prepared.input.Merger.Conflicts, func(c WorkflowMergerConflict) bool {
		return workflowMergerProcedureRequired(c) && (slices.Equal(command, c.RegenerationCommand) || slices.Equal(command, c.ValidationCommand))
	})
	if !allowed {
		return ErrWorkflowConflict
	}
	// Host procedures retain their own birth identity under the same capacity and
	// finite deadline. They do not dispatch another model call or refill its charge.
	return s.commitWorkflowState(run)
}
