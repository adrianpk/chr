package workflow

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeMergerCleanReplay(t *testing.T) (OSWorkflowPreparationGit, MergerRepairState) {
	t.Helper()

	f := nativeReplayFixture(t)
	query := workflowstate.PreparationQuery{IntegrationQuery: f.request.Git.IntegrationQuery, BaseCommit: f.request.Git.BaseCommit}
	query.SourceDirectory = f.request.Directory
	query.Lane = workflowstate.Lane{Unit: workflowstate.Ref{Kind: "slice", ID: "fixture"}, Branch: "refs/heads/chr/workflow/repair", Worktree: filepath.Join(t.TempDir(), "repair"), BaseCommit: query.BaseCommit, HeadCommit: query.BaseCommit}
	git := NewWorkflowPreparationGit(query.Repository, f.remote)
	git.committerTime = time.Unix(1700000000, 0).UTC()

	state, err := git.BeginWorkflowMergerRepair(t.Context(), query)
	if err != nil || state.Result == nil || len(state.Repairs) != 0 {
		t.Fatal("clean replay incorrectly required merger repair", err)
	}

	return git, state
}

// TestNativeMergerSemantic rewinds the unique owning source step and rebuilds
// its original attribution instead of appending an unexplained fix commit.
func TestNativeMergerSemantic(t *testing.T) {
	git, complete := nativeMergerCleanReplay(t)
	git.committerTime = time.Unix(1800000000, 0).UTC()
	_, input, now := nativeMergerInput(t)
	input.Repository, input.Source, input.Base, input.Method = complete.Query.Repository, complete.Query.SourceCommit, complete.Query.BaseCommit, "rebase"
	input.AllowedPaths, input.Conflicts[0].Path = []string{"next.txt"}, "next.txt"

	input, err := workflowmerge.SealInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := PlanWorkflowMergerRepair(t.Context(), complete.Query.Lane.Worktree, input, *complete.Result, nil)
	checkNativeMergerSemanticPlan(t, plan, err, complete.Query.SourceCommit)

	query := complete.Query
	query.SourceDirectory, query.Lane.Worktree = complete.Query.Lane.Worktree, filepath.Join(t.TempDir(), "semantic")

	state, err := git.BeginWorkflowMergerSemanticRepair(t.Context(), query, *complete.Result, plan.OriginalCommit)
	if err != nil || state.ReplayIndex != 2 {
		t.Fatal("semantic checkpoint differs from original step", err)
	}

	repair := nativeMergerTreeRepair(t, state, "next.txt", "Required source and target behavior\n")

	state, err = git.ContinueWorkflowMergerRepair(t.Context(), state, repair)
	requireNativeMergerResult(t, state, err)

	if len(state.Result.Provenance.IntegrationFixes) != 0 {
		t.Fatal("semantic replay became an unexplained integration fix", err)
	}

	checkNativeReplayMessages(t, query.Lane.Worktree, *state.Result)

	if state.History[0] != complete.History[0] || state.History[1] != complete.History[1] {
		t.Fatal("semantic repair changed its verified prefix")
	}
}

func checkNativeMergerSemanticPlan(t *testing.T, plan workflowstate.MergerRepairPlan, err error, original string) {
	t.Helper()

	if err != nil || plan.Outcome != "replay_step" || plan.OriginalCommit != original {
		t.Fatal("semantic finding lost its original replay owner", err)
	}
}

// TestNativeMergerFix requires an actionable explicit finding when no original
// source step owns the repair, and preserves every original replay mapping.
func TestNativeMergerFix(t *testing.T) {
	git, state := nativeMergerCleanReplay(t)
	before := *state.Result
	_, input, now := nativeMergerInput(t)
	input.Repository, input.Source, input.Base, input.Method = state.Query.Repository, state.Query.SourceCommit, state.Query.BaseCommit, "rebase"
	input.AllowedPaths, input.Conflicts[0].Path = []string{"repair.txt"}, "repair.txt"

	input, err := workflowmerge.SealInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	decision, err := PlanWorkflowMergerRepair(t.Context(), state.Query.Lane.Worktree, input, before, nil)
	if err != nil || decision.Outcome != "needs_decision" || len(decision.Decisions) != 1 {
		t.Fatal("unowned repair silently invented a replay owner", err)
	}

	finding := workflowstate.ReviewFinding{Category: workflowstate.AgenticReviewFindingCorrectness, Title: "Retain integration output", Detail: "Combined validation requires this explicit repair output."}

	plan, err := PlanWorkflowMergerRepair(t.Context(), state.Query.Lane.Worktree, input, before, &finding)
	if err != nil || plan.Outcome != "integration_fix" {
		t.Fatal("declared actionable integration fix was rejected", err)
	}

	if err := os.WriteFile(filepath.Join(state.Query.Lane.Worktree, "repair.txt"), []byte("Required output\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, state.Query.Lane.Worktree, "add", "repair.txt")

	tree := workflowIntegrationCommand(t, state.Query.Lane.Worktree, "write-tree")

	state, err = git.PrepareWorkflowMergerIntegrationFix(t.Context(), state, tree, *plan.IntegrationFix)
	requireNativeMergerResult(t, state, err)

	if !workflowNativeJSONEqual(before.History, state.Result.History) || len(state.Result.Provenance.IntegrationFixes) != 1 {
		t.Fatal("explicit fix lost source replay provenance", err)
	}
}
