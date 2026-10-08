package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func replaceWorkflowManualStatus(t *testing.T, data []byte) []byte {
	t.Helper()

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	// Preserve valid JSON while altering execution-owned ticket state directly.
	fields["workflow"] = replaceWorkflowManualField(t, fields["workflow"], "phase", "validated")

	changed, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}

	return changed
}

func replaceWorkflowManualField(t *testing.T, data []byte, field, value string) []byte {
	t.Helper()

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	fields[field] = encoded

	changed, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}

	return changed
}

// TestWorkflowManualFailure proves reported work undergoes the same failing gate
// and retained-attempt budget as automatic work, without accepting a status claim.
func TestWorkflowManualFailure(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, []string{"sh", "-c", "exit 7"}, 30)
	request.Policy.MaxWorkerAttempts = 1
	run := workflowTestStart(t, store, request)
	claim, task := workflowTestManual(t, store, run.Invocation.ID, nil)
	candidate := workflowManualTestCandidate(t, store, task)

	proof, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate)
	if err == nil || proof.Successful || len(proof.Validation) != 1 || proof.Validation[0].ExitCode != 7 {
		t.Fatal("manual gate failure was accepted")
	}

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || retained.Status != "paused" || retained.Cursor != 0 || retained.Attempts[0].ResultRef == "" {
		t.Fatal("manual failure lost retained evidence")
	}

	if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); err == nil {
		t.Fatal("manual failure reset budget")
	}
}

// TestWorkflowManualPrerequisite refuses a prerequisite changed after handoff.
func TestWorkflowManualPrerequisite(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	first := workflowTestStart(t, store, request)
	workflowTestExecute(t, store, first.Invocation.ID, "worker")

	request.RequestKey = "manual-next"
	request.Selection.Selector.Identity = hierarchy.Tasks[1].Task.Key
	next := workflowTestStart(t, store, request)
	claim, task := workflowTestManual(t, store, next.Invocation.ID, nil)
	candidate := workflowManualTestCandidate(t, store, task)

	path := filepath.Join(store.layout.Root(), filepath.Dir(hierarchy.Tasks[0].Path), "artifact.md")
	if err := os.WriteFile(path, []byte("Changed prerequisite\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("prerequisite edit accepted: %v", err)
	}
}

// TestWorkflowManualJournal reuses an accepted result after partial result replay
// instead of creating another attempt or rerunning the frozen host gate.
func TestWorkflowManualJournal(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, []string{"true"}, 30)
	run := workflowTestStart(t, store, request)
	claim, task := workflowTestManual(t, store, run.Invocation.ID, nil)
	candidate := workflowManualTestCandidate(t, store, task)

	interrupted := interruptWorkflowManualResult(claim, store, hierarchy.Tasks[0].Path)
	if _, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate); err == nil || !*interrupted {
		t.Fatal("journal interruption not exercised")
	}

	claim.store.afterReplace = nil

	proof, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate)
	if err != nil || !proof.Successful {
		t.Fatalf("manual replay: %v", err)
	}

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || len(retained.Attempts) != 1 || len(retained.Attempts[0].Processes) != 1 {
		t.Fatal("manual replay duplicated validation")
	}
}

func interruptWorkflowManualResult(claim *WorkflowClaim, store *WorkflowStore, taskPath string) *bool {
	interrupted := false

	claim.store.afterReplace = func(index int) error {
		if index != 0 || interrupted {
			return nil
		}

		record, err := store.readRecord(taskPath)
		if err != nil {
			return err
		}

		if record.Task.Phase == "validated" {
			interrupted = true

			return fmt.Errorf("owned manual result interruption")
		}

		return nil
	}

	return &interrupted
}

// TestWorkflowManualFencing rejects a yielded token and automatic inputs through
// the manual result API, retaining the replacement claim and operation lock.
func TestWorkflowManualFencing(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, []string{"true"}, 30)
	run := workflowTestStart(t, store, request)
	old, task := workflowTestManual(t, store, run.Invocation.ID, nil)

	candidate := workflowManualTestCandidate(t, store, task)
	if err := store.CheckpointWorkflow(t.Context(), old); err != nil {
		t.Fatal(err)
	}

	replacement := workflowTestClaim(t, store, run.Invocation.ID)
	if _, err := store.ReconcileWorkflowTask(t.Context(), old, candidate); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("old owner accepted: %v", err)
	}

	if err := store.CheckWorkflowClaim(t.Context(), replacement); err != nil {
		t.Fatal("old owner released replacement")
	}

	if _, err := store.prepareWorkflowTask(t.Context(), replacement, []string{"true"}); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ExecuteWorkflowManualTask(t.Context(), replacement); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("automatic attempt executed manually")
	}
}
