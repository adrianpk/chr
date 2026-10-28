package workflow

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) prepareWorkflowMerger(ctx context.Context, request MergerRequest, harness MergerHarness, provider PullRequestProvider) (claimed workflowClaimedMerger, err error) {
	request, err = freezeWorkflowMergerRequest(request)
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		admission, readErr := invocationLocation.Owner.workflowMergerAdmission(ctx, invocationLocation.Run, request, harness, provider)
		if readErr != nil {
			return readErr
		}

		claimed, readErr = invocationLocation.Owner.claimWorkflowMerger(ctx, admission, request.Input, harness)

		return readErr
	})

	return claimed, err
}

func (s *Store) claimWorkflowMerger(ctx context.Context, admission workflowMergerAdmissionResult, input workflowstate.MergerInput, harness MergerHarness) (claimed workflowClaimedMerger, err error) {
	run, entry, budget, seed := admission.Run, admission.Entry, admission.Budget, admission.Seed

	if err := checkWorkflowIntegrationCounterCharge(budget, IntegrationReservation{Kind: "agent_call", Role: "merger", Unit: entry.Unit}, "merger"); err != nil {
		return workflowClaimedMerger{}, err
	}

	claim, err := s.reserveWorkflowRole(ctx, run, workflowMergerPurpose)
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.Close())
		}
	}()

	claim.Token, err = newWorkflowClaimToken()
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	workspace, err := (workflowMergerHost{store: s}).workspace(ctx, input, seed)

	lane := workspace.Lane
	if err == nil && input.RepairCheckpointRef != nil {
		lane, err = (workflowMergerHost{store: s}).seedCheckpoint(ctx, input, lane)
	}

	if err != nil {
		return workflowClaimedMerger{}, err
	}

	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	record := observation.tasks[len(observation.tasks)-1]
	if err := s.beginWorkflowAttempt(&run, &record, claim); err != nil {
		return workflowClaimedMerger{}, err
	}

	workspace.Lane = lane

	return s.persistWorkflowMergerClaim(ctx, claim, workflowMergerClaimInput{Observation: workflowClaimObservation{Run: run, Record: record}, Budget: budget, Input: input, Harness: harness, Workspace: workspace})
}

func workflowMergerReservation(run *workflowInvocationFile, budget IntegrationBudget, input *workflowstate.MergerInput, claim *Claim) (IntegrationReservation, error) {
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

	reservation := IntegrationReservation{ID: input.BudgetReservationID, Kind: "agent_call", Role: "merger", EntryID: input.EntryID, Unit: input.Unit, ClaimGeneration: input.ClaimGeneration, AttemptID: claim.AttemptID, State: "reserved", RuntimeDeadline: input.Deadline}
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
