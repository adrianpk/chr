package workflow

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

func (s *Store) workflowNativeReceiptApplication(submission Submission) (*workflowstate.EvidenceRef, error) {
	native := submission.Native
	for _, ref := range native.GitObservationRefs {
		var observation NativeGitObservation
		if err := s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation); err != nil {
			return nil, err
		}

		if observation.Git != nil && observation.SourceCommit == native.PreparedResult {
			return &ref, nil
		}
	}

	return nil, nil
}

func (s *Store) checkWorkflowNativeReceiptApplication(receipt Receipt, transaction GitPublicationResult) error {
	native := receipt.Native
	if native.GitApplicationRef == nil {
		if !validWorkflowNativeReceiptTransaction(*native, transaction) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	var observation NativeGitObservation
	if err := s.readWorkflowNativeEvidence(native.InvocationID, *native.GitApplicationRef, &observation); err != nil {
		return err
	}

	if observation.SubmissionID != receipt.SubmissionID {
		return workflowstate.ErrConflict
	}

	var candidate IntegrationCandidate
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CandidateRef, &candidate); err != nil {
		return err
	}

	binding := NativePublicationBinding{Git: candidate.Input.Git, Publication: transaction, GitApplication: &observation}
	if !workflowNativeApplicationAttributed(binding) {
		return workflowstate.ErrConflict
	}

	intent := NativeSubmission{NativeRecordBinding: native.NativeRecordBinding}
	if !validWorkflowNativeResponse(intent, transaction) || transaction.State != "applied" && transaction.State != "unknown" {
		return workflowstate.ErrConflict
	}

	return nil
}
