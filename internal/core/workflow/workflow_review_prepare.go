package workflow

import (
	"context"
	"errors"
	"slices"
	"time"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) prepareWorkflowReview(ctx context.Context, id string, provider PullRequestProvider, role workflowReviewRole) (result workflowReviewPreparation, err error) {
	var (
		claim    *Claim
		prepared workflowPreparedTask
		binding  workflowstate.ReviewBinding
	)

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

		reviewInputs, readErr := invocationLocation.Owner.workflowReviewInputs(ctx, invocationLocation.Run, provider)

		observation, input := reviewInputs.Observation, reviewInputs.Input

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
		if binding.State == workflowstate.ReviewRepairLimitExceeded {
			return ErrWorkflowReviewRepairBudget
		}

		if binding.State != workflowstate.ReviewPending {
			return nil
		}

		reviewClaim, readErr := invocationLocation.Owner.applyWorkflowReviewAdmission(ctx, &invocationLocation.Run, workflowReviewObservation{Observation: observation, Input: input}, role)

		claim, prepared = reviewClaim.Claim, reviewClaim.Prepared
		binding = *invocationLocation.Run.Review

		return readErr
	})

	return workflowReviewPreparation{Claim: claim, Prepared: prepared, Binding: binding}, err
}

func (s *Store) ensureWorkflowReviewInput(run *workflowInvocationFile, observation workflowSliceObservation, input workflowstate.ReviewInput) error {
	hash, err := workflowreview.InputHash(input)
	if err != nil {
		return err
	}

	if run.Review == nil {
		return s.replaceWorkflowReviewInput(run, observation, input)
	}

	if _, err := s.readWorkflowReview(*run); err != nil {
		return err
	}

	if run.Review.State == workflowstate.ReviewRunning {
		return workflowstate.ErrOwned
	}

	if hash == run.Review.InputHash && run.Review.State != workflowstate.ReviewStale {
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

func (s *Store) applyWorkflowReviewAdmission(ctx context.Context, run *workflowInvocationFile, source workflowReviewObservation, role workflowReviewRole) (workflowClaimedReview, error) {
	admission, err := workflowreview.Prepare(source.Input.Policy.Mode, run.Review, role.Command)
	if err != nil {
		return workflowClaimedReview{Claim: nil, Prepared: workflowPreparedTask{}}, err
	}

	switch admission.State {
	case workflowstate.ReviewWaitingApproval:
		run.Review.State, run.Outcome = admission.State, admission.Outcome

		return workflowClaimedReview{Claim: nil, Prepared: workflowPreparedTask{}}, s.commitWorkflowState(run)
	case workflowstate.ReviewAuthorized:
		return workflowClaimedReview{Claim: nil, Prepared: workflowPreparedTask{}}, s.persistWorkflowReviewDecision(run, source.Observation, admission.Decision, admission.State)
	case workflowstate.ReviewRunning:
		return s.claimWorkflowReview(ctx, run, source, role)
	default:
		return workflowClaimedReview{Claim: nil, Prepared: workflowPreparedTask{}}, workflowstate.ErrConflict
	}
}

func (s *Store) claimWorkflowReview(ctx context.Context, run *workflowInvocationFile, source workflowReviewObservation, role workflowReviewRole) (result workflowClaimedReview, err error) {
	observation, input := source.Observation, source.Input

	var prepared workflowPreparedTask

	claim, err := s.reserveWorkflowRole(ctx, *run, workflowSliceReviewPurpose)
	if err != nil {
		return workflowClaimedReview{Claim: nil, Prepared: prepared}, err
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
		return workflowClaimedReview{Claim: nil, Prepared: prepared}, err
	}

	record := observation.tasks[len(observation.tasks)-1]
	if !workflowTaskClaimable(record.Task) {
		return workflowClaimedReview{Claim: nil, Prepared: prepared}, workflowstate.ErrOwned
	}

	if err := s.beginWorkflowAttempt(run, &record, claim); err != nil {
		return workflowClaimedReview{Claim: nil, Prepared: prepared}, err
	}

	taskInput, err := s.workflowTaskInput(ctx, *run, record, role.Command)
	if err != nil {
		return workflowClaimedReview{Claim: nil, Prepared: prepared}, err
	}

	files, err := s.workflowClaimRecordChanges(run, &record)
	if err != nil {
		return workflowClaimedReview{Claim: nil, Prepared: prepared}, err
	}

	taskInput.SourceHash, taskInput.ExecutionMode, taskInput.Review = record.SourceHash, "automatic", &input
	taskInput.ReviewHash = run.Review.InputHash
	run.Review.State, run.Outcome = workflowstate.ReviewRunning, "slice_review_running"

	encodedAgentInput, encodedAgentInputErr := s.encodeWorkflowAgentInput(ctx, run, taskInput, role.Harness)

	prepared, file, budgetFile, err := encodedAgentInput.Prepared, encodedAgentInput.Input, encodedAgentInput.Budget, encodedAgentInputErr
	if err != nil {
		return workflowClaimedReview{Claim: nil, Prepared: prepared}, err
	}

	files = appendWorkflowAgentBudgetFile(files, budgetFile)

	err = s.commitWorkflowPullRequestFiles(run, files, file)

	return workflowClaimedReview{Claim: claim, Prepared: prepared}, err
}

func (s *Store) checkWorkflowReviewAdmission(ctx context.Context, run workflowInvocationFile) error {
	if !slices.Contains([]workflowstate.InvocationStatus{workflowstate.InvocationWaitingReview, workflowstate.InvocationWaitingIntegration, workflowstate.InvocationBlocked}, run.Status) {
		return workflowstate.ErrOwned
	}

	return s.guardWorkflowPublicationWork(ctx, run)
}
