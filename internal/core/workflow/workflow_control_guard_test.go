package workflow

import (
	"errors"
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowControlCAS proves two stale control submissions cannot both mutate
// one invocation, including acceptance before a continuation leaves exclusion.
func TestWorkflowControlCAS(t *testing.T) {
	for _, action := range []string{"pause", "continuation"} {
		t.Run(action, func(t *testing.T) {
			testWorkflowControlCAS(t, action)
		})
	}
}

// TestWorkflowGuardedAdmission rejects changed or foreign selected identities
// before retaining an invocation or changing the selected ticket image.
func TestWorkflowGuardedAdmission(t *testing.T) {
	for _, name := range []string{"revision", "identity", "zero"} {
		t.Run(name, func(t *testing.T) {
			testWorkflowGuardedAdmission(t, name)
		})
	}
}

func testWorkflowControlCAS(t *testing.T, action string) {
	t.Helper()

	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request).Invocation
	results := make(chan error, 2)

	for range 2 {
		go func() { results <- workflowCASControl(t, store, run, action) }()
	}

	accepted, conflicted := 0, 0

	for range 2 {
		err := <-results
		if err == nil {
			accepted++
		} else if errors.Is(err, workflowstate.ErrConflict) {
			conflicted++
		} else {
			t.Fatal(err)
		}
	}

	if accepted != 1 || conflicted != 1 {
		t.Fatalf("accepted %d; conflicts %d", accepted, conflicted)
	}

	current, err := store.InspectWorkflow(t.Context(), run.ID)
	if err != nil || current.Invocation.Cursor != 0 || len(current.Invocation.Attempts) != 0 {
		t.Fatal("control fabricated task execution", err)
	}
}

func testWorkflowGuardedAdmission(t *testing.T, name string) {
	t.Helper()

	store, hierarchy, request := workflowClaimFixture(t)
	selected := hierarchy.Tasks[0]
	request.Selected = &workflowstate.SelectionGuard{Key: selected.Task.Key, Revision: selected.Revision}

	switch name {
	case "revision":
		request.Selected.Revision++
	case "identity":
		request.Selected.Key = hierarchy.Tasks[1].Task.Key
	case "zero":
		request.Selected.Revision = 0
	}

	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale selected ticket was admitted", err)
	}

	invocationPage, err := store.InspectWorkflowPage(t.Context(), 0, 8)

	if err != nil || len(invocationPage.Invocations) != 0 {
		t.Fatal("rejected admission retained authorization", err)
	}

	current, err := store.Read(t.Context(), selected.Path)
	if err != nil || !reflect.DeepEqual(current.Task, selected.Task) || current.Revision != selected.Revision {
		t.Fatal("rejected admission changed a ticket", err)
	}
}

func workflowCASControl(t *testing.T, store *Store, run Invocation, action string) error {
	t.Helper()

	if action == "pause" {
		_, err := store.StopWorkflow(t.Context(), run.ID, "pause", run.Revision)

		return err
	}

	_, err := store.acceptWorkflowContinuation(t.Context(), run.ID, run.Revision)

	return err
}

// TestWorkflowTerminalControl keeps guarded continuation observational once
// cancellation is terminal, without a new acceptance checkpoint or attempt.
func TestWorkflowTerminalControl(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request).Invocation

	terminal, err := store.StopWorkflow(t.Context(), run.ID, "cancel", run.Revision)
	if err != nil {
		t.Fatal(err)
	}

	observed, err := store.acceptWorkflowContinuation(t.Context(), run.ID, terminal.Revision)
	if err != nil || !reflect.DeepEqual(observed, terminal) {
		t.Fatal("terminal control retained another acceptance", err)
	}
}
