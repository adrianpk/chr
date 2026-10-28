package forge

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativeSourceResultContract retains a forge's original-source attribution while
// requiring independent exact-result publication and complete rebase history.
const NativeSourceResultContract = "github_same_pr_git_source_attribution/v2"

// NativeObservationContractMatches preserves the certified original-source attribution extension.
func NativeObservationContractMatches(expected, actual string) bool {
	return expected == actual || expected == "github_same_pr_git_merge/v1" && actual == NativeSourceResultContract
}

// NativeSourceResultCompatible requires certified source attribution and independently checked result history.
func NativeSourceResultCompatible(contract string, proof workflowstate.IntegrationProof, pr workflowstate.PullRequestSnapshot, proofs NativeProofs) bool {
	if contract != NativeSourceResultContract || proof.Method != "rebase" || pr.ResultCommit != proof.SourceCommit || proof.SourceCommit == proof.ResultCommit {
		return false
	}

	if !nativeSourceResultRepository(pr, proof.Repository) || len(proof.History) == 0 {
		return false
	}

	// The existing method validator binds the complete original replay and any
	// separately attributed integration repairs to the exact prepared result.
	return proofs.CheckNativeProof(proof.IntegrationQuery, proof) == nil
}

func nativeSourceResultRepository(pr workflowstate.PullRequestSnapshot, repository string) bool {
	reference, err := ParseReference(pr.URL)

	return err == nil && reference.Provider == GitHub && pr.Repository == repository && pr.HeadRepository == repository
}

// NativeSourceResultObserved binds compatible original-source metadata to the observed published source.
func NativeSourceResultObserved(observation workflowstate.NativePublicationObservation, proofs NativeProofs) bool {
	proof := observation.Git
	if proof == nil || !NativeSourceResultCompatible(observation.ObservationContract, *proof, observation.PullRequest, proofs) {
		return false
	}

	source := observation.ObservedSourceCommit
	if source == proof.ResultCommit {
		return observation.PullRequest.HeadCommit == proof.ResultCommit
	}

	return workflowstate.ValidCommitID(source) && source != proof.SourceCommit && source == observation.PullRequest.HeadCommit
}

// NativeReceiptResultMatches accepts only exact result equality or the certified rebase attribution form.
func NativeReceiptResultMatches(resultCommit string, native workflowstate.NativeReceipt) bool {
	if native.ActualMergeResult == resultCommit {
		return true
	}

	reference, err := ParseReference(native.CanonicalPR)

	return err == nil && reference.Provider == GitHub && native.ObservationContract == NativeSourceResultContract && native.Method == "rebase" && native.ActualMergeResult == native.OriginalSource
}
