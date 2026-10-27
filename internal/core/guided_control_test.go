package core

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCaseControlCAS accepts one concurrent control, retains its receipt
// and prevents further input until an explicit revision-bound resume.
func TestGuidedUseCaseControlCAS(t *testing.T) {
	for _, action := range []string{"pause", "cancel"} {
		t.Run(action, func(t *testing.T) { testGuidedControlCAS(t, action) })
	}
}

func testGuidedControlCAS(t *testing.T, action string) {
	t.Helper()

	root, session := guidedCaptureFixture(t)
	scope := authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}
	command := authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: scope, ExpectedRevision: session.Revision, Action: action}

	var calls atomic.Int32

	apply := func(context.Context, *authoringcore.Session) error { calls.Add(1); return nil }

	var wait sync.WaitGroup

	results := make(chan error, 2)

	for _, id := range []string{"control-one", "control-two"} {
		request := command
		request.OperationID = id

		wait.Add(1)
		go func() {
			defer wait.Done()

			_, err := ApplyGuidedUseCaseControl(t.Context(), root, request, apply)
			results <- err
		}()
	}

	wait.Wait()
	close(results)
	assertGuidedControlRace(t, results)

	current, err := ReadGuidedUseCase(t.Context(), root, scope)
	if err != nil || calls.Load() != 1 || GuidedDispatchAllowed(current) {
		t.Fatal("concurrent control repeated owner effects or allowed dispatch", err)
	}

	command.OperationID = current.Guided.Control.CompletedOperations[0].ID

	replay, err := ApplyGuidedUseCaseControl(t.Context(), root, command, apply)
	if err != nil || replay.Revision != current.Revision || calls.Load() != 1 {
		t.Fatal("control replay repeated its effect", err)
	}

	assertGuidedStoppedInput(t, root, current, command)
}

func assertGuidedControlRace(t *testing.T, results <-chan error) {
	t.Helper()

	successes, conflicts := 0, 0

	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, workflowstate.ErrConflict):
			conflicts++
		default:
			t.Fatal(err)
		}
	}

	if successes != 1 || conflicts != 1 {
		t.Fatalf("controls: %d accepted, %d conflicts", successes, conflicts)
	}
}

func assertGuidedStoppedInput(t *testing.T, root string, session authoringcore.Session, command authoringcore.GuidedUseCaseCommand) {
	t.Helper()

	command.ExpectedRevision, command.OperationID, command.Action, command.Content = session.Revision, "stopped-input", "append", "Cannot enter the frozen scope."
	if _, err := MutateGuidedUseCase(t.Context(), root, command); !errors.Is(err, ErrGuidedUseCaseStopped) {
		t.Fatal("stopped run accepted new input", err)
	}

	command.OperationID, command.Action, command.Content = "resume", "resume", ""

	resumed, err := ApplyGuidedUseCaseControl(t.Context(), root, command, func(context.Context, *authoringcore.Session) error { return nil })
	if session.Guided.Status == "cancelled" {
		if !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal("cancelled run resumed", err)
		}

		return
	}

	if err != nil || resumed.Guided.Status != "waiting_input" || !GuidedDispatchAllowed(resumed) {
		t.Fatal("resume lost original capture state", err)
	}
}

// TestGuidedUseCaseControlUncertain retains stop intent after an ambiguous
// owner error. Only a verified rejection allows a fresh control request.
func TestGuidedUseCaseControlUncertain(t *testing.T) {
	for _, rejected := range []bool{false, true} {
		name := "uncertain"
		if rejected {
			name = "rejected"
		}

		t.Run(name, func(t *testing.T) { testGuidedUncertainControl(t, rejected) })
	}
}

func testGuidedUncertainControl(t *testing.T, rejected bool) {
	t.Helper()

	root, session := guidedCaptureFixture(t)
	command := authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}, ExpectedRevision: session.Revision, OperationID: "original-stop", Action: "pause"}

	cause := workflowstate.ErrConflict
	if rejected {
		cause = errors.Join(cause, ErrGuidedControlRejected)
	}

	retained, err := ApplyGuidedUseCaseControl(t.Context(), root, command, func(context.Context, *authoringcore.Session) error { return cause })
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("owner failure was lost", err)
	}

	if rejected {
		if retained.Guided.Control.StopIntent != "" || retained.Guided.Control.PendingOperation != nil {
			t.Fatal("verified rejection blocked fresh controls")
		}

		return
	}

	if retained.Guided.Control.StopIntent != "pause" || retained.Guided.Control.PendingOperation == nil {
		t.Fatal("ambiguous owner failure discarded stop intent")
	}

	assertGuidedControlReconciled(t, root, command)
}

func assertGuidedControlReconciled(t *testing.T, root string, command authoringcore.GuidedUseCaseCommand) {
	t.Helper()

	recovered, err := ApplyGuidedUseCaseControl(t.Context(), root, command, func(context.Context, *authoringcore.Session) error { return nil })
	if err != nil || recovered.Guided.Status != "paused" || recovered.Guided.Control.PendingOperation != nil {
		t.Fatal("original uncertain control did not reconcile", err)
	}
}
