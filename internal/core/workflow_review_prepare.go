package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

func (s *WorkflowStore) prepareWorkflowReview(ctx context.Context, id string, provider WorkflowPullRequestProvider, reviewer []string) (claim *WorkflowClaim, prepared workflowPreparedTask, binding WorkflowReviewBinding, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if readErr := owner.checkWorkflowReviewAdmission(ctx, run); readErr != nil {
			return readErr
		}

		observation, input, readErr := owner.workflowReviewInputs(ctx, run, provider)
		if readErr != nil {
			return readErr
		}

		if readErr = owner.ensureWorkflowReviewInput(&run, observation, input); readErr != nil {
			return readErr
		}

		// Replacing review input can change task and slice wait images. A policy
		// decision must use those current images rather than the pre-journal read.
		observation, readErr = owner.observeWorkflowSlice(ctx, run)
		if readErr != nil {
			return readErr
		}

		binding = *run.Review
		if binding.State == "repair_limit_exceeded" {
			return ErrWorkflowReviewRepairBudget
		}

		if binding.State != "pending" {
			return nil
		}

		claim, prepared, readErr = owner.prepareWorkflowReviewPolicy(ctx, &run, observation, input, reviewer)
		binding = *run.Review

		return readErr
	})

	return claim, prepared, binding, err
}

func (s *WorkflowStore) ensureWorkflowReviewInput(run *workflowInvocationFile, observation workflowSliceObservation, input WorkflowReviewInput) error {
	hash, err := workflowReviewInputHash(input)
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

func (s *WorkflowStore) replaceWorkflowReviewInput(run *workflowInvocationFile, observation workflowSliceObservation, input WorkflowReviewInput) error {
	if err := s.persistWorkflowReviewInput(run, observation, input); err != nil {
		return err
	}

	var err error

	*run, err = s.readWorkflowInvocation(run.ID)

	return err
}

func (s *WorkflowStore) prepareWorkflowReviewPolicy(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, input WorkflowReviewInput, reviewer []string) (*WorkflowClaim, workflowPreparedTask, error) {
	switch input.Policy.Mode {
	case ReviewWorkflowModeHumanReview:
		run.Review.State, run.Outcome = "waiting_approval", "slice_review_waiting_approval"

		return nil, workflowPreparedTask{}, s.commitWorkflowState(run)
	case ReviewWorkflowModeTrustedAutoMerge:
		decision := WorkflowReviewDecision{BindingHash: run.Review.InputHash, Result: AgenticReviewResult{Decision: IntegrationReviewDecisionPass, Summary: "Required host validation accepted under current trusted policy."}}

		return nil, workflowPreparedTask{}, s.persistWorkflowReviewDecision(run, observation, decision, "authorized")
	case ReviewWorkflowModeAgenticReview:
		if err := validateWorkflowCommand(reviewer); err != nil {
			return nil, workflowPreparedTask{}, fmt.Errorf("separate workflow review command required: %w", err)
		}

		return s.claimWorkflowReview(ctx, run, observation, input, reviewer)
	default:
		return nil, workflowPreparedTask{}, ErrWorkflowConflict
	}
}

func (s *WorkflowStore) claimWorkflowReview(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, input WorkflowReviewInput, reviewer []string) (claim *WorkflowClaim, prepared workflowPreparedTask, err error) {
	claim, err = s.reserveWorkflowClaim(ctx, *run)
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

	prepared, file, err := encodeWorkflowTaskInput(run, taskInput)
	if err != nil {
		return nil, prepared, err
	}

	err = s.commitWorkflowPullRequestFiles(run, files, file)

	return claim, prepared, err
}

func (s *WorkflowStore) checkWorkflowReviewAdmission(ctx context.Context, run workflowInvocationFile) error {
	if !slices.Contains([]string{"waiting_review", "waiting_integration", "blocked"}, run.Status) {
		return ErrWorkflowOwned
	}

	return s.guardWorkflowPublicationWork(ctx, run)
}
