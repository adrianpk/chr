package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

type nativeSubmissionFixture struct {
	store     *WorkflowStore
	entry     WorkflowIntegrationEntry
	candidate WorkflowIntegrationCandidate
	provider  *nativeSubmissionProviderFixture
}

type nativeSubmissionProviderFixture struct {
	*workflowReviewProviderFixture
	capability WorkflowPublicationCapability
}

func (p *nativeSubmissionProviderFixture) WorkflowIntegrationCapability(context.Context, WorkflowPullRequestSnapshot, string) (WorkflowPublicationCapability, error) {
	return p.capability, nil
}

func newNativeSubmissionFixture(t *testing.T) nativeSubmissionFixture {
	t.Helper()

	store, candidate := nativeRecordCandidate(t, "rebase")
	run := workflowTestRead(t, store, candidate.InvocationID)
	entry := nativeEntryAdmit(t, store, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: candidate.Native.Unit, AdmissionKey: run.RequestKey, TargetRef: candidate.Native.TargetRef})
	candidate = nativeSubmissionInput(t, store, candidate)
	candidate.Native.EntryID = entry.ID

	var err error

	candidate.Native.CandidateHash, err = WorkflowNativeCandidateHash(candidate)
	if err != nil {
		t.Fatal(err)
	}

	ref := bindNativeSubmissionCandidate(t, store, candidate)
	request := nativeEntryTransition(t, store, entry, "preparing")
	request.CandidateRef = &ref

	entry, err = store.TransitionWorkflowIntegration(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	pr := candidate.Input.Source.PullRequest
	provider := &nativeSubmissionProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}, target: pr.BaseCommit}, capability: candidate.Input.Capability}

	for _, state := range []string{"validating", "reviewing"} {
		entry, err = store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, entry, state))
		if err != nil {
			t.Fatal(err)
		}
	}

	if _, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil); err != nil {
		t.Fatal(err)
	}

	entry, err = store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, entry, "ready_to_publish"))
	if err != nil {
		t.Fatal(err)
	}

	return nativeSubmissionFixture{store, entry, candidate, provider}
}

func bindNativeSubmissionCandidate(t *testing.T, store *WorkflowStore, candidate WorkflowIntegrationCandidate) WorkflowEvidenceRef {
	t.Helper()

	run, err := store.readWorkflowInvocation(candidate.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}

	before := run.Preparation.ResultHash
	run.Preparation.ResultHash = workflowHash(data)

	changes := []workflowFileChange{{Path: run.Preparation.ResultRef, Before: before, After: data}}
	for i := range run.Attempts {
		attempt := &run.Attempts[i]
		if attempt.ID != candidate.Validation.AttemptID {
			continue
		}

		result, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}

		changes = append(changes, workflowFileChange{Path: attempt.ResultRef, Before: attempt.ResultHash, After: result})
		attempt.ResultHash, attempt.InputRef, attempt.InputHash = run.Preparation.ResultHash, candidate.Validation.InputRef, candidate.Validation.InputHash
	}

	invocation, err := workflowInvocationChange(run)
	if err != nil {
		t.Fatal(err)
	}

	changes = append(changes, invocation)
	if err := store.commitFiles(changes); err != nil {
		t.Fatal(err)
	}

	return WorkflowEvidenceRef{Path: run.Preparation.ResultRef, Hash: run.Preparation.ResultHash}
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

func nativeSubmissionInput(t *testing.T, store *WorkflowStore, candidate WorkflowIntegrationCandidate) WorkflowIntegrationCandidate {
	t.Helper()

	run, err := store.readWorkflowInvocation(candidate.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	data, err := store.readWorkflowReviewImage(candidate.Validation.InputRef, candidate.Validation.InputHash)
	if err != nil {
		t.Fatal(err)
	}

	for _, attempt := range run.Attempts {
		if attempt.ID != candidate.Validation.AttemptID {
			continue
		}

		file := workflowFileChange{Path: attempt.InputRef, Before: attempt.InputHash, After: data}
		if err := store.commitFiles([]workflowFileChange{file}); err != nil {
			t.Fatal(err)
		}

		candidate.Validation.InputRef, candidate.Validation.InputHash = file.Path, workflowHash(data)
	}

	candidate.Native.ValidationRefs = []WorkflowEvidenceRef{retainNativeRecordEvidence(t, store, candidate.InvocationID, "dispatch-validation", candidate.Validation)}

	return candidate
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
	changed.Hash = workflowHash([]byte("another candidate"))
	request.CandidateRef = &changed

	request.Blocker = &WorkflowIntegrationBlocker{Reason: "pending", RequiredAction: "observe original dispatch", ResumeState: "observing"}
	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), request); err == nil {
		t.Fatal("uncertain intent allowed candidate replacement")
	}
}
