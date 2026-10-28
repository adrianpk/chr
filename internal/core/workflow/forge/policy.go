package forge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowSubmissionResponseState preserves the retained submission identity and observation contract.
func WorkflowSubmissionResponseState(response workflowstate.SubmissionResponse, err error) string {
	if err != nil {
		return "error_unknown"
	}

	if response.Accepted {
		return "accepted"
	}

	return "unaccepted"
}

// ValidWorkflowSubmissionState preserves the retained submission identity and observation contract.
func ValidWorkflowSubmissionState(state string) bool {
	return slices.Contains([]string{"awaiting_observation", "unpublished", "confirmed"}, state)
}

// ValidateWorkflowUnpublishedProof preserves the retained submission identity and observation contract.
func ValidateWorkflowUnpublishedProof(submission workflowstate.Submission) error {
	if submission.State == "unpublished" {
		if submission.Unpublished == nil || submission.Unpublished.SubmissionID != submission.ID || !submission.Unpublished.Unpublished {
			return workflowstate.ErrConflict
		}
	} else if submission.Unpublished != nil {
		return workflowstate.ErrConflict
	}

	return nil
}
