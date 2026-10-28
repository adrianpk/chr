package workflow

import (
	"testing"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type nativeClosingFixture struct {
	fixture nativeCanonicalFixture
	receipt workflowstate.Record
	setPath string
}

func newNativeClosingFixture(t *testing.T, method string, gate workflowstate.Validation) nativeClosingFixture {
	t.Helper()

	return newNativeClosingTransportFixture(t, method, gate, false)
}

func newNativeClosingTransportFixture(t *testing.T, method string, gate workflowstate.Validation, lost bool) nativeClosingFixture {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)
	opts.Plan.Slices = opts.Plan.Slices[:1]
	opts.Plan.ClosingValidation = gate
	f := workflowPreparationPlanFixture(t, method, "slice_set", workflowPreparationGate("test -f result.txt"), store, opts)
	writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "work", "integration.publication_profile", workspaceconfig.WorkflowPublicationGitAtomicDualRef)

	entry := nativeRoleAdmit(t, store, f.run.ID)
	entry = nativeRoleTransition(t, store, entry, "preparing")
	nativeRoleTransition(t, store, entry, "validating")

	provider := &nativePreparationProviderFixture{f.provider, "policy-v1"}

	candidate, err := PrepareIntegration(t.Context(), store, f.run.ID, provider, f.git)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ReviewSlice(t.Context(), store, f.run.ID, provider, nil); err != nil {
		t.Fatal(err)
	}

	entry = nativeRoleTransition(t, store, nativeReviewEntry(t, store, f.run.ID), "ready_to_publish")
	fixture, receipt := nativeFanoutReceipt(t, nativeCanonicalSubmissionFixture(t, nativeSubmissionFixture{store, entry, candidate, provider}, lost))

	return nativeClosingFixture{fixture, receipt, f.hierarchy.Set.Path}
}

func finalizeNativeClosingFixture(t *testing.T, f nativeClosingFixture) FrontierPorts {
	t.Helper()

	store, id := f.fixture.f.store, f.fixture.f.entry.InvocationID
	if _, err := store.FinalizeWorkflowSlice(t.Context(), id, nil, nil); err != nil {
		t.Fatal(err)
	}

	ports := nativeReceiptFrontierPorts(t, f.fixture, f.receipt.Receipt.ResultCommit)
	if _, err := store.AdvanceWorkflowFrontier(t.Context(), id, ports); err != nil {
		t.Fatal(err)
	}

	return ports
}

// TestNativeClosingGate runs the exact host gate after real atomic publication
// and settled fanout. All methods retain the original source, native observation
// and member receipts rather than substituting a legacy verification image.
func TestNativeClosingGate(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { nativeClosingGate(t, method) })
	}
}

func nativeClosingGate(t *testing.T, method string) {
	t.Helper()

	f := newNativeClosingFixture(t, method, workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"git", "rev-parse", "HEAD"}}})
	ports := finalizeNativeClosingFixture(t, f)
	store, id := f.fixture.f.store, f.fixture.f.entry.InvocationID

	closed, err := CloseInvocationSet(t.Context(), store, id, "dev", ports)
	if err != nil {
		t.Fatal("native receipt could not enter its exact set gate", err)
	}

	if closed.Outcome != "completed" || len(closed.Receipt.Members) != 3 || len(closed.Receipt.Slices) != 1 || len(closed.Receipt.Validation) != 1 || closed.Receipt.Validation[0].Stdout != f.receipt.Receipt.ResultCommit+"\n" {
		t.Fatal("closure lost intended receipts or ran another target")
	}

	checkNativeClosingGateEvidence(t, f, closed)
}

func checkNativeClosingGateEvidence(t *testing.T, f nativeClosingFixture, closed workflowcloseout.Result) {
	t.Helper()

	proof := closed.Receipt.Slices[0]
	expected := [4]string{f.receipt.Path, f.receipt.SourceHash, f.receipt.Receipt.VerificationRef, f.receipt.Receipt.SourceCommit}
	actual := [4]string{proof.Receipt.Ref, proof.Receipt.Hash, proof.Verification.Ref, proof.Proof.SourceCommit}

	if actual != expected {
		t.Fatal("set closure replaced native receipt identity")
	}

	status, err := f.fixture.f.store.QueryWorkflowIntegrationSetClosure(t.Context(), f.setPath)
	if err != nil || !status.CanClose || !status.ReceiptsVerified || !status.GateVerified || f.fixture.publisher.calls.Load() != 1 {
		t.Fatal("native closure status lost proof or repeated publication", status, err)
	}
}
