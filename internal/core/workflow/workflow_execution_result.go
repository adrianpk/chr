package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) finishWorkflowTask(ctx context.Context, claim *Claim, prepared workflowPreparedTask, result *workflowstate.TaskResult, interrupted bool) error {
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

func (s *Store) checkWorkflowTaskCandidate(ctx context.Context, run *workflowInvocationFile, prepared workflowPreparedTask, result *workflowstate.TaskResult) error {
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

func applyWorkflowTaskResult(run *workflowInvocationFile, record *Record, result workflowstate.TaskResult, ref string) {
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.ResultRef, attempt.FinishedAt = ref, time.Now().UTC()
	record.Task.AttemptID = ""

	if !result.Successful {
		attempt.Outcome, run.Status = "failed", "paused"
		run.Outcome = result.Outcome
		record.Task.Phase, record.Task.Status = "pending", string(ticketcore.StatusReadyForDev)

		return
	}

	attempt.Outcome = "validated"
	record.Task.Phase, record.Task.Status = "validated", string(ticketcore.StatusInProgress)
	record.Task.Evidence = append(record.Task.Evidence, ref)
	run.Cursor++

	run.Status = "pending"
	workflowTaskBoundary(run, *record)
}

func workflowExecutionStopped(run workflowInvocationFile, interrupted bool) bool {
	return run.StopIntent != "" || interrupted || !time.Now().Before(run.Attempts[len(run.Attempts)-1].LeaseUntil)
}

func (s *Store) finishStoppedWorkflowTask(run workflowInvocationFile, record Record, claim *Claim, result *workflowstate.TaskResult) error {
	result.Successful, result.Outcome = false, "interrupted"
	if run.StopIntent != "" {
		result.Outcome = workflowstate.StopStatus(run.StopIntent)
	}

	if err := s.checkpointWorkflowAttempt(run, record, result.Outcome, result); err != nil {
		return err
	}

	return errors.Join(ErrWorkflowStop, claim.releaseReservation())
}
