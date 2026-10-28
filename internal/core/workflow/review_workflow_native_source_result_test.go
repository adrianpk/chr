package workflow

import (
	"strings"
	"testing"
	"time"
)

// TestNativeSourceResult reproduces a GitHub indirect rebase whose original PR
// reports S while its actual source, target and independently verified history are P.
func TestNativeSourceResult(t *testing.T) {
	fixture, binding, pr := nativeSourceResultFixture(t)
	setNativePRFixture(t, fixture, pr)

	observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote))
	if err != nil || observed.State != "merged" || observed.Git == nil {
		t.Fatalf("exact Git result and original PR attribution were not reconciled: %s, %s, %v", observed.State, observed.Condition, err)
	}

	if observed.PullRequest.ResultCommit != binding.Git.SourceCommit || observed.Git.ResultCommit != binding.Git.ResultCommit {
		t.Fatal("canonical source attribution was substituted for the delivered Git result")
	}

	if !workflowNativeCanonicalResultValid(binding, observed) {
		t.Fatal("canonical receipt observation rejected the admitted source attribution")
	}

	checkNativeReceiptCompatibility(t, &nativeCertificationEvidence{Native: nativeLiveEvidence{Binding: binding, Observations: []NativePublicationObservation{observed}}})

	observed.ObservationContract = "github_same_pr_git_merge/v1"
	if workflowNativeCanonicalResultValid(binding, observed) {
		t.Fatal("canonical observation accepted a different interpretation")
	}
}

func nativeSourceResultFixture(t *testing.T) (nativeCapabilityFixture, NativePublicationBinding, PullRequestSnapshot) {
	t.Helper()

	fixture := newNativeCapabilityFixture(t, "github")
	fixture.policy.policy.ObservationContract = workflowNativeSourceResultContract
	binding := nativeObservationFixture(t, fixture)
	pr := binding.PullRequest
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, binding.Git.SourceCommit
	pr.HeadCommit, pr.BaseCommit = binding.Git.ResultCommit, binding.Git.ResultCommit

	return fixture, binding, pr
}

// TestNativeSourceResultGuards retains reconciliation for foreign results and
// identities, and waits for source metadata rather than prematurely confirming it.
func TestNativeSourceResultGuards(t *testing.T) {
	cases := []struct {
		name, state string
		change      func(*NativePublicationBinding, *PullRequestSnapshot)
	}{
		{"source_metadata_lag", "observing", func(b *NativePublicationBinding, pr *PullRequestSnapshot) {
			pr.HeadCommit = b.Git.SourceCommit
		}},
		{"source_metadata_deadline", "reconciliation_required", func(b *NativePublicationBinding, pr *PullRequestSnapshot) {
			pr.HeadCommit, b.ObservationDeadline = b.Git.SourceCommit, time.Now().Add(-time.Minute)
		}},
		{"foreign_result", "reconciliation_required", func(_ *NativePublicationBinding, pr *PullRequestSnapshot) {
			pr.ResultCommit = strings.Repeat("e", 40)
		}},
		{"retargeted", "reconciliation_required", func(_ *NativePublicationBinding, pr *PullRequestSnapshot) { pr.BaseBranch = "other" }},
		{"closed_unmerged", "reconciliation_required", func(_ *NativePublicationBinding, pr *PullRequestSnapshot) { pr.Merged = false }},
		{"foreign_pr", "reconciliation_required", func(_ *NativePublicationBinding, pr *PullRequestSnapshot) { pr.URL += "0" }},
		{"lost_response", "merged", func(b *NativePublicationBinding, _ *PullRequestSnapshot) {
			b.Publication.State, b.Publication.PublishedSourceCommit = "unknown", ""
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture, binding, pr := nativeSourceResultFixture(t)
			test.change(&binding, &pr)
			setNativePRFixture(t, fixture, pr)

			observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote))
			if err != nil || observed.State != test.state {
				t.Fatalf("source attribution guard: %s, %s, %v", observed.State, observed.Condition, err)
			}

			if observed.State == "observing" && !workflowNativeCanonicalResultValid(binding, observed) {
				t.Fatal("source metadata lag could not retain its pending observation")
			}
		})
	}
}

// TestNativeSourceRepair accepts an exact rebase result containing a declared
// integration repair without erasing original source attribution or replay proof.
func TestNativeSourceRepair(t *testing.T) {
	fixture := newNativeCapabilityFixture(t, "github")
	fixture.policy.policy.ObservationContract = workflowNativeSourceResultContract
	f := workflowGitFixture{request: fixture.request.Publication, remote: fixture.remote}
	proof, lane := nativeCombinedPreparation(t, f)
	proof = nativeReplayFix(t, proof, lane)
	fixture.request.Publication.Git, fixture.request.Publication.Directory = proof, lane.Worktree
	sealWorkflowGitFixture(t, &fixture.request.Publication)

	binding := nativeObservationFixture(t, fixture)
	pr := binding.PullRequest
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, proof.SourceCommit
	pr.HeadCommit, pr.BaseCommit = proof.ResultCommit, proof.ResultCommit
	setNativePRFixture(t, fixture, pr)

	observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, NewWorkflowIntegrationGit(proof.Repository, fixture.remote))
	if err != nil || observed.State != "merged" {
		t.Fatal("declared repair lost exact-result source attribution", err)
	}

	checkNativeReceiptCompatibility(t, &nativeCertificationEvidence{Native: nativeLiveEvidence{Binding: binding, Observations: []NativePublicationObservation{observed}}})
}

// TestNativeSourceLater preserves P after both refs advance, while an uncertain
// push still requires retained application proof before attributing a later head.
func TestNativeSourceLater(t *testing.T) {
	fixture, binding, pr := nativeSourceResultFixture(t)
	git := NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote)

	prior, err := verifyWorkflowNativeTarget(t.Context(), binding, git)
	if err != nil {
		t.Fatal(err)
	}

	later := workflowIntegrationCommit(t, binding.SourceDirectory, "later.txt", "Later external work")
	workflowIntegrationCommand(t, binding.SourceDirectory, "push", fixture.remote, later+":"+binding.Capability.SourceRef, later+":"+binding.Capability.TargetRef)

	pr.HeadCommit, pr.BaseCommit = later, later
	setNativePRFixture(t, fixture, pr)

	fixture.runner.outputs[fixture.prefix+"/branches/chr%2Fworkflow%2Ffixture"] = []byte(`{"commit":{"sha":"` + later + `"}}`)

	observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git)
	if err != nil || observed.State != "merged" || observed.LaterSourceCommit != later || observed.Git.ObservedTarget != later {
		t.Fatal("later refs lost the original verified result", err)
	}

	receipt := nativeSourceResultReceipt(binding, observed)
	if validateWorkflowNativeReceiptObservation(receipt, observed) != nil {
		t.Fatal("later refs substituted for the original receipt proof")
	}

	binding.Publication.State, binding.Publication.PublishedSourceCommit = "unknown", ""

	observed, err = fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git)
	if err != nil || observed.State != "reconciliation_required" {
		t.Fatal("uncertain transaction attributed later source without application proof", err)
	}

	checkNativeLaterAttribution(t, fixture, binding, prior, later)
}

// TestNativeSourceContract pins receipt interpretation to immutable capability
// evidence after reopening the store, and rejects a different certified contract.
func TestNativeSourceContract(t *testing.T) {
	_, binding, _ := nativeSourceResultFixture(t)
	store := workflowTestStore(t)
	native := NativeReceipt{ObservationContract: workflowNativeSourceResultContract}
	native.InvocationID, native.SourceRepository = "source-contract", binding.Git.Repository
	native.SourceRef, native.TargetRef = binding.Capability.SourceRef, binding.Capability.TargetRef
	native.CapabilityRef = retainNativeRecordEvidence(t, store, native.InvocationID, "capability", binding.Capability)
	native.HistoryManifestRef = retainNativeRecordEvidence(t, store, native.InvocationID, "history", NativeHistoryManifest{Schema: "chr-native-history/v1", Git: binding.Git})
	store = NewWorkflowStore(store.layout, store.projectID)

	if err := store.checkWorkflowNativeTransactionRevision(native, binding.Publication); err != nil {
		t.Fatal("restart lost the admitted observation contract", err)
	}

	native.ObservationContract = "github_same_pr_git_merge/v1"
	if err := store.checkWorkflowNativeTransactionRevision(native, binding.Publication); err == nil {
		t.Fatal("receipt changed the immutable observation contract")
	}
}

// TestNativeSourceCapability requires explicit GitHub certification; the same
// contract is denied for another forge before any publication inspection.
func TestNativeSourceCapability(t *testing.T) {
	for _, forge := range []string{"github", "forgejo"} {
		t.Run(forge, func(t *testing.T) {
			fixture := newNativeCapabilityFixture(t, forge)
			fixture.policy.policy.ObservationContract = workflowNativeSourceResultContract

			capability, err := fixture.adapter.WorkflowNativeIntegrationCapability(t.Context(), fixture.request, fixture.policy, fixture.git)
			if err != nil {
				t.Fatal(err)
			}

			if forge == "forgejo" {
				if capability.AtomicSourceBase || fixture.git.calls != 0 || len(capability.Unsupported) == 0 {
					t.Fatal("GitHub source attribution admitted on another forge")
				}

				return
			}

			fixture.policy.policy.ObservationContract = "github_same_pr_git_merge/v1"
			if err := fixture.adapter.RecheckWorkflowNativeCapability(t.Context(), fixture.request, capability, fixture.policy, fixture.git); err == nil {
				t.Fatal("changed certification reused the old admission")
			}
		})
	}
}

// TestNativeSourceReceipt rejects fabricated or incomplete source attribution
// even when the caller retains an otherwise valid exact-result Git observation.
func TestNativeSourceReceipt(t *testing.T) {
	fixture, binding, pr := nativeSourceResultFixture(t)
	setNativePRFixture(t, fixture, pr)

	observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote))
	if err != nil || observed.State != "merged" {
		t.Fatal("source receipt fixture is not verified", err)
	}

	receipt := nativeSourceResultReceipt(binding, observed)
	if !workflowNativeReceiptResultMatches(receipt) || validateWorkflowNativeReceiptObservation(receipt, observed) != nil {
		t.Fatal("verified source attribution did not preserve an exact-result receipt")
	}

	checkNativeSourceReceiptDenials(t, receipt, observed)
}

func nativeSourceResultReceipt(binding NativePublicationBinding, observed NativePublicationObservation) Receipt {
	proof := *observed.Git
	native := &NativeReceipt{
		NativeRecordBinding: NativeRecordBinding{CanonicalPR: binding.PullRequest.URL, SourceRepository: proof.Repository, SourceRef: binding.Capability.SourceRef, TargetRef: binding.Capability.TargetRef, Method: proof.Method, OriginalSource: proof.SourceCommit, PublishedSource: proof.ResultCommit, ValidatedBase: proof.BaseCommit, PreparedResult: proof.ResultCommit, PreparedTree: proof.PreparedTree},
		ActualMergeResult:   observed.PullRequest.ResultCommit, ObservationContract: observed.ObservationContract,
	}

	return Receipt{Native: native, Repository: proof.Repository, Method: proof.Method, SourceCommit: proof.SourceCommit, BaseCommit: proof.BaseCommit, PreparedTree: proof.PreparedTree, ResultCommit: proof.ResultCommit, VerifiedTree: proof.VerifiedTree, ObservedTarget: proof.ObservedTarget, HistoryProof: proof.HistoryHash}
}

func checkNativeSourceReceiptDenials(t *testing.T, receipt Receipt, observed NativePublicationObservation) {
	t.Helper()

	cases := []struct {
		name   string
		change func(*Receipt, *NativePublicationObservation)
	}{
		{"missing_receipt_contract", func(r *Receipt, _ *NativePublicationObservation) { r.Native.ObservationContract = "" }},
		{"missing_observation_contract", func(_ *Receipt, o *NativePublicationObservation) { o.ObservationContract = "" }},
		{"unverified_source", func(_ *Receipt, o *NativePublicationObservation) {
			o.ObservedSourceCommit = o.Git.SourceCommit
		}},
		{"stale_source_metadata", func(_ *Receipt, o *NativePublicationObservation) {
			o.PullRequest.HeadCommit = o.Git.SourceCommit
		}},
		{"missing_git", func(_ *Receipt, o *NativePublicationObservation) { o.Git = nil }},
		{"foreign_replay", func(_ *Receipt, o *NativePublicationObservation) {
			o.Git.History[len(o.Git.History)-1].Source = strings.Repeat("e", 40)
		}},
		{"wrong_method", func(_ *Receipt, o *NativePublicationObservation) { o.Git.Method = "squash" }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			copyReceipt, copyObservation := receipt, observed
			native, proof := *receipt.Native, *observed.Git
			proof.History = append([]IntegrationCommit(nil), proof.History...)
			copyReceipt.Native, copyObservation.Git = &native, &proof
			test.change(&copyReceipt, &copyObservation)

			if validateWorkflowNativeReceiptObservation(copyReceipt, copyObservation) == nil {
				t.Fatal("incomplete source attribution authorized receipt proof")
			}
		})
	}
}
