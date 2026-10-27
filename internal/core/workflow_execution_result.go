package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) finishWorkflowTask(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, result *WorkflowTaskResult, interrupted bool) error {
	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowCredential(ctx, claim, false)
		if err != nil {
			return err
		}

		if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
			return err
		}

		if err := checkWorkflowExecutionBinding(run, record, prepared); err != nil {
			return err
		}

		if workflowExecutionStopped(run, interrupted) {
			return claim.store.finishStoppedWorkflowTask(run, record, claim, result)
		}

		if err := claim.store.checkWorkflowTaskCandidate(ctx, &run, prepared, result); err != nil {
			return err
		}

		ref := workflowAttemptFile(run.ID, claim.AttemptID, "result.json")
		applyWorkflowTaskResult(&run, &record, *result, ref)

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil || len(data) >= workflowstate.FileLimit {
			return fmt.Errorf("workflow result exceeds file bounds: %w", errorsOrConflict(err))
		}

		data = append(data, '\n')
		run.Attempts[len(run.Attempts)-1].ResultHash = contenthash.Sum(data)

		if err := claim.store.commitWorkflowClaim(run, record, workflowstate.FileChange{Path: ref, After: data}); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}

func (s *WorkflowStore) checkWorkflowTaskCandidate(ctx context.Context, run *workflowInvocationFile, prepared workflowPreparedTask, result *WorkflowTaskResult) error {
	commit, tree, err := s.workflowLaneCandidate(ctx, prepared.input.Lane)
	if result.Successful && (err != nil || commit != result.Commit || tree != result.Tree) {
		return fmt.Errorf("%w: validated candidate changed before journaling", workflowstate.ErrConflict)
	}

	if err == nil {
		if result.Commit == "" {
			result.Commit, result.Tree = commit, tree
		}

		if result.Commit == commit && result.Tree == tree {
			lane := prepared.input.Lane
			lane.HeadCommit = commit
			run.Lane = &lane
		}
	}

	return ctx.Err()
}

func applyWorkflowTaskResult(run *workflowInvocationFile, record *WorkflowRecord, result WorkflowTaskResult, ref string) {
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.ResultRef, attempt.FinishedAt = ref, time.Now().UTC()
	record.Task.AttemptID = ""

	if !result.Successful {
		attempt.Outcome, run.Status = "failed", "paused"
		run.Outcome = result.Outcome
		record.Task.Phase, record.Task.Status = "pending", TicketStatusReadyForDev

		return
	}

	attempt.Outcome = "validated"
	record.Task.Phase, record.Task.Status = "validated", TicketStatusInProgress
	record.Task.Evidence = append(record.Task.Evidence, ref)
	run.Cursor++

	run.Status = "pending"
	workflowTaskBoundary(run, *record)
}

func workflowExecutionStopped(run workflowInvocationFile, interrupted bool) bool {
	return run.StopIntent != "" || interrupted || !time.Now().Before(run.Attempts[len(run.Attempts)-1].LeaseUntil)
}

func (s *WorkflowStore) finishStoppedWorkflowTask(run workflowInvocationFile, record WorkflowRecord, claim *WorkflowClaim, result *WorkflowTaskResult) error {
	result.Successful, result.Outcome = false, "interrupted"
	if run.StopIntent != "" {
		result.Outcome = workflowStopStatus(run.StopIntent)
	}

	if err := s.checkpointWorkflowAttempt(run, record, result.Outcome, result); err != nil {
		return err
	}

	return errors.Join(ErrWorkflowStop, claim.releaseReservation())
}
