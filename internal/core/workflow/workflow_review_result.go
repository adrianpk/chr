package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) finishWorkflowReview(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider, decision ReviewDecision, executionErr error, interrupted bool) (result ReviewBinding, err error) {
	err = s.withLock(ctx, func() error {
		run, record, readErr := s.checkWorkflowCredential(ctx, claim, false)
		if readErr != nil {
			return readErr
		}

		owner := claim.store
		if readErr = owner.checkWorkflowPreparedInput(prepared); readErr != nil {
			return readErr
		}

		if readErr = checkWorkflowExecutionBinding(run, record, prepared); readErr != nil {
			return readErr
		}

		if workflowExecutionStopped(run, interrupted) {
			readErr = owner.finishStoppedWorkflowTask(run, record, claim, &TaskResult{})

			return errors.Join(executionErr, ErrWorkflowStop, readErr)
		}

		result, readErr = owner.acceptWorkflowReview(ctx, claim, run, record, provider, decision, executionErr)

		return readErr
	})

	return result, err
}

func (s *Store) acceptWorkflowReview(ctx context.Context, claim *Claim, run workflowInvocationFile, record Record, provider PullRequestProvider, decision ReviewDecision, executionErr error) (ReviewBinding, error) {
	if err := workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]); err != nil {
		return ReviewBinding{}, err
	}

	observation, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err == nil {
		err = checkWorkflowReviewInputHash(input, run.Review.InputHash)
	}

	if failure := errors.Join(executionErr, err); failure != nil {
		return ReviewBinding{}, s.checkpointFailedWorkflowReview(claim, run, record, failure)
	}

	state := workflowReviewDecisionState(decision.Result.Decision)
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.Outcome, attempt.FinishedAt = "reviewed", time.Now().UTC()

	if err := s.persistWorkflowReviewDecision(&run, observation, decision, state); err != nil {
		return ReviewBinding{}, err
	}

	return *run.Review, claim.releaseReservation()
}

func checkWorkflowReviewInputHash(input ReviewInput, expected string) error {
	hash, err := workflowReviewInputHash(input)
	if err != nil || hash != expected {
		return fmt.Errorf("%w: review inputs changed during execution", workflowstate.ErrConflict)
	}

	return nil
}

func (s *Store) checkpointFailedWorkflowReview(claim *Claim, run workflowInvocationFile, record Record, failure error) error {
	run.Review.State = "stale"

	err := s.checkpointWorkflowAttempt(run, record, "interrupted", nil)
	if err == nil {
		err = claim.releaseReservation()
	}

	return errors.Join(failure, err)
}

func workflowReviewDecisionState(decision workflowstate.IntegrationReviewDecision) string {
	switch decision {
	case workflowstate.IntegrationReviewDecisionPass:
		return "authorized"
	case workflowstate.IntegrationReviewDecisionNeedsChanges:
		return "needs_repair"
	case workflowstate.IntegrationReviewDecisionHumanRequired:
		return "human_required"
	default:
		return "insufficient_evidence"
	}
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
