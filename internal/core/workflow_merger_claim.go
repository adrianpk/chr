package core

import (
	"context"
	"errors"
	"time"
)

func (s *WorkflowStore) prepareWorkflowMerger(ctx context.Context, request WorkflowMergerRequest, harness WorkflowMergerHarness, provider WorkflowPullRequestProvider) (claim *WorkflowClaim, prepared workflowPreparedTask, err error) {
	request, err = freezeWorkflowMergerRequest(request)
	if err != nil {
		return nil, prepared, err
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		entry, budget, seed, readErr := owner.workflowMergerAdmission(ctx, run, request, harness, provider)
		if readErr != nil {
			return readErr
		}

		claim, prepared, readErr = owner.claimWorkflowMerger(ctx, run, entry, budget, request.Input, harness, seed)

		return readErr
	})

	return claim, prepared, err
}

func (s *WorkflowStore) claimWorkflowMerger(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, input WorkflowMergerInput, harness WorkflowMergerHarness, seed string) (claim *WorkflowClaim, prepared workflowPreparedTask, err error) {
	if err := checkWorkflowIntegrationCounterCharge(budget, WorkflowIntegrationReservation{Kind: "agent_call", Role: "merger", Unit: entry.Unit}, "merger"); err != nil {
		return nil, prepared, err
	}

	claim, err = s.reserveWorkflowRole(ctx, run, workflowMergerPurpose)
	if err != nil {
		return nil, prepared, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.Close())
		}
	}()

	claim.Token, err = newWorkflowClaimToken()
	if err != nil {
		return nil, prepared, err
	}

	lane, root, err := s.workflowMergerWorkspace(ctx, input, seed)
	if err == nil && input.RepairCheckpointRef != nil {
		lane, err = s.seedWorkflowMergerCheckpoint(ctx, input, lane)
	}

	if err != nil {
		return nil, prepared, err
	}

	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return nil, prepared, err
	}

	record := observation.tasks[len(observation.tasks)-1]
	if err := s.beginWorkflowAttempt(&run, &record, claim); err != nil {
		return nil, prepared, err
	}

	return s.persistWorkflowMergerClaim(ctx, run, record, claim, budget, input, harness, lane, root)
}

func workflowMergerReservation(run *workflowInvocationFile, budget WorkflowIntegrationBudget, input *WorkflowMergerInput, claim *WorkflowClaim) (WorkflowIntegrationReservation, error) {
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.LeaseUntil = minWorkflowMergerDeadline(attempt.LeaseUntil, input.Deadline)
	input.AdmittedAt, input.Deadline = attempt.StartedAt, attempt.LeaseUntil
	input.CapacityReservationID = "worker-slot-" + workflowWorkerSlotLabel(claim.reservation)
	input.BudgetReservationID = claim.AttemptID

	remaining := int64(0)
	for _, counter := range budget.UnitCounters {
		remaining += counter.Merger
	}

	input.RemainingAttempts = int(min(int64(10), budget.LimitsSnapshot.Merger-remaining))

	reservation := WorkflowIntegrationReservation{ID: input.BudgetReservationID, Kind: "agent_call", Role: "merger", EntryID: input.EntryID, Unit: input.Unit, ClaimGeneration: input.ClaimGeneration, AttemptID: claim.AttemptID, State: "reserved", RuntimeDeadline: input.Deadline}
	if input.TokenAllowance != nil {
		input.TokenAllowance.Deadline = input.Deadline
		reservation.Allowance, reservation.RuntimeDeadline = *input.TokenAllowance, time.Time{}
	} else if budget.TokenLimit != nil {
		return reservation, ErrWorkflowIntegrationBudget
	}

	return reservation, validateWorkflowIntegrationReservation(reservation)
}

func minWorkflowMergerDeadline(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}

	return a
}
