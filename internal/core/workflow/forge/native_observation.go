package forge

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativePullRequests observes canonical PR metadata and branch identities.
type NativePullRequests interface {
	ReadWorkflowPullRequest(context.Context, string) (workflowstate.PullRequestSnapshot, error)
	ReadWorkflowBranch(context.Context, string, string) (string, error)
}

// NativeGit observes exact canonical Git history; it cannot publish refs.
type NativeGit interface {
	VerifyWorkflowIntegration(context.Context, workflowstate.IntegrationQuery) (workflowstate.IntegrationProof, error)
}

// PreparedNativeGit additionally verifies retained attributed history.
type PreparedNativeGit interface {
	VerifyPreparedWorkflowIntegration(context.Context, workflowstate.IntegrationQuery, workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error)
}

// NativeProofs checks immutable method/history proof through host composition.
type NativeProofs interface {
	CheckNativeProof(workflowstate.IntegrationQuery, workflowstate.IntegrationProof) error
}
type nativeObservationReaders struct {
	Provider NativePullRequests
	Proofs   NativeProofs
}

// ObserveNativePublication reads the original PR through the existing
// forge adapter and independently fetches the canonical target's exact history.
// It never submits another push, merges/closes a PR, or rolls target history back.
func ObserveNativePublication(ctx context.Context, binding workflowstate.NativePublicationBinding, adapter NativePullRequests, git NativeGit, proofs NativeProofs) (workflowstate.NativePublicationObservation, error) {
	observed := workflowstate.NativePublicationObservation{ObservationContract: binding.Capability.ObservationContract, CanonicalPR: binding.PullRequest.URL, State: workflowstate.NativeObserving, OriginalSourceCommit: binding.Git.SourceCommit, PublishedSourceCommit: binding.Git.ResultCommit}
	if err := CheckNativeBinding(binding, proofs); err != nil || git == nil {
		return observed, fmt.Errorf("invalid native observation binding")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	pr, err := adapter.ReadWorkflowPullRequest(ctx, binding.PullRequest.URL)
	if err != nil {
		return observed, err
	}

	observed.PullRequest, observed.ObservedSourceCommit = pr, pr.HeadCommit
	if condition := NativeObservationCondition(binding, pr, proofs); condition != "" {
		observed.State, observed.Condition = workflowstate.NativeReconciliationRequired, condition

		return observed, nil
	}

	return verifyNativeObservation(ctx, binding, nativeObservationReaders{Provider: adapter, Proofs: proofs}, git, observed)
}

// CheckNativeBinding checks capability, transport attribution and the original source identity.
func CheckNativeBinding(binding workflowstate.NativePublicationBinding, proofs NativeProofs) error {
	if err := CheckNativeCapability(binding.Capability); err != nil {
		return err
	}

	if err := checkNativeResult(binding); err != nil {
		return err
	}

	if binding.GitApplication != nil && !NativeApplicationAttributed(binding, proofs) {
		return workflowstate.ErrConflict
	}

	capability, proof := binding.Capability, binding.Git
	expected := [6]string{capability.Repository, capability.PullRequestURL, capability.Method, capability.OriginalSourceCommit, capability.ExpectedBaseCommit, capability.PreparedResultCommit}

	actual := [6]string{proof.Repository, binding.PullRequest.URL, proof.Method, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit}
	if expected != actual {
		return workflowstate.ErrConflict
	}

	request := PublicationIdentity{Repository: capability.Repository, SourceRepository: binding.PullRequest.HeadRepository, SourceRef: capability.SourceRef, TargetRef: capability.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: proof}

	revision, err := PublicationRevision(request)
	if err != nil || revision != binding.Publication.CandidateRevision || !NativeSourceBinding(NativeCapabilityInput{PullRequest: binding.PullRequest, Publication: request}) {
		return workflowstate.ErrConflict
	}

	return nil
}

// CheckNativeCapability rejects a changed seal or an unsupported publication mechanism.
func CheckNativeCapability(capability workflowstate.NativeCapability) error {
	sealed, err := SealNativeCapability(capability)
	if err != nil || sealed.Revision != capability.Revision || !capability.AtomicSourceBase || !capability.ExactResult || len(capability.Unsupported) != 0 || capability.AdapterVersion != NativeAdapterVersion || capability.Mechanism != nativePublicationMechanism {
		return workflowstate.ErrConflict
	}

	return nil
}

func checkNativeResult(binding workflowstate.NativePublicationBinding) error {
	proof, publication := binding.Git, binding.Publication
	if !publication.CommandsSent || (publication.State != workflowstate.GitPublicationApplied && publication.State != workflowstate.GitPublicationUnknown) {
		return workflowstate.ErrConflict
	}

	expected := [5]string{"chr-git-publication-result/v1", proof.SourceCommit, proof.ResultCommit, proof.SourceCommit, proof.BaseCommit}

	actual := [5]string{publication.Schema, publication.OriginalSourceCommit, publication.ResultCommit, publication.AdvertisedSourceCommit, publication.AdvertisedTargetCommit}
	if expected != actual {
		return workflowstate.ErrConflict
	}

	if publication.State == workflowstate.GitPublicationApplied && publication.PublishedSourceCommit != proof.ResultCommit {
		return workflowstate.ErrConflict
	}

	return nil
}

// NativeObservationCondition classifies original PR identity, target, result and source changes.
func NativeObservationCondition(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot, proofs NativeProofs) string {
	original := binding.PullRequest
	expected := [4]string{original.URL, original.Repository, original.HeadRepository, original.HeadBranch}

	actual := [4]string{pr.URL, pr.Repository, pr.HeadRepository, pr.HeadBranch}
	if expected != actual {
		return "canonical_pr_identity_changed"
	}

	if pr.BaseBranch != original.BaseBranch {
		return "canonical_pr_retargeted"
	}

	return nativeMetadataCondition(binding, pr, proofs)
}

func nativeMetadataCondition(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot, proofs NativeProofs) string {
	if condition := nativeMergeCondition(binding, pr, proofs); condition != "" {
		return condition
	}

	if !nativeSourceCompatible(binding, pr, proofs) {
		return "canonical_source_changed"
	}

	if !pr.Merged && pr.State != "open" {
		return "canonical_pr_state_unknown"
	}

	return ""
}

func nativeMergeCondition(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot, proofs NativeProofs) string {
	if !pr.Merged && pr.State == "closed" {
		return "canonical_pr_closed_unmerged"
	}

	if binding.ReceiptConfirmed && !pr.Merged {
		return "canonical_receipt_pr_unmerged"
	}

	compatible := pr.ResultCommit == binding.Git.ResultCommit || NativeSourceResultCompatible(binding.Capability.ObservationContract, binding.Git, pr, proofs)
	if pr.Merged && (pr.State != "closed" || !compatible) {
		return "canonical_pr_result_mismatch"
	}

	return ""
}

func nativeSourceCompatible(binding workflowstate.NativePublicationBinding, pr workflowstate.PullRequestSnapshot, proofs NativeProofs) bool {
	if pr.HeadCommit == binding.Git.SourceCommit || pr.HeadCommit == binding.Git.ResultCommit {
		return true
	}

	return workflowstate.ValidCommitID(pr.HeadCommit) && (binding.Publication.State == workflowstate.GitPublicationApplied || NativeApplicationAttributed(binding, proofs)) && pr.Merged
}

// NativeApplicationAttributed requires the exact journaled Git application and complete method proof.
func NativeApplicationAttributed(binding workflowstate.NativePublicationBinding, proofs NativeProofs) bool {
	observation := binding.GitApplication
	if !nativeApplicationHeader(binding, observation) {
		return false
	}

	proof := *observation.Git
	expected := [6]string{binding.Git.Repository, binding.Git.Method, binding.Git.SourceCommit, binding.Git.BaseCommit, binding.Git.ResultCommit, binding.Git.HistoryHash}
	actual := [6]string{proof.Repository, proof.Method, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit, proof.HistoryHash}

	return expected == actual && proof.PreparedTree == binding.Git.PreparedTree && proof.VerifiedTree == binding.Git.VerifiedTree && proofs.CheckNativeProof(proof.IntegrationQuery, proof) == nil
}

func nativeApplicationHeader(binding workflowstate.NativePublicationBinding, observation *workflowstate.NativeGitObservation) bool {
	if observation == nil || observation.Git == nil {
		return false
	}

	return observation.SchemaVersion == 1 && observation.Condition == "result_verified" && observation.CandidateRevision == binding.Publication.CandidateRevision && observation.SourceCommit == binding.Git.ResultCommit && observation.TargetCommit == observation.Git.ObservedTarget
}

func verifyNativeObservation(ctx context.Context, binding workflowstate.NativePublicationBinding, readers nativeObservationReaders, git NativeGit, observed workflowstate.NativePublicationObservation) (workflowstate.NativePublicationObservation, error) {
	adapter, proofs := readers.Provider, readers.Proofs

	proof, err := VerifyNativeTarget(ctx, binding, git, proofs)
	if err != nil {
		observed.State, observed.Condition = workflowstate.NativeReconciliationRequired, "canonical_target_unverified"
		if nativeReceiptHistoryDiverged(binding, proof, err) {
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
	if binding.Publication.State == workflowstate.GitPublicationUnknown && source != binding.Git.ResultCommit && !NativeApplicationAttributed(binding, proofs) {
		observed.State, observed.Condition = workflowstate.NativeReconciliationRequired, "published_source_unattributed"

		return observed, nil
	}

	pr, err := adapter.ReadWorkflowPullRequest(ctx, binding.PullRequest.URL)
	if err != nil {
		return observed, err
	}

	observed.PullRequest = pr

	return finishNativeObservation(binding, observed, proofs), nil
}

func nativeReceiptHistoryDiverged(binding workflowstate.NativePublicationBinding, proof workflowstate.IntegrationProof, err error) bool {
	return binding.ReceiptConfirmed && errors.Is(err, workflowstate.ErrHistoryDivergence) && workflowstate.ValidCommitID(proof.ObservedTarget)
}

// VerifyNativeTarget re-proves the retained method and history against the fetched canonical target.
func VerifyNativeTarget(ctx context.Context, binding workflowstate.NativePublicationBinding, git NativeGit, proofs NativeProofs) (workflowstate.IntegrationProof, error) {
	query := binding.Git.IntegrationQuery
	query.SourceDirectory = binding.SourceDirectory

	proof, err := ObservePreparedGit(ctx, git, query, binding.Git)
	if err != nil {
		return proof, err
	}

	if err := proofs.CheckNativeProof(query, proof); err != nil {
		return proof, err
	}

	expected := [4]string{binding.Git.BaseCommit, binding.Git.VerifiedTree, binding.Git.PreparedTree, binding.Git.HistoryHash}

	actual := [4]string{proof.BaseCommit, proof.VerifiedTree, proof.PreparedTree, proof.HistoryHash}
	if expected != actual {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

func finishNativeObservation(binding workflowstate.NativePublicationBinding, observed workflowstate.NativePublicationObservation, proofs NativeProofs) workflowstate.NativePublicationObservation {
	pr := observed.PullRequest
	if condition := NativeObservationCondition(binding, pr, proofs); condition != "" {
		observed.State, observed.Condition = workflowstate.NativeReconciliationRequired, condition

		return observed
	}

	if pr.Merged && (pr.ResultCommit == binding.Git.ResultCommit || NativeSourceResultObserved(observed, proofs)) {
		observed.State = workflowstate.NativeMerged
		if observed.ObservedSourceCommit != binding.Git.SourceCommit && observed.ObservedSourceCommit != binding.Git.ResultCommit {
			observed.LaterSourceCommit = observed.ObservedSourceCommit
		}

		return observed
	}

	if !binding.ObservationDeadline.IsZero() && !time.Now().Before(binding.ObservationDeadline) {
		observed.State, observed.Condition = workflowstate.NativeReconciliationRequired, "canonical_pr_observation_deadline"
	}

	return observed
}

// ObservePreparedGit preserves the prepared-history observer and legacy-proof fallback contract.
func ObservePreparedGit(ctx context.Context, git NativeGit, query workflowstate.IntegrationQuery, expected workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error) {
	if prepared, ok := git.(PreparedNativeGit); ok {
		return prepared.VerifyPreparedWorkflowIntegration(ctx, query, expected)
	}

	if expected.Provenance != nil {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	return git.VerifyWorkflowIntegration(ctx, query)
}
