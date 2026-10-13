package core

import (
	"errors"
	"os"
	"testing"
)

// TestNativeDispatchGuard refuses stale generations, policy changes and stop
// intent before either a publication reservation or actual external dispatch.
func TestNativeDispatchGuard(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, nativeSubmissionFixture, *WorkflowIntegrationBudgetGuard)
	}{
		{"revision", func(_ *testing.T, _ nativeSubmissionFixture, g *WorkflowIntegrationBudgetGuard) { g.EntryRevision++ }},
		{"generation", func(_ *testing.T, _ nativeSubmissionFixture, g *WorkflowIntegrationBudgetGuard) { g.ClaimGeneration++ }},
		{"policy", func(_ *testing.T, f nativeSubmissionFixture, _ *WorkflowIntegrationBudgetGuard) {
			f.provider.policyRevision = "policy-v2"
		}},
		{"stop", nativeDispatchStopped},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newNativeSubmissionFixture(t)
			publisher, _ := nativeDispatchPublisher(t, f)
			guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)
			test.change(t, f, &guard)

			if _, _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, publisher); err == nil || publisher.calls.Load() != 0 {
				t.Fatal("stale or stopped owner reached dispatch", err)
			}

			if _, err := f.store.ReadWorkflowSubmission(t.Context(), f.entry.InvocationID); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("invalid authorization retained a publication intent", err)
			}
		})
	}
}

func nativeDispatchStopped(t *testing.T, f nativeSubmissionFixture, guard *WorkflowIntegrationBudgetGuard) {
	t.Helper()

	if _, err := f.store.StopWorkflow(t.Context(), f.entry.InvocationID, "cancel"); err != nil {
		t.Fatal(err)
	}

	*guard = nativeBudgetGuard(t, f.store, f.entry.InvocationID)
}

// TestNativeDispatchPolicy changes live policy after intent is durable but
// before host dispatch. The final observation denies IO and preserves exclusion.
func TestNativeDispatchPolicy(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, _ := nativeDispatchPublisher(t, f)
	f.store.afterReplace = func(index int) error {
		if index == -1 {
			f.provider.policyRevision = "policy-after-intent"
		}

		return nil
	}

	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)

	submission, _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, publisher)
	if !errors.Is(err, ErrWorkflowConflict) || submission.ID == "" || publisher.calls.Load() != 0 {
		t.Fatal("policy race reached publication", err)
	}

	f.store.afterReplace = nil

	retained, err := f.store.ReadWorkflowSubmission(t.Context(), f.entry.InvocationID)
	if err != nil || retained.ID != submission.ID || retained.Native.DispatchState != "not_started" {
		t.Fatal("policy race discarded the durable intent", err)
	}

	nativeSubmissionCandidateExcluded(t, f)
}
