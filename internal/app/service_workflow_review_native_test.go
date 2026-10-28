package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type serviceNativeReviewFixture struct {
	*serviceWorkflowPreparationFixture
	policy string
}

func (p *serviceNativeReviewFixture) WorkflowIntegrationCapability(ctx context.Context, pr workflowcore.PullRequestSnapshot, method string) (workflowcore.PublicationCapability, error) {
	capability, err := p.serviceWorkflowPreparationFixture.WorkflowIntegrationCapability(ctx, pr, method)
	capability.Mechanism = workspaceconfig.WorkflowPublicationGitAtomicDualRef

	return capability, err
}

func (p *serviceNativeReviewFixture) WorkflowPreparedNativeCapability(_ context.Context, pr workflowcore.PullRequestSnapshot, proof workflowcore.IntegrationProof) (workflowcore.NativeCapability, error) {
	capability := workflowcore.NativeCapability{
		PublicationCapability: workflowcore.PublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: proof.Method, Mechanism: workspaceconfig.WorkflowPublicationGitAtomicDualRef, AtomicSourceBase: true, ExactResult: true},
		SourceRef:             "refs/heads/" + pr.HeadBranch, TargetRef: "refs/heads/" + pr.BaseBranch, OwnershipRevision: "service-review/v1",
		OriginalSourceCommit: proof.SourceCommit, ExpectedBaseCommit: proof.BaseCommit, PreparedResultCommit: proof.ResultCommit,
		PolicyFingerprint: serviceNativeReviewHash([]byte(p.policy)), ServerCapabilityFingerprint: serviceNativeReviewHash([]byte("server")), CredentialIdentityFingerprint: serviceNativeReviewHash([]byte("publisher")),
		ObservationContract: "forgejo_same_pr_git_auto_merge/v1", AdapterVersion: "chr-native-publication/v1",
	}

	data, err := json.Marshal(capability)
	capability.Revision = serviceNativeReviewHash(data)

	return capability, err
}

func serviceNativeReviewHash(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

// TestServiceNativeReview exercises the shared service with real native host
// preparation. A changed publication policy rejects every adapter-facing
// approval and guard operation without creating a second PR or repeating work.
func TestServiceNativeReview(t *testing.T) {
	operations, id, original := serviceWorkflowPreparationReady(t)
	provider := &serviceNativeReviewFixture{original, "policy-v1"}
	operations.workflowPullRequests = provider
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "review:\n  mode: human_review\nwork:\n  merge_strategy: rebase\n  integration:\n    publication_profile: git_atomic_dual_ref\n")

	if _, err := operations.ReviewWorkflowSlice(t.Context(), id, nil); err == nil {
		t.Fatal("service granted native source-only review")
	}

	serviceNativeReviewAdmit(t, operations, id)

	candidate, err := operations.PrepareWorkflowSliceIntegration(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	binding, err := operations.ReviewWorkflowSlice(t.Context(), id, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ApproveWorkflowSlice(t.Context(), id, binding.InputHash); err != nil {
		t.Fatal(err)
	}

	input, err := operations.CheckWorkflowReviewGrant(t.Context(), id)
	if err != nil || input.Native == nil || input.Native.Candidate.CandidateHash != candidate.Native.CandidateHash {
		t.Fatal("service lost native review binding", err)
	}

	if _, err := operations.CheckWorkflowPublicationGuard(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	provider.policy = "policy-v2"

	serviceNativeReviewReject(t, operations, id, binding)

	if provider.creates != 1 {
		t.Fatal("service native review repeated PR creation")
	}
}

func serviceNativeReviewAdmit(t *testing.T, operations service, id string) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := store.AdmitWorkflowIntegration(t.Context(), workflowcore.IntegrationAdmission{InvocationID: id, ExpectedRevision: run.Revision, Unit: run.Lane.Unit, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"})
	if err != nil {
		t.Fatal(err)
	}

	for _, state := range []string{"preparing", "validating"} {
		run, err = store.ReadWorkflow(t.Context(), id)
		if err != nil {
			t.Fatal(err)
		}

		entry, err = store.TransitionWorkflowIntegration(t.Context(), workflowcore.IntegrationTransition{InvocationID: id, EntryID: entry.ID, InvocationRevision: run.Revision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, State: state})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func serviceNativeReviewReject(t *testing.T, operations service, id string, binding workflowcore.ReviewBinding) {
	t.Helper()

	if _, err := operations.ApproveWorkflowSlice(t.Context(), id, binding.InputHash); err == nil {
		t.Fatal("service reused changed native approval")
	}

	if _, err := operations.ReviewWorkflowSlice(t.Context(), id, nil); err == nil {
		t.Fatal("service reused changed native review")
	}

	if _, err := operations.CheckWorkflowReviewGrant(t.Context(), id); err == nil {
		t.Fatal("service reused changed native grant")
	}

	if _, err := operations.CheckWorkflowPublicationGuard(t.Context(), id); err == nil {
		t.Fatal("service reused changed native publication guard")
	}
}
