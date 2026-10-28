package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowSliceRepairPurpose = "slice_repair"

// RepairWorkflowSlice invalidates the affected task and its ordered successors.
// Historical results remain immutable; unchanged earlier task proof is reusable.
func (s *Store) RepairWorkflowSlice(ctx context.Context, id, taskKey string) (result Invocation, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		restoreWorkflowCapacityWait(&invocationLocation.Run)

		if invocationLocation.Run.CandidateRepair != nil {
			return workflowstate.ErrConflict
		}

		if !workflowSliceRepairReady(invocationLocation.Run.Invocation) || invocationLocation.Run.StopIntent != "" {
			return ErrWorkflowOwned
		}

		if readErr := invocationLocation.Owner.guardWorkflowPublicationWork(ctx, invocationLocation.Run); readErr != nil {
			return readErr
		}

		if readErr := invocationLocation.Owner.checkWorkflowSliceRepairPR(invocationLocation.Run); readErr != nil {
			return readErr
		}

		observation, readErr := invocationLocation.Owner.readWorkflowSliceMembership(ctx, invocationLocation.Run)
		if readErr != nil {
			return readErr
		}

		readErr = invocationLocation.Owner.repairWorkflowSlice(ctx, &invocationLocation.Run, observation, taskKey)
		if errors.Is(readErr, ErrWorkflowReviewRepairBudget) {
			invocationLocation.Run.Review.State = workflowstate.ReviewRepairLimitExceeded
			invocationLocation.Run.Status, invocationLocation.Run.Outcome = workflowstate.InvocationWaitingReview, "review_repair_limit_exceeded"
			invocationLocation.Run.Blocker = &Blocker{Dependency: invocationLocation.Run.Review.Unit, Reason: invocationLocation.Run.Outcome, NextActionOwner: "workflow_operator"}
			readErr = errors.Join(readErr, invocationLocation.Owner.commitWorkflowState(&invocationLocation.Run))
		}

		result = invocationLocation.Run.Invocation

		return readErr
	})

	return result, err
}

func (s *Store) checkWorkflowSliceRepairPR(run workflowInvocationFile) error {
	if run.PullRequest == nil || run.PullRequest.ObservationRef == "" {
		return nil
	}

	observation, err := s.readWorkflowPullRequestObservation(*run.PullRequest)
	if err != nil {
		return err
	}

	for _, pr := range observation.Snapshots {
		if pr.Merged {
			return fmt.Errorf("%w: observed delivery requires verification before repair", ErrWorkflowOwned)
		}
	}

	return nil
}

func (s *Store) repairWorkflowSlice(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, taskKey string) (err error) {
	index := slices.IndexFunc(observation.tasks, func(record workflowstate.Record) bool { return record.Task.Key == taskKey })
	if index < 0 {
		return fmt.Errorf("repair task is outside the retained slice")
	}

	if err := s.prepareWorkflowReviewRepair(run, taskKey); err != nil {
		return err
	}

	for _, record := range observation.tasks {
		if !workflowTaskClaimable(record.Task) {
			return ErrWorkflowOwned
		}
	}

	claim, err := s.reserveWorkflowRole(ctx, *run, workflowSliceRepairPurpose)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, claim.releaseReservation()) }()

	laneImage, err := s.workflowLaneCandidate(ctx, *run.Lane)
	if err != nil {
		return err
	}

	if err := s.beginWorkflowSliceRepair(run, observation, claim); err != nil {
		return err
	}

	lane := *run.Lane
	lane.HeadCommit = laneImage.Commit
	run.Lane = &lane
	run.Cursor -= len(observation.tasks) - index
	run.Status, run.Outcome, run.ResumeStatus = workflowstate.InvocationPending, "slice_repair_requested", workflowstate.InvocationUnspecified

	run.CandidateRef, run.CandidateHash = "", ""
	if run.Preparation != nil {
		run.Preparation.State = "stale"
	}

	if run.Review != nil {
		run.Review.State = workflowstate.ReviewStale
	}

	return s.commitWorkflowSliceRepair(run, observation, index)
}

func workflowSliceRepairReady(run Invocation) bool {
	if run.PullRequest != nil && !slices.Contains([]string{"prepared", "bound"}, run.PullRequest.State) {
		return false
	}

	return workflowSliceValidationReady(run) || run.Status == workflowstate.InvocationWaitingReview
}

func (s *Store) commitWorkflowSliceRepair(run *workflowInvocationFile, observation workflowSliceObservation, index int) error {
	files := []workflowstate.FileChange{}

	for i, record := range observation.tasks {
		if i < index {
			if record.Task.Phase == "validated" {
				continue
			}

			record.Task.Phase, record.Task.Status = "validated", string(ticketcore.StatusInProgress)
		} else {
			record.Task.Phase, record.Task.Status, record.Task.Evidence, record.Task.AttemptID = "pending", string(ticketcore.StatusReadyForDev), nil, ""
		}

		changes, err := s.workflowClaimRecordChanges(run, &record)
		if err != nil {
			return err
		}

		files = append(files, changes...)
	}

	changes, err := s.workflowRepairSetChanges(run, observation.set)
	if err != nil {
		return err
	}

	files = append(files, changes...)

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	return s.commitFiles(append(files, state))
}

func (s *Store) workflowRepairSetChanges(run *workflowInvocationFile, set workflowstate.Record) ([]workflowstate.FileChange, error) {
	if set.Set == nil {
		return nil, nil
	}

	for i := range set.Set.Slices {
		descriptor := &set.Set.Slices[i]
		if descriptor.ID != run.Lane.Unit.ID || (descriptor.CandidateRef == "" && descriptor.PullRequest == "") {
			continue
		}

		descriptor.CandidateRef, descriptor.Phase = "", "developing"

		return s.workflowSetRecordChanges(run, &set)
	}

	return nil, nil
}

func (s *Store) beginWorkflowSliceRepair(run *workflowInvocationFile, observation workflowSliceObservation, claim *Claim) error {
	claim.purpose = workflowSliceRepairPurpose

	var err error

	claim.Token, err = newWorkflowClaimToken()
	if err != nil {
		return err
	}

	anchor := &observation.tasks[len(observation.tasks)-1]
	if err := s.beginWorkflowAttempt(run, anchor, claim); err != nil {
		return err
	}

	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.ContractHash = workflowUnitContractHash(observation)
	attempt.Outcome, attempt.FinishedAt = "repaired", time.Now().UTC()

	return nil
}
