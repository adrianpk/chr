package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

const workflowSliceRepairPurpose = "slice_repair"

// RepairWorkflowSlice invalidates the affected task and its ordered successors.
// Historical results remain immutable; unchanged earlier task proof is reusable.
func (s *WorkflowStore) RepairWorkflowSlice(ctx context.Context, id, taskKey string) (result WorkflowInvocation, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		restoreWorkflowCapacityWait(&run)

		if !workflowSliceRepairReady(run.WorkflowInvocation) || run.StopIntent != "" {
			return ErrWorkflowOwned
		}

		if readErr := owner.guardWorkflowPublicationWork(ctx, run); readErr != nil {
			return readErr
		}

		if readErr := owner.checkWorkflowSliceRepairPR(run); readErr != nil {
			return readErr
		}

		observation, readErr := owner.readWorkflowSliceMembership(ctx, run)
		if readErr != nil {
			return readErr
		}

		readErr = owner.repairWorkflowSlice(ctx, &run, observation, taskKey)
		if errors.Is(readErr, ErrWorkflowReviewRepairBudget) {
			run.Review.State = "repair_limit_exceeded"
			run.Status, run.Outcome = "waiting_review", "review_repair_limit_exceeded"
			run.Blocker = &WorkflowBlocker{Dependency: run.Review.Unit, Reason: run.Outcome, NextActionOwner: "workflow_operator"}
			readErr = errors.Join(readErr, owner.commitWorkflowState(&run))
		}

		result = run.WorkflowInvocation

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) checkWorkflowSliceRepairPR(run workflowInvocationFile) error {
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

func (s *WorkflowStore) repairWorkflowSlice(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, taskKey string) (err error) {
	index := slices.IndexFunc(observation.tasks, func(record WorkflowRecord) bool { return record.Task.Key == taskKey })
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

	claim, err := s.reserveWorkflowClaim(ctx, *run)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, claim.releaseReservation()) }()

	commit, _, err := s.workflowLaneCandidate(ctx, *run.Lane)
	if err != nil {
		return err
	}

	if err := s.beginWorkflowSliceRepair(run, observation, claim); err != nil {
		return err
	}

	lane := *run.Lane
	lane.HeadCommit = commit
	run.Lane = &lane
	run.Cursor -= len(observation.tasks) - index
	run.Status, run.Outcome, run.ResumeStatus = "pending", "slice_repair_requested", ""

	run.CandidateRef, run.CandidateHash = "", ""
	if run.Preparation != nil {
		run.Preparation.State = "stale"
	}

	if run.Review != nil {
		run.Review.State = "stale"
	}

	return s.commitWorkflowSliceRepair(run, observation, index)
}

func workflowSliceRepairReady(run WorkflowInvocation) bool {
	if run.PullRequest != nil && !slices.Contains([]string{"prepared", "bound"}, run.PullRequest.State) {
		return false
	}

	return workflowSliceValidationReady(run) || run.Status == "waiting_review"
}

func (s *WorkflowStore) commitWorkflowSliceRepair(run *workflowInvocationFile, observation workflowSliceObservation, index int) error {
	files := []workflowFileChange{}

	for i, record := range observation.tasks {
		if i < index {
			if record.Task.Phase == "validated" {
				continue
			}

			record.Task.Phase, record.Task.Status = "validated", TicketStatusInProgress
		} else {
			record.Task.Phase, record.Task.Status, record.Task.Evidence, record.Task.AttemptID = "pending", TicketStatusReadyForDev, nil, ""
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

func (s *WorkflowStore) workflowRepairSetChanges(run *workflowInvocationFile, set WorkflowRecord) ([]workflowFileChange, error) {
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

func (s *WorkflowStore) beginWorkflowSliceRepair(run *workflowInvocationFile, observation workflowSliceObservation, claim *WorkflowClaim) error {
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
