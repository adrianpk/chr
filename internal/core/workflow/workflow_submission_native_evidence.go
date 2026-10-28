package workflow

import (
	"context"
	"slices"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) checkWorkflowNativeSubmissionEvidence(ctx context.Context, submission workflowstate.Submission) error {
	native := submission.Native
	if native == nil {
		return workflowstate.ErrConflict
	}

	var candidate workflowstate.IntegrationCandidate
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, native.CandidateRef, &candidate); err != nil {
		return err
	}

	if !workflowNativeSubmissionCandidateMatches(submission, candidate) {
		return workflowstate.ErrConflict
	}

	if err := s.checkWorkflowNativeCandidate(candidate); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeSubmissionReview(submission); err != nil {
		return err
	}

	return s.checkWorkflowNativeTransportEvidence(ctx, submission)
}

func workflowNativeSubmissionCandidateMatches(submission workflowstate.Submission, candidate workflowstate.IntegrationCandidate) bool {
	native := submission.Native

	return candidate.Native != nil && candidate.Native.CandidateHash == native.CandidateHash && workflowNativeJSONEqual(candidate.Native.NativeRecordBinding, native.NativeRecordBinding) && workflowNativeJSONEqual(candidate.Input.Git, submission.Guard.Git) && native.FastForwardProofRef == candidate.Native.HistoryManifestRef
}

func (s *Store) checkWorkflowNativeSubmissionReview(submission workflowstate.Submission) error {
	var review workflowReviewEvidence
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, submission.Native.ReviewGrantRef, &review); err != nil {
		return err
	}

	if review.Decision.BindingHash != submission.Guard.ReviewHash || review.Decision.Result.Decision != workflowstate.IntegrationReviewDecisionPass {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowNativeTransportEvidence(ctx context.Context, submission workflowstate.Submission) error {
	if err := s.checkWorkflowNativeAdvertisement(submission); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeCommands(submission); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeResponse(submission); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeDispatchCompletion(submission); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeGitObservations(submission); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeJournalObservations(submission); err != nil {
		return err
	}

	switch submission.Native.ApplicationState {
	case "git_applied":
		return s.checkWorkflowNativeAppliedEvidence(submission)
	case "not_applied":
		return s.checkWorkflowNativeNonApplication(ctx, submission)
	default:
		return nil
	}
}

func (s *Store) checkWorkflowNativeAdvertisement(submission workflowstate.Submission) error {
	native := submission.Native
	if native.AdvertisementRef == nil {
		return nil
	}

	var advertisement GitPublicationAdvertisement
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.AdvertisementRef, &advertisement); err != nil {
		return err
	}

	var capability workflowstate.NativeCapability
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, native.CapabilityRef, &capability); err != nil {
		return err
	}

	if advertisement.SourceCommit != native.OriginalSource || advertisement.TargetCommit != native.ValidatedBase || advertisement.Fingerprint != capability.ServerCapabilityFingerprint {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowNativeCommands(submission workflowstate.Submission) error {
	native := submission.Native
	if native.CommandManifestRef == nil {
		return nil
	}

	var manifest NativeCommandManifest
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.CommandManifestRef, &manifest); err != nil {
		return err
	}

	expected := workflowNativeCommandManifest(submission)
	if manifest != expected {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowNativeResponse(submission workflowstate.Submission) error {
	native := submission.Native

	refs := slices.Clone(native.TransportResponseHistory)
	if native.TransportResponseRef != nil {
		refs = append(refs, *native.TransportResponseRef)
	}

	for _, ref := range refs {
		var response GitPublicationResult
		if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &response); err != nil {
			return err
		}

		if !validWorkflowNativeResponse(*native, response) {
			return workflowstate.ErrConflict
		}

		if err := s.checkWorkflowNativeTransactionRevision(NativeReceipt{NativeRecordBinding: native.NativeRecordBinding}, response); err != nil {
			return err
		}
	}

	return nil
}

func validWorkflowNativeResponse(native workflowstate.NativeSubmission, response GitPublicationResult) bool {
	if [3]string{response.Schema, response.OriginalSourceCommit, response.ResultCommit} != [3]string{"chr-git-publication-result/v1", native.OriginalSource, native.PreparedResult} {
		return false
	}

	if response.State == "not_sent" {
		return !response.CommandsSent && response.PublishedSourceCommit == ""
	}

	if !response.CommandsSent || response.AdvertisedSourceCommit != native.OriginalSource || response.AdvertisedTargetCommit != native.ValidatedBase {
		return false
	}

	if response.State == "applied" {
		return response.PublishedSourceCommit == native.PreparedResult
	}

	return slices.Contains([]string{"unknown", "rejected"}, response.State) && response.PublishedSourceCommit == ""
}

func (s *Store) checkWorkflowNativeJournalObservations(submission workflowstate.Submission) error {
	native := submission.Native
	for _, ref := range native.ObservationRefs {
		var observation NativePublicationObservation
		if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &observation); err != nil {
			return err
		}

		if !workflowNativeJournalObservationMatches(*native, observation) || !s.workflowNativeCanonicalObservationValid(submission, observation) {
			return workflowstate.ErrConflict
		}

		if observation.Git == nil {
			continue
		}

		if !workflowNativeRecordGitMatches(native.NativeRecordBinding, *observation.Git) || observation.Git.HistoryHash != submission.Guard.Git.HistoryHash {
			return workflowstate.ErrConflict
		}

		if err := workflowintegrate.CheckProof(observation.Git.IntegrationQuery, *observation.Git, workflowIntegrationRepairs{}); err != nil {
			return err
		}
	}

	return nil
}

func workflowNativeJournalObservationMatches(native workflowstate.NativeSubmission, observation NativePublicationObservation) bool {
	if [2]string{native.OriginalSource, native.PublishedSource} != [2]string{observation.OriginalSourceCommit, observation.PublishedSourceCommit} {
		return false
	}

	bound := observation.CanonicalPR == native.CanonicalPR
	if observation.CanonicalPR == "" {
		bound = observation.PullRequest.URL == native.CanonicalPR && observation.PullRequest.Repository == native.SourceRepository
	}

	return bound && slices.Contains([]string{"observing", "merged", "reconciliation_required"}, observation.State)
}

func (s *Store) checkWorkflowNativeAppliedEvidence(submission workflowstate.Submission) error {
	native := submission.Native
	if native.TransportResponseRef == nil {
		return workflowstate.ErrConflict
	}

	var result GitPublicationResult
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.TransportResponseRef, &result); err != nil {
		return err
	}

	receipt := NativeReceipt{NativeRecordBinding: native.NativeRecordBinding}
	if !validWorkflowNativeReceiptTransaction(receipt, result) {
		if result.State != "unknown" || !result.CommandsSent || !s.workflowNativeGitApplied(submission) {
			return workflowstate.ErrConflict
		}
	}

	return s.checkWorkflowNativeTransactionRevision(receipt, result)
}

func (s *Store) checkWorkflowNativeNonApplication(_ context.Context, submission workflowstate.Submission) error {
	native := submission.Native
	if native.NonApplicationRef == nil || native.DispatchState != "quiescent" || native.DispatchIdentity == nil {
		return workflowstate.ErrConflict
	}

	alive, err := workflowProcessAlive(*native.DispatchIdentity)
	if err != nil {
		return err
	}

	if alive && native.DispatchCompletionRef == nil {
		return ErrWorkflowObservationRequired
	}

	var proof NativeNonApplicationProof
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.NonApplicationRef, &proof); err != nil {
		return err
	}

	if native.DispatchCompletionRef != nil {
		return s.checkWorkflowNativeDefiniteNonApplication(submission, proof)
	}

	if !validWorkflowNativeNonApplication(submission, proof) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowNativeNonApplication(submission workflowstate.Submission, proof NativeNonApplicationProof) bool {
	native := submission.Native

	return proof.SchemaVersion == 1 && proof.SubmissionID == submission.ID && proof.CannotApply && proof.InputChanged && workflowstate.ValidCommitID(proof.ObservedSource) && workflowstate.ValidCommitID(proof.ObservedTarget) && (proof.ObservedSource != native.OriginalSource || proof.ObservedTarget != native.ValidatedBase)
}

func (s *Store) applyWorkflowNativeSubmissionEvidence(ctx context.Context, submission *workflowstate.Submission) error {
	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, *submission); err != nil {
		return err
	}

	native := submission.Native
	if native.ApplicationState == "not_applied" {
		var proof NativeNonApplicationProof
		if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.NonApplicationRef, &proof); err != nil {
			return err
		}

		if proof.InputChanged {
			submission.State = "unpublished"
			submission.Unpublished = &workflowstate.SubmissionObservation{SubmissionID: submission.ID, Unpublished: true}
		}
	}

	if native.ReceiptRef != nil {
		if native.ApplicationState != "git_applied" || native.TransportResponseRef == nil {
			return workflowstate.ErrConflict
		}

		return s.confirmWorkflowNativeSubmission(ctx, submission)
	}

	return validateWorkflowNativeSubmission(*submission, "chr-integration-journal/v2")
}

func (s *Store) confirmWorkflowNativeSubmission(_ context.Context, submission *workflowstate.Submission) error {
	native := submission.Native

	record, err := s.readRecord(native.ReceiptRef.Path)
	if err != nil {
		return err
	}

	if _, err := s.readWorkflowReviewImage(native.ReceiptRef.Path, native.ReceiptRef.Hash); err != nil {
		return err
	}

	if !workflowNativeSubmissionReceiptMatches(*submission, record) {
		return workflowstate.ErrConflict
	}

	if err := s.checkWorkflowNativeReceipt(record); err != nil {
		return err
	}

	var observation NativePublicationObservation

	ref := record.Receipt.Native.ObservationRefs[len(record.Receipt.Native.ObservationRefs)-1]
	if !slices.Contains(native.ObservationRefs, ref) {
		return workflowstate.ErrConflict
	}

	if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &observation); err != nil {
		return err
	}

	submission.State, submission.ObservedPR, submission.Proof = "confirmed", &observation.PullRequest, observation.Git

	return validateWorkflowSubmissionProof(*submission)
}

func workflowNativeSubmissionReceiptMatches(submission workflowstate.Submission, record Record) bool {
	receipt, native := record.Receipt, submission.Native

	return receipt != nil && receipt.Native != nil && receipt.SubmissionID == submission.ID && receipt.Native.CandidateRef == native.CandidateRef && receipt.Native.TransactionProofRef == *native.TransportResponseRef && receipt.Native.CandidateHash == native.CandidateHash
}
