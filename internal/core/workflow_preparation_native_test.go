package core

import (
	"context"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativePreparationFixture struct {
	f        workflowPreparationFixture
	provider *nativePreparationProviderFixture
	entry    WorkflowIntegrationEntry
}

type nativePreparationProviderFixture struct {
	*workflowPreparationProviderFixture
	policyRevision string
}

func (p *nativePreparationProviderFixture) WorkflowIntegrationCapability(ctx context.Context, pr WorkflowPullRequestSnapshot, method string) (WorkflowPublicationCapability, error) {
	capability, err := p.workflowPreparationProviderFixture.WorkflowIntegrationCapability(ctx, pr, method)
	capability.Mechanism = WorkflowPublicationGitAtomicDualRef

	return capability, err
}

func (p *nativePreparationProviderFixture) WorkflowPreparedNativeCapability(_ context.Context, pr WorkflowPullRequestSnapshot, proof WorkflowIntegrationProof) (WorkflowNativeCapability, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fingerprint, err := workflowGitCapabilityFingerprint(map[string]bool{"atomic": true, "report-status": true})
	if err != nil {
		return WorkflowNativeCapability{}, err
	}

	capability := WorkflowNativeCapability{
		WorkflowPublicationCapability: WorkflowPublicationCapability{Repository: proof.Repository, PullRequestURL: pr.URL, Method: proof.Method, Mechanism: WorkflowPublicationGitAtomicDualRef, AtomicSourceBase: true, ExactResult: true},
		SourceRef:                     "refs/heads/" + pr.HeadBranch, TargetRef: "refs/heads/" + pr.BaseBranch, OwnershipRevision: "native-preparation/v1",
		PolicyFingerprint: contenthash.Sum([]byte(p.policyRevision)), ServerCapabilityFingerprint: fingerprint, CredentialIdentityFingerprint: contenthash.Sum([]byte("publisher")),
		ObservationContract: "forgejo_same_pr_git_auto_merge/v1", AdapterVersion: workflowNativeAdapterVersion,
		OriginalSourceCommit: proof.SourceCommit, ExpectedBaseCommit: proof.BaseCommit, PreparedResultCommit: proof.ResultCommit,
	}

	return sealWorkflowNativeCapability(capability)
}

func newNativePreparationFixture(t *testing.T, method string, gate workflowstate.Validation) nativePreparationFixture {
	t.Helper()

	return nativePreparationExtentFixture(t, method, "slice", gate)
}

func nativePreparationExtentFixture(t *testing.T, method, extent string, gate workflowstate.Validation) nativePreparationFixture {
	t.Helper()

	f := workflowPreparationExtentFixture(t, method, extent, gate)
	writeWorkflowTestConfigValue(t, f.store.layout.ProjectDir(), "work", "integration.publication_profile", WorkflowPublicationGitAtomicDualRef)

	entry := nativeRoleAdmit(t, f.store, f.run.ID)
	entry = nativeRoleTransition(t, f.store, entry, "preparing")
	entry = nativeRoleTransition(t, f.store, entry, "validating")

	return nativePreparationFixture{f, &nativePreparationProviderFixture{f.provider, "policy-v1"}, entry}
}

// TestNativePreparationMethods runs the real host gate on all configured Git
// histories. Immutable evidence and the queue pointer survive restart without
// repeating task work or using a worker-reported validation outcome.
func TestNativePreparationMethods(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { nativePreparationMethod(t, method) })
	}
}

func nativePreparationMethod(t *testing.T, method string) {
	t.Helper()

	fixture := newNativePreparationFixture(t, method, workflowPreparationGate("test -f result.txt && if test -f target.txt; then printf combined; else printf source; fi"))
	f := fixture.f

	candidate, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, fixture.provider, f.git)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowPreparationMethodFacts(t, f, candidate, method)
	assertWorkflowPreparationSource(t, f)
	checkNativePreparationEvidence(t, fixture, candidate)

	before := workflowTestRead(t, f.store, f.run.ID)
	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	again, err := restarted.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, fixture.provider, f.git)
	if err != nil || again.Validation.AttemptID != candidate.Validation.AttemptID || workflowTestRead(t, restarted, f.run.ID).Revision != before.Revision {
		t.Fatal("native retry repeated its gate or lost retained evidence", err)
	}
}

func checkNativePreparationEvidence(t *testing.T, fixture nativePreparationFixture, candidate WorkflowIntegrationCandidate) {
	t.Helper()

	f := fixture.f
	if candidate.Native == nil || candidate.Input.Native == nil || candidate.Native.EntryID != fixture.entry.ID || candidate.Input.Native.ClaimGeneration != fixture.entry.ClaimGeneration {
		t.Fatal("native gate lost its queue generation or full capability")
	}

	if err := f.store.checkWorkflowNativeCandidate(candidate); err != nil {
		t.Fatal("native gate evidence was not independently readable", err)
	}

	checkNativePreparationRetained(t, fixture, candidate)
}

func checkNativePreparationRetained(t *testing.T, fixture nativePreparationFixture, candidate WorkflowIntegrationCandidate) {
	t.Helper()

	f := fixture.f

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := f.store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil || entry.State != "reviewing" || !workflowNativePreparedEntryMatches(run, entry) || entry.Revision != fixture.entry.Revision+1 {
		t.Fatal("candidate and queue pointer did not advance together", err)
	}

	retained, err := f.store.readWorkflowPreparation(run)
	if err != nil || !workflowNativeJSONEqual(retained, candidate) {
		t.Fatal("retained native preparation changed its frozen host input", err)
	}
}

// TestNativePreparationFailure rejects clean textual integrations that fail the
// combined semantic gate, and gates that edit their prepared checkout. Failed
// observations remain retained without a candidate or publication grant.
func TestNativePreparationFailure(t *testing.T) {
	cases := []struct{ name, script string }{
		{"semantic", "test ! -f target.txt || grep -q required target.txt"},
		{"mutation", "if test -f target.txt; then printf changed >> target.txt; fi"},
		{"mutation failure", "if test -f target.txt; then printf changed >> target.txt; exit 1; fi"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newNativePreparationFixture(t, "rebase", workflowPreparationGate(test.script))
			f := fixture.f

			candidate, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, fixture.provider, f.git)
			if err == nil || candidate.Validation.Successful || candidate.Native != nil {
				t.Fatal("failed native gate produced a publishable candidate", err)
			}

			checkNativePreparationRejected(t, fixture)

			if test.name != "semantic" && workflowTestRead(t, f.store, f.run.ID).CandidateRepair != nil {
				t.Fatal("checkout mutation was routed as actionable semantic repair")
			}
		})
	}
}

func checkNativePreparationRejected(t *testing.T, fixture nativePreparationFixture) {
	t.Helper()

	f := fixture.f

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := f.store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil || entry.CandidateRef != nil || run.Preparation.State != "stale" || run.Preparation.ResultRef != "" {
		t.Fatal("rejected gate retained candidate authorization", err)
	}

	last := run.Attempts[len(run.Attempts)-1]
	if last.CheckpointRef == "" || last.Outcome != "interrupted" {
		t.Fatal("rejected gate lost its attempt observations")
	}

	if _, err := f.store.readWorkflowReviewImage(last.CheckpointRef, last.CheckpointHash); err != nil {
		t.Fatal("failed validation evidence was not retained", err)
	}

	if _, err := f.store.CheckWorkflowPublicationGuard(t.Context(), f.run.ID, fixture.provider); err == nil {
		t.Fatal("rejected gate authorized publication")
	}

	assertWorkflowPreparationSource(t, f)
}
