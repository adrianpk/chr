package core

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

func (s *WorkflowStore) workflowNativeReceiptApplication(submission WorkflowSubmission) (*WorkflowEvidenceRef, error) {
	native := submission.Native
	for _, ref := range native.GitObservationRefs {
		var observation WorkflowNativeGitObservation
		if err := s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation); err != nil {
			return nil, err
		}

		if observation.Git != nil && observation.SourceCommit == native.PreparedResult {
			return &ref, nil
		}
	}

	return nil, nil
}

func (s *WorkflowStore) checkWorkflowNativeReceiptApplication(receipt WorkflowReceipt, transaction WorkflowGitPublicationResult) error {
	native := receipt.Native
	if native.GitApplicationRef == nil {
		if !validWorkflowNativeReceiptTransaction(*native, transaction) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	var observation WorkflowNativeGitObservation
	if err := s.readWorkflowNativeEvidence(native.InvocationID, *native.GitApplicationRef, &observation); err != nil {
		return err
	}

	if observation.SubmissionID != receipt.SubmissionID {
		return workflowstate.ErrConflict
	}

	var candidate WorkflowIntegrationCandidate
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CandidateRef, &candidate); err != nil {
		return err
	}

	binding := WorkflowNativePublicationBinding{Git: candidate.Input.Git, Publication: transaction, GitApplication: &observation}
	if !workflowNativeApplicationAttributed(binding) {
		return workflowstate.ErrConflict
	}

	intent := WorkflowNativeSubmission{WorkflowNativeRecordBinding: native.WorkflowNativeRecordBinding}
	if !validWorkflowNativeResponse(intent, transaction) || transaction.State != "applied" && transaction.State != "unknown" {
		return workflowstate.ErrConflict
	}

	return nil
}
