package workflow

import (
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowNativeReceiptApplication(submission workflowstate.Submission) (*workflowstate.EvidenceRef, error) {
	native := submission.Native
	for _, ref := range native.GitObservationRefs {
		var observation workflowstate.NativeGitObservation
		if err := s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation); err != nil {
			return nil, err
		}

		if observation.Git != nil && observation.SourceCommit == native.PreparedResult {
			return &ref, nil
		}
	}

	return nil, nil
}

func (s *Store) checkWorkflowNativeReceiptApplication(receipt workflowstate.Receipt, transaction workflowstate.GitPublicationResult) error {
	native := receipt.Native
	if native.GitApplicationRef == nil {
		if !validWorkflowNativeReceiptTransaction(*native, transaction) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	var observation workflowstate.NativeGitObservation
	if err := s.readWorkflowNativeEvidence(native.InvocationID, *native.GitApplicationRef, &observation); err != nil {
		return err
	}

	if observation.SubmissionID != receipt.SubmissionID {
		return workflowstate.ErrConflict
	}

	var candidate workflowstate.IntegrationCandidate
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CandidateRef, &candidate); err != nil {
		return err
	}

	binding := workflowstate.NativePublicationBinding{Git: candidate.Input.Git, Publication: transaction, GitApplication: &observation}
	if !workflowforge.NativeApplicationAttributed(binding, workflowNativePublicationProofs{}) {
		return workflowstate.ErrConflict
	}

	intent := workflowstate.NativeSubmission{NativeRecordBinding: native.NativeRecordBinding}
	if !validWorkflowNativeResponse(intent, transaction) || transaction.State != workflowstate.GitPublicationApplied && transaction.State != workflowstate.GitPublicationUnknown {
		return workflowstate.ErrConflict
	}

	return nil
}
