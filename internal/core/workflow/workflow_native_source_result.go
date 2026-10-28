package workflow

import (
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// The additional contract retains a forge's original-source attribution while
// requiring independent exact-result publication and complete rebase history.
const workflowNativeSourceResultContract = "github_same_pr_git_source_attribution/v2"

func workflowNativeObservationContractMatches(expected, actual string) bool {
	return expected == actual || expected == "github_same_pr_git_merge/v1" && actual == workflowNativeSourceResultContract
}

func workflowNativeSourceResultCompatible(contract string, proof workflowstate.IntegrationProof, pr workflowstate.PullRequestSnapshot) bool {
	if contract != workflowNativeSourceResultContract || proof.Method != "rebase" || pr.ResultCommit != proof.SourceCommit || proof.SourceCommit == proof.ResultCommit {
		return false
	}

	if !workflowNativeSourceResultRepository(pr, proof.Repository) || len(proof.History) == 0 {
		return false
	}

	// The existing method validator binds the complete original replay and any
	// separately attributed integration repairs to the exact prepared result.
	return workflowintegrate.ValidateWorkflowIntegrationProof(proof.IntegrationQuery, proof, workflowIntegrationRepairs{}) == nil
}

func workflowNativeSourceResultRepository(pr workflowstate.PullRequestSnapshot, repository string) bool {
	reference, err := DetectReviewProviderReference(pr.URL)

	return err == nil && reference.Provider == ReviewProviderGitHub && pr.Repository == repository && pr.HeadRepository == repository
}

func workflowNativeSourceResultObserved(observation NativePublicationObservation) bool {
	proof := observation.Git
	if proof == nil || !workflowNativeSourceResultCompatible(observation.ObservationContract, *proof, observation.PullRequest) {
		return false
	}

	source := observation.ObservedSourceCommit
	if source == proof.ResultCommit {
		return observation.PullRequest.HeadCommit == proof.ResultCommit
	}

	return workflowstate.ValidCommitID(source) && source != proof.SourceCommit && source == observation.PullRequest.HeadCommit
}

func workflowNativeReceiptResultMatches(receipt Receipt) bool {
	native := receipt.Native
	if native.ActualMergeResult == receipt.ResultCommit {
		return true
	}

	reference, err := DetectReviewProviderReference(native.CanonicalPR)

	return err == nil && reference.Provider == ReviewProviderGitHub && native.ObservationContract == workflowNativeSourceResultContract && native.Method == "rebase" && native.ActualMergeResult == native.OriginalSource
}
