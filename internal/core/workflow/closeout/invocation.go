package closeout

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SetID reads the owning set identity from its frozen contract path.
func SetID(image workflowstate.InputImage) string {
	return filepath.Base(filepath.Dir(image.Path))
}

// NextSet selects the first admitted set without confirmed closure.
func NextSet(run workflowstate.Invocation) string {
	if len(run.Closures) >= len(run.SetImages) {
		return ""
	}

	return SetID(run.SetImages[len(run.Closures)])
}

// NeedsClosing detects a selected set boundary before a successor starts.
func NeedsClosing(run workflowstate.Invocation) bool {
	if run.Selection.Selector.Extent != "slice_set" && run.Selection.Selector.Extent != "use_case" {
		return false
	}

	next := NextSet(run)

	return next != "" && (run.Cursor == len(run.Members) || run.Members[run.Cursor].Admission.SetID != next)
}

// AtFrontier requires confirmed progress without active integration bindings.
func AtFrontier(run workflowstate.Invocation, setID string) bool {
	return run.Lane == nil && run.Submission == nil && run.Cursor == run.ConfirmedCursor && NextSet(run) == setID
}

func runStopped(run workflowstate.Invocation, ownerRoot, root string) bool {
	return workflowstate.InvocationTerminal(run.Status) || run.StopIntent != "" || ownerRoot != root || run.Status == workflowstate.InvocationPaused
}

// CheckReady preserves ordered eligibility, process and current-input checks.
func CheckReady(ctx context.Context, request Readiness, guards RunGuard) error {
	run := request.Run
	if runStopped(run, request.OwnerRoot, request.Root) {
		return workflowstate.ErrStop
	}

	if !AtFrontier(run, request.SetID) {
		return fmt.Errorf("%w: invocation must reach this set's closing frontier", workflowstate.ErrOwned)
	}

	if run.Cursor < len(run.Members) && run.Members[run.Cursor].Admission.SetID == request.SetID {
		return fmt.Errorf("%w: selected set work remains", workflowstate.ErrConflict)
	}

	for _, attempt := range run.Attempts {
		if slices.Contains([]string{"running", "checkpointed"}, attempt.Outcome) {
			return workflowstate.ErrOwned
		}

		if err := guards.Attempt(attempt); err != nil {
			return err
		}
	}

	return guards.Inputs(ctx)
}

// ProjectInvocation applies one closing outcome to the supplied checkpoint.
func ProjectInvocation(run *workflowstate.Invocation, projection Projection) {
	run.ClosingSet = ""

	run.Status, run.Outcome = workflowstate.InvocationWaitingIntegration, "set_closure_required"
	switch projection.State.Outcome {
	case "running":
		run.ClosingSet, run.Status, run.Outcome = projection.SetID, workflowstate.InvocationRunning, "set_closing"
	case "completed":
		run.Closures = append(run.Closures, workflowstate.InvocationClosure{SetID: projection.SetID, Cursor: run.Cursor, Receipt: projection.State.Result, Target: *projection.Target})
		run.ExecutionTarget = projection.Target

		run.Status, run.Outcome = workflowstate.InvocationPending, "set_closed"
		if len(run.Closures) == len(run.SetImages) {
			run.Status, run.Outcome = workflowstate.InvocationCompleted, "sets_closed"
		} else if NeedsClosing(*run) {
			run.Status, run.Outcome = workflowstate.InvocationWaitingIntegration, "set_closure_required"
		}
	default:
		run.Outcome = projection.State.Outcome
	}
}
