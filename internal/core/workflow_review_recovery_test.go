package core

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// TestWorkflowReviewCrashProcess is the owned child writer used to terminate at
// a real input, claim, approval or role-decision journal replacement.
func TestWorkflowReviewCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_REVIEW_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id, stage := os.Getenv("CHR_REVIEW_CRASH_ID"), os.Getenv("CHR_REVIEW_CRASH_STAGE")

	index, err := strconv.Atoi(os.Getenv("CHR_REVIEW_CRASH_INDEX"))
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

	provider := &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: observation.Snapshots[0]}, target: observation.Snapshots[0].BaseCommit}

	store.afterReplace = func(replaced int) error {
		if replaced == index && workflowReviewCrashReady(t, store, id, stage) {
			os.Exit(17)
		}

		return nil
	}
	if stage == "approval" {
		_, err = store.ApproveWorkflowSlice(t.Context(), id, run.Review.InputHash, provider)
	} else {
		_, err = store.ReviewWorkflowSlice(t.Context(), id, provider, workflowReviewWorker("pass", "", ""))
	}

	t.Fatalf("owned review writer did not terminate: %v", err)
}

func workflowReviewCrashReady(t *testing.T, store *WorkflowStore, id, stage string) bool {
	t.Helper()

	if stage == "input" || stage == "approval" || stage == "repair" {
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
		return record.Task.AttemptID != ""
	}

	first, err := store.readRecord(run.Members[0].Admission.Path)
	if err != nil {
		t.Fatal(err)
	}

	return first.Task.Status == TicketStatusReadyForMerge && run.Attempts[len(run.Attempts)-1].Purpose == workflowSliceReviewPurpose
}

func workflowReviewCrash(t *testing.T, store *WorkflowStore, id, stage string, index int) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowReviewCrashProcess$")

	cmd.Env = append(os.Environ(), "CHR_REVIEW_CRASH_DIR="+store.layout.ProjectDir(), "CHR_REVIEW_CRASH_ID="+id, "CHR_REVIEW_CRASH_STAGE="+stage, "CHR_REVIEW_CRASH_INDEX="+strconv.Itoa(index))
	output, err := cmd.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 17 {
		t.Fatalf("owned review crash: %v: %s", err, output)
	}
}

// TestWorkflowReviewJournal replays actual host death at each approval fan-out
// and immutable input write, without a second approval or task execution.
func TestWorkflowReviewJournal(t *testing.T) {
	for _, stage := range []string{"input", "approval"} {
		indices := []int{0, 1}
		if stage == "approval" {
			indices = []int{0, 2, 4, 6, 8, 9}
		}

		for _, index := range indices {
			t.Run(stage+" "+strconv.Itoa(index), func(t *testing.T) {
				testWorkflowReviewJournal(t, stage, index)
			})
		}
	}
}

// TestWorkflowReviewRoleJournal recovers a killed fenced role claim and replays
// every accepted decision replacement without repeating a completed review.
func TestWorkflowReviewRoleJournal(t *testing.T) {
	for _, stage := range []string{"claim", "decision"} {
		indices := []int{0, 2, 3}
		if stage == "decision" {
			indices = []int{0, 2, 4, 6, 8, 9, 10}
		}

		for _, index := range indices {
			t.Run(stage+" "+strconv.Itoa(index), func(t *testing.T) {
				testWorkflowReviewRoleJournal(t, stage, index)
			})
		}
	}
}

// TestWorkflowReviewRepairJournal proves that a killed automatic repair consumes
// its budget exactly once when the task/set/run fan-out is replayed.
func TestWorkflowReviewRepairJournal(t *testing.T) {
	for _, index := range []int{0, 8} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			store, hierarchy, run, provider := workflowReviewFixture(t, ReviewWorkflowModeAgenticReview)
			workflowReviewRepairCheckpoint(t, store, run.ID, hierarchy.Tasks[2].Task.Key, provider)
			workflowReviewCrash(t, store, run.ID, "repair", index)

			retained, err := store.ReconcileWorkflow(t.Context(), run.ID)
			if err != nil {
				t.Fatal(err)
			}

			if retained.Status != "pending" || retained.Cursor != 2 || len(retained.ReviewRepairs) != 1 || retained.ReviewRepairs[0].Consumed != 1 {
				t.Fatal("repair replay lost its budget or ordered frontier")
			}

			again, err := store.ReconcileWorkflow(t.Context(), run.ID)
			if err != nil || again.Revision != retained.Revision || again.ReviewRepairs[0].Consumed != 1 {
				t.Fatal("repair replay repeated consumption")
			}
		})
	}
}

func workflowReviewRepairCheckpoint(t *testing.T, store *WorkflowStore, id, key string, provider *workflowReviewProviderFixture) {
	t.Helper()

	interrupted := errors.New("fixture stops after accepted review and before repair")
	store.afterReplace = func(int) error {
		run, err := store.readWorkflowInvocation(id)
		if err != nil {
			return err
		}

		if run.Review != nil && run.Review.State == "needs_repair" {
			return interrupted
		}

		return nil
	}
	_, err := store.ReviewWorkflowSlice(t.Context(), id, provider, workflowReviewWorker("needs_changes", key, ""))
	store.afterReplace = nil

	if !errors.Is(err, interrupted) {
		t.Fatalf("repair checkpoint: %v", err)
	}

	if _, err := store.ReconcileWorkflow(t.Context(), id); err != nil {
		t.Fatal(err)
	}
}

func testWorkflowReviewJournal(t *testing.T, stage string, index int) {
	t.Helper()

	store, _, run, provider := workflowReviewFixture(t, ReviewWorkflowModeHumanReview)
	if stage == "approval" {
		if _, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil); err != nil {
			t.Fatal(err)
		}
	}

	workflowReviewCrash(t, store, run.ID, stage, index)

	retained, err := store.ReconcileWorkflow(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	before := retained.Revision

	checkWorkflowReviewJournalRetry(t, store, run.ID, provider, retained, before, stage)

	retained = workflowTestRead(t, store, run.ID)
	if retained.ReviewCount != 1 || len(retained.Attempts) != len(run.Attempts) {
		t.Fatal("input/approval recovery repeated source work")
	}
}

func testWorkflowReviewRoleJournal(t *testing.T, stage string, index int) {
	t.Helper()

	store, _, run, provider := workflowReviewFixture(t, ReviewWorkflowModeAgenticReview)
	workflowReviewCrash(t, store, run.ID, stage, index)

	retained, err := store.ReconcileWorkflow(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowReviewClaimRecovered(t, store, run.ID, retained, stage)

	reviewer := workflowReviewWorker("pass", "", "")
	if stage == "decision" {
		reviewer = []string{"false"}
	}

	binding, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, reviewer)
	if err != nil || binding.State != "authorized" {
		t.Fatalf("role replay: %+v, %v", binding, err)
	}

	expected := len(run.Attempts) + 1
	if stage == "claim" {
		expected++
	}

	if len(workflowTestRead(t, store, run.ID).Attempts) != expected {
		t.Fatal("role recovery repeated source work or accepted review")
	}
}

func checkWorkflowReviewJournalRetry(t *testing.T, store *WorkflowStore, id string, provider *workflowReviewProviderFixture, retained WorkflowInvocation, before uint64, stage string) {
	t.Helper()

	if stage == "input" {
		if _, err := store.ReviewWorkflowSlice(t.Context(), id, provider, nil); err != nil {
			t.Fatal(err)
		}

		return
	}

	if _, err := store.ApproveWorkflowSlice(t.Context(), id, retained.Review.InputHash, provider); err != nil {
		t.Fatal(err)
	}

	if workflowTestRead(t, store, id).Revision != before {
		t.Fatal("approval replay allocated another decision")
	}
}

func checkWorkflowReviewClaimRecovered(t *testing.T, store *WorkflowStore, id string, retained WorkflowInvocation, stage string) {
	t.Helper()

	if stage == "claim" {
		if retained.Status != "paused" || retained.ResumeStatus != "waiting_review" {
			t.Fatal("review claim recovery rewound tasks")
		}

		if _, err := store.ResumeWorkflow(t.Context(), id); err != nil {
			t.Fatal(err)
		}
	}
}
