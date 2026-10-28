package workflow

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) checkWorkflowNativeDispatchCompletion(submission Submission) error {
	native := submission.Native
	if native.DispatchCompletionRef == nil {
		return nil
	}

	var completion NativeDispatchCompletion
	if err := s.readWorkflowNativeEvidence(native.InvocationID, *native.DispatchCompletionRef, &completion); err != nil {
		return err
	}

	if !workflowNativeCompletionMatches(submission, completion) {
		return workflowstate.ErrConflict
	}

	refs := slices.Clone(native.TransportResponseHistory)
	if native.TransportResponseRef != nil {
		refs = append(refs, *native.TransportResponseRef)
	}

	if !slices.Contains(refs, completion.ResponseRef) {
		return workflowstate.ErrConflict
	}

	var result GitPublicationResult
	if err := s.readWorkflowNativeEvidence(native.InvocationID, completion.ResponseRef, &result); err != nil {
		return err
	}

	if completion.CandidateRevision != result.CandidateRevision {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowNativeCompletionMatches(submission Submission, completion NativeDispatchCompletion) bool {
	native := submission.Native
	if native.DispatchIdentity == nil {
		return false
	}

	return completion.SchemaVersion == 1 && completion.SubmissionID == submission.ID && completion.Identity == *native.DispatchIdentity && native.DispatchState == "quiescent" && !completion.FinishedAt.Before(native.IntentTime)
}

func (s *Store) checkWorkflowNativeGitObservations(submission Submission) error {
	for _, ref := range submission.Native.GitObservationRefs {
		var observation NativeGitObservation
		if err := s.readWorkflowNativeEvidence(submission.InvocationID, ref, &observation); err != nil {
			return err
		}

		if err := validateWorkflowNativeGitObservation(submission, observation); err != nil {
			return err
		}

		if err := s.checkWorkflowNativeTransactionRevision(NativeReceipt{NativeRecordBinding: submission.Native.NativeRecordBinding}, GitPublicationResult{CandidateRevision: observation.CandidateRevision}); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowNativeGitObservation(submission Submission, observation NativeGitObservation) error {
	if !validWorkflowNativeGitObservationHeader(submission, observation) {
		return workflowstate.ErrConflict
	}

	proof := observation.Git
	if (proof != nil) != (observation.Condition == "result_verified") || !slices.Contains([]string{"result_verified", "result_unverified"}, observation.Condition) {
		return workflowstate.ErrConflict
	}

	if proof == nil {
		return nil
	}

	if !workflowNativeRecordGitMatches(submission.Native.NativeRecordBinding, *proof) || proof.HistoryHash != submission.Guard.Git.HistoryHash || proof.ObservedTarget != observation.TargetCommit {
		return workflowstate.ErrConflict
	}

	return validateWorkflowIntegrationProof(proof.IntegrationQuery, *proof)
}

func validWorkflowNativeGitObservationHeader(submission Submission, observation NativeGitObservation) bool {
	return observation.SchemaVersion == 1 && observation.SubmissionID == submission.ID && workflowCommitID(observation.SourceCommit) && workflowCommitID(observation.TargetCommit) && validWorkflowNativeFingerprint(observation.CandidateRevision)
}

func (s *Store) workflowNativeGitApplied(submission Submission) bool {
	native := submission.Native
	for _, ref := range native.GitObservationRefs {
		var observation NativeGitObservation
		if s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation) == nil && observation.Git != nil && observation.SourceCommit == native.PreparedResult {
			return true
		}
	}

	return false
}

func (s *Store) checkWorkflowNativeDefiniteNonApplication(submission Submission, proof NativeNonApplicationProof) error {
	native := submission.Native
	if !validWorkflowNativeNonApplicationHeader(submission, proof) {
		return workflowstate.ErrConflict
	}

	var result GitPublicationResult
	if err := s.readWorkflowNativeEvidence(native.InvocationID, *native.TransportResponseRef, &result); err != nil {
		return err
	}

	if !workflowNativeDefiniteResponse(*native, result) {
		return ErrWorkflowObservationRequired
	}

	var observation NativeGitObservation

	ref := native.GitObservationRefs[len(native.GitObservationRefs)-1]
	if err := s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation); err != nil {
		return err
	}

	if !workflowNativeNonApplicationObservation(*native, proof, observation) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowNativeNonApplicationHeader(submission Submission, proof NativeNonApplicationProof) bool {
	native := submission.Native

	return proof.SchemaVersion == 1 && proof.SubmissionID == submission.ID && proof.CannotApply && native.TransportResponseRef != nil && len(native.GitObservationRefs) > 0
}

func workflowNativeDefiniteResponse(native NativeSubmission, result GitPublicationResult) bool {
	if result.State == "not_sent" {
		return !result.CommandsSent
	}

	return result.State == "rejected" && result.CommandsSent && native.CommandManifestRef != nil && native.AdvertisementRef != nil
}

func workflowNativeNonApplicationObservation(native NativeSubmission, proof NativeNonApplicationProof, observation NativeGitObservation) bool {
	changed := observation.SourceCommit != native.OriginalSource || observation.TargetCommit != native.ValidatedBase
	refsMatch := [2]string{proof.ObservedSource, proof.ObservedTarget} == [2]string{observation.SourceCommit, observation.TargetCommit}

	return refsMatch && proof.InputChanged == changed && observation.SourceCommit != native.PreparedResult && observation.TargetCommit != native.PreparedResult && observation.Git == nil
}
