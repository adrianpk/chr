package core

import (
	"context"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) checkWorkflowNativeSubmissionEvidence(ctx context.Context, submission WorkflowSubmission) error {
	native := submission.Native
	if native == nil {
		return workflowstate.ErrConflict
	}

	var candidate WorkflowIntegrationCandidate
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

func workflowNativeSubmissionCandidateMatches(submission WorkflowSubmission, candidate WorkflowIntegrationCandidate) bool {
	native := submission.Native

	return candidate.Native != nil && candidate.Native.CandidateHash == native.CandidateHash && workflowNativeJSONEqual(candidate.Native.WorkflowNativeRecordBinding, native.WorkflowNativeRecordBinding) && workflowNativeJSONEqual(candidate.Input.Git, submission.Guard.Git) && native.FastForwardProofRef == candidate.Native.HistoryManifestRef
}

func (s *WorkflowStore) checkWorkflowNativeSubmissionReview(submission WorkflowSubmission) error {
	var review workflowReviewEvidence
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, submission.Native.ReviewGrantRef, &review); err != nil {
		return err
	}

	if review.Decision.BindingHash != submission.Guard.ReviewHash || review.Decision.Result.Decision != IntegrationReviewDecisionPass {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowNativeTransportEvidence(ctx context.Context, submission WorkflowSubmission) error {
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

func (s *WorkflowStore) checkWorkflowNativeAdvertisement(submission WorkflowSubmission) error {
	native := submission.Native
	if native.AdvertisementRef == nil {
		return nil
	}

	var advertisement WorkflowGitPublicationAdvertisement
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.AdvertisementRef, &advertisement); err != nil {
		return err
	}

	var capability WorkflowNativeCapability
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, native.CapabilityRef, &capability); err != nil {
		return err
	}

	if advertisement.SourceCommit != native.OriginalSource || advertisement.TargetCommit != native.ValidatedBase || advertisement.Fingerprint != capability.ServerCapabilityFingerprint {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowNativeCommands(submission WorkflowSubmission) error {
	native := submission.Native
	if native.CommandManifestRef == nil {
		return nil
	}

	var manifest WorkflowNativeCommandManifest
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.CommandManifestRef, &manifest); err != nil {
		return err
	}

	expected := workflowNativeCommandManifest(submission)
	if manifest != expected {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowNativeResponse(submission WorkflowSubmission) error {
	native := submission.Native

	refs := slices.Clone(native.TransportResponseHistory)
	if native.TransportResponseRef != nil {
		refs = append(refs, *native.TransportResponseRef)
	}

	for _, ref := range refs {
		var response WorkflowGitPublicationResult
		if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &response); err != nil {
			return err
		}

		if !validWorkflowNativeResponse(*native, response) {
			return workflowstate.ErrConflict
		}

		if err := s.checkWorkflowNativeTransactionRevision(WorkflowNativeReceipt{WorkflowNativeRecordBinding: native.WorkflowNativeRecordBinding}, response); err != nil {
			return err
		}
	}

	return nil
}

func validWorkflowNativeResponse(native WorkflowNativeSubmission, response WorkflowGitPublicationResult) bool {
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

func (s *WorkflowStore) checkWorkflowNativeJournalObservations(submission WorkflowSubmission) error {
	native := submission.Native
	for _, ref := range native.ObservationRefs {
		var observation WorkflowNativePublicationObservation
		if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &observation); err != nil {
			return err
		}

		if !workflowNativeJournalObservationMatches(*native, observation) || !s.workflowNativeCanonicalObservationValid(submission, observation) {
			return workflowstate.ErrConflict
		}

		if observation.Git == nil {
			continue
		}

		if !workflowNativeRecordGitMatches(native.WorkflowNativeRecordBinding, *observation.Git) || observation.Git.HistoryHash != submission.Guard.Git.HistoryHash {
			return workflowstate.ErrConflict
		}

		if err := validateWorkflowIntegrationProof(observation.Git.WorkflowIntegrationQuery, *observation.Git); err != nil {
			return err
		}
	}

	return nil
}

func workflowNativeJournalObservationMatches(native WorkflowNativeSubmission, observation WorkflowNativePublicationObservation) bool {
	if [2]string{native.OriginalSource, native.PublishedSource} != [2]string{observation.OriginalSourceCommit, observation.PublishedSourceCommit} {
		return false
	}

	bound := observation.CanonicalPR == native.CanonicalPR
	if observation.CanonicalPR == "" {
		bound = observation.PullRequest.URL == native.CanonicalPR && observation.PullRequest.Repository == native.SourceRepository
	}

	return bound && slices.Contains([]string{"observing", "merged", "reconciliation_required"}, observation.State)
}

func (s *WorkflowStore) checkWorkflowNativeAppliedEvidence(submission WorkflowSubmission) error {
	native := submission.Native
	if native.TransportResponseRef == nil {
		return workflowstate.ErrConflict
	}

	var result WorkflowGitPublicationResult
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.TransportResponseRef, &result); err != nil {
		return err
	}

	receipt := WorkflowNativeReceipt{WorkflowNativeRecordBinding: native.WorkflowNativeRecordBinding}
	if !validWorkflowNativeReceiptTransaction(receipt, result) {
		if result.State != "unknown" || !result.CommandsSent || !s.workflowNativeGitApplied(submission) {
			return workflowstate.ErrConflict
		}
	}

	return s.checkWorkflowNativeTransactionRevision(receipt, result)
}

func (s *WorkflowStore) checkWorkflowNativeNonApplication(_ context.Context, submission WorkflowSubmission) error {
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

	var proof WorkflowNativeNonApplicationProof
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

func validWorkflowNativeNonApplication(submission WorkflowSubmission, proof WorkflowNativeNonApplicationProof) bool {
	native := submission.Native

	return proof.SchemaVersion == 1 && proof.SubmissionID == submission.ID && proof.CannotApply && proof.InputChanged && workflowCommitID(proof.ObservedSource) && workflowCommitID(proof.ObservedTarget) && (proof.ObservedSource != native.OriginalSource || proof.ObservedTarget != native.ValidatedBase)
}

func (s *WorkflowStore) applyWorkflowNativeSubmissionEvidence(ctx context.Context, submission *WorkflowSubmission) error {
	if err := s.checkWorkflowNativeSubmissionEvidence(ctx, *submission); err != nil {
		return err
	}

	native := submission.Native
	if native.ApplicationState == "not_applied" {
		var proof WorkflowNativeNonApplicationProof
		if err := s.readWorkflowNativeEvidence(submission.InvocationID, *native.NonApplicationRef, &proof); err != nil {
			return err
		}

		if proof.InputChanged {
			submission.State = "unpublished"
			submission.Unpublished = &WorkflowSubmissionObservation{SubmissionID: submission.ID, Unpublished: true}
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

func (s *WorkflowStore) confirmWorkflowNativeSubmission(_ context.Context, submission *WorkflowSubmission) error {
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

	var observation WorkflowNativePublicationObservation

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

func workflowNativeSubmissionReceiptMatches(submission WorkflowSubmission, record WorkflowRecord) bool {
	receipt, native := record.Receipt, submission.Native

	return receipt != nil && receipt.Native != nil && receipt.SubmissionID == submission.ID && receipt.Native.CandidateRef == native.CandidateRef && receipt.Native.TransactionProofRef == *native.TransportResponseRef && receipt.Native.CandidateHash == native.CandidateHash
}
