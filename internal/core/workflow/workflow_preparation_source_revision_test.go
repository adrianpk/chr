package workflow

import (
	"context"
	"errors"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeDriftProviderFixture struct {
	*nativePreparationProviderFixture
	remote string
}

func (p nativeDriftProviderFixture) ReadWorkflowPullRequest(ctx context.Context, url string) (workflowstate.PullRequestSnapshot, error) {
	pr, err := p.nativePreparationProviderFixture.ReadWorkflowPullRequest(ctx, url)
	if err != nil {
		return pr, err
	}

	pr.BaseCommit, err = p.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)

	return pr, err
}

func (p nativeDriftProviderFixture) ReadWorkflowBranch(ctx context.Context, repository, branch string) (string, error) {
	if repository != p.pr.Repository {
		return "", workflowstate.ErrConflict
	}

	output, err := outputWorktreeGit(ctx, p.remote, "rev-parse", "refs/heads/"+branch)

	return strings.TrimSpace(output), err
}

// TestNativeSourceDrift freezes appended content, changed ownership and stale
// PR observations without expanding the frozen task scope or replacing proofs.
func TestNativeSourceDrift(t *testing.T) {
	cases := []struct{ name, path string }{
		{"owned path", "result.txt"},
		{"foreign scope", "unapproved.txt"},
		{"stale PR", "result.txt"},
		{"foreign repository", ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) { nativeSourceDrift(t, test.name, test.path) })
	}
}

func nativeSourceDrift(t *testing.T, name, path string) {
	t.Helper()

	f, _ := nativeReviewFixture(t, workflowstate.ReviewTrustedAutoMerge)
	nativeReviewAuthorize(t, f)

	store, id := f.f.store, f.f.run.ID
	before := workflowTestRead(t, store, id)

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	nativeDriftMoveSource(t, &f, name, path)

	provider := nativeDriftProviderFixture{f.provider, f.f.remote}
	if _, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, provider, f.f.git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed source did not require a scope decision", err)
	}

	after := nativeSourceDriftFacts(t, f, before, budget)
	nativeSourceDriftRetry(t, f, provider, after)
}

func nativeSourceDriftFacts(t *testing.T, f nativePreparationFixture, before workflowstate.Invocation, budget IntegrationBudget) workflowstate.Invocation {
	t.Helper()

	store, id := f.f.store, f.f.run.ID
	after := workflowTestRead(t, store, id)

	entry := nativeReviewEntry(t, store, id)
	if entry.State != "blocked" || entry.Blocker.Reason != "source_revision_requires_reconciliation" || after.Preparation.State != "stale" || after.Review.State != workflowstate.ReviewStale {
		t.Fatal("changed source retained authority or changed its frozen owner")
	}

	if after.Cursor != before.Cursor || after.PreparationCount != before.PreparationCount || after.CandidateHash != before.CandidateHash {
		t.Fatal("source drift silently admitted work or reran validation")
	}

	nativeSourceDriftOwnership(t, f, entry, budget)

	return after
}

func nativeSourceDriftOwnership(t *testing.T, f nativePreparationFixture, entry IntegrationEntry, budget IntegrationBudget) {
	t.Helper()

	if entry.FrozenScopeHash != f.entry.FrozenScopeHash || entry.ClaimGeneration != f.entry.ClaimGeneration {
		t.Fatal("source drift changed the frozen ownership or scope")
	}

	store, id := f.f.store, f.f.run.ID

	retained, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || !workflowNativeJSONEqual(budget, retained) {
		t.Fatal("scope decision changed retained consumption", err)
	}
}

func nativeDriftMoveSource(t *testing.T, f *nativePreparationFixture, name, path string) {
	t.Helper()

	if path == "" {
		f.provider.pr.HeadRepository = "https://forge.example/foreign/repo"

		return
	}

	workflowIntegrationCommand(t, f.f.actor, "checkout", "-b", "source-writer", f.f.run.Lane.HeadCommit)

	head := workflowIntegrationCommit(t, f.f.actor, path, "External source content")
	workflowIntegrationCommand(t, f.f.actor, "push", "origin", head+":"+f.f.run.Lane.Branch)

	if name != "stale PR" {
		f.provider.pr.HeadCommit = head
	}
}

func nativeSourceDriftRetry(t *testing.T, f nativePreparationFixture, provider nativeDriftProviderFixture, before workflowstate.Invocation) {
	t.Helper()

	store := NewWorkflowStore(f.f.store.layout, f.f.store.projectID)

	id := f.f.run.ID
	if _, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, provider, f.f.git); err == nil || workflowTestRead(t, store, id).Revision != before.Revision {
		t.Fatal("restart admitted unverified source or repeated a decision", err)
	}

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, provider); err == nil {
		t.Fatal("unverified source retained publication authority")
	}

	assertWorkflowPreparationSource(t, f.f)
}
