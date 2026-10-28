package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ObserveWorkflowNativePublication reads the original PR through the existing
// forge adapter and independently fetches the canonical target's exact history.
// It never submits another push, merges/closes a PR, or rolls target history back.
func (adapter RemoteReviewAdapter) ObserveWorkflowNativePublication(ctx context.Context, binding workflowstate.NativePublicationBinding, git IntegrationGit) (workflowstate.NativePublicationObservation, error) {
	observed := workflowstate.NativePublicationObservation{ObservationContract: binding.Capability.ObservationContract, CanonicalPR: binding.PullRequest.URL, State: "observing", OriginalSourceCommit: binding.Git.SourceCommit, PublishedSourceCommit: binding.Git.ResultCommit}
	if err := validateWorkflowNativeBinding(binding); err != nil || git == nil {
		return observed, fmt.Errorf("invalid native observation binding")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	pr, err := adapter.ReadWorkflowPullRequest(ctx, binding.PullRequest.URL)
	if err != nil {
		return observed, err
	}

	observed.PullRequest, observed.ObservedSourceCommit = pr, pr.HeadCommit
	if condition := workflowNativeObservationCondition(binding, pr); condition != "" {
		observed.State, observed.Condition = "reconciliation_required", condition

		return observed, nil
	}

	return adapter.verifyWorkflowNativeObservation(ctx, binding, git, observed)
}

func validateWorkflowNativeBinding(binding workflowstate.NativePublicationBinding) error {
	if err := validateWorkflowNativeCapability(binding.Capability); err != nil {
		return err
	}

	if err := validateWorkflowNativeResult(binding); err != nil {
		return err
	}

	if binding.GitApplication != nil && !workflowNativeApplicationAttributed(binding) {
		return workflowstate.ErrConflict
	}

	capability, proof := binding.Capability, binding.Git
	expected := [6]string{capability.Repository, capability.PullRequestURL, capability.Method, capability.OriginalSourceCommit, capability.ExpectedBaseCommit, capability.PreparedResultCommit}

	actual := [6]string{proof.Repository, binding.PullRequest.URL, proof.Method, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit}
	if expected != actual {
		return workflowstate.ErrConflict
	}

	request := GitPublication{Repository: capability.Repository, SourceRepository: binding.PullRequest.HeadRepository, Directory: binding.SourceDirectory, SourceRef: capability.SourceRef, TargetRef: capability.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: proof}

	revision, err := GitPublicationRevision(request)
	if err != nil || revision != binding.Publication.CandidateRevision || !workflowNativeSourceBinding(NativeCapabilityRequest{PullRequest: binding.PullRequest, Publication: request}) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowNativeCapability(capability workflowstate.NativeCapability) error {
	sealed, err := sealWorkflowNativeCapability(capability)
	if err != nil || sealed.Revision != capability.Revision || !capability.AtomicSourceBase || !capability.ExactResult || len(capability.Unsupported) != 0 || capability.AdapterVersion != workflowNativeAdapterVersion || capability.Mechanism != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowNativeResult(binding workflowstate.NativePublicationBinding) error {
	proof, publication := binding.Git, binding.Publication
	if !publication.CommandsSent || (publication.State != "applied" && publication.State != "unknown") {
		return workflowstate.ErrConflict
	}

	expected := [5]string{"chr-git-publication-result/v1", proof.SourceCommit, proof.ResultCommit, proof.SourceCommit, proof.BaseCommit}

	actual := [5]string{publication.Schema, publication.OriginalSourceCommit, publication.ResultCommit, publication.AdvertisedSourceCommit, publication.AdvertisedTargetCommit}
	if expected != actual {
		return workflowstate.ErrConflict
	}

	if publication.State == "applied" && publication.PublishedSourceCommit != proof.ResultCommit {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowNativeObservationCondition(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot) string {
	original := binding.PullRequest
	expected := [4]string{original.URL, original.Repository, original.HeadRepository, original.HeadBranch}

	actual := [4]string{pr.URL, pr.Repository, pr.HeadRepository, pr.HeadBranch}
	if expected != actual {
		return "canonical_pr_identity_changed"
	}

	if pr.BaseBranch != original.BaseBranch {
		return "canonical_pr_retargeted"
	}

	return workflowNativeMetadataCondition(binding, pr)
}

func workflowNativeMetadataCondition(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot) string {
	if condition := workflowNativeMergeCondition(binding, pr); condition != "" {
		return condition
	}

	if !workflowNativeSourceCompatible(binding, pr) {
		return "canonical_source_changed"
	}

	if !pr.Merged && pr.State != "open" {
		return "canonical_pr_state_unknown"
	}

	return ""
}

func workflowNativeMergeCondition(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot) string {
	if !pr.Merged && pr.State == "closed" {
		return "canonical_pr_closed_unmerged"
	}

	if binding.ReceiptConfirmed && !pr.Merged {
		return "canonical_receipt_pr_unmerged"
	}

	compatible := pr.ResultCommit == binding.Git.ResultCommit || workflowNativeSourceResultCompatible(binding.Capability.ObservationContract, binding.Git, pr)
	if pr.Merged && (pr.State != "closed" || !compatible) {
		return "canonical_pr_result_mismatch"
	}

	return ""
}

func workflowNativeSourceCompatible(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot) bool {
	if pr.HeadCommit == binding.Git.SourceCommit || pr.HeadCommit == binding.Git.ResultCommit {
		return true
	}

	return workflowstate.ValidCommitID(pr.HeadCommit) && (binding.Publication.State == "applied" || workflowNativeApplicationAttributed(binding)) && pr.Merged
}

func workflowNativeApplicationAttributed(binding workflowstate.NativePublicationBinding) bool {
	observation := binding.GitApplication
	if !workflowNativeApplicationHeader(binding, observation) {
		return false
	}

	proof := *observation.Git
	expected := [6]string{binding.Git.Repository, binding.Git.Method, binding.Git.SourceCommit, binding.Git.BaseCommit, binding.Git.ResultCommit, binding.Git.HistoryHash}
	actual := [6]string{proof.Repository, proof.Method, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit, proof.HistoryHash}

	return expected == actual && proof.PreparedTree == binding.Git.PreparedTree && proof.VerifiedTree == binding.Git.VerifiedTree && workflowintegrate.CheckProof(proof.IntegrationQuery, proof, workflowIntegrationRepairs{}) == nil
}

func workflowNativeApplicationHeader(binding workflowstate.NativePublicationBinding, observation *workflowstate.NativeGitObservation) bool {
	if observation == nil || observation.Git == nil {
		return false
	}

	return observation.SchemaVersion == 1 && observation.Condition == "result_verified" && observation.CandidateRevision == binding.Publication.CandidateRevision && observation.SourceCommit == binding.Git.ResultCommit && observation.TargetCommit == observation.Git.ObservedTarget
}

func (adapter RemoteReviewAdapter) verifyWorkflowNativeObservation(ctx context.Context, binding workflowstate.NativePublicationBinding, git IntegrationGit, observed workflowstate.NativePublicationObservation) (workflowstate.NativePublicationObservation, error) {
	proof, err := verifyWorkflowNativeTarget(ctx, binding, git)
	if err != nil {
		observed.State, observed.Condition = "reconciliation_required", "canonical_target_unverified"
		if workflowNativeReceiptHistoryDiverged(binding, proof, err) {
			observed.Condition = "canonical_target_history_diverged"
			observed.ObservedTargetCommit = proof.ObservedTarget
		}

		return observed, err
	}

	observed.Git = &proof
	observed.ObservedTargetCommit = proof.ObservedTarget

	source, err := adapter.ReadWorkflowBranch(ctx, binding.PullRequest.Repository, binding.PullRequest.HeadBranch)
	if err != nil {
		return observed, err
	}

	observed.ObservedSourceCommit = source
	if binding.Publication.State == "unknown" && source != binding.Git.ResultCommit && !workflowNativeApplicationAttributed(binding) {
		observed.State, observed.Condition = "reconciliation_required", "published_source_unattributed"

		return observed, nil
	}

	pr, err := adapter.ReadWorkflowPullRequest(ctx, binding.PullRequest.URL)
	if err != nil {
		return observed, err
	}

	observed.PullRequest = pr

	return finishWorkflowNativeObservation(binding, observed), nil
}

func workflowNativeReceiptHistoryDiverged(binding workflowstate.NativePublicationBinding, proof workflowstate.IntegrationProof, err error) bool {
	return binding.ReceiptConfirmed && errors.Is(err, workflowstate.ErrHistoryDivergence) && workflowstate.ValidCommitID(proof.ObservedTarget)
}

func verifyWorkflowNativeTarget(ctx context.Context, binding workflowstate.NativePublicationBinding, git IntegrationGit) (workflowstate.IntegrationProof, error) {
	query := binding.Git.IntegrationQuery
	query.SourceDirectory = binding.SourceDirectory

	proof, err := observeWorkflowPreparedGit(ctx, git, query, binding.Git)
	if err != nil {
		return proof, err
	}

	if err := workflowintegrate.CheckProof(query, proof, workflowIntegrationRepairs{}); err != nil {
		return proof, err
	}

	expected := [4]string{binding.Git.BaseCommit, binding.Git.VerifiedTree, binding.Git.PreparedTree, binding.Git.HistoryHash}

	actual := [4]string{proof.BaseCommit, proof.VerifiedTree, proof.PreparedTree, proof.HistoryHash}
	if expected != actual {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

func finishWorkflowNativeObservation(binding workflowstate.NativePublicationBinding, observed workflowstate.NativePublicationObservation) workflowstate.NativePublicationObservation {
	pr := observed.PullRequest
	if condition := workflowNativeObservationCondition(binding, pr); condition != "" {
		observed.State, observed.Condition = "reconciliation_required", condition

		return observed
	}

	if pr.Merged && (pr.ResultCommit == binding.Git.ResultCommit || workflowNativeSourceResultObserved(observed)) {
		observed.State = "merged"
		if observed.ObservedSourceCommit != binding.Git.SourceCommit && observed.ObservedSourceCommit != binding.Git.ResultCommit {
			observed.LaterSourceCommit = observed.ObservedSourceCommit
		}

		return observed
	}

	if !binding.ObservationDeadline.IsZero() && !time.Now().Before(binding.ObservationDeadline) {
		observed.State, observed.Condition = "reconciliation_required", "canonical_pr_observation_deadline"
	}

	return observed
}

func observeWorkflowPreparedGit(ctx context.Context, git IntegrationGit, query workflowstate.IntegrationQuery, expected workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error) {
	if prepared, ok := git.(PreparedIntegrationGit); ok {
		return prepared.VerifyPreparedWorkflowIntegration(ctx, query, expected)
	}

	if expected.Provenance != nil {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	return git.VerifyWorkflowIntegration(ctx, query)
}
