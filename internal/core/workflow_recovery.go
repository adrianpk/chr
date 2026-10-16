package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// ReconcileWorkflow fences an expired or dead owner after proving its group dead
// and acquiring capacity. It never creates an attempt or resets a retry budget.
func (s *WorkflowStore) ReconcileWorkflow(ctx context.Context, id string, expected ...uint64) (result WorkflowInvocation, err error) {
	return s.reconcileWorkflowChecked(ctx, id, nil, expected...)
}

func (s *WorkflowStore) reconcileWorkflowChecked(ctx context.Context, id string, check WorkflowAdmissionCheck, expected ...uint64) (result WorkflowInvocation, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if err := checkWorkflowRevision(run.WorkflowInvocation, expected); err != nil {
			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		if readErr = owner.reconcileWorkflowInvocation(ctx, run); readErr != nil {
			return readErr
		}

		updated, readErr := owner.readWorkflowInvocation(run.ID)
		result = updated.WorkflowInvocation

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) reconcileWorkflowInvocation(ctx context.Context, run workflowInvocationFile) error {
	if run.ClosingSet != "" {
		set, err := s.readRecord(run.SetImages[len(run.Closures)].Path)
		if err != nil {
			return err
		}

		_, err = s.reconcileWorkflowClosing(ctx, set)

		return err
	}

	if len(run.Attempts) == 0 {
		return nil
	}

	return s.reconcileWorkflowOwner(ctx, run)
}

func workflowRecoverableOwner(attempt WorkflowAttempt) error {
	if attempt.OwnerProcess == nil {
		return ErrWorkflowConflict
	}

	alive, err := workflowProcessAlive(*attempt.OwnerProcess)
	if err != nil {
		return err
	}

	if alive && time.Now().Before(attempt.LeaseUntil) {
		return ErrWorkflowOwned
	}

	return nil
}

func (s *WorkflowStore) recoveredWorkflowLocks(ctx context.Context, run workflowInvocationFile) ([]*AgentOperationLock, error) {
	layouts, err := s.workflowOperationLayouts(ctx)
	if err != nil {
		return nil, err
	}

	locks := []*AgentOperationLock{}

	attempt := run.Attempts[len(run.Attempts)-1]
	for _, layout := range layouts {
		lock, err := recoveredWorkflowLock(filepath.Join(layout.Root(), "lock", "agent-operation.lock"), run.ID, attempt)
		if err != nil {
			return nil, err
		}

		if lock != nil {
			locks = append(locks, lock)
		}
	}

	return locks, nil
}

func (s *WorkflowStore) reconcileWorkflowOwner(ctx context.Context, run workflowInvocationFile) (err error) {
	attempt := run.Attempts[len(run.Attempts)-1]
	if err := workflowAttemptQuiescent(attempt); err != nil {
		return err
	}

	if err := s.checkRecoverableWorkflow(ctx, run, attempt); err != nil {
		return err
	}

	reservation, err := acquireWorkflowRecoveryGuard(ctx, s.layout.ProjectDir(), attempt.CapacitySlot)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, reservation.releaseReservation()) }()

	locks, err := s.recoveredWorkflowLocks(ctx, run)
	if err != nil {
		return err
	}

	if run.Status == "running" {
		if err := s.recoverWorkflowAttempt(run, attempt); err != nil {
			return err
		}
	}

	return releaseWorkflowAdoptionLocks(locks)
}

func (s *WorkflowStore) checkRecoverableWorkflow(ctx context.Context, run workflowInvocationFile, attempt WorkflowAttempt) error {
	if run.Status != "running" {
		return nil
	}

	if err := workflowRecoverableOwner(attempt); err != nil {
		return err
	}

	return s.checkWorkflowInputs(ctx, run)
}

func (s *WorkflowStore) recoverWorkflowAttempt(run workflowInvocationFile, attempt WorkflowAttempt) error {
	record, err := s.readWorkflowAttemptTask(run, attempt)
	if err != nil {
		return err
	}

	if record.Task.AttemptID != attempt.ID || record.Task.ClaimGeneration != attempt.Generation {
		return ErrWorkflowConflict
	}

	return s.checkpointWorkflowAttempt(run, record, "recovered", nil)
}

func recoveredWorkflowLock(path, invocationID string, attempt WorkflowAttempt) (*AgentOperationLock, error) {
	data, readErr := readWorkflowFile(path)
	if errors.Is(readErr, os.ErrNotExist) {
		return nil, nil
	}

	if readErr != nil {
		return nil, readErr
	}

	var lock agentOperationLockRecord
	if readErr = json.Unmarshal(data, &lock); readErr != nil {
		return nil, readErr
	}

	if lock.Operation != "workflow_execution" || lock.Owner != invocationID {
		return nil, nil
	}

	if lock.AttemptID != attempt.ID || lock.Process == nil || attempt.OwnerProcess == nil || *lock.Process != *attempt.OwnerProcess {
		return nil, ErrWorkflowConflict
	}

	return &AgentOperationLock{path: path, guard: workflowHash(data)}, nil
}
