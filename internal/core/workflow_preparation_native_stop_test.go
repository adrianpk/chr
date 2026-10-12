package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

// TestNativePreparationStop projects a queue stop while the host gate owns its
// live credential. The owner must retain stopped observations and release shared
// capacity only after its process group exits, without rewinding task evidence.
func TestNativePreparationStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { nativePreparationStop(t, intent) })
	}
}

func nativePreparationStop(t *testing.T, intent string) {
	t.Helper()

	marker := filepath.Join(t.TempDir(), "gate-started")
	script := fmt.Sprintf("if test -f target.txt; then printf started > '%s'; sleep 30; fi", marker)
	fixture := newNativePreparationFixture(t, "merge", workflowPreparationGate(script))
	f := fixture.f
	done := make(chan error, 1)

	go func() {
		_, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, fixture.provider, f.git)
		done <- err
	}()

	waitWorkflowClaimFile(t, marker)

	request := nativeEntryTransition(t, f.store, fixture.entry, workflowStopStatus(intent))
	request.StopIntent = intent

	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), request); err != nil {
		t.Fatal(err)
	}

	if err := <-done; !errors.Is(err, ErrWorkflowStop) {
		t.Fatal("native queue stop was not retained", err)
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Status != workflowStopStatus(intent) || run.Preparation.State != "stale" || run.ResumeStatus != "waiting_integration" {
		t.Fatal("native queue stop lost the integration frontier")
	}

	last := run.Attempts[len(run.Attempts)-1]
	if last.CheckpointRef == "" || workflowAttemptQuiescent(last) != nil {
		t.Fatal("native queue stop discarded process observations")
	}

	assertWorkflowPreparationSource(t, f)
	checkWorkflowReviewCapacity(t, f.store, marker)
}
