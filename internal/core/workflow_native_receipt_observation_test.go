package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestNativeReceiptDescendant observes actual target advancement after native
// fanout. The same submission remains confirmed without rewriting its receipt.
func TestNativeReceiptDescendant(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { nativeReceiptDescendant(t, method) })
	}
}

func nativeReceiptDescendant(t *testing.T, method string) {
	t.Helper()

	fixture, receipt := nativeReceiptObservedFixture(t, method, "slice")
	f := fixture.f
	before := nativeReceiptPreservedFiles(t, fixture, receipt)

	target := nativeReceiptAdvanceTarget(t, fixture)

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if err != nil || submission.State != "confirmed" {
		t.Fatal("confirmed receipt could not observe its later descendant", err)
	}

	observation := nativeCanonicalLast(t, f.store, submission)
	if observation.Git == nil || observation.Git.ObservedTarget != target || submission.Native.ReceiptRef.Hash != receipt.SourceHash {
		t.Fatal("later target replaced immutable delivery")
	}

	checkNativeReceiptPreserved(t, fixture, receipt, before)
}

// TestNativeReceiptDivergence distinguishes a fetched target excluding P from
// IO uncertainty. Receipt bytes and completed members survive the typed blocker,
// repeat observation and restoration, with one publication and unchanged budget.
func TestNativeReceiptDivergence(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { nativeReceiptDivergence(t, method) })
	}
}

func nativeReceiptDivergence(t *testing.T, method string) {
	t.Helper()

	fixture, receipt := nativeReceiptObservedFixture(t, method, "slice")
	f, b := fixture.f, fixture.f.candidate.Native.WorkflowNativeRecordBinding
	before := nativeReceiptPreservedFiles(t, fixture, receipt)
	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if !errors.Is(err, ErrWorkflowHistoryDivergence) || submission.State != "confirmed" {
		t.Fatal("removed result was not retained as history divergence", err)
	}

	observation := nativeCanonicalLast(t, f.store, submission)

	entry := nativeReviewEntry(t, f.store, f.entry.InvocationID)
	if observation.Condition != "canonical_target_history_diverged" || observation.ObservedTargetCommit != b.ValidatedBase || entry.State != "blocked" || entry.Blocker.Reason != observation.Condition {
		t.Fatal("divergence lost actual target or visible blocker")
	}

	if _, err := f.store.FinalizeWorkflowSlice(t.Context(), f.entry.InvocationID, nil, nil); err == nil {
		t.Fatal("divergence silently recreated integration")
	}

	checkNativeReceiptPreserved(t, fixture, receipt, before)
	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.PreparedResult, b.ValidatedBase)

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git); err != nil {
		t.Fatal(err)
	}

	if nativeReviewEntry(t, f.store, f.entry.InvocationID).State != "integrated" {
		t.Fatal("restored target lost completed fanout")
	}

	checkNativeReceiptPreserved(t, fixture, receipt, before)
}

func nativeReceiptObservedFixture(t *testing.T, method, extent string) (nativeCanonicalFixture, WorkflowRecord) {
	t.Helper()

	f := nativePreparationExtentFixture(t, method, extent, workflowPreparationGate("test -f result.txt"))

	candidate, err := f.f.store.PrepareWorkflowSliceIntegration(t.Context(), f.f.run.ID, f.provider, f.f.git)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.f.store.ReviewWorkflowSlice(t.Context(), f.f.run.ID, f.provider, nil); err != nil {
		t.Fatal(err)
	}

	entry := nativeRoleTransition(t, f.f.store, nativeReviewEntry(t, f.f.store, f.f.run.ID), "ready_to_publish")
	submission := nativeSubmissionFixture{f.f.store, entry, candidate, f.provider}

	fixture, receipt := nativeFanoutReceipt(t, nativeCanonicalSubmissionFixture(t, submission, false))
	if _, err := f.f.store.FinalizeWorkflowSlice(t.Context(), f.f.run.ID, nil, nil); err != nil {
		t.Fatal(err)
	}

	return fixture, receipt
}

func nativeReceiptPreservedFiles(t *testing.T, fixture nativeCanonicalFixture, receipt WorkflowRecord) map[string]string {
	t.Helper()

	run := nativeFanoutRun(t, fixture)

	paths := []string{receipt.Path, run.DeliveryProofRef}
	for _, member := range run.Members {
		paths = append(paths, member.Admission.Path)
	}

	files := map[string]string{}

	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(fixture.f.store.layout.Root(), path))
		if err != nil {
			t.Fatal(err)
		}

		files[path] = contenthash.Sum(data)
	}

	return files
}

func checkNativeReceiptPreserved(t *testing.T, fixture nativeCanonicalFixture, receipt WorkflowRecord, before map[string]string) {
	t.Helper()

	if !reflect.DeepEqual(before, nativeReceiptPreservedFiles(t, fixture, receipt)) || fixture.publisher.calls.Load() != 1 {
		t.Fatal("receipt observation rewrote delivery or repeated publication")
	}

	budget, err := fixture.f.store.ReadWorkflowIntegrationBudget(t.Context(), fixture.f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("receipt observation changed consumed submission", err)
	}
}

// TestNativeReceiptFrontier revalidates the prepared method under later target
// advancement before releasing the next ordered slice. Native metadata may show
// P while the retained source and receipt still name original S.
func TestNativeReceiptFrontier(t *testing.T) {
	fixture, receipt := nativeReceiptObservedFixture(t, "rebase", "slice_set")
	f := fixture.f
	target := nativeReceiptAdvanceTarget(t, fixture)
	ports := nativeReceiptFrontierPorts(t, fixture, target)

	advanced, err := f.store.AdvanceWorkflowFrontier(t.Context(), f.entry.InvocationID, ports)
	if err != nil || advanced.Lane != nil || len(advanced.Deliveries) != 1 || advanced.ExecutionTarget.Commit != target {
		t.Fatal("native receipt could not release ordered continuation", err)
	}

	if advanced.Deliveries[0].DeliveryRef != receipt.Path || advanced.Deliveries[0].Proof.SourceCommit != receipt.Receipt.SourceCommit {
		t.Fatal("frontier replaced native receipt or original source")
	}

	read := workflowTestRead(t, f.store, advanced.ID)
	if read.PreparationCount != 1 || read.ReviewCount != 1 || read.Status != "pending" {
		t.Fatal("frontier reset consumed role counters or lost its retained unit")
	}
}

// TestNativeReceiptFrontierBlock retains divergence before returning a failed
// continuation. The displayed invocation remains at its delivered unit, with
// completed member images and an explicit operator reconciliation blocker.
func TestNativeReceiptFrontierBlock(t *testing.T) {
	fixture, receipt := nativeReceiptObservedFixture(t, "rebase", "slice_set")
	f, b := fixture.f, fixture.f.candidate.Native.WorkflowNativeRecordBinding
	before := nativeReceiptPreservedFiles(t, fixture, receipt)
	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	ports := nativeReceiptFrontierPorts(t, fixture, b.ValidatedBase)

	blocked, err := f.store.AdvanceWorkflowFrontier(t.Context(), f.entry.InvocationID, ports)
	if !errors.Is(err, ErrWorkflowHistoryDivergence) || blocked.Status != "blocked" || blocked.Blocker == nil || len(blocked.Deliveries) != 0 {
		t.Fatal("frontier hid or bypassed removed delivery history", err)
	}

	checkNativeReceiptPreserved(t, fixture, receipt, before)
}

// TestNativeReceiptRoleCounters rejects cumulative role counts with no retained
// owner. The bounded frontier exception cannot authorize missing role history.
func TestNativeReceiptRoleCounters(t *testing.T) {
	fixture, _ := nativeReceiptObservedFixture(t, "rebase", "slice")
	run := nativeFanoutRun(t, fixture)
	clearWorkflowSliceBindings(&run)

	if err := validateWorkflowPreparationBinding(run); err == nil {
		t.Fatal("preparation count with no retained owner was accepted")
	}

	if err := validateWorkflowReviewBinding(run); err == nil {
		t.Fatal("review count with no retained owner was accepted")
	}
}

func nativeReceiptFrontierPorts(t *testing.T, fixture nativeCanonicalFixture, target string) WorkflowFrontierPorts {
	t.Helper()

	setNativePRFixture(t, fixture.forge, nativeCanonicalMergedPR(fixture, target))

	data, err := json.Marshal(map[string]any{"commit": map[string]string{"sha": target}})
	if err != nil {
		t.Fatal(err)
	}

	fixture.forge.runner.outputs[fixture.forge.prefix+"/branches/dev"] = data

	return WorkflowFrontierPorts{Provider: fixture.forge.adapter, Observer: fixture.git, Target: fixture.git}
}

func nativeCanonicalMergedPR(fixture nativeCanonicalFixture, target string) WorkflowPullRequestSnapshot {
	pr := fixture.forge.request.PullRequest
	pr.State, pr.Merged, pr.ResultCommit, pr.HeadCommit, pr.BaseCommit = "closed", true, fixture.f.candidate.Native.PreparedResult, fixture.f.candidate.Native.PublishedSource, target

	return pr
}

func nativeReceiptAdvanceTarget(t *testing.T, fixture nativeCanonicalFixture) string {
	t.Helper()

	actor := filepath.Join(t.TempDir(), "target-writer")
	workflowIntegrationCommand(t, fixture.remote, "clone", "--no-checkout", fixture.remote, actor)
	workflowIntegrationCommand(t, actor, "config", "user.name", "Test")
	workflowIntegrationCommand(t, actor, "config", "user.email", "test@example.com")
	workflowIntegrationCommand(t, actor, "checkout", "-b", "dev", "origin/dev")

	target := workflowIntegrationCommit(t, actor, "later.txt", "Later target content")
	workflowIntegrationCommand(t, actor, "push", "origin", "dev")

	return target
}
