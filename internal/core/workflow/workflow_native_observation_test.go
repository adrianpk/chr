package workflow

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type nativeCanonicalFixture struct {
	f         nativeSubmissionFixture
	forge     nativeCapabilityFixture
	publisher *nativeDispatchPublisherFixture
	remote    string
	git       OSWorkflowIntegrationGit
}

func newNativeCanonicalFixture(t *testing.T) nativeCanonicalFixture {
	t.Helper()

	return nativeCanonicalPublicationFixture(t, false)
}

func nativeCanonicalPublicationFixture(t *testing.T, lost bool) nativeCanonicalFixture {
	t.Helper()

	f := newNativeSubmissionFixture(t)

	return nativeCanonicalSubmissionFixture(t, f, lost)
}

func nativeCanonicalSubmissionFixture(t *testing.T, f nativeSubmissionFixture, lost bool) nativeCanonicalFixture {
	t.Helper()

	publisher, remote := nativeDispatchPublisher(t, f)
	if lost {
		publisher.publisher = nativeLostResponsePublisher(t, f.candidate.Native.SourceRepository, remote)
	}

	submission, _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)
	if !lost && err != nil || lost && err == nil {
		t.Fatal("native canonical fixture publication", err)
	}

	b := f.candidate.Native.NativeRecordBinding

	git := NewWorkflowIntegrationGit(b.SourceRepository, remote)
	if _, err := f.store.ReconcileWorkflowNativeGitSubmission(t.Context(), nativeBudgetGuard(t, f.store, submission.InvocationID), NewWorkflowGitPublisher(b.SourceRepository, remote), git); err != nil {
		t.Fatal("native canonical fixture Git reconciliation", err)
	}

	runner := &workflowPRCommandFixture{fakeRemoteReviewRunner: newFakeRemoteReviewRunner()}
	forge := nativeCapabilityFixture{adapter: NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner), runner: runner, request: NativeCapabilityRequest{PullRequest: submission.Guard.PullRequest}}
	forge.prefix = nativePolicyCommand(t, forge, "")
	setNativePRFixture(t, forge, forge.request.PullRequest)

	forge.runner.outputs[forge.prefix+"/branches/"+url.PathEscape(forge.request.PullRequest.HeadBranch)] = []byte(`{"commit":{"sha":"` + b.PreparedResult + `"}}`)

	return nativeCanonicalFixture{f, forge, publisher, remote, git}
}

// TestNativeCanonicalOutcomes retains real adapter observations without a receipt
// or rollback. Typed PR conflicts and target divergence keep the same exclusion.
func TestNativeCanonicalOutcomes(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	f, b := fixture.f, fixture.f.candidate.Native.NativeRecordBinding

	before, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name, condition string
		change          func(*workflowstate.PullRequestSnapshot)
	}{
		{"pending", "", func(_ *workflowstate.PullRequestSnapshot) {}},
		{"closed", "canonical_pr_closed_unmerged", func(pr *workflowstate.PullRequestSnapshot) { pr.State = "closed" }},
		{"retargeted", "canonical_pr_retargeted", func(pr *workflowstate.PullRequestSnapshot) { pr.BaseBranch = "main" }},
		{"identity", "canonical_pr_identity_changed", func(pr *workflowstate.PullRequestSnapshot) { pr.URL += "0" }},
		{"wrong_result", "canonical_pr_result_mismatch", func(pr *workflowstate.PullRequestSnapshot) {
			pr.State, pr.Merged, pr.ResultCommit = "closed", true, b.OriginalSource
		}},
		{"merged", "", func(pr *workflowstate.PullRequestSnapshot) {
			pr.State, pr.Merged, pr.ResultCommit = "closed", true, b.PreparedResult
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			pr := fixture.forge.request.PullRequest
			test.change(&pr)
			setNativePRFixture(t, fixture.forge, pr)

			submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
			if err != nil {
				t.Fatal(err)
			}

			observation := nativeCanonicalLast(t, f.store, submission)
			if observation.Condition != test.condition || observation.PullRequest.URL != pr.URL {
				t.Fatal("canonical facts were replaced", observation)
			}

			checkNativeCanonicalBlocker(t, f.store, submission, test.condition)

			nativeCanonicalPending(t, fixture, submission, before)
		})
	}

	checkNativeCanonicalDivergence(t, fixture, before)
}

func checkNativeCanonicalBlocker(t *testing.T, store *Store, submission workflowstate.Submission, condition string) {
	t.Helper()

	entry := nativeReviewEntry(t, store, submission.InvocationID)
	if condition != "" && (entry.State != "blocked" || entry.Blocker == nil || entry.Blocker.Reason != condition) || condition == "" && entry.State != "observing" {
		t.Fatal("canonical outcome did not project its retained blocker", entry)
	}
}

func checkNativeCanonicalDivergence(t *testing.T, fixture nativeCanonicalFixture, before IntegrationBudget) {
	t.Helper()

	f, b := fixture.f, fixture.f.candidate.Native.NativeRecordBinding

	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if err == nil || nativeCanonicalLast(t, f.store, submission).Condition != "canonical_target_unverified" {
		t.Fatal("rewritten target did not retain reconciliation", err)
	}

	if target := workflowIntegrationCommand(t, fixture.remote, "rev-parse", b.TargetRef); target != b.ValidatedBase {
		t.Fatal("observer rolled target history back or republished", target)
	}

	nativeCanonicalPending(t, fixture, submission, before)
}

func nativeCanonicalLast(t *testing.T, store *Store, submission workflowstate.Submission) NativePublicationObservation {
	t.Helper()

	refs := submission.Native.ObservationRefs
	if len(refs) == 0 {
		t.Fatal("canonical observation was not retained")
	}

	var observation NativePublicationObservation
	if err := store.readWorkflowNativeEvidence(submission.InvocationID, refs[len(refs)-1], &observation); err != nil {
		t.Fatal(err)
	}

	return observation
}

func nativeCanonicalPending(t *testing.T, fixture nativeCanonicalFixture, submission workflowstate.Submission, before IntegrationBudget) {
	t.Helper()

	if submission.State != "awaiting_observation" || submission.Native.ApplicationState != "git_applied" || submission.Native.ReceiptRef != nil || fixture.publisher.calls.Load() != 1 {
		t.Fatal("canonical observation created delivery or dispatched again")
	}

	nativeSubmissionPendingExcluded(t, fixture.f, submission)

	if after, err := fixture.f.store.ReadWorkflowIntegrationBudget(t.Context(), submission.InvocationID); err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("canonical observation changed persistent budgets", err)
	}
}

type nativeCanonicalObserverFixture struct {
	observer NativePublicationObserver
	before   func(context.Context, NativePublicationBinding)
	calls    int
}

func (f *nativeCanonicalObserverFixture) ObserveWorkflowNativePublication(ctx context.Context, binding NativePublicationBinding, git IntegrationGit) (NativePublicationObservation, error) {
	f.calls++
	f.before(ctx, binding)

	return f.observer.ObserveWorkflowNativePublication(ctx, binding, git)
}

// TestNativeCanonicalPoll observes metadata convergence with bounded backoff and
// deduplicates unchanged facts. Restart retains the deadline and one publication.
func TestNativeCanonicalPoll(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	f := fixture.f
	before := workflowTestRead(t, f.store, f.entry.InvocationID)

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), before.ID)
	if err != nil {
		t.Fatal(err)
	}

	observer := &nativeCanonicalObserverFixture{observer: fixture.forge.adapter}
	observer.before = func(_ context.Context, binding NativePublicationBinding) {
		if observer.calls == 3 {
			pr := binding.PullRequest
			pr.State, pr.Merged, pr.ResultCommit = "closed", true, binding.Git.ResultCommit
			setNativePRFixture(t, fixture.forge, pr)
		}
	}

	submission, err := f.store.PollWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, before.ID), observer, fixture.git)
	if err != nil || observer.calls != 3 || len(submission.Native.ObservationRefs) != 2 || nativeCanonicalLast(t, f.store, submission).State != "merged" {
		t.Fatal("bounded canonical polling did not converge", err)
	}

	f.store = NewWorkflowStore(f.store.layout, f.store.projectID)
	fixture.f = f

	retained, err := f.store.ReadWorkflowSubmission(t.Context(), before.ID)
	if err != nil || !workflowNativeJSONEqual(retained, submission) {
		t.Fatal("restart lost canonical observation or persisted deadline", err)
	}

	nativeCanonicalPending(t, fixture, retained, budget)
}

// TestNativeCanonicalFence rejects metadata whose owner changed during external
// IO, including a duplicate callback. No stale observation replaces the journal.
func TestNativeCanonicalFence(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	f := fixture.f

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if err != nil {
		t.Fatal(err)
	}

	observer := &nativeCanonicalObserverFixture{observer: fixture.forge.adapter, before: func(ctx context.Context, _ NativePublicationBinding) {
		entry := nativeReviewEntry(t, f.store, f.entry.InvocationID)

		run, _, err := f.store.locateWorkflowInvocation(ctx, f.entry.InvocationID)
		if err != nil {
			t.Fatal(err)
		}

		request := IntegrationTransition{InvocationID: run.ID, EntryID: entry.ID, InvocationRevision: run.Revision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: entry.CandidateRef.Hash, State: "blocked"}

		request.Blocker = &IntegrationBlocker{Reason: "observation_hold", RequiredAction: "observe current owner", ResumeState: "observing"}
		if _, err := f.store.TransitionWorkflowIntegration(ctx, request); err != nil {
			t.Fatal(err)
		}
	}}
	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), observer, fixture.git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale canonical callback changed the journal", err)
	}

	retained, err := f.store.ReadWorkflowSubmission(t.Context(), submission.InvocationID)
	if err != nil || !workflowNativeJSONEqual(retained, submission) {
		t.Fatal("stale canonical callback replaced retained evidence", err)
	}
}

func nativeCanonicalDeadline(t *testing.T, fixture nativeCanonicalFixture, deadline time.Time) {
	t.Helper()

	err := fixture.f.store.withLock(t.Context(), func() error {
		run, err := fixture.f.store.readWorkflowInvocation(fixture.f.entry.InvocationID)
		if err != nil {
			return err
		}

		journal, err := fixture.f.store.readWorkflowSubmissionJournal(t.Context(), run)
		if err != nil {
			return err
		}

		journal.Attempts[len(journal.Attempts)-1].Native.ObservationDeadline = deadline
		if err := fixture.f.store.writeWorkflowSubmissionJournal(t.Context(), &journal); err != nil {
			return err
		}

		return fixture.f.store.projectWorkflowNativeSubmission(&run, journal)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestNativeCanonicalDeadline stops automatic polling at the persisted bound.
// An explicit fresh observation can later reconcile the same PR without resend.
func TestNativeCanonicalDeadline(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	f := fixture.f
	deadline := time.Now().UTC().Add(600 * time.Millisecond)
	nativeCanonicalDeadline(t, fixture, deadline)

	observer := &nativeCanonicalObserverFixture{observer: fixture.forge.adapter, before: func(_ context.Context, _ NativePublicationBinding) {}}

	submission, err := f.store.PollWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), observer, fixture.git)
	// The persisted bound may expire during a real Git read or during backoff.
	// Both paths retain the same deadline fact; only the read returns an IO error.
	if err != nil && !errors.Is(err, context.DeadlineExceeded) || nativeCanonicalLast(t, f.store, submission).Condition != "canonical_pr_observation_deadline" || !submission.Native.ObservationDeadline.Equal(deadline) {
		t.Fatal("polling renewed its deadline or lost reconciliation", err)
	}

	calls := observer.calls
	if _, err := f.store.PollWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), observer, fixture.git); err != nil || observer.calls != calls {
		t.Fatal("expired automatic polling performed more external IO", err)
	}

	checkNativeDeadlineRecovery(t, fixture, submission)
}

func checkNativeDeadlineRecovery(t *testing.T, fixture nativeCanonicalFixture, submission workflowstate.Submission) {
	t.Helper()

	f := fixture.f
	pr := fixture.forge.request.PullRequest
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, submission.Native.PreparedResult
	setNativePRFixture(t, fixture.forge, pr)

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if err != nil || nativeCanonicalLast(t, f.store, submission).State != "merged" || nativeReviewEntry(t, f.store, submission.InvocationID).State != "observing" || fixture.publisher.calls.Load() != 1 {
		t.Fatal("explicit reconciliation did not recover the same publication", err)
	}
}

type nativeCanonicalBlockedObserver struct{}

func (nativeCanonicalBlockedObserver) ObserveWorkflowNativePublication(ctx context.Context, _ NativePublicationBinding, _ IntegrationGit) (NativePublicationObservation, error) {
	<-ctx.Done()

	return NativePublicationObservation{}, ctx.Err()
}

// TestNativeCanonicalTimeout bounds a stalled adapter by the retained deadline,
// then durably records uncertainty despite the completed IO context expiring.
func TestNativeCanonicalTimeout(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	f := fixture.f
	deadline := time.Now().UTC().Add(600 * time.Millisecond)
	nativeCanonicalDeadline(t, fixture, deadline)

	submission, err := f.store.PollWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), nativeCanonicalBlockedObserver{}, fixture.git)
	if !errors.Is(err, context.DeadlineExceeded) || nativeCanonicalLast(t, f.store, submission).Condition != "canonical_pr_observation_deadline" {
		t.Fatal("stalled canonical IO lost its persisted deadline", err)
	}

	checkNativeCanonicalBlocker(t, f.store, submission, "canonical_pr_observation_deadline")
	nativeSubmissionPendingExcluded(t, f, submission)
}

// TestNativeCanonicalLost recovers actual lost SSH output before later source
// advancement. Canonical observation retains the raw unknown transport result.
func TestNativeCanonicalLost(t *testing.T) {
	fixture := nativeCanonicalPublicationFixture(t, true)
	f, b := fixture.f, fixture.f.candidate.Native.NativeRecordBinding
	later := workflowIntegrationCommit(t, f.candidate.Lane.Worktree, "later.txt", "Later source work")
	workflowIntegrationCommand(t, f.candidate.Lane.Worktree, "push", fixture.remote, later+":"+b.SourceRef, later+":"+b.TargetRef)

	pr := fixture.forge.request.PullRequest
	pr.State, pr.Merged, pr.ResultCommit, pr.HeadCommit = "closed", true, b.PreparedResult, later
	setNativePRFixture(t, fixture.forge, pr)

	fixture.forge.runner.outputs[fixture.forge.prefix+"/branches/"+url.PathEscape(pr.HeadBranch)] = []byte(`{"commit":{"sha":"` + later + `"}}`)

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if err != nil || nativeCanonicalLast(t, f.store, submission).LaterSourceCommit != later {
		t.Fatal("lost response attribution was replaced by later source", err)
	}

	var result GitPublicationResult
	if err := f.store.readWorkflowNativeEvidence(submission.InvocationID, *submission.Native.TransportResponseRef, &result); err != nil || result.State != "unknown" || result.PublishedSourceCommit != "" {
		t.Fatal("canonical observation fabricated a transport response", err)
	}

	nativeSubmissionPendingExcluded(t, f, submission)
}
