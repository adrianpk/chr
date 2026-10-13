package core

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type nativeDispatchWaitingPublisher struct {
	publisher WorkflowGitPublisher
	started   chan struct{}
	release   chan struct{}
	once      sync.Once
}

func (p *nativeDispatchWaitingPublisher) PublishWorkflowGit(ctx context.Context, request WorkflowGitPublication) (WorkflowGitPublicationResult, error) {
	p.once.Do(func() { close(p.started) })

	select {
	case <-ctx.Done():
		return WorkflowGitPublicationResult{}, ctx.Err()
	case <-p.release:
		return p.publisher.PublishWorkflowGit(ctx, request)
	}
}

// TestNativeDispatchConcurrent keeps an earlier live host in the dispatch port.
// A restarted store cannot replace it, even with a fresh current revision guard.
func TestNativeDispatchConcurrent(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, _ := nativeDispatchPublisher(t, f)
	waiting := &nativeDispatchWaitingPublisher{publisher: publisher, started: make(chan struct{}), release: make(chan struct{})}

	var release sync.Once

	t.Cleanup(func() { release.Do(func() { close(waiting.release) }) })

	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)
	finished := make(chan error, 1)

	go func() {
		_, _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, waiting)
		finished <- err
	}()

	select {
	case <-t.Context().Done():
		t.Fatal("owned dispatch did not reach its port")
	case err := <-finished:
		t.Fatal("owned dispatcher failed before publication", err)
	case <-waiting.started:
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)
	current := nativeBudgetGuard(t, restarted, f.entry.InvocationID)

	submission, _, err := restarted.PublishWorkflowNativeSubmission(t.Context(), current, f.provider, publisher)
	if !errors.Is(err, ErrWorkflowObservationRequired) || submission.Native.DispatchIdentity == nil || publisher.calls.Load() != 0 {
		t.Fatal("replacement bypassed a live prior dispatch", err)
	}

	release.Do(func() { close(waiting.release) })

	if err := <-finished; err != nil || publisher.calls.Load() != 1 {
		t.Fatal("the original dispatcher did not finish exactly once", err)
	}
}
