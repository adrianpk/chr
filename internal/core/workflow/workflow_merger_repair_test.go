package workflow

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeMergerRepairFixture(t *testing.T, method string) (workflowGitFixture, PreparationQuery) {
	t.Helper()

	f := newWorkflowGitFixture(t, method)
	path := "shared.txt"

	if method == "rebase" {
		f = nativeReplayFixture(t)
		path = "next.txt"
	} else {
		workflowIntegrationCommand(t, f.request.Directory, "checkout", "source")

		f.request.Git.SourceCommit = workflowIntegrationCommit(t, f.request.Directory, path, "Source intent")
	}

	workflowIntegrationCommand(t, f.request.Directory, "checkout", "-B", "dev", f.request.Git.BaseCommit)

	f.request.Git.BaseCommit = workflowIntegrationCommit(t, f.request.Directory, path, "Target intent")
	workflowIntegrationCommand(t, f.request.Directory, "checkout", "source")

	if method == "rebase" {
		f.request.Git.SourceCommit = workflowIntegrationCommit(t, f.request.Directory, "suffix.txt", "Later intent")
	}

	f.request.Git.SourceTree = workflowIntegrationCommand(t, f.request.Directory, "rev-parse", f.request.Git.SourceCommit+"^{tree}")
	workflowIntegrationCommand(t, f.request.Directory, "push", f.remote, f.request.Git.SourceCommit+":"+f.request.SourceRef, f.request.Git.BaseCommit+":refs/heads/dev")

	query := PreparationQuery{IntegrationQuery: f.request.Git.IntegrationQuery, BaseCommit: f.request.Git.BaseCommit}
	query.SourceDirectory = f.request.Directory
	query.Lane = workflowstate.Lane{Unit: workflowstate.Ref{Kind: "slice", ID: "fixture"}, Branch: "refs/heads/chr/workflow/repair", Worktree: filepath.Join(t.TempDir(), "repair"), BaseCommit: query.BaseCommit, HeadCommit: query.BaseCommit}

	return f, query
}

func nativeMergerTreeRepair(t *testing.T, state MergerRepairState, path, content string) workflowstate.IntegrationTreeRepair {
	t.Helper()

	_, input, now := nativeMergerInput(t)
	input.Repository, input.Source, input.Base, input.Method = state.Query.Repository, state.Query.SourceCommit, state.Query.BaseCommit, state.Query.Method
	input.AllowedPaths = []string{path}

	input.Conflicts[0].Path = path
	if len(state.Conflicts) == 0 {
		input.Reason, input.Conflicts[0].Kind = "actionable_review_finding", "semantic"
		input.FindingRefs = slices.Clone(input.IntentRefs)
	}

	if state.Query.Method == "rebase" {
		input.HistoryCheckpoint = nativeMergerReplayCheckpoint(t, state)
	}

	input, err := SealWorkflowMergerInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	proposal := nativeMergerResolved(input)

	proposal.ChangedPaths, proposal.ResolutionNotes[0].Path = []string{path}, path
	if err := os.WriteFile(filepath.Join(state.Query.Lane.Worktree, path), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, state.Query.Lane.Worktree, "add", "--", path)

	tree := workflowIntegrationCommand(t, state.Query.Lane.Worktree, "write-tree")

	return workflowstate.IntegrationTreeRepair{Input: input, Proposal: proposal, CompletedAt: now.Add(time.Millisecond), OriginalParent: state.OriginalParent, PreparedParent: state.PreparedParent, Source: state.CurrentSource, BaselineTree: state.Tree, ResolvedTree: tree}
}

func nativeMergerReplayCheckpoint(t *testing.T, state MergerRepairState) *workflowstate.MergerCheckpoint {
	t.Helper()

	checkpoint := &workflowstate.MergerCheckpoint{ReplayIndex: state.ReplayIndex, OriginalCommit: state.CurrentSource, PreparedParent: state.PreparedParent, Evidence: workflowstate.EvidenceRef{}}
	_, template, _ := nativeMergerInput(t)
	checkpoint.Evidence = template.IntentRefs[0]

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: state.Query.Lane.Worktree}}
	for _, entry := range state.History {
		source, err := reader.commitAttribution(t.Context(), entry.Source)
		if err != nil {
			t.Fatal(err)
		}

		prepared, err := reader.commitAttribution(t.Context(), entry.Commit)
		if err != nil {
			t.Fatal(err)
		}

		decision, err := reader.replayEmptyDecision(t.Context(), source, prepared)
		if err != nil {
			t.Fatal(err)
		}

		checkpoint.Predecessors = append(checkpoint.Predecessors, workflowstate.IntegrationReplay{Source: entry.Source, Commit: entry.Commit, EmptyDecision: decision})
	}

	return checkpoint
}

// TestNativeMergerRepair preserves the original replay step, exact empty prefix
// and subsequent intent while merge and squash retain their configured parents.
func TestNativeMergerRepair(t *testing.T) {
	for _, method := range []string{"merge", "squash", "rebase"} {
		t.Run(method, func(t *testing.T) {
			f, query := nativeMergerRepairFixture(t, method)
			git := NewWorkflowPreparationGit(query.Repository, f.remote)
			before := workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD")

			state, err := git.BeginWorkflowMergerRepair(t.Context(), query)
			if err != nil || state.Result != nil || len(state.Conflicts) != 1 {
				t.Fatal("conflict checkpoint was lost", err)
			}

			prefix := slices.Clone(state.History)
			repair := nativeMergerTreeRepair(t, state, state.Conflicts[0], "Source intent and target intent\n")

			state, err = git.ContinueWorkflowMergerRepair(t.Context(), state, repair)
			if err != nil || state.Result == nil || !slices.Equal(state.History[:len(prefix)], prefix) {
				t.Fatal("original replay prefix changed or repair failed", err)
			}

			checkNativeMergerRepairedProof(t, state)
			checkWorkflowGitRefs(t, f, query.SourceCommit, query.BaseCommit)

			if workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD") != before || workflowIntegrationCommand(t, f.request.Directory, "status", "--porcelain") != "" {
				t.Fatal("repair changed the source checkout")
			}
		})
	}
}

func checkNativeMergerRepairedProof(t *testing.T, state MergerRepairState) {
	t.Helper()

	proof := *state.Result
	if validateWorkflowIntegrationProof(proof.IntegrationQuery, proof) != nil || len(proof.Provenance.Repairs) != 1 {
		t.Fatal("repair provenance was not bound to final method proof")
	}

	if proof.Method != "rebase" {
		checkNativeCombinedObjects(t, state.Query.Lane.Worktree, proof)

		return
	}

	if len(proof.History) != 4 || proof.Provenance.Replays[0].EmptyDecision != "already_applied" || proof.Provenance.Replays[1].EmptyDecision != "source_empty" {
		t.Fatal("repair collapsed original empty intent or later source step")
	}

	checkNativeReplayMessages(t, state.Query.Lane.Worktree, proof)
}

func requireNativeMergerResult(t *testing.T, state MergerRepairState, err error) {
	t.Helper()

	if err != nil || state.Result == nil {
		t.Fatal("trusted repair has no complete method proof", err)
	}
}

func requireNativeMergerCheckpoint(t *testing.T, state MergerRepairState, index int) {
	t.Helper()

	if state.Result != nil || state.ReplayIndex != index {
		t.Fatal("repair skipped the original replay checkpoint")
	}
}

// TestNativeMergerRepairGuards rejects incorrect checkpoint identity and an
// undeclared actual diff before the trusted host creates a replay commit.
func TestNativeMergerRepairGuards(t *testing.T) {
	cases := []struct {
		name   string
		change func(*MergerRepairState, *workflowstate.IntegrationTreeRepair)
	}{
		{"original step", func(_ *MergerRepairState, r *workflowstate.IntegrationTreeRepair) { r.Source = r.Input.Source }},
		{"prepared parent", func(_ *MergerRepairState, r *workflowstate.IntegrationTreeRepair) { r.PreparedParent = r.Input.Base }},
		{"baseline", func(_ *MergerRepairState, r *workflowstate.IntegrationTreeRepair) { r.BaselineTree = r.ResolvedTree }},
		{"actual diff", func(_ *MergerRepairState, r *workflowstate.IntegrationTreeRepair) { r.Proposal.ChangedPaths = nil }},
		{"prefix object", func(s *MergerRepairState, _ *workflowstate.IntegrationTreeRepair) {
			s.History[0].Commit = s.Query.BaseCommit
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, query := nativeMergerRepairFixture(t, "rebase")
			git := NewWorkflowPreparationGit(query.Repository, f.remote)

			state, err := git.BeginWorkflowMergerRepair(t.Context(), query)
			if err != nil {
				t.Fatal(err)
			}

			repair := nativeMergerTreeRepair(t, state, state.Conflicts[0], "Both intents\n")
			tc.change(&state, &repair)

			if result, err := git.ContinueWorkflowMergerRepair(t.Context(), state, repair); err == nil || result.Result != nil {
				t.Fatal("invalid repair produced a host result")
			}

			checkWorkflowGitRefs(t, f, query.SourceCommit, query.BaseCommit)
		})
	}
}
