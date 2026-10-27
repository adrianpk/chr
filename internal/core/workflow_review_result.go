package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) finishWorkflowReview(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, provider WorkflowPullRequestProvider, decision WorkflowReviewDecision, executionErr error, interrupted bool) (result WorkflowReviewBinding, err error) {
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
			readErr = owner.finishStoppedWorkflowTask(run, record, claim, &WorkflowTaskResult{})

			return errors.Join(executionErr, ErrWorkflowStop, readErr)
		}

		result, readErr = owner.acceptWorkflowReview(ctx, claim, run, record, provider, decision, executionErr)

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) acceptWorkflowReview(ctx context.Context, claim *WorkflowClaim, run workflowInvocationFile, record WorkflowRecord, provider WorkflowPullRequestProvider, decision WorkflowReviewDecision, executionErr error) (WorkflowReviewBinding, error) {
	if err := workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]); err != nil {
		return WorkflowReviewBinding{}, err
	}

	observation, input, err := s.workflowReviewInputs(ctx, run, provider)
	if err == nil {
		err = checkWorkflowReviewInputHash(input, run.Review.InputHash)
	}

	if failure := errors.Join(executionErr, err); failure != nil {
		return WorkflowReviewBinding{}, s.checkpointFailedWorkflowReview(claim, run, record, failure)
	}

	state := workflowReviewDecisionState(decision.Result.Decision)
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.Outcome, attempt.FinishedAt = "reviewed", time.Now().UTC()

	if err := s.persistWorkflowReviewDecision(&run, observation, decision, state); err != nil {
		return WorkflowReviewBinding{}, err
	}

	return *run.Review, claim.releaseReservation()
}

func checkWorkflowReviewInputHash(input WorkflowReviewInput, expected string) error {
	hash, err := workflowReviewInputHash(input)
	if err != nil || hash != expected {
		return fmt.Errorf("%w: review inputs changed during execution", workflowstate.ErrConflict)
	}

	return nil
}

func (s *WorkflowStore) checkpointFailedWorkflowReview(claim *WorkflowClaim, run workflowInvocationFile, record WorkflowRecord, failure error) error {
	run.Review.State = "stale"

	err := s.checkpointWorkflowAttempt(run, record, "interrupted", nil)
	if err == nil {
		err = claim.releaseReservation()
	}

	return errors.Join(failure, err)
}

func workflowReviewDecisionState(decision IntegrationReviewDecision) string {
	switch decision {
	case IntegrationReviewDecisionPass:
		return "authorized"
	case IntegrationReviewDecisionNeedsChanges:
		return "needs_repair"
	case IntegrationReviewDecisionHumanRequired:
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

	run.ReviewRepairs = append(run.ReviewRepairs, WorkflowReviewRepairBudget{Unit: run.Review.Unit, Consumed: 1})
	run.Review.State = "stale"

	return nil
}

func (s *WorkflowStore) prepareWorkflowReviewRepair(run *workflowInvocationFile, taskKey string) error {
	if run.Review != nil {
		if _, _, err := s.readWorkflowReview(*run); err != nil {
			return err
		}
	}

	return consumeWorkflowReviewRepair(run, taskKey)
}
