package core

import (
	"context"
	"fmt"
	"time"
)

// WorkflowNativePublicationBinding retains the original PR/source and exact
// prepared history independently of the rewritten published source head.
type WorkflowNativePublicationBinding struct {
	Capability          WorkflowNativeCapability     `json:"capability"`
	PullRequest         WorkflowPullRequestSnapshot  `json:"original_pull_request"`
	Git                 WorkflowIntegrationProof     `json:"prepared_git"`
	Publication         WorkflowGitPublicationResult `json:"publication"`
	SourceDirectory     string                       `json:"-"`
	ObservationDeadline time.Time                    `json:"observation_deadline"`
}

// WorkflowNativePublicationObservation distinguishes Git proof from canonical
// merged metadata. Only merged supplies both; no outcome issues a receipt.
type WorkflowNativePublicationObservation struct {
	State                 string                      `json:"state"`
	Condition             string                      `json:"condition,omitempty"`
	PullRequest           WorkflowPullRequestSnapshot `json:"pull_request"`
	Git                   *WorkflowIntegrationProof   `json:"git,omitempty"`
	OriginalSourceCommit  string                      `json:"original_source_commit"`
	PublishedSourceCommit string                      `json:"published_source_commit"`
	ObservedSourceCommit  string                      `json:"observed_source_commit,omitempty"`
	LaterSourceCommit     string                      `json:"later_source_commit,omitempty"`
}

// ObserveWorkflowNativePublication reads the original PR through the existing
// forge adapter and independently fetches the canonical target's exact history.
// It never submits another push, merges/closes a PR, or rolls target history back.
func (adapter RemoteReviewAdapter) ObserveWorkflowNativePublication(ctx context.Context, binding WorkflowNativePublicationBinding, git WorkflowIntegrationGit) (WorkflowNativePublicationObservation, error) {
	observed := WorkflowNativePublicationObservation{State: "observing", OriginalSourceCommit: binding.Git.SourceCommit, PublishedSourceCommit: binding.Git.ResultCommit}
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

func validateWorkflowNativeBinding(binding WorkflowNativePublicationBinding) error {
	if err := validateWorkflowNativeCapability(binding.Capability); err != nil {
		return err
	}

	if err := validateWorkflowNativeResult(binding); err != nil {
		return err
	}

	capability, proof := binding.Capability, binding.Git
	expected := [6]string{capability.Repository, capability.PullRequestURL, capability.Method, capability.OriginalSourceCommit, capability.ExpectedBaseCommit, capability.PreparedResultCommit}

	actual := [6]string{proof.Repository, binding.PullRequest.URL, proof.Method, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit}
	if expected != actual {
		return ErrWorkflowConflict
	}

	request := WorkflowGitPublication{Repository: capability.Repository, SourceRepository: binding.PullRequest.HeadRepository, Directory: binding.SourceDirectory, SourceRef: capability.SourceRef, TargetRef: capability.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: proof}

	revision, err := WorkflowGitPublicationRevision(request)
	if err != nil || revision != binding.Publication.CandidateRevision || !workflowNativeSourceBinding(WorkflowNativeCapabilityRequest{PullRequest: binding.PullRequest, Publication: request}) {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowNativeCapability(capability WorkflowNativeCapability) error {
	sealed, err := sealWorkflowNativeCapability(capability)
	if err != nil || sealed.Revision != capability.Revision || !capability.AtomicSourceBase || !capability.ExactResult || len(capability.Unsupported) != 0 || capability.AdapterVersion != workflowNativeAdapterVersion || capability.Mechanism != WorkflowPublicationGitAtomicDualRef {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowNativeResult(binding WorkflowNativePublicationBinding) error {
	proof, publication := binding.Git, binding.Publication
	if !publication.CommandsSent || (publication.State != "applied" && publication.State != "unknown") {
		return ErrWorkflowConflict
	}

	expected := [5]string{"chr-git-publication-result/v1", proof.SourceCommit, proof.ResultCommit, proof.SourceCommit, proof.BaseCommit}

	actual := [5]string{publication.Schema, publication.OriginalSourceCommit, publication.ResultCommit, publication.AdvertisedSourceCommit, publication.AdvertisedTargetCommit}
	if expected != actual {
		return ErrWorkflowConflict
	}

	if publication.State == "applied" && publication.PublishedSourceCommit != proof.ResultCommit {
		return ErrWorkflowConflict
	}

	return nil
}

func workflowNativeObservationCondition(binding WorkflowNativePublicationBinding, pr WorkflowPullRequestSnapshot) string {
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

func workflowNativeMetadataCondition(binding WorkflowNativePublicationBinding, pr WorkflowPullRequestSnapshot) string {
	if !pr.Merged && pr.State == "closed" {
		return "canonical_pr_closed_unmerged"
	}

	if pr.Merged && (pr.State != "closed" || pr.ResultCommit != binding.Git.ResultCommit) {
		return "canonical_pr_result_mismatch"
	}

	if !workflowNativeSourceCompatible(binding, pr) {
		return "canonical_source_changed"
	}

	if !pr.Merged && pr.State != "open" {
		return "canonical_pr_state_unknown"
	}

	return ""
}

func workflowNativeSourceCompatible(binding WorkflowNativePublicationBinding, pr WorkflowPullRequestSnapshot) bool {
	if pr.HeadCommit == binding.Git.SourceCommit || pr.HeadCommit == binding.Git.ResultCommit {
		return true
	}

	return workflowCommitID(pr.HeadCommit) && binding.Publication.State == "applied" && pr.Merged
}

func (adapter RemoteReviewAdapter) verifyWorkflowNativeObservation(ctx context.Context, binding WorkflowNativePublicationBinding, git WorkflowIntegrationGit, observed WorkflowNativePublicationObservation) (WorkflowNativePublicationObservation, error) {
	proof, err := verifyWorkflowNativeTarget(ctx, binding, git)
	if err != nil {
		return observed, err
	}

	observed.Git = &proof

	source, err := adapter.ReadWorkflowBranch(ctx, binding.PullRequest.Repository, binding.PullRequest.HeadBranch)
	if err != nil {
		return observed, err
	}

	observed.ObservedSourceCommit = source
	if binding.Publication.State == "unknown" && source != binding.Git.ResultCommit {
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

func verifyWorkflowNativeTarget(ctx context.Context, binding WorkflowNativePublicationBinding, git WorkflowIntegrationGit) (WorkflowIntegrationProof, error) {
	query := binding.Git.WorkflowIntegrationQuery
	query.SourceDirectory = binding.SourceDirectory

	proof, err := observeWorkflowPreparedGit(ctx, git, query, binding.Git)
	if err != nil {
		return proof, err
	}

	if err := validateWorkflowIntegrationProof(query, proof); err != nil {
		return proof, err
	}

	expected := [4]string{binding.Git.BaseCommit, binding.Git.VerifiedTree, binding.Git.PreparedTree, binding.Git.HistoryHash}

	actual := [4]string{proof.BaseCommit, proof.VerifiedTree, proof.PreparedTree, proof.HistoryHash}
	if expected != actual {
		return proof, ErrWorkflowConflict
	}

	return proof, nil
}

func finishWorkflowNativeObservation(binding WorkflowNativePublicationBinding, observed WorkflowNativePublicationObservation) WorkflowNativePublicationObservation {
	pr := observed.PullRequest
	if condition := workflowNativeObservationCondition(binding, pr); condition != "" {
		observed.State, observed.Condition = "reconciliation_required", condition

		return observed
	}

	if pr.Merged {
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

func observeWorkflowPreparedGit(ctx context.Context, git WorkflowIntegrationGit, query WorkflowIntegrationQuery, expected WorkflowIntegrationProof) (WorkflowIntegrationProof, error) {
	if prepared, ok := git.(WorkflowPreparedIntegrationGit); ok {
		return prepared.VerifyPreparedWorkflowIntegration(ctx, query, expected)
	}

	if expected.Provenance != nil {
		return WorkflowIntegrationProof{}, ErrWorkflowConflict
	}

	return git.VerifyWorkflowIntegration(ctx, query)
}
