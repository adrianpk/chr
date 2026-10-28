package forge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ResponseState classifies a provider reply without turning acceptance into confirmed delivery.
func ResponseState(response workflowstate.SubmissionResponse, err error) workflowstate.SubmissionResponseState {
	if err != nil {
		return workflowstate.SubmissionResponseUnknown
	}

	if response.Accepted {
		return workflowstate.SubmissionResponseAccepted
	}

	return workflowstate.SubmissionResponseUnaccepted
}

// ValidSubmissionState recognizes the original retained publication state spellings.
func ValidSubmissionState(state workflowstate.SubmissionState) bool {
	return slices.Contains([]workflowstate.SubmissionState{workflowstate.SubmissionAwaitingObservation, workflowstate.SubmissionUnpublished, workflowstate.SubmissionConfirmed}, state)
}

// CheckUnpublishedProof requires a sealed, identity-bound unpublished observation.
func CheckUnpublishedProof(submission workflowstate.Submission) error {
	if submission.State == workflowstate.SubmissionUnpublished {
		if submission.Unpublished == nil || submission.Unpublished.SubmissionID != submission.ID || !submission.Unpublished.Unpublished {
			return workflowstate.ErrConflict
		}
	} else if submission.Unpublished != nil {
		return workflowstate.ErrConflict
	}

	return nil
}
