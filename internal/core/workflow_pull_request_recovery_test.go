package core

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// TestWorkflowSlicePRCrashProcess is an owned writer that exits after provider
// creation or a selected actual replacement in the review-binding journal.
func TestWorkflowSlicePRCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_PR_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id, stage := os.Getenv("CHR_PR_CRASH_ID"), os.Getenv("CHR_PR_CRASH_STAGE")

	candidate, err := store.ReadWorkflowSliceCandidate(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	provider := &workflowPRProviderFixture{pr: WorkflowPullRequestSnapshot{URL: "https://forge.example/test/repo/pulls/7", Repository: candidate.Target.Repository, HeadRepository: candidate.Target.Repository, HeadBranch: strings.TrimPrefix(candidate.Lane.Branch, "refs/heads/"), BaseBranch: "dev", HeadCommit: candidate.Result.Commit, BaseCommit: candidate.Target.Commit, State: "open"}}
	provider.createHook = func() {
		data, err := json.Marshal(provider.pr)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(os.Getenv("CHR_PR_CRASH_REMOTE"), data, 0o600); err != nil {
			t.Fatal(err)
		}

		if stage == "created" {
			os.Exit(17)
		}
	}

	index, err := strconv.Atoi(os.Getenv("CHR_PR_CRASH_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	store.afterReplace = func(replaced int) error {
		if replaced == index && workflowPRCrashJournalState(t, store, stage) {
			os.Exit(17)
		}

		return nil
	}
	_, err = store.EnsureWorkflowSlicePullRequest(t.Context(), id, workflowPRContent(), provider)
	t.Fatalf("owned PR writer did not exit: %v", err)
}

func workflowPRCrashJournalState(t *testing.T, store *WorkflowStore, state string) bool {
	t.Helper()

	path, err := store.safePath(workflowJournalPath)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var journal workflowJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		t.Fatal(err)
	}

	var run workflowInvocationFile
	if err := json.Unmarshal(journal.Changes[len(journal.Changes)-1].After, &run); err != nil {
		return false
	}

	return run.PullRequest != nil && run.PullRequest.State == state
}

// TestWorkflowSlicePRCrash recovers the sole provider PR after host death, and
// replays every review-task/set/observation/state boundary without duplicate IO.
func TestWorkflowSlicePRCrash(t *testing.T) {
	cases := []struct {
		stage string
		index int
	}{{"created", 0}, {"bound", 0}, {"bound", 2}, {"bound", 4}, {"bound", 6}, {"bound", 8}, {"bound", 9}}
	for _, tc := range cases {
		t.Run(tc.stage+strconv.Itoa(tc.index), func(t *testing.T) {
			store, hierarchy, run, provider := workflowPRFixture(t)
			remote := t.TempDir() + "/provider-pr.json"
			workflowPRCrash(t, store, run.ID, remote, tc.stage, tc.index)

			data, err := os.ReadFile(remote)
			if err != nil {
				t.Fatal(err)
			}

			if err := json.Unmarshal(data, &provider.pr); err != nil {
				t.Fatal(err)
			}

			provider.found = []WorkflowPullRequestSnapshot{provider.pr}
			restarted := NewWorkflowStore(store.layout, store.projectID)

			pr := workflowPREnsure(t, restarted, run.ID, provider)
			if provider.creates != 0 {
				t.Fatal("host crash created a replacement PR")
			}

			checkWorkflowPRBinding(t, restarted, hierarchy, run.ID, pr)
		})
	}
}

func workflowPRCrash(t *testing.T, store *WorkflowStore, id, remote, stage string, index int) {
	t.Helper()

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), program, "-test.run=^TestWorkflowSlicePRCrashProcess$")

	command.Env = append(os.Environ(), "CHR_PR_CRASH_DIR="+store.layout.ProjectDir(), "CHR_PR_CRASH_ID="+id, "CHR_PR_CRASH_REMOTE="+remote, "CHR_PR_CRASH_STAGE="+stage, "CHR_PR_CRASH_INDEX="+strconv.Itoa(index))

	output, err := command.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("owned writer: %v: %s", err, output)
	}
}
