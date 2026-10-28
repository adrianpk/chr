package workflow

import (
	"testing"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
)

// TestNativeMergerAmbiguity requests ownership rather than silently picking a
// source step when semantic findings span different original commits.
func TestNativeMergerAmbiguity(t *testing.T) {
	_, state := nativeMergerCleanReplay(t)
	_, input, now := nativeMergerInput(t)
	input.Repository, input.Source, input.Base, input.Method = state.Query.Repository, state.Query.SourceCommit, state.Query.BaseCommit, "rebase"
	input.AllowedPaths = []string{"source.txt", "next.txt"}
	input.Conflicts[0].Path = "source.txt"
	other := input.Conflicts[0]
	other.ID, other.Path = "conflict-2", "next.txt"
	input.Conflicts = append(input.Conflicts, other)

	input, err := workflowmerge.SealInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := PlanWorkflowMergerRepair(t.Context(), state.Query.Lane.Worktree, input, *state.Result, nil)
	if err != nil || plan.Outcome != "needs_decision" || len(plan.Decisions) != 1 || plan.OriginalCommit != "" {
		t.Fatal("ambiguous source ownership was silently chosen", err)
	}
}
