package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeMergerAcceptanceFixture(t *testing.T, id, method string) (workflowPreparationFixture, MergerRequest, MergerHarness, MergerRepairState) {
	t.Helper()

	f, request, harness := nativeMergerExecution(t, id)
	request.Input.Method = method
	writeWorkspaceConfig(t, f.store.layout.ProjectDir(), "review:\n  mode: trusted_auto_merge\nwork:\n  merge_strategy: "+method+"\n")

	query := workflowstate.PreparationQuery{IntegrationQuery: workflowstate.IntegrationQuery{Repository: request.Input.Repository, TargetBranch: "dev", Method: method, SourceCommit: request.Input.Source, SourceTree: workflowIntegrationCommand(t, f.run.Lane.Worktree, "rev-parse", request.Input.Source+"^{tree}"), OriginCommit: f.run.Lane.BaseCommit, SourceDirectory: f.run.Lane.Worktree}, BaseCommit: f.base}
	query.Lane = workflowstate.Lane{Unit: request.Input.Unit, Branch: "refs/heads/chr/workflow/checkpoint", Worktree: filepath.Join(t.TempDir(), "checkpoint"), BaseCommit: f.base, HeadCommit: f.base}

	state, err := f.git.BeginWorkflowMergerRepair(t.Context(), query)
	if err != nil || state.Result != nil {
		t.Fatal("real conflict checkpoint missing", err)
	}

	ref, err := f.store.RetainWorkflowMergerCheckpoint(t.Context(), request.IntegrationBudgetGuard, state, f.git)
	if err != nil {
		t.Fatal("checkpoint retention failed", err)
	}

	request.Input.RepairCheckpointRef = &ref

	if method == "rebase" {
		checkpoint := nativeMergerReplayCheckpoint(t, state)
		payload := *checkpoint
		payload.Evidence = workflowstate.EvidenceRef{}
		checkpoint.Evidence = retainNativeRecordEvidence(t, f.store, f.run.ID, "host-checkpoint", payload)
		request.Input.HistoryCheckpoint = checkpoint
	}

	request.Input.Deadline = time.Now().UTC().Add(25 * time.Second)

	request.Input, err = workflowmerge.SealInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	return f, request, harness, state
}

// TestNativeMergerAcceptance uses real conflicts, a broker driver and trusted
// Git commits; host acceptance remains separate from validation and review.
func TestNativeMergerAcceptance(t *testing.T) {
	for _, method := range []string{"merge", "squash", "rebase"} {
		t.Run(method, func(t *testing.T) { testNativeMergerAcceptance(t, method) })
	}
}

func checkNativeMergerAcceptanceRetained(t *testing.T, f workflowPreparationFixture, expected MergerAttempt) {
	t.Helper()

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Outcome != "merger_repair_pending_validation" {
		t.Fatal("accepted repair was not left pending validation")
	}

	if _, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, f.provider); err == nil {
		t.Fatal("host acceptance substituted for gate/review")
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)
	current := run.Attempts[len(run.Attempts)-1]

	data, err := restarted.readWorkflowReviewImage(current.ResultRef, current.ResultHash)
	if err != nil {
		t.Fatal(err)
	}

	var retained MergerAttempt
	if err := json.Unmarshal(data, &retained); err != nil {
		t.Fatal(err)
	}

	if !workflowNativeJSONEqual(retained, expected) {
		t.Fatal("accepted/rejected evidence did not survive restart")
	}

	checkNativeMergerProcesses(t, current)
	checkNativeMergerBudget(t, f)
}

type nativeMergerAcceptanceGit struct {
	OSWorkflowPreparationGit
	beforeCheck func()
	afterCommit func()
}

func (git nativeMergerAcceptanceGit) CheckWorkflowMergerCheckpoint(ctx context.Context, state MergerRepairState) error {
	if git.beforeCheck != nil {
		git.beforeCheck()
	}

	return git.OSWorkflowPreparationGit.CheckWorkflowMergerCheckpoint(ctx, state)
}

func (git nativeMergerAcceptanceGit) ContinueWorkflowMergerRepair(ctx context.Context, state MergerRepairState, repair workflowstate.IntegrationTreeRepair) (MergerRepairState, error) {
	next, err := git.OSWorkflowPreparationGit.ContinueWorkflowMergerRepair(ctx, state, repair)
	if err == nil && git.afterCommit != nil {
		git.afterCommit()
	}

	return next, err
}

func nativeMergerProposalDirectory(t *testing.T, f workflowPreparationFixture, input workflowstate.MergerInput) string {
	t.Helper()

	common, err := projectcore.ResolveRepositoryDirectory(t.Context(), f.store.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Join(common, "chr-merger-workspaces", input.OwnedWorkspaceID, "checkout")
}

// TestNativeMergerAcceptanceReject checks actual worktree contents independently
// of a well-formed proposal, including undeclared files and whole-side choices.
func TestNativeMergerAcceptanceReject(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testing.T, workflowPreparationFixture, workflowstate.MergerInput, string)
	}{
		{"source side", func(t *testing.T, _ workflowPreparationFixture, input workflowstate.MergerInput, dir string) {
			content := workflowIntegrationCommand(t, dir, "show", input.Source+":result.txt")
			nativeMergerWrite(t, dir, "result.txt", content)
		}},
		{"target side", func(t *testing.T, _ workflowPreparationFixture, input workflowstate.MergerInput, dir string) {
			content := workflowIntegrationCommand(t, dir, "show", input.Base+":result.txt")
			nativeMergerWrite(t, dir, "result.txt", content)
		}},
		{"missing file", func(t *testing.T, _ workflowPreparationFixture, _ workflowstate.MergerInput, dir string) {
			if err := os.Remove(filepath.Join(dir, "result.txt")); err != nil {
				t.Fatal(err)
			}
		}},
		{"undeclared file", func(t *testing.T, _ workflowPreparationFixture, _ workflowstate.MergerInput, dir string) {
			nativeMergerWrite(t, dir, "extra.txt", "hidden edit")
		}},
		{"gate edit", func(t *testing.T, _ workflowPreparationFixture, _ workflowstate.MergerInput, dir string) {
			nativeMergerWrite(t, dir, "Makefile", "test:\n\ttrue\n")
		}},
		{"markers", func(t *testing.T, _ workflowPreparationFixture, _ workflowstate.MergerInput, dir string) {
			nativeMergerWrite(t, dir, "result.txt", "<<<<<<< ours\nA\n=======\nB\n>>>>>>> theirs\n")
		}},
		{"binary edit", func(t *testing.T, _ workflowPreparationFixture, _ workflowstate.MergerInput, dir string) {
			nativeMergerWrite(t, dir, "result.txt", "unexpected\x00binary")
		}},
		{"symlink", func(t *testing.T, _ workflowPreparationFixture, _ workflowstate.MergerInput, dir string) {
			if err := os.Remove(filepath.Join(dir, "result.txt")); err != nil {
				t.Fatal(err)
			}

			if err := os.Symlink("target.txt", filepath.Join(dir, "result.txt")); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { testNativeMergerAcceptanceReject(t, tc.mutate) })
	}
}

func nativeMergerWrite(t *testing.T, directory, path, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(directory, path), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestNativeMergerAcceptanceStale retains rejection if canonical B moves after
// private host commit, and never makes that private result publishable.
func TestNativeMergerAcceptanceStale(t *testing.T) {
	f, request, harness, _ := nativeMergerAcceptanceFixture(t, "merger-stale-accept", "merge")
	git := nativeMergerAcceptanceGit{OSWorkflowPreparationGit: f.git, afterCommit: func() { f.provider.target = strings.Repeat("f", 40) }}

	attempt, err := f.store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, git)
	if !errors.Is(err, workflowstate.ErrConflict) || attempt.Acceptance == nil || attempt.Acceptance.Outcome != "rejected" || attempt.Acceptance.Next == nil {
		t.Fatal("stale result accepted or private evidence lost", err)
	}

	if _, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, f.provider); err == nil {
		t.Fatal("stale result granted publication")
	}
}

func testNativeMergerAcceptance(t *testing.T, method string) {
	t.Helper()

	f, request, harness, state := nativeMergerAcceptanceFixture(t, "merger-accept", method)
	sourceStatus := workflowIntegrationCommand(t, f.run.Lane.Worktree, "status", "--porcelain")

	attempt, err := f.store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, f.git)
	if err != nil || attempt.Acceptance == nil || attempt.Acceptance.Outcome != "accepted_pending_validation" || attempt.Acceptance.Next == nil {
		t.Fatal("host acceptance failed", attempt.Failure, err)
	}

	checkNativeMergerAcceptedMethod(t, attempt, state, method, request.Input.Source)

	if workflowIntegrationCommand(t, f.remote, "rev-parse", "refs/heads/dev") != f.base || workflowIntegrationCommand(t, f.run.Lane.Worktree, "status", "--porcelain") != sourceStatus {
		t.Fatal("acceptance changed a canonical ref or worker tree")
	}

	checkNativeMergerAcceptanceRetained(t, f, attempt)
}

func testNativeMergerAcceptanceReject(t *testing.T, mutate func(*testing.T, workflowPreparationFixture, workflowstate.MergerInput, string)) {
	t.Helper()

	f, request, harness, state := nativeMergerAcceptanceFixture(t, "merger-reject", "merge")
	dir := nativeMergerProposalDirectory(t, f, request.Input)
	git := nativeMergerAcceptanceGit{OSWorkflowPreparationGit: f.git, beforeCheck: func() {
		if _, err := os.Stat(dir); err == nil {
			mutate(t, f, request.Input, dir)
		}
	}}

	attempt, err := f.store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, git)
	if !errors.Is(err, workflowstate.ErrConflict) || attempt.Acceptance == nil || attempt.Acceptance.Outcome != "rejected" || attempt.Acceptance.Next != nil {
		t.Fatal("invalid actual edit accepted", attempt.Failure, err)
	}

	if workflowIntegrationCommand(t, state.Query.Lane.Worktree, "rev-parse", "HEAD") != state.PreparedParent {
		t.Fatal("rejection committed a replay object")
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	latest := run.Attempts[len(run.Attempts)-1]

	data, readErr := f.store.readWorkflowReviewImage(latest.ResultRef, latest.ResultHash)
	if readErr != nil || !strings.Contains(string(data), "\"outcome\": \"rejected\"") {
		t.Fatal("rejection evidence was not retained", readErr)
	}
}

func checkNativeMergerAcceptedMethod(t *testing.T, attempt MergerAttempt, state MergerRepairState, method, source string) {
	t.Helper()

	next := attempt.Acceptance.Next
	if len(next.Repairs) != 1 || next.History[state.ReplayIndex].Source != state.CurrentSource {
		t.Fatal("host commit lost original replay attribution")
	}

	if method != "rebase" && (next.Result == nil || next.Result.ResultCommit == source) {
		t.Fatal("host committed no exact combined result")
	}
}
