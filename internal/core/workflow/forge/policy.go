package forge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ResponseState classifies a provider reply without turning acceptance into confirmed delivery.
func ResponseState(response workflowstate.SubmissionResponse, err error) string {
	if err != nil {
		return "error_unknown"
	}

	if response.Accepted {
		return "accepted"
	}

	return "unaccepted"
}

// ValidSubmissionState recognizes the original retained publication state spellings.
func ValidSubmissionState(state string) bool {
	return slices.Contains([]string{"awaiting_observation", "unpublished", "confirmed"}, state)
}

// CheckUnpublishedProof requires a sealed, identity-bound unpublished observation.
func CheckUnpublishedProof(submission workflowstate.Submission) error {
	if submission.State == "unpublished" {
		if submission.Unpublished == nil || submission.Unpublished.SubmissionID != submission.ID || !submission.Unpublished.Unpublished {
			return workflowstate.ErrConflict
		}
	} else if submission.Unpublished != nil {
		return workflowstate.ErrConflict
	}

	return nil
}
