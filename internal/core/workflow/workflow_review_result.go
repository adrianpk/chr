package workflow

import (
	"context"
	"errors"
	"time"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) finishWorkflowReview(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider, decision workflowstate.ReviewDecision, executionErr error, interrupted bool) (result workflowstate.ReviewBinding, err error) {
	err = s.withLock(ctx, func() error {
		claimObservation, readErr := s.checkWorkflowCredential(ctx, claim, workflowCredentialIdentity)
		if readErr != nil {
			return readErr
		}

		owner := claim.store
		if readErr = owner.checkWorkflowPreparedInput(prepared); readErr != nil {
			return readErr
		}

		if readErr = checkWorkflowExecutionBinding(claimObservation.Run, claimObservation.Record, prepared); readErr != nil {
			return readErr
		}

		if workflowExecutionStopped(claimObservation.Run, interrupted) {
			readErr = owner.finishStoppedWorkflowTask(claimObservation.Run, claimObservation.Record, claim, &workflowstate.TaskResult{})

			return errors.Join(executionErr, ErrWorkflowStop, readErr)
		}

		result, readErr = owner.acceptWorkflowReview(ctx, claim, claimObservation.Run, claimObservation.Record, provider, decision, executionErr)

		return readErr
	})

	return result, err
}

func (s *Store) acceptWorkflowReview(ctx context.Context, claim *Claim, run workflowInvocationFile, record Record, provider PullRequestProvider, decision workflowstate.ReviewDecision, executionErr error) (workflowstate.ReviewBinding, error) {
	if err := workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]); err != nil {
		return workflowstate.ReviewBinding{}, err
	}

	observation, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err == nil {
		err = workflowreview.CheckWorkflowReviewInputHash(input, run.Review.InputHash)
	}

	if failure := errors.Join(executionErr, err); failure != nil {
		return workflowstate.ReviewBinding{}, s.checkpointFailedWorkflowReview(claim, run, record, failure)
	}

	state := workflowreview.WorkflowReviewDecisionState(decision.Result.Decision)
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.Outcome, attempt.FinishedAt = "reviewed", time.Now().UTC()

	if err := s.persistWorkflowReviewDecision(&run, observation, decision, state); err != nil {
		return workflowstate.ReviewBinding{}, err
	}

	return *run.Review, claim.releaseReservation()
}

func (s *Store) checkpointFailedWorkflowReview(claim *Claim, run workflowInvocationFile, record Record, failure error) error {
	run.Review.State = "stale"

	err := s.checkpointWorkflowAttempt(run, record, "interrupted", nil)
	if err == nil {
		err = claim.releaseReservation()
	}

	return errors.Join(failure, err)
}

func consumeWorkflowReviewRepair(run *workflowInvocationFile, taskKey string) error {
	if run.Review == nil || run.Review.State != "needs_repair" {
		return nil
	}

	if run.Review.RepairTaskKey != taskKey {
		return workflowstate.ErrConflict
	}

	for i := range run.ReviewRepairs {
		budget := &run.ReviewRepairs[i]
		if budget.Unit != run.Review.Unit {
			continue
		}

		if budget.Consumed >= workflowReviewRepairLimit {
			return ErrWorkflowReviewRepairBudget
		}

		budget.Consumed++
		run.Review.State = "stale"

		return nil
	}

	run.ReviewRepairs = append(run.ReviewRepairs, ReviewRepairBudget{Unit: run.Review.Unit, Consumed: 1})
	run.Review.State = "stale"

	return nil
}

func (s *Store) prepareWorkflowReviewRepair(run *workflowInvocationFile, taskKey string) error {
	if run.Review != nil {
		if _, _, err := s.readWorkflowReview(*run); err != nil {
			return err
		}
	}

	return consumeWorkflowReviewRepair(run, taskKey)
}
