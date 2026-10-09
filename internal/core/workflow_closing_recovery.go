package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ReconcileWorkflowClosing records interruption only after the previous owner
// expires or dies, every descendant stops, and common capacity is acquired.
// It does not confirm a gate, remove attempt history or mutate task records.
func (s *WorkflowStore) ReconcileWorkflowClosing(ctx context.Context, setPath string) (result WorkflowRecord, err error) {
	err = s.withLock(ctx, func() error {
		set, readErr := s.readRecord(setPath)
		if readErr != nil {
			return readErr
		}

		if set.Set == nil || set.Set.Closing == nil {
			return ErrWorkflowConflict
		}

		result, readErr = s.reconcileWorkflowClosing(ctx, set)

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) reconcileWorkflowClosing(ctx context.Context, set WorkflowRecord) (result WorkflowRecord, err error) {
	state, err := s.readWorkflowClosingAttempt(set, len(set.Set.Closing.Attempts)-1)
	if err != nil || state.Outcome != "running" {
		return set, err
	}

	attempt := WorkflowAttempt{ID: workflowClosingAttemptID(set, state.Generation), OwnerProcess: &state.Owner, LeaseUntil: state.LeaseUntil, Processes: state.Processes}
	if state.OwnerRoot != s.layout.Root() {
		return set, ErrWorkflowOwned
	}

	if err := errors.Join(workflowRecoverableOwner(attempt), workflowAttemptQuiescent(attempt)); err != nil {
		return set, err
	}

	reservation, err := acquireWorkflowCapacity(ctx, s.layout.ProjectDir())
	if err != nil {
		return set, err
	}
	defer func() { err = errors.Join(err, reservation.Close()) }()

	run := workflowInvocationFile{WorkflowInvocation: WorkflowInvocation{WorkflowIdentity: WorkflowIdentity{ID: workflowClosingOwnerID(set)}, Attempts: []WorkflowAttempt{attempt}}}

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
	if err := s.commitWorkflowClosingState(ctx, claim); err != nil {
		return set, err
	}

	return claim.set, releaseWorkflowAdoptionLocks(locks)
}

func workflowClosingAttemptID(set WorkflowRecord, generation int) string {
	return fmt.Sprintf("%s-attempt-%d", workflowClosingOwnerID(set), generation)
}

func (s *WorkflowStore) workflowRecoveredClosingClaim(set WorkflowRecord, state workflowClosingAttempt) (*workflowClosingClaim, error) {
	data, err := s.workflowClosingRead(state.Input.Ref)
	if err != nil || workflowHash(data) != state.Input.Hash {
		return nil, errorsOrConflict(err)
	}

	claim := &workflowClosingClaim{set: set, state: state}
	if err := json.Unmarshal(data, &claim.input); err != nil {
		return nil, err
	}

	for i := range claim.state.Processes {
		process := &claim.state.Processes[i]

		process.State = "stopped"
		if err := s.hashWorkflowProcessOutput(process); err != nil {
			return nil, err
		}
	}

	return claim, nil
}
