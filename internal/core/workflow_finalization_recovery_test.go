package core

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowFinalizationCrashProcess exits at an actual replacement in the
// journal containing the shared receipt, task fan-out, set and final checkpoint.
func TestWorkflowFinalizationCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_FINALIZATION_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id := os.Getenv("CHR_FINALIZATION_CRASH_ID")

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	proof, err := store.readWorkflowDeliveryProof(run)
	if err != nil {
		t.Fatal(err)
	}

	index, err := strconv.Atoi(os.Getenv("CHR_FINALIZATION_CRASH_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	store.afterReplace = func(replaced int) error {
		if replaced == index && workflowFinalizationJournal(t, store) {
			os.Exit(17)
		}

		return nil
	}
	provider := &workflowPRProviderFixture{pr: proof.PullRequest}
	git := NewWorkflowIntegrationGit(proof.Git.Repository, os.Getenv("CHR_FINALIZATION_CRASH_REMOTE"))
	_, err = store.FinalizeWorkflowSlice(t.Context(), id, provider, git)
	t.Fatalf("owned finalization writer did not exit: %v", err)
}

func workflowFinalizationJournal(t *testing.T, store *WorkflowStore) bool {
	t.Helper()

	path, err := store.safePath(workflowstate.JournalPath)
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

	return (run.Outcome == "slice_integrated" || run.Outcome == "ticket_integrated") && run.DeliveryRef != ""
}

// TestWorkflowFinalizationCrash proves replay once after real host death at
// receipt introduction, each task metadata write, set write and final checkpoint.
func TestWorkflowFinalizationCrash(t *testing.T) {
	for _, index := range []int{0, 1, 3, 5, 7, 9} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			fixture := workflowVerifiedDelivery(t, "rebase")
			crashWorkflowFinalization(t, fixture, index)

			restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

			receipt, err := restarted.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			checkWorkflowFinalizedSlice(t, fixture, receipt)

			before := workflowTestRead(t, restarted, fixture.run.ID)
			if _, err := restarted.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, nil, nil); err != nil {
				t.Fatal(err)
			}

			if workflowTestRead(t, restarted, fixture.run.ID).Revision != before.Revision {
				t.Fatal("replayed fan-out ran twice")
			}
		})
	}
}

func crashWorkflowFinalization(t *testing.T, fixture workflowIntegrationFixture, index int) {
	t.Helper()

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowFinalizationCrashProcess$")

	command.Env = append(os.Environ(), "CHR_FINALIZATION_CRASH_DIR="+fixture.store.layout.ProjectDir(), "CHR_FINALIZATION_CRASH_ID="+fixture.run.ID, "CHR_FINALIZATION_CRASH_INDEX="+strconv.Itoa(index), "CHR_FINALIZATION_CRASH_REMOTE="+fixture.remote)
	output, err := command.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("owned finalization crash: %v: %s", err, output)
	}
}
