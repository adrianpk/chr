package workflow

import (
	"context"
	"errors"
	"slices"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) commitWorkflowProcessStart(ctx context.Context, run *workflowInvocationFile, prepared workflowPreparedTask) error {
	if prepared.input.AgentCall != nil {
		return s.commitWorkflowAgentStart(ctx, run, prepared)
	}

	if prepared.input.Merger == nil {
		if err := s.checkWorkflowAgentProcessBudget(*run, prepared); err != nil {
			return err
		}

		return s.commitWorkflowState(run)
	}

	return s.commitWorkflowMergerStart(ctx, run, prepared)
}

func (s *Store) commitWorkflowMergerStart(ctx context.Context, run *workflowInvocationFile, prepared workflowPreparedTask) error {
	budget, err := s.readWorkflowIntegrationBudget(*run)
	if err != nil {
		return err
	}

	input := prepared.input.Merger

	index := slices.IndexFunc(budget.Reservations, func(r IntegrationReservation) bool {
		return r.ID == input.BudgetReservationID && r.AttemptID == prepared.input.AttemptID && r.EntryID == input.EntryID && r.ClaimGeneration == input.ClaimGeneration
	})
	if index < 0 {
		return workflowstate.ErrConflict
	}

	if prepared.mergerProcedure {
		return s.commitWorkflowMergerProcedureStart(run, prepared, budget.Reservations[index])
	}

	if budget.Reservations[index].State != "reserved" {
		return workflowstate.ErrConflict
	}

	// The birth identity and launch reservation reach disk before the supervisor
	// receives its permit. A lost response cannot refund an uncertain dispatch.
	budget.Reservations[index].State = "launched"

	file, err := s.workflowIntegrationBudgetChange(ctx, run, &budget)
	if err != nil {
		return err
	}

	return s.commitWorkflowPullRequestFiles(run, []workflowstate.FileChange{file}, workflowstate.FileChange{})
}

func (s *Store) commitWorkflowMergerProcedureStart(run *workflowInvocationFile, prepared workflowPreparedTask, reservation IntegrationReservation) error {
	if reservation.State != "launched" {
		return workflowstate.ErrConflict
	}

	processes := run.Attempts[len(run.Attempts)-1].Processes
	if len(processes) < 2 {
		return workflowstate.ErrConflict
	}

	command := processes[len(processes)-1].Command

	allowed := slices.ContainsFunc(prepared.input.Merger.Conflicts, func(c workflowstate.MergerConflict) bool {
		return workflowmerge.WorkflowMergerProcedureRequired(c) && (slices.Equal(command, c.RegenerationCommand) || slices.Equal(command, c.ValidationCommand))
	})
	if !allowed {
		return workflowstate.ErrConflict
	}
	// Host procedures retain their own birth identity under the same capacity and
	// finite deadline. They do not dispatch another model call or refill its charge.
	return s.commitWorkflowState(run)
}

func (s *Store) commitWorkflowAgentStart(ctx context.Context, run *workflowInvocationFile, prepared workflowPreparedTask) error {
	if len(run.Attempts[len(run.Attempts)-1].Processes) > 1 {
		return s.commitWorkflowState(run)
	}

	budget, err := s.readWorkflowIntegrationBudget(*run)
	if err != nil {
		return err
	}

	reservation, err := workflowAgentReservation(&budget, prepared.input)
	if err != nil || reservation.State != "reserved" {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	reservation.State = "launched"

	file, err := s.workflowIntegrationBudgetChange(ctx, run, &budget)
	if err != nil {
		return err
	}

	return s.commitWorkflowPullRequestFiles(run, []workflowstate.FileChange{file}, workflowstate.FileChange{})
}
