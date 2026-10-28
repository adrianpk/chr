package workflow

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowSubmissionAtomicGuard changes a canonical source or target after
// host preflight. The server transaction rejects the entire publication and
// observation must seal that attempt before exclusion can be released.
func TestWorkflowSubmissionAtomicGuard(t *testing.T) {
	for _, changed := range []string{"source", "target"} {
		t.Run(changed, func(t *testing.T) { testWorkflowSubmissionAtomicGuard(t, changed) })
	}
}

func testWorkflowSubmissionAtomicGuard(t *testing.T, changed string) {
	t.Helper()

	f, provider := newWorkflowSubmissionFixture(t, "rebase")
	provider.before = func(ctx context.Context, s workflowstate.Submission) {
		moveWorkflowSubmissionRef(t, ctx, f.remote, changed, s)
	}

	submission, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))
	if err == nil || submission.State != "awaiting_observation" || submission.Response != "error_unknown" {
		t.Fatalf("guard rejection: %+v, %v", submission, err)
	}

	if workflowIntegrationCommand(t, f.remote, "rev-parse", "refs/heads/dev") == submission.Guard.Git.ResultCommit {
		t.Fatal("stale source or base was published")
	}

	observed, err := f.store.ReconcileWorkflowSubmission(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))
	if err != nil || observed.State != "unpublished" || observed.ID != submission.ID || provider.calls != 1 {
		t.Fatalf("sealed guard rejection: %+v, %v", observed, err)
	}
}

func moveWorkflowSubmissionRef(t *testing.T, ctx context.Context, remote, changed string, s workflowstate.Submission) {
	t.Helper()

	branch, commit := s.Guard.Git.TargetBranch, s.Guard.Git.OriginCommit
	if changed == "source" {
		branch = s.Guard.PullRequest.HeadBranch
	}

	if output, err := exec.CommandContext(ctx, "git", "-C", remote, "update-ref", "refs/heads/"+branch, commit).CombinedOutput(); err != nil {
		t.Fatalf("external ref movement: %v: %s", err, output)
	}
}

// TestWorkflowSubmissionExclusion retains unknown target ownership after pause,
// cancellation and loss of the invocation projection. A linked
// checkout and different namespace/project alias still see the durable owner.
func TestWorkflowSubmissionExclusion(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { testWorkflowSubmissionExclusion(t, intent) })
	}
}

func testWorkflowSubmissionExclusion(t *testing.T, intent string) {
	t.Helper()

	f, provider := newWorkflowSubmissionFixture(t, "merge")
	provider.unknown = true
	git := workflowSubmissionObserver(f)

	first, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, git)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}

	if _, err := f.store.StopWorkflow(t.Context(), f.run.ID, intent); err != nil {
		t.Fatal(err)
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)
	if _, err := restarted.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, git); !errors.Is(err, ErrWorkflowObservationRequired) || provider.calls != 1 {
		t.Fatal("stop intent or restart cleared an unknown submission", err)
	}

	alias := workflowSubmissionAlias(t, f)
	assertWorkflowSubmissionExcluded(t, alias, first)
	provider.seal(first.ID)

	sealed, err := restarted.ReconcileWorkflowSubmission(t.Context(), f.run.ID, provider, git)
	if err != nil || sealed.State != "unpublished" || workflowTestRead(t, restarted, f.run.ID).StopIntent != intent {
		t.Fatalf("observation changed the stop or failed to seal: %+v, %v", sealed, err)
	}

	if err := alias.withLock(t.Context(), func() error {
		key, err := workflowSubmissionTargetKey(first.Guard.PullRequest)
		if err != nil {
			return err
		}

		return alias.guardWorkflowSubmissionTarget(t.Context(), contenthash.Sum([]byte("another-unit")), key)
	}); err != nil {
		t.Fatal("sealed unpublished result retained exclusion", err)
	}
}

func workflowSubmissionAlias(t *testing.T, f workflowPreparationFixture) *Store {
	t.Helper()

	linked := filepath.Join(t.TempDir(), "linked")
	workflowIntegrationCommand(t, f.store.layout.ProjectDir(), "worktree", "add", "--detach", linked, "HEAD")

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: linked, Namespace: "other"})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, "project-alias")
}

func assertWorkflowSubmissionExcluded(t *testing.T, alias *Store, submission workflowstate.Submission) {
	t.Helper()

	err := alias.withLock(t.Context(), func() error {
		key, err := workflowSubmissionTargetKey(submission.Guard.PullRequest)
		if err != nil {
			return err
		}

		return alias.guardWorkflowSubmissionTarget(t.Context(), contenthash.Sum([]byte("another-unit")), key)
	})
	if !errors.Is(err, ErrWorkflowOwned) {
		t.Fatalf("linked worktree or namespace bypassed target exclusion: %v", err)
	}
}

// TestWorkflowSubmissionProjectionLoss proves that deleting only an owned
// invocation diagnostic cannot remove the common-directory exclusion journal.
func TestWorkflowSubmissionProjectionLoss(t *testing.T) {
	f, provider := newWorkflowSubmissionFixture(t, "squash")
	provider.unknown = true

	submission, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}

	alias := workflowSubmissionAlias(t, f)
	if err := os.Remove(filepath.Join(f.store.layout.Root(), workflowInvocationPath(f.run.ID))); err != nil {
		t.Fatal(err)
	}

	assertWorkflowSubmissionExcluded(t, alias, submission)
}

// TestWorkflowSubmissionStop cancels a live provider call and preserves its
// uncertain result. A concurrent publisher can only observe the same intent.
func TestWorkflowSubmissionStop(t *testing.T) {
	f, provider := newWorkflowSubmissionFixture(t, "rebase")
	entered := make(chan struct{})
	provider.before = func(ctx context.Context, _ workflowstate.Submission) {
		close(entered)
		<-ctx.Done()
	}
	provider.unknown = true

	finished := make(chan error, 1)

	go func() {
		_, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f))
		finished <- err
	}()

	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("owned provider call did not start")
	}

	if _, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, provider, workflowSubmissionObserver(f)); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("concurrent publisher repeated an uncertain intent", err)
	}

	if _, err := f.store.StopWorkflow(t.Context(), f.run.ID, "cancel"); err != nil {
		t.Fatal(err)
	}

	assertWorkflowSubmissionStopped(t, f, provider, finished)
}

func assertWorkflowSubmissionStopped(t *testing.T, f workflowPreparationFixture, provider *workflowSubmissionProviderFixture, finished <-chan error) {
	t.Helper()

	select {
	case err := <-finished:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("provider cancellation was not retained", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("provider call did not observe cancellation")
	}

	last, err := f.store.ReadWorkflowSubmission(t.Context(), f.run.ID)
	if err != nil || last.State != "awaiting_observation" || provider.calls != 1 || workflowTestRead(t, f.store, f.run.ID).Status != workflowstate.InvocationCancelled {
		t.Fatal("cancelled IO discarded exclusion or consumed another attempt")
	}
}
