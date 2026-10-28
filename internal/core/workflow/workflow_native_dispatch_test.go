package workflow

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
)

type nativeDispatchPublisherFixture struct {
	t         *testing.T
	store     *Store
	id        string
	publisher GitPublisher
	calls     atomic.Int32
}

func (p *nativeDispatchPublisherFixture) PublishWorkflowGit(ctx context.Context, request GitPublication) (GitPublicationResult, error) {
	p.calls.Add(1)

	submission, err := p.store.ReadWorkflowSubmission(ctx, p.id)
	if err != nil || submission.Native == nil {
		p.t.Fatal("publisher reached IO before durable intent", err)
	}

	native := submission.Native
	if native.DispatchState != "dispatching" || native.ApplicationState != "unknown" || native.DispatchIdentity == nil {
		p.t.Fatal("publisher reached IO before durable dispatch identity")
	}

	alive, err := workflowProcessAlive(*native.DispatchIdentity)
	if err != nil || !alive {
		p.t.Fatal("dispatch owner is not the live host", err)
	}

	if !workflowNativeJSONEqual(request.Git, submission.Guard.Git) {
		p.t.Fatal("transport did not receive the exact authorized proof")
	}

	return p.publisher.PublishWorkflowGit(ctx, request)
}

func nativeDispatchPublisher(t *testing.T, f nativeSubmissionFixture) (*nativeDispatchPublisherFixture, string) {
	t.Helper()

	b := f.candidate.Native.NativeRecordBinding
	remote := filepath.Join(t.TempDir(), "dispatch.git")
	workflowIntegrationCommand(t, f.candidate.Lane.Worktree, "init", "--bare", "--quiet", remote)
	workflowIntegrationCommand(t, f.candidate.Lane.Worktree, "push", remote, b.OriginalSource+":"+b.SourceRef, b.ValidatedBase+":"+b.TargetRef)

	return &nativeDispatchPublisherFixture{t: t, store: f.store, id: f.entry.InvocationID, publisher: NewWorkflowGitPublisher(b.SourceRepository, remote)}, remote
}

// TestNativeDispatch exercises actual atomic publication through the store.
// Intent and host identity precede IO; restart never repeats an uncertain call.
func TestNativeDispatch(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, remote := nativeDispatchPublisher(t, f)
	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)

	submission, result, err := f.store.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, publisher)
	if err != nil || result.State != "applied" || publisher.calls.Load() != 1 {
		t.Fatal("authorized atomic dispatch did not run once", err)
	}

	b := f.candidate.Native.NativeRecordBinding
	for _, ref := range []string{b.SourceRef, b.TargetRef} {
		if got := workflowIntegrationCommand(t, f.candidate.Lane.Worktree, "--git-dir="+remote, "rev-parse", ref); got != b.PreparedResult {
			t.Fatal("actual remote did not receive the exact result", got)
		}
	}

	if submission.State != "awaiting_observation" || submission.Native.ApplicationState != "unknown" {
		t.Fatal("raw transport response manufactured delivery")
	}

	nativeDispatchNoRepeat(t, f, publisher, submission)
}

func nativeDispatchNoRepeat(t *testing.T, f nativeSubmissionFixture, publisher *nativeDispatchPublisherFixture, submission Submission) {
	t.Helper()

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)
	guard := nativeBudgetGuard(t, restarted, f.entry.InvocationID)

	again, _, err := restarted.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, publisher)
	if !errors.Is(err, ErrWorkflowObservationRequired) || again.ID != submission.ID || publisher.calls.Load() != 1 {
		t.Fatal("restart repeated native dispatch", err)
	}

	budget, err := restarted.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("dispatch retry refilled or double charged submission", err)
	}
}
