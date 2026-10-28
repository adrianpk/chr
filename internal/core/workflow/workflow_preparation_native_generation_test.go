package workflow

import (
	"errors"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativePreparationGeneration keeps exact S/B/P unchanged while a new queue
// generation references the earlier candidate. Restart must refuse that cached
// gate without spending another preparation attempt or rewriting task evidence.
func TestNativePreparationGeneration(t *testing.T) {
	fixture := newNativePreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	f := fixture.f

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, fixture.provider, f.git); err != nil {
		t.Fatal(err)
	}

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := f.store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	for _, state := range []string{"queued", "preparing", "validating", "reviewing"} {
		entry = nativeRoleTransition(t, f.store, entry, state)
	}

	before := workflowTestRead(t, f.store, f.run.ID)
	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	if _, err := restarted.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, fixture.provider, f.git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("prior queue generation reused combined evidence", err)
	}

	after := workflowTestRead(t, restarted, f.run.ID)
	if after.Revision != before.Revision || len(after.Attempts) != len(before.Attempts) {
		t.Fatal("stale cached gate changed the new generation")
	}

	assertWorkflowPreparationSource(t, f)
}
