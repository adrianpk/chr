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

		if !workflowSliceValidationReady(run.WorkflowInvocation) || run.StopIntent != "" {
			return ErrWorkflowOwned
		}

		observation, readErr := owner.readWorkflowSliceMembership(ctx, run)
		if readErr != nil {
			return readErr
		}

		readErr = owner.repairWorkflowSlice(ctx, &run, observation, taskKey)
		result = run.WorkflowInvocation

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) repairWorkflowSlice(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, taskKey string) (err error) {
	index := slices.IndexFunc(observation.tasks, func(record WorkflowRecord) bool { return record.Task.Key == taskKey })
	if index < 0 {
		return fmt.Errorf("repair task is outside the retained slice")
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

	return s.commitWorkflowSliceRepair(run, observation.tasks[index:])
}

func (s *WorkflowStore) commitWorkflowSliceRepair(run *workflowInvocationFile, records []WorkflowRecord) error {
	files := []workflowFileChange{}

	for _, record := range records {
		record.Task.Phase, record.Task.Status, record.Task.Evidence, record.Task.AttemptID = "pending", TicketStatusReadyForDev, nil, ""

		changes, err := s.workflowClaimRecordChanges(run, &record)
		if err != nil {
			return err
		}

		files = append(files, changes...)
	}

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	return s.commitFiles(append(files, state))
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
	attempt.ContractHash = observation.descriptor.ContractHash
	attempt.Outcome, attempt.FinishedAt = "repaired", time.Now().UTC()

	return nil
}
