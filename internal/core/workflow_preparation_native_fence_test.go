package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativePreparationFence changes the durable queue owner or current policy
// during a real host gate. Its successful exit cannot authorize the old image.
func TestNativePreparationFence(t *testing.T) {
	for _, change := range []string{"phase", "policy"} {
		t.Run(change, func(t *testing.T) { nativePreparationFence(t, change) })
	}
}

func nativePreparationFence(t *testing.T, change string) {
	t.Helper()

	root := t.TempDir()
	marker, release := filepath.Join(root, "started"), filepath.Join(root, "release")
	script := fmt.Sprintf("if test -f target.txt; then printf started > '%s'; while ! test -f '%s'; do sleep 0.02; done; fi", marker, release)
	fixture := newNativePreparationFixture(t, "merge", workflowPreparationGate(script))
	f := fixture.f
	done := make(chan error, 1)

	go func() {
		_, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, fixture.provider, f.git)
		done <- err
	}()

	waitWorkflowClaimFile(t, marker)

	if change == "phase" {
		nativeRoleTransition(t, f.store, fixture.entry, "merging")
	} else {
		fixture.provider.mu.Lock()
		fixture.provider.policyRevision = "policy-v2"
		fixture.provider.mu.Unlock()
	}

	if err := os.WriteFile(release, []byte("release"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := <-done; !errors.Is(err, workflowstate.ErrConflict) && !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("stale gate succeeded", err)
	}

	checkNativePreparationRejected(t, fixture)
}
