package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (s *WorkflowStore) finishWorkflowTask(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, result *WorkflowTaskResult) error {
	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowClaim(ctx, claim)
		if err != nil {
			return err
		}

		if err := claim.store.checkWorkflowPreparedInput(prepared); err != nil {
			return err
		}

		if err := checkWorkflowExecutionBinding(run, record, prepared); err != nil {
			return err
		}

		if err := claim.store.checkWorkflowTaskCandidate(ctx, &run, prepared, result); err != nil {
			return err
		}

		ref := workflowAttemptFile(run.ID, claim.AttemptID, "result.json")
		applyWorkflowTaskResult(&run, &record, *result, ref)

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil || len(data) >= WorkflowFileLimit {
			return fmt.Errorf("workflow result exceeds file bounds: %w", errorsOrConflict(err))
		}

		data = append(data, '\n')
		run.Attempts[len(run.Attempts)-1].ResultHash = workflowHash(data)

		if err := claim.store.commitWorkflowClaim(run, record, workflowFileChange{Path: ref, After: data}); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}

func (s *WorkflowStore) checkWorkflowTaskCandidate(ctx context.Context, run *workflowInvocationFile, prepared workflowPreparedTask, result *WorkflowTaskResult) error {
	commit, tree, err := s.workflowLaneCandidate(ctx, prepared.input.Lane)
	if result.Successful && (err != nil || commit != result.Commit || tree != result.Tree) {
		return fmt.Errorf("%w: validated candidate changed before journaling", ErrWorkflowConflict)
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
