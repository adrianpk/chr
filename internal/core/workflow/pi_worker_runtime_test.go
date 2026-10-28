package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
)

// The container retains the existing admitted slot after the original host
// descriptors close. A fresh exclusive admission remains blocked until reaping.
func TestPiWorkerIsolationReservation(t *testing.T) {
	common := t.TempDir()

	slot, err := workflowCapacityFile(common, workflowWorkerSlotName(1), syscall.LOCK_EX)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = slot.Close() }() //nolint:errcheck // A fixture closes the descriptor early below.

	guard, err := workflowCapacityFile(common, "chr-workflow-worker.lock", syscall.LOCK_SH)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = guard.Close() }() //nolint:errcheck // A fixture closes the descriptor early below.

	if err := syscall.Flock(int(slot.Fd()), syscall.LOCK_SH); err != nil {
		t.Fatal(err)
	}

	release, err := harnessruntime.HoldReservation(t.Context(), harnessruntime.Reservation{Slot: slot.Name(), Guard: guard.Name()})
	if err != nil {
		t.Fatal(err)
	}

	if err := errors.Join(slot.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}

	assertRuntimeSlotOccupied(t, common)

	if err := release(); err != nil {
		t.Fatal(err)
	}

	free, err := workflowCapacityFile(common, workflowWorkerSlotName(1), syscall.LOCK_EX)
	if err != nil {
		t.Fatalf("reaped runtime retained capacity: %v", err)
	}

	if err := free.Close(); err != nil {
		t.Fatal(err)
	}
}

// A stale container cannot join a slot held by a later exclusive owner; its
// canceled host context stops that wait without running a worker.
func TestPiWorkerIsolationCanceled(t *testing.T) {
	common := t.TempDir()

	slot, err := workflowCapacityFile(common, workflowWorkerSlotName(1), syscall.LOCK_EX)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = slot.Close() }() //nolint:errcheck // Fixture cleanup.

	guard := filepath.Join(common, "chr-workflow-worker.lock")
	if err := os.WriteFile(guard, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if release, err := harnessruntime.HoldReservation(ctx, harnessruntime.Reservation{Slot: slot.Name(), Guard: guard}); !errors.Is(err, context.Canceled) || release != nil {
		t.Fatalf("abandoned runtime acquired capacity: %v", err)
	}
}

func assertRuntimeSlotOccupied(t *testing.T, common string) {
	t.Helper()

	if file, err := workflowCapacityFile(common, workflowWorkerSlotName(1), syscall.LOCK_EX); !errors.Is(err, ErrWorkflowCapacity) {
		if file != nil {
			_ = file.Close() //nolint:errcheck // Failure cleanup.
		}

		t.Fatalf("container lost shared capacity: %v", err)
	}
}
