package core

import (
	"context"
	"errors"
	"syscall"
)

// Quiescent recovery reacquires the attempt's exact slot. A still-open old
// descriptor must block fencing even after lease expiry. Its shared legacy
// guard permits unrelated slots to continue; unbound history stays exclusive.
func acquireWorkflowRecoveryGuard(ctx context.Context, projectDir string, slot int) (*WorkflowClaim, error) {
	if slot == 0 {
		file, err := acquireWorkflowCapacity(ctx, projectDir)

		return &WorkflowClaim{reservation: file}, err
	}

	common, err := workflowRepositoryDirectory(ctx, projectDir)
	if err != nil {
		return nil, err
	}

	if _, err := workflowOccupiedProcessSlots(common); err != nil {
		return nil, err
	}

	guard, err := workflowCapacityFile(common, "chr-workflow-worker.lock", syscall.LOCK_SH)
	if err != nil {
		return nil, err
	}

	file, err := workflowCapacityFile(common, workflowWorkerSlotName(slot), syscall.LOCK_EX)
	if err != nil {
		return nil, errors.Join(err, guard.Close())
	}

	return &WorkflowClaim{reservation: file, capacityGuard: guard}, nil
}
