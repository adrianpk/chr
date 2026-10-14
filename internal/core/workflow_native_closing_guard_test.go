package core

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestNativeClosingFanout refuses even a standalone set gate before every
// intended member receives its confirmed native receipt. No gate attempt or
// closure record can replace incomplete finalization.
func TestNativeClosingFanout(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "gate")
	f := newNativeClosingFixture(t, "rebase", WorkflowValidation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"touch", marker}}})
	store := f.fixture.f.store

	set, err := store.Read(t.Context(), f.setPath)
	if err != nil {
		t.Fatal(err)
	}

	ports := nativeReceiptFrontierPorts(t, f.fixture, f.receipt.Receipt.ResultCommit)
	request := WorkflowSetClosureRequest{SetPath: set.Path, ExpectedRevision: set.Revision, ExpectedSourceHash: set.SourceHash, Repository: f.receipt.Receipt.Repository, Branch: "dev"}

	if _, err := store.CloseWorkflowSet(t.Context(), request, ports); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("physical receipt bypassed unfinished member fanout", err)
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("unfinished fanout ran a closing gate", err)
	}

	checkNativeClosingAbsent(t, f)
}

// TestNativeClosingRetry retains a failed host gate and retries only closure
// after real native fanout. Task, receipt and frontier evidence stay byte exact;
// a confirmed retry attaches its own existing gate rather than running it again.
func TestNativeClosingRetry(t *testing.T) {
	counter := filepath.Join(t.TempDir(), "gate-count")
	gate := workflowPreparationGate("n=0; [ ! -f \"$1\" ] || read -r n <\"$1\"; n=$((n+1)); printf '%s\\n' \"$n\" >\"$1\"; [ \"$n\" -gt 1 ]")
	gate.Commands[0] = append(gate.Commands[0], "closing-gate", counter)
	f := newNativeClosingFixture(t, "rebase", gate)
	ports := finalizeNativeClosingFixture(t, f)
	before := nativeClosingPreservedFiles(t, f)
	store, id := f.fixture.f.store, f.fixture.f.entry.InvocationID

	failed, err := store.CloseWorkflowInvocationSet(t.Context(), id, "dev", ports)
	if err == nil || failed.Outcome != "closing_validation_failed" || failed.Set.Set.ClosureRef != "" || len(failed.Set.Set.Closing.Attempts) != 1 {
		t.Fatal("failed gate established closure or lost its attempt", err)
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)

	closed, err := restarted.CloseWorkflowInvocationSet(t.Context(), id, "dev", ports)
	if err != nil || closed.Outcome != "completed" || len(closed.Set.Set.Closing.Attempts) != 2 {
		t.Fatal("native closing retry did not finish its own gate", err)
	}

	checkNativeClosingRetry(t, f, before, closed, restarted, ports, counter)
}

func checkNativeClosingRetry(t *testing.T, f nativeClosingFixture, before map[string]string, closed WorkflowSetClosureResult, restarted *WorkflowStore, ports WorkflowFrontierPorts, counter string) {
	t.Helper()

	request := WorkflowSetClosureRequest{SetPath: closed.Set.Path, ExpectedRevision: closed.Set.Revision, ExpectedSourceHash: closed.Set.SourceHash, Repository: closed.Receipt.Target.Repository, Branch: "dev"}

	again, err := restarted.CloseWorkflowSet(t.Context(), request, ports)
	if err != nil || again.ReceiptRef != closed.ReceiptRef || again.Set.Revision != closed.Set.Revision || !reflect.DeepEqual(before, nativeClosingPreservedFiles(t, f)) || f.fixture.publisher.calls.Load() != 1 {
		t.Fatal("confirmed closure replayed publication, members or its gate", err)
	}

	data, err := os.ReadFile(counter)
	if err != nil || string(data) != "2\n" {
		t.Fatal("confirmed retry ran a third gate", err)
	}
}

// TestNativeClosingDivergence observes actual target history removal after
// frontier release. Neither an archived receipt nor completed task images can
// authorize a gate on a target that no longer contains their prepared result.
func TestNativeClosingDivergence(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "gate")
	f := newNativeClosingFixture(t, "merge", WorkflowValidation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"touch", marker}}})
	finalizeNativeClosingFixture(t, f)

	before := nativeClosingPreservedFiles(t, f)
	b := f.fixture.f.candidate.Native.WorkflowNativeRecordBinding
	workflowIntegrationCommand(t, f.fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	ports := nativeReceiptFrontierPorts(t, f.fixture, b.ValidatedBase)

	if _, err := f.fixture.f.store.CloseWorkflowInvocationSet(t.Context(), f.fixture.f.entry.InvocationID, "dev", ports); !errors.Is(err, ErrWorkflowHistoryDivergence) {
		t.Fatal("removed native result entered a closing gate", err)
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) || !reflect.DeepEqual(before, nativeClosingPreservedFiles(t, f)) {
		t.Fatal("divergence ran a gate or rewrote integrated evidence", err)
	}

	checkNativeClosingAbsent(t, f)
}

// TestNativeClosingSets separates original and corrective outcomes. A later
// explicitly empty corrective set still requires its own real host gate and
// retry history; the original native closure receipt remains unchanged.
func TestNativeClosingSets(t *testing.T) {
	f := newNativeClosingFixture(t, "squash", workflowPreparationGate("true"))
	ports := finalizeNativeClosingFixture(t, f)
	store := f.fixture.f.store

	original, err := store.CloseWorkflowInvocationSet(t.Context(), f.fixture.f.entry.InvocationID, "dev", ports)
	if err != nil {
		t.Fatal(err)
	}

	before := workflowReadTestBytes(t, store, original.ReceiptRef)
	set := nativeClosingCorrectiveSet(t, store)

	request := WorkflowSetClosureRequest{SetPath: set.Path, ExpectedRevision: set.Revision, ExpectedSourceHash: set.SourceHash, Repository: original.Receipt.Target.Repository, Branch: "dev"}

	failed, err := store.CloseWorkflowSet(t.Context(), request, ports)
	if err == nil || failed.Outcome != "closing_validation_failed" || failed.Set.Set.ClosureRef != "" {
		t.Fatal("original gate proof closed the corrective set", err)
	}

	request.ExpectedRevision, request.ExpectedSourceHash = failed.Set.Revision, failed.Set.SourceHash

	corrective, err := store.CloseWorkflowSet(t.Context(), request, ports)
	if err != nil || corrective.Outcome != "completed" {
		t.Fatal("corrective set lost its distinct verified outcome", err)
	}

	checkNativeClosingSetOutcomes(t, store, original, corrective, before)
}

func nativeClosingCorrectiveSet(t *testing.T, store *WorkflowStore) WorkflowRecord {
	t.Helper()

	plan := workflowPlanFixture()
	counter := filepath.Join(t.TempDir(), "corrective-count")
	gate := WorkflowValidation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"sh", "-c", "n=0; [ ! -f \"$1\" ] || read -r n <\"$1\"; n=$((n+1)); printf '%s\\n' \"$n\" >\"$1\"; [ \"$n\" -gt 1 ]", "corrective-gate", counter}}}

	set, err := store.Create(t.Context(), "Corrective closing set", WorkflowRecord{Kind: "plan", Set: &WorkflowSet{Schema: WorkflowSchema, Status: "ready", PlanRef: plan.PlanRef, PlanHash: plan.PlanHash, ClosingValidation: gate}})
	if err != nil {
		t.Fatal(err)
	}

	return set
}

func checkNativeClosingSetOutcomes(t *testing.T, store *WorkflowStore, original, corrective WorkflowSetClosureResult, before []byte) {
	t.Helper()

	if corrective.ReceiptRef == original.ReceiptRef || corrective.Receipt.SetID == original.Receipt.SetID || len(corrective.Set.Set.Closing.Attempts) != 2 {
		t.Fatal("corrective set lost its distinct verified outcome")
	}

	if !bytes.Equal(before, workflowReadTestBytes(t, store, original.ReceiptRef)) || len(corrective.Receipt.Validation) != 1 || len(corrective.Receipt.Slices) != 0 {
		t.Fatal("corrective closure replaced or borrowed original evidence")
	}
}

func nativeClosingPreservedFiles(t *testing.T, f nativeClosingFixture) map[string]string {
	t.Helper()

	store := f.fixture.f.store

	run, err := store.readWorkflowInvocation(f.fixture.f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	paths := []string{f.receipt.Path}
	for _, delivery := range run.Deliveries {
		paths = append(paths, delivery.EvidenceRef, delivery.DeliveryProofRef)
	}

	for _, member := range run.Members {
		paths = append(paths, member.Admission.Path)
	}

	files := map[string]string{}

	for _, path := range paths {
		files[path] = workflowHash(workflowReadTestBytes(t, store, path))
	}

	return files
}

func checkNativeClosingAbsent(t *testing.T, f nativeClosingFixture) {
	t.Helper()

	set, err := f.fixture.f.store.Read(t.Context(), f.setPath)
	if err != nil || set.Set.Closing != nil || set.Set.ClosureRef != "" {
		t.Fatal("rejected closure created a gate or current receipt", err)
	}
}

// nativeClosingChangedTarget injects a canonical metadata change only after a
// successful real gate. The closing role must read that same PR again before
// confirming, even when the actual Git target did not change.
type nativeClosingChangedTarget struct {
	WorkflowTargetGit
	calls  int
	change func()
}

func (target *nativeClosingChangedTarget) ObserveWorkflowTarget(ctx context.Context, query WorkflowTargetQuery) (WorkflowPreviewTarget, error) {
	target.calls++
	if target.calls == 2 {
		target.change()
	}

	return target.WorkflowTargetGit.ObserveWorkflowTarget(ctx, query)
}

// TestNativeClosingPRChange runs a successful real gate, then removes the
// canonical merged flag without moving Git. Fresh same-PR confirmation must
// retain a failed closure instead of treating gate output as delivery proof.
func TestNativeClosingPRChange(t *testing.T) {
	f := newNativeClosingFixture(t, "rebase", workflowPreparationGate("true"))
	ports := finalizeNativeClosingFixture(t, f)
	ports.Target = &nativeClosingChangedTarget{WorkflowTargetGit: ports.Target, change: func() {
		pr := nativeCanonicalMergedPR(f.fixture, f.receipt.Receipt.ResultCommit)
		pr.State, pr.Merged = "open", false
		setNativePRFixture(t, f.fixture.forge, pr)
	}}

	failed, err := f.fixture.f.store.CloseWorkflowInvocationSet(t.Context(), f.fixture.f.entry.InvocationID, "dev", ports)
	if !errors.Is(err, ErrWorkflowObservationRequired) || failed.Outcome != "closing_validation_failed" || failed.Set.Set.ClosureRef != "" || len(failed.Receipt.Validation) != 1 || failed.Receipt.Validation[0].ExitCode != 0 {
		t.Fatal("successful gate hid a changed native PR", err)
	}
}

// TestNativeClosingEvidence detects native transaction mutation after a real
// successful gate. The closure cannot omit a nested immutable proof from its
// frozen inputs or overwrite the original receipt to adopt changed evidence.
func TestNativeClosingEvidence(t *testing.T) {
	f := newNativeClosingFixture(t, "rebase", workflowPreparationGate("true"))
	ports := finalizeNativeClosingFixture(t, f)
	store := f.fixture.f.store
	before := nativeClosingPreservedFiles(t, f)
	path := filepath.Join(store.layout.Root(), f.receipt.Receipt.Native.TransactionProofRef.Path)
	ports.Target = &nativeClosingChangedTarget{WorkflowTargetGit: ports.Target, change: func() {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
			t.Fatal(err)
		}
	}}

	failed, err := store.CloseWorkflowInvocationSet(t.Context(), f.fixture.f.entry.InvocationID, "dev", ports)
	if err == nil || failed.Outcome != "closing_validation_failed" || failed.Set.Set.ClosureRef != "" || len(failed.Receipt.Validation) != 1 || failed.Receipt.Validation[0].ExitCode != 0 {
		t.Fatal("changed native proof confirmed closure", err)
	}

	if !reflect.DeepEqual(before, nativeClosingPreservedFiles(t, f)) {
		t.Fatal("failed evidence fence rewrote receipts, members or frontier")
	}
}

// TestNativeClosingLost closes a reconciled real lost SSH exchange using its
// independent exact Git application proof. An immutable unknown transport fact
// cannot be confused with an unresolved publication or trigger another push.
func TestNativeClosingLost(t *testing.T) {
	f := newNativeClosingTransportFixture(t, "rebase", workflowPreparationGate("true"), true)
	ports := finalizeNativeClosingFixture(t, f)
	store, id := f.fixture.f.store, f.fixture.f.entry.InvocationID

	closed, err := store.CloseWorkflowInvocationSet(t.Context(), id, "dev", ports)
	if err != nil || closed.Outcome != "completed" || f.receipt.Receipt.Native.GitApplicationRef == nil {
		t.Fatal("reconciled lost exchange remained a closure barrier", err)
	}

	var response WorkflowGitPublicationResult
	if err := store.readWorkflowNativeEvidence(id, f.receipt.Receipt.Native.TransactionProofRef, &response); err != nil || response.State != "unknown" || f.fixture.publisher.calls.Load() != 1 {
		t.Fatal("closure rewrote or repeated the lost exchange", err)
	}
}
