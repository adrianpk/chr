package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReconcileWorkflowClosing records interruption only after the previous owner
// expires or dies, every descendant stops, and common capacity is acquired.
// It does not confirm a gate, remove attempt history or mutate task records.
func (s *Store) ReconcileWorkflowClosing(ctx context.Context, setPath string) (result Record, err error) {
	err = s.withLock(ctx, func() error {
		set, readErr := s.readRecord(setPath)
		if readErr != nil {
			return readErr
		}

		if set.Set == nil || set.Set.Closing == nil {
			return workflowstate.ErrConflict
		}

		result, readErr = s.reconcileWorkflowClosing(ctx, set, nil)

		return readErr
	})

	return result, err
}

func (s *Store) reconcileWorkflowClosing(ctx context.Context, set Record, control *workflowControlCommit) (result Record, err error) {
	state, err := s.readWorkflowClosingAttempt(set, len(set.Set.Closing.Attempts)-1)
	if err != nil {
		return set, err
	}

	if state.Outcome != "running" {
		return set, s.retainWorkflowClosingControlNoop(ctx, control)
	}

	attempt := Attempt{ID: workflowClosingAttemptID(set, state.Generation), OwnerProcess: &state.Owner, LeaseUntil: state.LeaseUntil, Processes: state.Processes}
	if state.OwnerRoot != s.layout.Root() {
		return set, ErrWorkflowOwned
	}

	if err := errors.Join(workflowRecoverableOwner(attempt), workflowAttemptQuiescent(attempt)); err != nil {
		return set, err
	}

	reservation, err := acquireWorkflowRecoveryGuard(ctx, s.layout.ProjectDir(), state.CapacitySlot)
	if err != nil {
		return set, err
	}
	defer func() { err = errors.Join(err, reservation.releaseReservation()) }()

	run := workflowInvocationFile{Invocation: Invocation{Identity: workflowstate.Identity{ID: workflowClosingOwnerID(set)}, InvocationExecution: InvocationExecution{Attempts: []Attempt{attempt}}}}

	locks, err := s.recoveredWorkflowLocks(ctx, run)
	if err != nil {
		return set, err
	}

	claim, err := s.workflowRecoveredClosingClaim(set, state)
	if err != nil {
		return set, err
	}

	claim.state.Outcome, claim.state.FinishedAt = "interrupted", time.Now().UTC()

	claim.set.Set.Status = "blocked"

	files, err := s.workflowClosingControlFiles(ctx, claim, control)
	if err != nil {
		return set, err
	}

	if err := s.commitWorkflowClosingState(claim, files...); err != nil {
		return set, err
	}

	return claim.set, releaseWorkflowAdoptionLocks(locks)
}

func workflowClosingAttemptID(set Record, generation int) string {
	return fmt.Sprintf("%s-attempt-%d", workflowClosingOwnerID(set), generation)
}

func (s *Store) workflowRecoveredClosingClaim(set Record, state workflowClosingAttempt) (*workflowClosingClaim, error) {
	data, err := s.workflowClosingRead(state.Input.Ref)
	if err != nil || contenthash.Sum(data) != state.Input.Hash {
		return nil, errorsOrConflict(err)
	}

	claim := &workflowClosingClaim{set: set, state: state}
	if err := json.Unmarshal(data, &claim.input); err != nil {
		return nil, err
	}

	for i := range claim.state.Processes {
		process := &claim.state.Processes[i]

		process.State = ProcessStopped
		if err := s.hashWorkflowProcessOutput(process); err != nil {
			return nil, err
		}
	}

	return claim, nil
}
