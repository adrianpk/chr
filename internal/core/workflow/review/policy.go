package review

import (
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckPolicy checks the current retained mode, method, source and revision shape.
func CheckPolicy(policy workflowstate.ReviewPolicy) error {
	if !workflowstate.IntegrationCycleMode(policy.Mode).Valid() || !slices.Contains([]string{"merge", "rebase", "squash"}, policy.Method) || !slices.Contains([]string{"legacy_default", "version_default", "project", "invocation_restriction"}, policy.Source) || len(policy.Revision) != 64 {
		return fmt.Errorf("invalid current workflow review policy")
	}

	return nil
}

// DecisionState maps a role decision to its original retained review state.
func DecisionState(decision workflowstate.IntegrationReviewDecision) workflowstate.ReviewState {
	switch decision {
	case workflowstate.IntegrationReviewDecisionPass:
		return workflowstate.ReviewAuthorized
	case workflowstate.IntegrationReviewDecisionNeedsChanges:
		return workflowstate.ReviewNeedsRepair
	case workflowstate.IntegrationReviewDecisionHumanRequired:
		return workflowstate.ReviewHumanRequired
	default:
		return workflowstate.ReviewInsufficientEvidence
	}
}
