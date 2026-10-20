package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"forge.adrianpk.com/chr/internal/harness"
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

	release, err := HoldWorkerRuntimeReservation(t.Context(), WorkerRuntimeReservation{Slot: slot.Name(), Guard: guard.Name()})
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

	if release, err := HoldWorkerRuntimeReservation(ctx, WorkerRuntimeReservation{Slot: slot.Name(), Guard: guard}); !errors.Is(err, context.Canceled) || release != nil {
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

// Explicit project instructions must not become a route to host files through
// links, a blocking FIFO, invalid text or an oversized resource.
func TestPiWorkerInstructions(t *testing.T) {
	for _, scenario := range []string{"text", "missing", "symlink", "hard link", "FIFO", "invalid text", "oversized"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			prepareWorkerInstructionFixture(t, root, scenario)

			data, err := WorkerRuntimeInstructions(root)
			if scenario == "text" || scenario == "missing" {
				if err != nil || (scenario == "text" && string(data) != "Project instructions") {
					t.Fatalf("explicit instructions: %q, %v", data, err)
				}
			} else if err == nil || len(data) != 0 {
				t.Fatal("unsafe project instructions were imported")
			}
		})
	}
}

func prepareWorkerInstructionFixture(t *testing.T, root, scenario string) {
	t.Helper()

	path := filepath.Join(root, "AGENTS.md")

	var err error

	switch scenario {
	case "missing":
		return
	case "symlink", "hard link":
		outside := filepath.Join(t.TempDir(), "outside")
		if err := os.WriteFile(outside, []byte("outside fixture"), 0o600); err != nil {
			t.Fatal(err)
		}

		if scenario == "symlink" {
			err = os.Symlink(outside, path)
		} else {
			err = os.Link(outside, path)
		}
	case "FIFO":
		err = syscall.Mkfifo(path, 0o600)
	case "invalid text":
		err = os.WriteFile(path, []byte{0xff}, 0o600)
	case "oversized":
		err = os.WriteFile(path, make([]byte, harness.MaxInput+1), 0o600)
	default:
		err = os.WriteFile(path, []byte("Project instructions"), 0o600)
	}

	if err != nil {
		t.Fatal(err)
	}
}

// Even an otherwise explicit checkout cannot grant the entire developer home.
func TestPiWorkerHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := WorkerRuntimeMounts(t.Context(), home); err == nil || !strings.Contains(err.Error(), "home") {
		t.Fatalf("home mount was accepted: %v", err)
	}
}
