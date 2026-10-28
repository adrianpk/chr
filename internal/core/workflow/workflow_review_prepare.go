package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) prepareWorkflowReview(ctx context.Context, id string, provider PullRequestProvider, reviewer []string, harness *AgentHarness) (claim *Claim, prepared workflowPreparedTask, binding workflowstate.ReviewBinding, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		restoreWorkflowCapacityWait(&invocationLocation.Run)

		if readErr := invocationLocation.Owner.checkWorkflowReviewAdmission(ctx, invocationLocation.Run); readErr != nil {
			return readErr
		}

		observation, input, readErr := invocationLocation.Owner.workflowReviewInputs(ctx, invocationLocation.Run, provider)
		if readErr != nil {
			return readErr
		}

		if readErr = invocationLocation.Owner.ensureWorkflowReviewInput(&invocationLocation.Run, observation, input); readErr != nil {
			return readErr
		}

		// Replacing review input can change task and slice wait images. A policy
		// decision must use those current images rather than the pre-journal read.
		observation, readErr = invocationLocation.Owner.observeWorkflowSlice(ctx, invocationLocation.Run)
		if readErr != nil {
			return readErr
		}

		binding = *invocationLocation.Run.Review
		if binding.State == "repair_limit_exceeded" {
			return ErrWorkflowReviewRepairBudget
		}

		if binding.State != "pending" {
			return nil
		}

		claim, prepared, readErr = invocationLocation.Owner.prepareWorkflowReviewPolicy(ctx, &invocationLocation.Run, observation, input, reviewer, harness)
		binding = *invocationLocation.Run.Review

		return readErr
	})

	return claim, prepared, binding, err
}

func (s *Store) ensureWorkflowReviewInput(run *workflowInvocationFile, observation workflowSliceObservation, input workflowstate.ReviewInput) error {
	hash, err := workflowreview.InputHash(input)
	if err != nil {
		return err
	}

	if run.Review == nil {
		return s.replaceWorkflowReviewInput(run, observation, input)
	}

	if _, _, err := s.readWorkflowReview(*run); err != nil {
		return err
	}

	if run.Review.State == "running" {
		return ErrWorkflowOwned
	}

	if hash == run.Review.InputHash && run.Review.State != "stale" {
		return nil
	}

	return s.replaceWorkflowReviewInput(run, observation, input)
}

func (s *Store) replaceWorkflowReviewInput(run *workflowInvocationFile, observation workflowSliceObservation, input workflowstate.ReviewInput) error {
	if err := s.persistWorkflowReviewInput(run, observation, input); err != nil {
		return err
	}

	var err error

	*run, err = s.readWorkflowInvocation(run.ID)

	return err
}

func (s *Store) prepareWorkflowReviewPolicy(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, input workflowstate.ReviewInput, reviewer []string, harness *AgentHarness) (*Claim, workflowPreparedTask, error) {
	switch input.Policy.Mode {
	case workflowstate.ReviewHuman:
		run.Review.State, run.Outcome = "waiting_approval", "slice_review_waiting_approval"

		return nil, workflowPreparedTask{}, s.commitWorkflowState(run)
	case workflowstate.ReviewTrustedAutoMerge:
		decision := workflowstate.ReviewDecision{BindingHash: run.Review.InputHash, Result: workflowstate.AgenticReviewResult{Decision: workflowstate.IntegrationReviewDecisionPass, Summary: "Required host validation accepted under current trusted policy."}}

		return nil, workflowPreparedTask{}, s.persistWorkflowReviewDecision(run, observation, decision, "authorized")
	case workflowstate.ReviewAgentic:
		if err := workflowstate.ValidateCommand(reviewer); err != nil {
			return nil, workflowPreparedTask{}, fmt.Errorf("separate workflow review command required: %w", err)
		}

		return s.claimWorkflowReview(ctx, run, observation, input, reviewer, harness)
	default:
		return nil, workflowPreparedTask{}, workflowstate.ErrConflict
	}
}

func (s *Store) claimWorkflowReview(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, input workflowstate.ReviewInput, reviewer []string, harness *AgentHarness) (claim *Claim, prepared workflowPreparedTask, err error) {
	claim, err = s.reserveWorkflowRole(ctx, *run, workflowSliceReviewPurpose)
	if err != nil {
		return nil, prepared, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	claim.purpose = workflowSliceReviewPurpose

	claim.Token, err = newWorkflowClaimToken()
	if err != nil {
		return nil, prepared, err
	}

	record := observation.tasks[len(observation.tasks)-1]
	if !workflowTaskClaimable(record.Task) {
		return nil, prepared, ErrWorkflowOwned
	}

	if err := s.beginWorkflowAttempt(run, &record, claim); err != nil {
		return nil, prepared, err
	}

	taskInput, err := s.workflowTaskInput(ctx, *run, record, reviewer)
	if err != nil {
		return nil, prepared, err
	}

	files, err := s.workflowClaimRecordChanges(run, &record)
	if err != nil {
		return nil, prepared, err
	}

	taskInput.SourceHash, taskInput.ExecutionMode, taskInput.Review = record.SourceHash, "automatic", &input
	taskInput.ReviewHash = run.Review.InputHash
	run.Review.State, run.Outcome = "running", "slice_review_running"

	encodedAgentInput, encodedAgentInputErr := s.encodeWorkflowAgentInput(ctx, run, taskInput, harness)

	prepared, file, budgetFile, err := encodedAgentInput.Prepared, encodedAgentInput.Input, encodedAgentInput.Budget, encodedAgentInputErr
	if err != nil {
		return nil, prepared, err
	}

	files = appendWorkflowAgentBudgetFile(files, budgetFile)

	err = s.commitWorkflowPullRequestFiles(run, files, file)

	return claim, prepared, err
}

func (s *Store) checkWorkflowReviewAdmission(ctx context.Context, run workflowInvocationFile) error {
	if !slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationWaitingReview, workflowstate.InvocationWaitingIntegration, workflowstate.InvocationBlocked}, run.Status) {
		return ErrWorkflowOwned
	}

	return s.guardWorkflowPublicationWork(ctx, run)
}
