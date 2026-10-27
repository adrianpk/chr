package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPreparationGateFailure rejects failed, mutating and truncated
// host gates against combined content without rewinding accepted source tasks.
func TestWorkflowPreparationGateFailure(t *testing.T) {
	cases := []struct{ name, script string }{
		{"failed", "if test -f target.txt; then exit 17; fi"},
		{"mutated", "if test -f target.txt; then printf changed >> target.txt; fi"},
		{"truncated", "if test -f target.txt; then python3 -c 'print(\"x\"*20000)'; fi"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newWorkflowPreparationFixture(t, "rebase", workflowPreparationGate(test.script))
			if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); err == nil {
				t.Fatal("invalid combined gate was accepted")
			}

			run := workflowTestRead(t, f.store, f.run.ID)
			if run.Status != "paused" || run.ResumeStatus != "waiting_integration" || run.Preparation.State != "stale" {
				t.Fatal("failed gate did not retain a resumable combined-result wait")
			}

			if _, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, f.provider); err == nil {
				t.Fatal("failed gate authorized publication")
			}

			assertWorkflowPreparationSource(t, f)
		})
	}
}

// TestWorkflowPreparationStop keeps cancellation terminal and pause resumable,
// while a running validation excludes a concurrent preparer and retains capacity
// until its owned process group has exited.
func TestWorkflowPreparationStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { testWorkflowPreparationStop(t, intent) })
	}
}

func testWorkflowPreparationStop(t *testing.T, intent string) {
	t.Helper()

	marker := filepath.Join(t.TempDir(), "gate-started")
	script := fmt.Sprintf("if test -f target.txt; then printf started > '%s'; sleep 30; fi", marker)
	f := newWorkflowPreparationFixture(t, "merge", workflowPreparationGate(script))
	done := make(chan error, 1)

	go func() {
		_, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git)
		done <- err
	}()

	waitWorkflowClaimFile(t, marker)

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("concurrent gate: %v", err)
	}

	if _, err := f.store.StopWorkflow(t.Context(), f.run.ID, intent); err != nil {
		t.Fatal(err)
	}

	if err := <-done; !errors.Is(err, ErrWorkflowStop) {
		t.Fatalf("stop: %v", err)
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Status != workflowStopStatus(intent) || run.Preparation.State != "stale" || run.ResumeStatus != "waiting_integration" {
		t.Fatal("stop reset combined-result state")
	}

	assertWorkflowPreparationSource(t, f)
	checkWorkflowReviewCapacity(t, f.store, marker)

	if intent == "pause" {
		if _, err := f.store.ResumeWorkflow(t.Context(), f.run.ID); err != nil {
			t.Fatal(err)
		}
	}
}

// TestWorkflowPreparationConflict blocks a real conflicting combined tree and
// enforces the preparation bound without a fallback method or another worker.
func TestWorkflowPreparationConflict(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "squash", WorkflowValidation{WaiverReason: "Conflicting target fixture"})
	f.base = workflowIntegrationCommit(t, f.actor, "result.txt", "Conflicting target content")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	f.provider.target, f.provider.pr.BaseCommit = f.base, f.base
	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); err == nil {
		t.Fatal("conflicting combined content was accepted")
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Status != "blocked" || run.Preparation.State != "stale" || run.PreparationCount != 1 {
		t.Fatal("conflict lost its preparation bound")
	}

	err := f.store.withLock(t.Context(), func() error {
		current, err := f.store.readWorkflowInvocation(f.run.ID)
		if err != nil {
			return err
		}

		current.PreparationCount, current.Preparation.Round = workflowstate.PageLimit, workflowstate.PageLimit

		return f.store.commitWorkflowState(&current)
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); err == nil {
		t.Fatal("preparation bound was reset")
	}

	assertWorkflowPreparationSource(t, f)
}

// TestWorkflowPreparationApproval requires approval of the exact new combined
// validation; a previously approved source-only hash cannot approve it.
func TestWorkflowPreparationApproval(t *testing.T) {
	f := newWorkflowPreparationFixture(t, "rebase", WorkflowValidation{WaiverReason: "Manual combined approval fixture"})
	workflowReviewConfig(t, f.store, ReviewWorkflowModeHumanReview)

	old, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.ApproveWorkflowSlice(t.Context(), f.run.ID, old.InputHash, f.provider); err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); err != nil {
		t.Fatal(err)
	}

	fresh, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, nil)
	if err != nil || fresh.State != "waiting_approval" {
		t.Fatalf("combined approval wait: %+v, %v", fresh, err)
	}

	if _, err := f.store.ApproveWorkflowSlice(t.Context(), f.run.ID, old.InputHash, f.provider); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("source-only approval was reused")
	}

	if _, err := f.store.ApproveWorkflowSlice(t.Context(), f.run.ID, fresh.InputHash, f.provider); err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, f.provider); err != nil {
		t.Fatal(err)
	}

	assertWorkflowPreparationSource(t, f)
}
