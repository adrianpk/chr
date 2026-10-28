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

// TestWorkflowSliceCrashProcess terminates an owned writer between journaled
// replacements for aggregate claim, candidate acceptance or affected-task repair.
func TestWorkflowSliceCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_SLICE_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id := os.Getenv("CHR_SLICE_CRASH_ID")
	stage := os.Getenv("CHR_SLICE_CRASH_STAGE")

	index, err := strconv.Atoi(os.Getenv("CHR_SLICE_CRASH_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	store.afterReplace = func(replaced int) error {
		if replaced == index && workflowSliceCrashReady(t, store, id, stage) {
			os.Exit(17)
		}

		return nil
	}

	if stage == "repair" {
		_, err = store.RepairWorkflowSlice(t.Context(), id, os.Getenv("CHR_SLICE_CRASH_TASK"))
	} else {
		_, err = ValidateSlice(t.Context(), store, id)
	}

	t.Fatalf("owned writer did not terminate: %v", err)
}

func workflowSliceCrashReady(t *testing.T, store *Store, id, stage string) bool {
	t.Helper()

	if stage != "candidate" {
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

	return record.Task.AttemptID == ""
}

func workflowSliceCrash(t *testing.T, store *Store, id, stage, task string, index int) {
	t.Helper()

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowSliceCrashProcess$")

	command.Env = append(os.Environ(), "CHR_SLICE_CRASH_DIR="+store.layout.ProjectDir(), "CHR_SLICE_CRASH_ID="+id, "CHR_SLICE_CRASH_STAGE="+stage, "CHR_SLICE_CRASH_INDEX="+strconv.Itoa(index), "CHR_SLICE_CRASH_TASK="+task)

	output, err := command.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 17 {
		t.Fatalf("owned slice crash: %v: %s", err, output)
	}
}

// TestWorkflowSliceJournal replays every aggregate claim/candidate replacement
// after real host death, retaining task proof and avoiding duplicate accepted gates.
func TestWorkflowSliceJournal(t *testing.T) {
	for _, stage := range []string{"claim", "candidate"} {
		for _, index := range []int{0, 2, 3} {
			t.Run(stage+" "+strconv.Itoa(index), func(t *testing.T) {
				store, hierarchy, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
				run := workflowTestStart(t, store, request)
				workflowSliceBeforeGate(t, store, run.Invocation.ID)
				workflowSliceCrash(t, store, run.Invocation.ID, stage, "", index)

				retained, err := store.ReconcileWorkflow(t.Context(), run.Invocation.ID)
				if err != nil {
					t.Fatal(err)
				}

				checkWorkflowSliceJournal(t, store, hierarchy, retained, stage)
			})
		}
	}
}

func checkWorkflowSliceJournal(t *testing.T, store *Store, hierarchy MaterializationResult, retained workflowstate.Invocation, stage string) {
	t.Helper()

	workflowTestValidatedSlice(t, store, hierarchy, retained.ID)

	if stage == "claim" {
		if retained.Status != workflowstate.InvocationPaused || retained.ResumeStatus != workflowstate.InvocationWaitingIntegration || retained.Attempts[3].Outcome != "recovered" {
			t.Fatal("aggregate claim recovery reset validated tasks")
		}

		if _, err := store.ResumeWorkflow(t.Context(), retained.ID); err != nil {
			t.Fatal(err)
		}
	}

	candidate := workflowTestExecuteSlice(t, store, retained.ID, "fail")

	replayed := workflowTestValidatedSlice(t, store, hierarchy, retained.ID)
	if !candidate.Result.Successful || stage == "candidate" && len(replayed.Attempts) != 4 {
		t.Fatal("candidate journal recovery repeated accepted validation")
	}
}

// TestWorkflowSliceRepairJournal replays a killed fan-out repair at every write,
// keeping the unaffected prefix and invalidating all affected successor evidence.
func TestWorkflowSliceRepairJournal(t *testing.T) {
	for _, index := range []int{0, 2, 4} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			store, hierarchy, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
			run := workflowTestStart(t, store, request)
			candidate := workflowTestExecuteSlice(t, store, run.Invocation.ID, "worker")
			workflowSliceCrash(t, store, run.Invocation.ID, "repair", hierarchy.Tasks[1].Task.Key, index)

			retained, err := store.ReconcileWorkflow(t.Context(), run.Invocation.ID)
			if err != nil || retained.Status != workflowstate.InvocationPending || retained.Cursor != 1 || retained.CandidateRef != "" {
				t.Fatalf("repair journal recovery: %+v, %v", retained, err)
			}

			checkWorkflowSliceRepair(t, store, hierarchy, candidate)
		})
	}
}
