package workflow

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowClosingBudget retains exhaustion across restart and rejects before
// another attempt. A distinct explicitly admitted invocation receives its own
// approved budget while all prior set-owned attempt images remain intact.
func TestWorkflowClosingBudget(t *testing.T) {
	fixture, set, run := workflowClosingEmptyOwner(t)
	claim := workflowClosingBeginTest(t, fixture, run, workflowFrontierPorts(fixture))
	expireWorkflowClosingTest(t, fixture.store, claim)

	if _, err := fixture.store.ReconcileWorkflow(t.Context(), run.ID); err != nil {
		t.Fatal(err)
	}

	seedWorkflowClosingBudget(t, fixture.store, set.Path, run.ID)

	restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)
	if _, err := CloseInvocationSet(t.Context(), restarted, run.ID, "dev", workflowFrontierPorts(fixture)); err == nil || !strings.Contains(err.Error(), "attempt budget exhausted") {
		t.Fatal("restart reset an exhausted set invocation budget", err)
	}

	if _, err := restarted.StopWorkflow(t.Context(), run.ID, "cancel"); err != nil {
		t.Fatal(err)
	}

	start := StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "new-closing-budget",
		Policy:     workflowstate.ExecutionPolicy{CheckpointStepLimit: 1},
		Selection:  run.Selection,
	}}

	next := workflowTestStart(t, restarted, start)
	if next.Invocation.ID == run.ID {
		t.Fatal("distinct policy request did not create a new set invocation")
	}

	closed, err := CloseInvocationSet(t.Context(), restarted, next.Invocation.ID, "dev", workflowFrontierPorts(fixture))
	if err != nil || closed.Outcome != "completed" || len(closed.Set.Set.Closing.Attempts) != workflowstate.PageLimit+1 {
		t.Fatal("new invocation lost prior attempts or inherited another invocation's exhausted budget", err)
	}
}

func seedWorkflowClosingBudget(t *testing.T, store *Store, path, id string) {
	t.Helper()

	err := store.withLock(t.Context(), func() error {
		set, err := store.readRecord(path)
		if err != nil {
			return err
		}

		state, err := workflowcloseout.ReadAttempt(set.ID, set.Set.Closing, 0, workflowClosingEvidence{store: store})
		if err != nil {
			return err
		}

		input, err := (workflowClosingEvidence{store: store}).Read(state.Input.Ref)
		if err != nil {
			return err
		}

		files := workflowClosingBudgetImages(t, &set, state, input)
		set.Revision++
		set.UpdatedAt = time.Now().UTC()

		projection, err := store.encodeChanges(set)
		if err != nil {
			return err
		}

		set.SourceHash = workflowClaimRecordHash(set, projection)

		run, err := store.readWorkflowInvocation(id)
		if err != nil {
			return err
		}

		run.SetImages[0].Revision, run.SetImages[0].SourceHash = set.Revision, set.SourceHash
		run.Revision++

		change, err := workflowInvocationChange(run)
		if err != nil {
			return err
		}

		return store.commitFiles(append(append(files, projection...), change))
	})
	if err != nil {
		t.Fatal(err)
	}
}

func workflowClosingBudgetImages(t *testing.T, set *workflowstate.Record, state workflowcloseout.Attempt, input []byte) []workflowstate.FileChange {
	t.Helper()

	files := []workflowstate.FileChange{}

	for generation := 2; generation <= workflowstate.PageLimit; generation++ {
		state.Generation = generation
		state.Input.Ref = workflowcloseout.File(set.ID, generation, "input.json")

		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			t.Fatal(err)
		}

		data = append(data, '\n')
		ref := workflowcloseout.File(set.ID, generation, "state.json")
		set.Set.Closing.Attempts = append(set.Set.Closing.Attempts, workflowstate.ClosingEvidence{Ref: ref, Hash: contenthash.Sum(data)})
		files = append(files, workflowstate.FileChange{Path: ref, After: data}, workflowstate.FileChange{Path: state.Input.Ref, After: input})
	}

	return files
}
