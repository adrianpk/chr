package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeReviewFence changes queue ownership or native policy while a real
// reviewer is running. Its passing answer must retain a failed checkpoint,
// preserve task proof and release capacity only after the owned process exits.
func TestNativeReviewFence(t *testing.T) {
	for _, change := range []string{"phase", "policy"} {
		t.Run(change, func(t *testing.T) { nativeReviewFence(t, change) })
	}
}

func nativeReviewFence(t *testing.T, change string) {
	t.Helper()

	f, _ := nativeReviewFixture(t, workflowstate.ReviewAgentic)
	store, id := f.f.store, f.f.run.ID
	root := t.TempDir()
	marker, release := filepath.Join(root, "started"), filepath.Join(root, "release")
	before := fmt.Sprintf("import pathlib,time\npathlib.Path(%q).write_text('started')\nwhile not pathlib.Path(%q).exists(): time.sleep(.02)", marker, release)
	done := make(chan error, 1)

	go func() {
		_, err := store.ReviewWorkflowSlice(t.Context(), id, f.provider, workflowReviewWorker("pass", "", before))
		done <- err
	}()

	waitWorkflowClaimFile(t, marker)

	if change == "phase" {
		entry := nativeReviewEntry(t, store, id)
		nativeRoleTransition(t, store, entry, "merging")
	} else {
		f.provider.mu.Lock()
		f.provider.policyRevision = "changed"
		f.provider.mu.Unlock()
	}

	if err := os.WriteFile(release, []byte("release"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := <-done; !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale review succeeded or lost its owner", err)
	}

	nativeReviewFailedCheckpoint(t, f, marker)
}

func nativeReviewFailedCheckpoint(t *testing.T, f nativePreparationFixture, marker string) {
	t.Helper()

	run := workflowTestRead(t, f.f.store, f.f.run.ID)

	last := run.Attempts[len(run.Attempts)-1]
	if run.Review.State != "stale" || last.CheckpointRef == "" || last.Outcome != "interrupted" {
		t.Fatal("stale reviewer lost its checkpoint")
	}

	if _, err := f.f.store.readWorkflowReviewImage(last.CheckpointRef, last.CheckpointHash); err != nil {
		t.Fatal(err)
	}

	checkWorkflowReviewCapacity(t, f.f.store, marker)
	assertWorkflowPreparationSource(t, f.f)
}
