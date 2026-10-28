package workflow

import (
	"context"
	"errors"
	"time"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) finishWorkflowReview(ctx context.Context, operation workflowReviewDecisions, completion workflowreview.Completion) (result workflowstate.ReviewBinding, err error) {
	claim, prepared := operation.claim, operation.prepared
	executionErr, interrupted := completion.Failure, completion.Interrupted

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

		result, readErr = owner.acceptWorkflowReview(ctx, claimObservation, operation, completion)

		return readErr
	})

	return result, err
}

func (s *Store) acceptWorkflowReview(ctx context.Context, claimed workflowClaimObservation, operation workflowReviewDecisions, completion workflowreview.Completion) (workflowstate.ReviewBinding, error) {
	claim, run, record := operation.claim, claimed.Run, claimed.Record
	provider, decision, executionErr := operation.provider, completion.Decision, completion.Failure

	if err := workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]); err != nil {
		return workflowstate.ReviewBinding{}, err
	}

	observation, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err == nil {
		err = workflowreview.CheckInputHash(input, run.Review.InputHash)
	}

	if failure := errors.Join(executionErr, err); failure != nil {
		return workflowstate.ReviewBinding{}, s.checkpointFailedWorkflowReview(claim, run, record, failure)
	}

	state := workflowreview.DecisionState(decision.Result.Decision)
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
