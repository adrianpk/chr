package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowContinuationRecovery interrupts both checkpoint replacements and
// resumes the same paused invocation without duplicate checkpoint events or work.
func TestWorkflowContinuationRecovery(t *testing.T) {
	for _, index := range []int{0, 1} {
		t.Run(string(rune('0'+index)), func(t *testing.T) {
			store, _, run := workflowContinuationFixture(t)
			if _, err := store.StopWorkflow(t.Context(), run.ID, "pause"); err != nil {
				t.Fatal(err)
			}

			store.afterReplace = func(position int) error {
				if position == index {
					return errors.New("interrupted continuation checkpoint")
				}

				return nil
			}

			if _, err := store.ContinueWorkflow(t.Context(), run.ID, WorkflowContinuationRequest{StepLimit: 1}, WorkflowContinuationPorts{}); err == nil {
				t.Fatal("checkpoint replacement was not interrupted")
			}

			restarted := NewWorkflowStore(store.layout, store.projectID)
			first := workflowContinueOne(t, restarted, run.ID, WorkflowContinuationPorts{})

			second := workflowContinueOne(t, restarted, run.ID, WorkflowContinuationPorts{})
			if first.Steps != 0 || first.Reason != "stopped" || first.Invocation.Continuation.Sequence != 1 || first.Invocation.Revision != second.Invocation.Revision || len(first.Invocation.Attempts) != 0 {
				t.Fatal("checkpoint replay reset policy, repeated an event or executed paused work")
			}
		})
	}
}

// TestWorkflowContinuationCorruption rejects altered immutable checkpoint bytes
// before allocating the next task, rather than replacing the evidence on retry.
func TestWorkflowContinuationCorruption(t *testing.T) {
	store, _, run := workflowContinuationFixture(t)
	first := workflowContinueOne(t, store, run.ID, WorkflowContinuationPorts{})

	path := filepath.Join(store.layout.Root(), first.Invocation.Continuation.EvidenceRef)
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ContinueWorkflow(t.Context(), run.ID, WorkflowContinuationRequest{Worker: workflowPRWorker(), StepLimit: 1}, WorkflowContinuationPorts{}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("corrupt continuation admitted more work", err)
	}

	retained := workflowTestRead(t, store, run.ID)
	if retained.Cursor != 1 || len(retained.Attempts) != 1 || retained.Revision != first.Invocation.Revision {
		t.Fatal("corrupt checkpoint was silently replaced or repeated a task")
	}
}

// TestWorkflowContinuationTerminal keeps cancelled requests and checkpoints
// immutable on both explicit continuation and equivalent retry.
func TestWorkflowContinuationTerminal(t *testing.T) {
	store, request, run := workflowContinuationFixture(t)
	first := workflowContinueOne(t, store, run.ID, WorkflowContinuationPorts{})

	cancelled, err := store.StopWorkflow(t.Context(), run.ID, "cancel")
	if err != nil {
		t.Fatal(err)
	}

	again := workflowContinueOne(t, store, run.ID, WorkflowContinuationPorts{})

	attached := workflowTestStart(t, store, request)
	if again.Steps != 0 || again.Reason != "cancelled" || again.Invocation.Revision != cancelled.Revision || again.Invocation.Continuation.EvidenceHash != first.Invocation.Continuation.EvidenceHash || !attached.Attached || attached.Invocation.ID != run.ID {
		t.Fatal("terminal continuation mutated old scope or restarted work")
	}
}
