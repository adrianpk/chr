package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeRecordReceipt(t *testing.T, store *WorkflowStore, candidate WorkflowIntegrationCandidate) WorkflowRecord {
	t.Helper()

	b := candidate.Native.WorkflowNativeRecordBinding
	remote := filepath.Join(t.TempDir(), "receipt.git")
	workflowIntegrationCommand(t, candidate.Lane.Worktree, "init", "--bare", "--quiet", remote)
	workflowIntegrationCommand(t, candidate.Lane.Worktree, "push", remote, b.OriginalSource+":"+b.SourceRef, b.ValidatedBase+":"+b.TargetRef)

	var capability WorkflowNativeCapability
	if err := store.readWorkflowNativeEvidence(b.InvocationID, b.CapabilityRef, &capability); err != nil {
		t.Fatal(err)
	}

	request := WorkflowGitPublication{Repository: b.SourceRepository, SourceRepository: b.SourceRepository, Directory: candidate.Lane.Worktree, SourceRef: b.SourceRef, TargetRef: b.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: candidate.Input.Git}
	sealWorkflowGitFixture(t, &request)

	transaction, err := NewWorkflowGitPublisher(b.SourceRepository, remote).PublishWorkflowGit(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	query := candidate.Input.Git.WorkflowIntegrationQuery
	query.SourceDirectory = candidate.Lane.Worktree

	proof, err := NewWorkflowIntegrationGit(b.SourceRepository, remote).VerifyPreparedWorkflowIntegration(t.Context(), query, candidate.Input.Git)
	if err != nil {
		t.Fatal(err)
	}

	pr := candidate.Input.Source.PullRequest
	pr.Merged, pr.State, pr.HeadCommit, pr.ResultCommit = true, "closed", b.PublishedSource, b.PreparedResult
	observation := WorkflowNativePublicationObservation{State: "merged", PullRequest: pr, Git: &proof, OriginalSourceCommit: b.OriginalSource, PublishedSourceCommit: b.PublishedSource}
	native := &WorkflowNativeReceipt{WorkflowNativeRecordBinding: b, CandidateHash: candidate.Native.CandidateHash, ActualMergeResult: b.PreparedResult}
	native.CandidateRef = retainNativeRecordEvidence(t, store, b.InvocationID, "candidate", candidate)
	native.TransactionProofRef = retainNativeRecordEvidence(t, store, b.InvocationID, "transaction", transaction)
	native.ObservationRefs = []WorkflowEvidenceRef{retainNativeRecordEvidence(t, store, b.InvocationID, "observation", observation)}
	r := workflowTestTask("native-delivery")
	r.Kind, r.Task, r.Path = "receipt", nil, "run/"+b.InvocationID+"/evidence/native-delivery.json"
	r.Receipt = &WorkflowReceipt{Schema: workflowNativeReceiptSchema, Native: native, Unit: b.Unit, CandidateID: candidate.Input.Source.CandidateID, SubmissionID: "native-submission", Provider: "forgejo", Repository: b.SourceRepository, PullRequest: b.CanonicalPR, Method: b.Method, SourceCommit: b.OriginalSource, BaseCommit: b.ValidatedBase, PreparedTree: b.PreparedTree, HistoryProof: proof.HistoryHash, ResultCommit: b.PreparedResult, VerifiedTree: proof.VerifiedTree, ObservedTarget: proof.ObservedTarget, VerificationRef: native.ObservationRefs[0].Path, ValidationRefs: []string{native.CandidateRef.Path}, ConfirmedAt: candidate.Validation.FinishedAt}

	return r
}

// TestNativeReceiptProof combines a real atomic Git application with retained
// original-PR proof and refuses incomplete receipts before writing their bytes.
func TestNativeReceiptProof(t *testing.T) {
	store, candidate := nativeRecordCandidate(t, "rebase")
	original := nativeRecordReceipt(t, store, candidate)

	cases := []struct {
		name   string
		change func(*WorkflowReceipt)
	}{
		{"version", func(r *WorkflowReceipt) { r.Schema = "chr-workflow-receipt/v3" }},
		{"missing", func(r *WorkflowReceipt) { r.Native = nil }},
		{"source", func(r *WorkflowReceipt) { r.SourceCommit = r.ResultCommit }},
		{"transaction", func(r *WorkflowReceipt) { r.Native.TransactionProofRef.Hash = "" }},
		{"observation", func(r *WorkflowReceipt) { r.Native.ObservationRefs = nil }},
		{"merge", func(r *WorkflowReceipt) { r.Native.ActualMergeResult = r.SourceCommit }},
		{"candidate", func(r *WorkflowReceipt) { r.CandidateID = "another-candidate" }},
		{"profile", func(r *WorkflowReceipt) { r.Native.CapabilityHash = contenthash.Sum([]byte("another-profile")) }},
		{"mechanism", func(r *WorkflowReceipt) { r.Native.Mechanism = "another-mechanism" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record := original
			receipt := *original.Receipt
			native := *receipt.Native
			receipt.Native, record.Receipt = &native, &receipt
			tc.change(record.Receipt)

			if _, err := store.CreateReceipt(t.Context(), record); err == nil {
				t.Fatal("created unproved native receipt")
			}

			if _, err := os.Stat(filepath.Join(store.layout.Root(), record.Path)); !os.IsNotExist(err) {
				t.Fatal("invalid receipt mutated durable evidence")
			}
		})
	}

	created, err := store.CreateReceipt(t.Context(), original)
	if err != nil {
		t.Fatal(err)
	}

	again, err := store.Read(t.Context(), created.Path)
	if err != nil || again.Receipt.SourceCommit != candidate.Native.OriginalSource || again.Receipt.Native.PublishedSource != candidate.Native.PreparedResult {
		t.Fatalf("lost original/published identity: %v", err)
	}

	testNativeReceiptPending(t, store, original)
}

func testNativeReceiptPending(t *testing.T, store *WorkflowStore, record WorkflowRecord) {
	t.Helper()

	var observation WorkflowNativePublicationObservation

	ref := record.Receipt.Native.ObservationRefs[0]
	if err := store.readWorkflowNativeEvidence(record.Receipt.Native.InvocationID, ref, &observation); err != nil {
		t.Fatal(err)
	}

	observation.State, observation.PullRequest.Merged = "observing", false
	ref = retainNativeRecordEvidence(t, store, record.Receipt.Native.InvocationID, "pending", observation)
	receipt := *record.Receipt
	native := *receipt.Native
	native.ObservationRefs, receipt.Native = []WorkflowEvidenceRef{ref}, &native

	record.Receipt, record.Path = &receipt, "run/"+native.InvocationID+"/evidence/pending-delivery.json"
	if _, err := store.CreateReceipt(t.Context(), record); err == nil {
		t.Fatal("pending original PR created a receipt")
	}
}

// TestNativeReceiptLegacy preserves historical receipt bytes and verification
// without inventing publication fields or upgrading their schema on a read.
func TestNativeReceiptLegacy(t *testing.T) {
	store := workflowTestStore(t)
	workflowTestCreate(t, store, "legacy-task")

	r := workflowTestTask("legacy-receipt")
	r.Kind, r.Task, r.Path = "receipt", nil, "run/legacy/evidence/delivery.json"
	object := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	r.Receipt = &WorkflowReceipt{Schema: WorkflowSchema, Unit: workflowstate.Ref{Kind: "ticket", ID: "legacy-task"}, CandidateID: "old-candidate", SubmissionID: "old-submission", Provider: "forgejo", Repository: "https://forge.example/team/project", PullRequest: "https://forge.example/team/project/pulls/1", Method: "rebase", SourceCommit: object, BaseCommit: object, PreparedTree: object, HistoryProof: "old-history", ResultCommit: object, VerifiedTree: object, ObservedTarget: object, VerificationRef: "old-proof", ValidationRefs: []string{"old-validation"}, ConfirmedAt: r.CreatedAt}

	created, err := store.CreateReceipt(t.Context(), r)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.layout.Root(), created.Path)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Snapshot(t.Context()); err != nil {
		t.Fatal(err)
	}

	again, err := store.Read(t.Context(), created.Path)
	if err != nil || again.Receipt.Native != nil || again.Receipt.Schema != WorkflowSchema {
		t.Fatalf("historical receipt changed: %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(before) != string(after) {
		t.Fatal("read rewrote historical receipt bytes")
	}

	assertNativeReceiptLegacyFields(t, again.Receipt)
}

func assertNativeReceiptLegacyFields(t *testing.T, receipt *WorkflowReceipt) {
	t.Helper()

	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}

	if _, exists := fields["native"]; exists {
		t.Fatal("read invented native proof")
	}
}
