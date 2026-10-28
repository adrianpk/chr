package workflow

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPreparationCrashProcess is the owned writer killed during real
// intent, fenced validation input or accepted-result journal replacement.
func TestWorkflowPreparationCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_PREPARATION_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id, stage := os.Getenv("CHR_PREPARATION_CRASH_ID"), os.Getenv("CHR_PREPARATION_CRASH_STAGE")

	index, err := strconv.Atoi(os.Getenv("CHR_PREPARATION_CRASH_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	observation, err := store.readWorkflowPullRequestObservation(*run.PullRequest)
	if err != nil {
		t.Fatal(err)
	}

	pr := observation.Snapshots[0]
	remote := os.Getenv("CHR_PREPARATION_CRASH_REMOTE")
	pr.BaseCommit = workflowIntegrationCommand(t, remote, "rev-parse", "refs/heads/dev")
	provider := &workflowPreparationProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}, target: pr.BaseCommit}, guarded: true, revision: "guard-v1"}
	store.afterReplace = func(replaced int) error {
		if replaced == index && workflowPreparationCrashReady(t, store, id, stage) {
			os.Exit(17)
		}

		return nil
	}
	_, err = store.PrepareWorkflowSliceIntegration(t.Context(), id, provider, NewWorkflowPreparationGit(pr.Repository, remote))
	t.Fatalf("owned preparation writer did not terminate: %v", err)
}

func workflowPreparationCrashReady(t *testing.T, store *Store, id, stage string) bool {
	t.Helper()

	if stage == "intent" {
		return true
	}

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	record, err := store.readRecord(run.Members[run.Cursor-1].Admission.Path)
	if err != nil {
		t.Fatal(err)
	}

	if stage == "claim" {
		attempt := run.Attempts[len(run.Attempts)-1]
		_, err := os.Stat(store.layout.Root() + "/" + workflowAttemptFile(run.ID, attempt.ID, "input.json"))

		return record.Task.AttemptID != "" && attempt.Purpose == workflowIntegrationValidationPurpose && err == nil
	}

	return record.Task.AttemptID == "" && run.Preparation != nil && (run.Preparation.State == "validating" || run.Preparation.State == "validated")
}

func workflowPreparationCrash(t *testing.T, f workflowPreparationFixture, stage string, index int) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowPreparationCrashProcess$")

	cmd.Env = append(os.Environ(), "CHR_PREPARATION_CRASH_DIR="+f.store.layout.ProjectDir(), "CHR_PREPARATION_CRASH_ID="+f.run.ID, "CHR_PREPARATION_CRASH_STAGE="+stage, "CHR_PREPARATION_CRASH_INDEX="+strconv.Itoa(index), "CHR_PREPARATION_CRASH_REMOTE="+f.remote)
	output, err := cmd.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 17 {
		t.Fatalf("owned preparation crash: %v: %s", err, output)
	}
}

// TestWorkflowPreparationJournal recovers real process death at preparation
// intent, input and result fan-out. Accepted validation replays exactly once;
// an unstarted fenced gate resumes without resetting task proofs or the bound.
func TestWorkflowPreparationJournal(t *testing.T) {
	cases := []struct {
		stage   string
		indices []int
	}{{"intent", []int{0, 2}}, {"claim", []int{2, 3}}, {"result", []int{0, 2, 3, 4}}}
	for _, test := range cases {
		for _, index := range test.indices {
			t.Run(test.stage+" "+strconv.Itoa(index), func(t *testing.T) { testWorkflowPreparationJournal(t, test.stage, index) })
		}
	}
}

func testWorkflowPreparationJournal(t *testing.T, stage string, index int) {
	t.Helper()

	f := newWorkflowPreparationFixture(t, "rebase", workflowstate.Validation{WaiverReason: "Preparation crash fixture"})
	workflowPreparationCrash(t, f, stage, index)

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	run, err := restarted.ReconcileWorkflow(t.Context(), f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowPreparationRecovery(t, restarted, run, stage)

	before := workflowTestRead(t, restarted, f.run.ID)

	candidate, err := restarted.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git)
	if err != nil {
		t.Fatal(err)
	}

	after := workflowTestRead(t, restarted, f.run.ID)
	if stage == "result" && (after.Revision != before.Revision || candidate.Round != 1) {
		t.Fatal("result replay repeated combined validation")
	}

	if stage != "result" && candidate.Round != 2 {
		t.Fatal("interrupted preparation reset its retained bound")
	}

	assertWorkflowPreparationSource(t, f)
}

func checkWorkflowPreparationRecovery(t *testing.T, store *Store, run Invocation, stage string) {
	t.Helper()

	if stage != "result" {
		if run.Status != workflowstate.InvocationPaused || run.ResumeStatus != workflowstate.InvocationWaitingIntegration || run.Preparation.State != "stale" {
			t.Fatal("killed gate lost its retained frontier")
		}

		if _, err := store.ResumeWorkflow(t.Context(), run.ID); err != nil {
			t.Fatal(err)
		}
	}
}
