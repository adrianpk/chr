package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNativeMergerReplaySteps retains each successive real conflict checkpoint
// and its exact prepared parent instead of squashing repairs into the tip.
func TestNativeMergerReplaySteps(t *testing.T) {
	f, query := nativeMergerRepairFixture(t, "rebase")
	workflowIntegrationCommand(t, f.request.Directory, "checkout", "-B", "dev", query.BaseCommit)

	query.BaseCommit = workflowIntegrationCommit(t, f.request.Directory, "suffix.txt", "Target suffix intent")
	query.Lane.BaseCommit, query.Lane.HeadCommit = query.BaseCommit, query.BaseCommit
	workflowIntegrationCommand(t, f.request.Directory, "push", f.remote, query.BaseCommit+":refs/heads/dev")
	workflowIntegrationCommand(t, f.request.Directory, "checkout", "source")

	git := NewWorkflowPreparationGit(query.Repository, f.remote)

	state, err := git.BeginWorkflowMergerRepair(t.Context(), query)
	if err != nil {
		t.Fatal(err)
	}

	for index, path := range []string{"next.txt", "suffix.txt"} {
		requireNativeMergerCheckpoint(t, state, index+2)

		if len(state.Conflicts) != 1 || state.Conflicts[0] != path {
			t.Fatal("repair skipped the original replay step")
		}

		repair := nativeMergerTreeRepair(t, state, path, "Source and target intent\n")

		state, err = git.ContinueWorkflowMergerRepair(t.Context(), state, repair)
		if err != nil {
			t.Fatal(err)
		}
	}

	if state.Result == nil || len(state.Result.Provenance.Repairs) != 2 || len(state.Result.History) != 4 {
		t.Fatal("successive conflict repairs lost original history")
	}

	checkNativeReplayMessages(t, query.Lane.Worktree, *state.Result)
	checkWorkflowGitRefs(t, f, query.SourceCommit, query.BaseCommit)
}

// TestNativeMergerRepairProof rejects repaired history whose host repair record
// was removed even when a caller recomputes the outer history hash.
func TestNativeMergerRepairProof(t *testing.T) {
	f, query := nativeMergerRepairFixture(t, "squash")
	git := NewWorkflowPreparationGit(query.Repository, f.remote)

	state, err := git.BeginWorkflowMergerRepair(t.Context(), query)
	if err != nil {
		t.Fatal(err)
	}

	repair := nativeMergerTreeRepair(t, state, state.Conflicts[0], "Both original intents\n")

	state, err = git.ContinueWorkflowMergerRepair(t.Context(), state, repair)
	if err != nil {
		t.Fatal(err)
	}

	proof := *state.Result
	proof.Provenance.Repairs = nil

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil {
		t.Fatal(err)
	}

	reader := workflowIntegrationReader{dir: query.Lane.Worktree}
	if reader.verifyPublicationHistory(t.Context(), proof) == nil {
		t.Fatal("recomputed hash hid a nondeterministic repaired tree")
	}

	checkWorkflowGitRefs(t, f, query.SourceCommit, query.BaseCommit)
}

// TestNativeMergerBinary rejects a text claim for an actual binary Git conflict
// before the host can turn the proposal into a prepared commit.
func TestNativeMergerBinary(t *testing.T) {
	f, query := nativeMergerRepairFixture(t, "squash")
	workflowIntegrationCommand(t, f.request.Directory, "checkout", "-B", "dev", query.BaseCommit)

	if err := os.WriteFile(filepath.Join(f.request.Directory, "shared.txt"), []byte("Target\x00binary intent"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, f.request.Directory, "add", "shared.txt")
	workflowIntegrationCommand(t, f.request.Directory, "commit", "-m", "test(integration): retain binary target intent")

	query.BaseCommit = workflowIntegrationCommand(t, f.request.Directory, "rev-parse", "HEAD")
	query.Lane.BaseCommit, query.Lane.HeadCommit = query.BaseCommit, query.BaseCommit
	workflowIntegrationCommand(t, f.request.Directory, "push", f.remote, query.BaseCommit+":refs/heads/dev")
	workflowIntegrationCommand(t, f.request.Directory, "checkout", "source")

	git := NewWorkflowPreparationGit(query.Repository, f.remote)

	state, err := git.BeginWorkflowMergerRepair(t.Context(), query)
	if err != nil || len(state.Conflicts) == 0 {
		t.Fatal("actual binary conflict checkpoint was lost", err)
	}

	repair := nativeMergerTreeRepair(t, state, state.Conflicts[0], "A text-only proposal\n")
	if _, err := git.ContinueWorkflowMergerRepair(t.Context(), state, repair); err == nil {
		t.Fatal("model text claim concealed an unsupported binary resolution")
	}

	checkWorkflowGitRefs(t, f, query.SourceCommit, query.BaseCommit)
}
