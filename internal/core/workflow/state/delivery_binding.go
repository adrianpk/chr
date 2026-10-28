package state

import (
	"slices"
)

// ReceiptMatchesProof checks the retained native or legacy delivery binding.
func ReceiptMatchesProof(record Record, proof SliceDeliveryProof, ref string) bool {
	if record.Receipt == nil || record.Revision != 1 {
		return false
	}

	receipt := record.Receipt
	if proof.NativeReceiptRef != nil {
		return nativeReceiptMatchesProof(record, proof)
	}

	return legacyReceiptMatchesProof(*receipt, proof, ref)
}

func legacyReceiptMatchesProof(receipt Receipt, proof SliceDeliveryProof, ref string) bool {
	if receipt.Unit != proof.Unit || receipt.CandidateID != proof.CandidateID || receipt.SubmissionID != proof.SubmissionID || receipt.VerificationRef != ref || !receipt.ConfirmedAt.Equal(proof.ConfirmedAt) {
		return false
	}

	return ReceiptGitMatches(receipt, proof.Git) && receipt.PullRequest == proof.PullRequest.URL && slices.Contains(receipt.ValidationRefs, proof.CandidateRef)
}

func nativeReceiptMatchesProof(record Record, proof SliceDeliveryProof) bool {
	receipt := record.Receipt
	if receipt.Native == nil || *proof.NativeReceiptRef != (EvidenceRef{Path: record.Path, Hash: record.SourceHash}) || receipt.Unit != proof.Unit || receipt.CandidateID != proof.CandidateID || receipt.SubmissionID != proof.SubmissionID {
		return false
	}

	return receipt.ConfirmedAt.Equal(proof.ConfirmedAt) && ReceiptGitMatches(*receipt, proof.Git) && receipt.PullRequest == proof.PullRequest.URL
}

// ReceiptGitMatches requires the exact retained Git identity, trees and history.
func ReceiptGitMatches(receipt Receipt, git IntegrationProof) bool {
	if receipt.Repository != git.Repository || receipt.Method != git.Method || receipt.SourceCommit != git.SourceCommit || receipt.BaseCommit != git.BaseCommit {
		return false
	}

	return receipt.PreparedTree == git.PreparedTree && receipt.VerifiedTree == git.VerifiedTree && receipt.HistoryProof == git.HistoryHash && receipt.ResultCommit == git.ResultCommit && receipt.ObservedTarget == git.ObservedTarget
}

// FinalizedTaskMatches requires an integrated task's exact delivery projection.
func FinalizedTaskMatches(record Record, branch string, receipt Record) bool {
	return record.Task != nil && record.Task.Phase == "integrated" && record.Task.Status == "done" && record.Task.DeliveryRef == receipt.Path && taskDeliveryMatches(record.Task.Delivery, branch, *receipt.Receipt)
}

func taskDeliveryMatches(delivery *TaskDelivery, branch string, receipt Receipt) bool {
	return delivery != nil && delivery.PullRequest == receipt.PullRequest && delivery.Branch == branch && delivery.Method == receipt.Method && delivery.ResultCommit == receipt.ResultCommit && delivery.ConfirmedAt.Equal(receipt.ConfirmedAt)
}

// ReceiptTargetProofMatches binds observed target, history and verified tree.
func ReceiptTargetProofMatches(proof IntegrationProof, receipt Record, target PreviewTarget) bool {
	return proof.ObservedTarget == target.Commit && proof.HistoryHash == receipt.Receipt.HistoryProof && proof.VerifiedTree == receipt.Receipt.VerifiedTree
}
