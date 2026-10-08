package core

import (
	"errors"
	"path/filepath"
	"testing"
)

// TestWorkflowSliceChanged refuses candidate reuse after proof, contract or Git
// mutation; no fresh aggregate attempt hides the incompatible retained image.
func TestWorkflowSliceChanged(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testing.T, *WorkflowStore, WorkflowInvocation, WorkflowSliceCandidate)
	}{
		{"candidate", func(t *testing.T, s *WorkflowStore, run WorkflowInvocation, _ WorkflowSliceCandidate) {
			workflowProcessWrite(t, filepath.Join(s.layout.Root(), run.CandidateRef), "Changed\n")
		}},
		{"contract", func(t *testing.T, s *WorkflowStore, _ WorkflowInvocation, candidate WorkflowSliceCandidate) {
			workflowProcessWrite(t, filepath.Join(s.layout.Root(), candidate.Slice.ContractRef), "Changed\n")
		}},
		{"task proof", func(t *testing.T, s *WorkflowStore, _ WorkflowInvocation, candidate WorkflowSliceCandidate) {
			workflowProcessWrite(t, filepath.Join(s.layout.Root(), candidate.Slice.Tasks[0].EvidenceRef), "Changed\n")
		}},
		{"task input", func(t *testing.T, s *WorkflowStore, _ WorkflowInvocation, candidate WorkflowSliceCandidate) {
			workflowProcessWrite(t, filepath.Join(s.layout.Root(), candidate.Slice.Tasks[0].InputRef), "Changed\n")
		}},
		{"dirty lane", func(t *testing.T, _ *WorkflowStore, _ WorkflowInvocation, candidate WorkflowSliceCandidate) {
			workflowProcessWrite(t, filepath.Join(candidate.Lane.Worktree, "dirty.txt"), "Changed\n")
		}},
		{"new commit", func(t *testing.T, _ *WorkflowStore, _ WorkflowInvocation, candidate WorkflowSliceCandidate) {
			workflowProcessWrite(t, filepath.Join(candidate.Lane.Worktree, "later.txt"), "Changed\n")
			workflowProcessCommit(t, candidate.Lane.Worktree)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
			run := workflowTestStart(t, store, request)
			candidate := workflowTestExecuteSlice(t, store, run.Invocation.ID, "worker")

			retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
			if err != nil {
				t.Fatal(err)
			}

			tc.mutate(t, store, retained, candidate)

			if _, err := store.ValidateWorkflowSlice(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatalf("changed aggregate image reused: %v", err)
			}

			retained, err = store.ReadWorkflow(t.Context(), run.Invocation.ID)
			if err != nil || len(retained.Attempts) != 4 {
				t.Fatal("changed image created another attempt")
			}
		})
	}
}

// TestWorkflowSliceGateMutation rejects a host gate that changes the clean Git
// candidate and retains task proof pending an explicit affected-task repair.
func TestWorkflowSliceGateMutation(t *testing.T) {
	for _, mode := range []string{"gate dirty", "gate commit"} {
		t.Run(mode, func(t *testing.T) {
			store, hierarchy, request := workflowSliceFixture(t, workflowSliceGate(t, mode))
			run := workflowTestStart(t, store, request)

			candidate, err := store.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, workflowTaskCommand(t, "worker"))
			if !errors.Is(err, ErrWorkflowConflict) || candidate.Result.Successful {
				t.Fatalf("mutable aggregate candidate accepted: %v", err)
			}

			retained := workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
			checkWorkflowAggregateFailure(t, retained)

			if _, err := store.ValidateWorkflowSlice(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("changed lane reran the aggregate gate")
			}
		})
	}
}

// TestWorkflowSliceStop proves supervised aggregate pause/cancel preserves the
// completed task cursor and evidence; a pause resumes only aggregate validation.
func TestWorkflowSliceStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { testWorkflowSliceStop(t, intent) })
	}
}

func testWorkflowSliceStop(t *testing.T, intent string) {
	t.Helper()

	marker := filepath.Join(t.TempDir(), "aggregate-started")
	gate := WorkflowValidation{Required: true, Commands: [][]string{{"sh", "-c", "test -f \"$1\" && exit 0; printf started > \"$1\"; sleep 20", "gate", marker}}, TimeoutSeconds: 30}
	store, hierarchy, request := workflowSliceFixture(t, gate)
	run := workflowTestStart(t, store, request)
	done := make(chan error, 1)

	go func() {
		_, err := store.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, workflowTaskCommand(t, "worker"))
		done <- err
	}()

	waitWorkflowClaimFile(t, marker)

	if _, err := store.ValidateWorkflowSlice(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("competing aggregate validator was admitted")
	}

	if _, err := store.StopWorkflow(t.Context(), run.Invocation.ID, intent); err != nil {
		t.Fatal(err)
	}

	if err := <-done; !errors.Is(err, ErrWorkflowStop) {
		t.Fatalf("aggregate stop: %v", err)
	}

	retained := workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
	if retained.Status != workflowStopStatus(intent) || retained.CandidateRef != "" || retained.Attempts[3].CheckpointRef == "" {
		t.Fatal("aggregate stop fabricated acceptance or lost its checkpoint")
	}

	checkWorkflowSliceStopResume(t, store, retained, intent)
}

func checkWorkflowSliceStopResume(t *testing.T, store *WorkflowStore, retained WorkflowInvocation, intent string) {
	t.Helper()

	resumed, err := store.ResumeWorkflow(t.Context(), retained.ID)
	if intent == "cancel" {
		if !errors.Is(err, ErrWorkflowOwned) {
			t.Fatal("cancelled aggregate resumed")
		}

		return
	}

	if err != nil || resumed.Status != "waiting_integration" || resumed.Cursor != 3 {
		t.Fatal("pause lost the aggregate barrier")
	}

	candidate := workflowTestExecuteSlice(t, store, retained.ID, "fail")
	if !candidate.Result.Successful || candidate.Result.Commit != retained.Lane.HeadCommit {
		t.Fatal("aggregate resume reran tasks or changed the accepted tree")
	}
}

// TestWorkflowSliceWaiver retains an explicit approved aggregate waiver without
// substituting per-task validation or launching an unnecessary aggregate process.
func TestWorkflowSliceWaiver(t *testing.T) {
	store, hierarchy, request := workflowSliceFixture(t, WorkflowValidation{WaiverReason: "Aggregate accepted by task checks"})
	run := workflowTestStart(t, store, request)
	candidate := workflowTestExecuteSlice(t, store, run.Invocation.ID, "worker")
	retained := workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)

	if candidate.Result.WaiverReason == "" || len(candidate.Result.Validation) != 0 || len(retained.Attempts[3].Processes) != 0 {
		t.Fatal("aggregate waiver was discarded or invented command evidence")
	}
}

func workflowSliceBeforeGate(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	for range 3 {
		workflowTestExecute(t, store, id, "worker")
	}
}

// TestWorkflowSlicePartial compares complete actual membership against a partial
// frozen extent before permitting an aggregate acceptance command.
func TestWorkflowSlicePartial(t *testing.T) {
	store, _, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))

	run := workflowTestStart(t, store, request)
	if err := store.withLock(t.Context(), func() error {
		retained, err := store.readWorkflowInvocation(run.Invocation.ID)
		if err != nil {
			return err
		}

		retained.Members = retained.Members[:2]

		return store.commitWorkflowState(&retained)
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, workflowTaskCommand(t, "worker")); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("partial slice formed an aggregate candidate: %v", err)
	}

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || retained.CandidateRef != "" || len(retained.Attempts) != 2 {
		t.Fatal("partial slice launched aggregate validation")
	}
}

// TestWorkflowSliceBudget preserves the finite aggregate retry budget while
// retaining all task results; an exhausted retry cannot create another attempt.
func TestWorkflowSliceBudget(t *testing.T) {
	store, hierarchy, request := workflowSliceFixture(t, WorkflowValidation{Required: true, Commands: [][]string{{"false"}}, TimeoutSeconds: 30})
	request.Policy.MaxWorkerAttempts = 1

	run := workflowTestStart(t, store, request)
	if _, err := store.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, workflowTaskCommand(t, "worker")); err == nil {
		t.Fatal("fixture aggregate gate did not fail")
	}

	if _, err := store.ValidateWorkflowSlice(t.Context(), run.Invocation.ID); err == nil {
		t.Fatal("aggregate retry reset its consumed budget")
	}

	retained := workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
	checkWorkflowAggregateFailure(t, retained)
}

// TestWorkflowSliceFence preserves validated task images across aggregate yield
// and refuses old credentials without releasing the replacement claim's locks.
func TestWorkflowSliceFence(t *testing.T) {
	store, hierarchy, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
	run := workflowTestStart(t, store, request)
	workflowSliceBeforeGate(t, store, run.Invocation.ID)

	old, prepared, _, err := store.prepareWorkflowSliceValidation(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.CheckpointWorkflow(t.Context(), old); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ResumeWorkflow(t.Context(), run.Invocation.ID); err != nil {
		t.Fatal(err)
	}

	replacement, _, _, err := store.prepareWorkflowSliceValidation(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.finishWorkflowSliceValidation(t.Context(), old, prepared, &WorkflowSliceCandidate{}, false); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("old aggregate claim finalized or released its replacement")
	}

	if err := store.CheckWorkflowClaim(t.Context(), replacement); err != nil {
		t.Fatal("replacement aggregate claim lost ownership")
	}

	if err := store.CheckpointWorkflow(t.Context(), replacement); err != nil {
		t.Fatal(err)
	}

	workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
}
