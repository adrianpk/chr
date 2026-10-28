package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeDriftGateProvider struct {
	nativeDriftProviderFixture
	marker string
	once   sync.Once
	move   func()
}

func (p *nativeDriftGateProvider) ReadWorkflowPullRequest(ctx context.Context, url string) (workflowstate.PullRequestSnapshot, error) {
	if _, err := os.Stat(p.marker); err == nil {
		p.once.Do(p.move)
	}

	return p.nativeDriftProviderFixture.ReadWorkflowPullRequest(ctx, url)
}

// TestNativeDriftDuringGate moves the physical target before the host completion
// check. The finished gate becomes retained stale evidence, then restart runs a
// replacement against the new base without accepting the stale result.
func TestNativeDriftDuringGate(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "gate-finished")
	gate := workflowPreparationGate(fmt.Sprintf("test -f result.txt && if test -f target.txt; then touch %q; fi", marker))
	f := newNativePreparationFixture(t, "rebase", gate)
	store, id := f.f.store, f.f.run.ID
	provider := &nativeDriftGateProvider{nativeDriftProviderFixture: nativeDriftProviderFixture{f.provider, f.f.remote}, marker: marker, move: func() { nativeDriftMoveTarget(t, &f, "Target changed during gate") }}

	old, err := PrepareIntegration(t.Context(), store, id, provider, f.f.git)
	if err == nil || old.Native != nil {
		t.Fatal("gate completion authorized a stale base", err)
	}

	before := workflowTestRead(t, store, id)
	if before.CandidateRepair != nil || before.Preparation.State != "stale" {
		t.Fatal("target drift was routed as a semantic repair")
	}

	store = NewWorkflowStore(store.layout, store.projectID)
	f.f.store = store

	fresh, err := PrepareIntegration(t.Context(), store, id, provider, f.f.git)
	if err != nil || fresh.Input.Git.BaseCommit != f.f.base || fresh.Input.Native.ClaimGeneration != old.Input.Native.ClaimGeneration+1 {
		t.Fatal("stale completed gate was not replaced after restart", err)
	}

	nativeDriftRetained(t, f, before, fresh)
}

// TestNativeSourceDuringGate retains a typed ownership/scope decision when an
// external source writer wins before completion, without accepting stale gate
// evidence or routing unapproved content to automatic semantic repair.
func TestNativeSourceDuringGate(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "source-moved")
	gate := workflowPreparationGate(fmt.Sprintf("test -f result.txt && if test -f target.txt; then touch %q; fi", marker))
	f := newNativePreparationFixture(t, "rebase", gate)
	store, id := f.f.store, f.f.run.ID
	provider := &nativeDriftGateProvider{nativeDriftProviderFixture: nativeDriftProviderFixture{f.provider, f.f.remote}, marker: marker, move: func() { nativeDriftMoveSource(t, &f, "foreign scope", "unapproved.txt") }}

	candidate, err := PrepareIntegration(t.Context(), store, id, provider, f.f.git)
	if err == nil || candidate.Native != nil {
		t.Fatal("changed source received candidate authority", err)
	}

	run := workflowTestRead(t, store, id)

	entry := nativeReviewEntry(t, store, id)
	if entry.State != "blocked" || entry.Blocker.Reason != "source_revision_requires_reconciliation" || run.CandidateRepair != nil {
		t.Fatal("unapproved source did not retain an ownership/scope decision")
	}

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, provider); err == nil {
		t.Fatal("source drift retained publication authority")
	}

	assertWorkflowPreparationSource(t, f.f)
}
