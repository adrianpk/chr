package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReconcileWorkflow fences an expired or dead owner after proving its group dead
// and acquiring capacity. It never creates an attempt or resets a retry budget.
func (s *Store) ReconcileWorkflow(ctx context.Context, id string, expected ...uint64) (result Invocation, err error) {
	return s.reconcileWorkflowChecked(ctx, id, nil, expected...)
}

func (s *Store) reconcileWorkflowChecked(ctx context.Context, id string, check AdmissionCheck, expected ...uint64) (result Invocation, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if err := checkWorkflowRevision(run.Invocation, expected); err != nil {
			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		if readErr = owner.reconcileWorkflowInvocation(ctx, run, nil); readErr != nil {
			return readErr
		}

		updated, readErr := owner.readWorkflowInvocation(run.ID)
		result = updated.Invocation

		return readErr
	})

	return result, err
}

func (s *Store) reconcileWorkflowInvocation(ctx context.Context, run workflowInvocationFile, control *workflowControlCommit) error {
	if run.ClosingSet != "" {
		set, err := s.readRecord(run.SetImages[len(run.Closures)].Path)
		if err != nil {
			return err
		}

		_, err = s.reconcileWorkflowClosing(ctx, set, control)

		return err
	}

	if len(run.Attempts) == 0 {
		return s.retainWorkflowControlNoop(ctx, run, control)
	}

	return s.reconcileWorkflowOwner(ctx, run, control)
}

func workflowRecoverableOwner(attempt Attempt) error {
	if attempt.OwnerProcess == nil {
		return workflowstate.ErrConflict
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

func (s *Store) recoveredWorkflowLocks(ctx context.Context, run workflowInvocationFile) ([]*AgentOperationLock, error) {
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

func (s *Store) reconcileWorkflowOwner(ctx context.Context, run workflowInvocationFile, control *workflowControlCommit) (err error) {
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
		if err := s.recoverWorkflowAttempt(ctx, run, attempt, control); err != nil {
			return err
		}
	} else if err := s.retainWorkflowControlNoop(ctx, run, control); err != nil {
		return err
	}

	return releaseWorkflowAdoptionLocks(locks)
}

func (s *Store) checkRecoverableWorkflow(ctx context.Context, run workflowInvocationFile, attempt Attempt) error {
	if run.Status != "running" {
		return nil
	}

	if err := workflowRecoverableOwner(attempt); err != nil {
		return err
	}

	return s.checkWorkflowInputs(ctx, run)
}

func (s *Store) recoverWorkflowAttempt(ctx context.Context, run workflowInvocationFile, attempt Attempt, control *workflowControlCommit) error {
	record, err := s.readWorkflowAttemptTask(run, attempt)
	if err != nil {
		return err
	}

	if record.Task.AttemptID != attempt.ID || record.Task.ClaimGeneration != attempt.Generation {
		return workflowstate.ErrConflict
	}

	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return err
	}

	return s.checkpointWorkflowAttempt(run, record, "recovered", nil, files...)
}

func recoveredWorkflowLock(path, invocationID string, attempt Attempt) (*AgentOperationLock, error) {
	data, readErr := opsfile.Read(path)
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
		return nil, workflowstate.ErrConflict
	}

	return &AgentOperationLock{path: path, guard: contenthash.Sum(data)}, nil
}
