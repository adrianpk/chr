package core

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeSubmissionFixture struct {
	store     *WorkflowStore
	entry     WorkflowIntegrationEntry
	candidate WorkflowIntegrationCandidate
	provider  *nativePreparationProviderFixture
}

func newNativeSubmissionFixture(t *testing.T) nativeSubmissionFixture {
	t.Helper()

	fixture, candidate := nativeReviewFixture(t, workflowstate.ReviewWorkflowModeTrustedAutoMerge)
	store, id := fixture.f.store, fixture.f.run.ID

	if _, err := store.ReviewWorkflowSlice(t.Context(), id, fixture.provider, nil); err != nil {
		t.Fatal(err)
	}

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	entry = nativeRoleTransition(t, store, entry, "ready_to_publish")

	return nativeSubmissionFixture{store, entry, candidate, fixture.provider}
}

func beginNativeSubmission(t *testing.T, f nativeSubmissionFixture) WorkflowSubmission {
	t.Helper()

	submission, dispatch, err := f.store.BeginWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider)
	if err != nil || !dispatch {
		t.Fatal("native intent not retained", err)
	}

	return submission
}

func nativeSubmissionUpdate(t *testing.T, f nativeSubmissionFixture, submission WorkflowSubmission) WorkflowNativeSubmissionUpdate {
	t.Helper()

	run, err := f.store.readWorkflowInvocation(f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	journal, err := f.store.readWorkflowSubmissionJournal(t.Context(), run)
	if err != nil {
		t.Fatal(err)
	}

	return WorkflowNativeSubmissionUpdate{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, f.store, f.entry.InvocationID), SubmissionID: submission.ID, JournalHash: journal.sourceHash, Facts: *submission.Native}
}

// TestNativeSubmissionUnknown retains one common intent, target exclusion and
// consumed submission after restart, without granting a second dispatch.
func TestNativeSubmissionUnknown(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	submission := beginNativeSubmission(t, f)

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("intent did not consume budget", err)
	}

	update := nativeSubmissionUpdate(t, f, submission)

	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	update.Facts.DispatchState, update.Facts.ApplicationState, update.Facts.DispatchIdentity = "dispatching", "unknown", &identity

	submission, err = f.store.RetainWorkflowNativeSubmission(t.Context(), update)
	if err != nil {
		t.Fatal(err)
	}

	f.store = NewWorkflowStore(f.store.layout, f.store.projectID)

	again, dispatch, err := f.store.BeginWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider)
	if err != nil || dispatch || !workflowNativeJSONEqual(again, submission) {
		t.Fatal("restart re-dispatched uncertain Git", err)
	}

	nativeSubmissionCandidateExcluded(t, f)

	if after, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID); err != nil || !reflect.DeepEqual(budget, after) {
		t.Fatal("observation refilled counters", err)
	}
}

// TestNativeSubmissionRecovery recovers a common journal written before a crash
// in the invocation projection; retries do not reserve another mutation attempt.
func TestNativeSubmissionRecovery(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)
	injected := errors.New("interrupted native common intent")
	f.store.afterReplace = func(index int) error {
		if index == -1 {
			return injected
		}

		return nil
	}

	submission, dispatch, err := f.store.BeginWorkflowNativeSubmission(t.Context(), guard, f.provider)
	if !errors.Is(err, injected) || dispatch {
		t.Fatal("common intent fault not reached", err)
	}

	f.store.afterReplace = nil

	again, dispatch, err := f.store.BeginWorkflowNativeSubmission(t.Context(), guard, f.provider)
	if err != nil || dispatch || again.ID != submission.ID {
		t.Fatal("projection recovery allocated another intent", err)
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("recovery changed consumed attempts", err)
	}
}

func nativeSubmissionCandidateExcluded(t *testing.T, f nativeSubmissionFixture) {
	t.Helper()

	run, err := f.store.readWorkflowInvocation(f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := f.store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	request := nativeEntryTransition(t, f.store, entry, "blocked")
	changed := *entry.CandidateRef
	changed.Hash = contenthash.Sum([]byte("another candidate"))
	request.CandidateRef = &changed

	request.Blocker = &WorkflowIntegrationBlocker{Reason: "pending", RequiredAction: "observe original dispatch", ResumeState: "observing"}
	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), request); err == nil {
		t.Fatal("uncertain intent allowed candidate replacement")
	}
}
