package workflow

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeDispatchPublisherFixture struct {
	t         *testing.T
	store     *Store
	id        string
	publisher GitPublisher
	calls     atomic.Int32
}

func (p *nativeDispatchPublisherFixture) PublishWorkflowGit(ctx context.Context, request GitPublication) (workflowstate.GitPublicationResult, error) {
	p.calls.Add(1)

	submission, err := p.store.ReadWorkflowSubmission(ctx, p.id)
	if err != nil || submission.Native == nil {
		p.t.Fatal("publisher reached IO before durable intent", err)
	}

	native := submission.Native
	if native.DispatchState != workflowstate.NativeDispatching || native.ApplicationState != workflowstate.NativeApplicationUnknown || native.DispatchIdentity == nil {
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

	nativePublication, err := f.store.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, publisher)

	submission, result := nativePublication.Submission, nativePublication.Transport
	if err != nil || result.State != workflowstate.GitPublicationApplied || publisher.calls.Load() != 1 {
		t.Fatal("authorized atomic dispatch did not run once", err)
	}

	b := f.candidate.Native.NativeRecordBinding
	for _, ref := range []string{b.SourceRef, b.TargetRef} {
		if got := workflowIntegrationCommand(t, f.candidate.Lane.Worktree, "--git-dir="+remote, "rev-parse", ref); got != b.PreparedResult {
			t.Fatal("actual remote did not receive the exact result", got)
		}
	}

	if submission.State != workflowstate.SubmissionAwaitingObservation || submission.Native.ApplicationState != workflowstate.NativeApplicationUnknown {
		t.Fatal("raw transport response manufactured delivery")
	}

	nativeDispatchNoRepeat(t, f, publisher, submission)
}

func nativeDispatchNoRepeat(t *testing.T, f nativeSubmissionFixture, publisher *nativeDispatchPublisherFixture, submission workflowstate.Submission) {
	t.Helper()

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)
	guard := nativeBudgetGuard(t, restarted, f.entry.InvocationID)

	nativePublication2, err := restarted.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, publisher)

	again := nativePublication2.Submission
	if !errors.Is(err, ErrWorkflowObservationRequired) || again.ID != submission.ID || publisher.calls.Load() != 1 {
		t.Fatal("restart repeated native dispatch", err)
	}

	budget, err := restarted.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("dispatch retry refilled or double charged submission", err)
	}
}
