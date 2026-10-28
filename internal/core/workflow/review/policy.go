package review

import (
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidateWorkflowReviewPolicy preserves the retained review input and decision contract.
func ValidateWorkflowReviewPolicy(policy workflowstate.ReviewPolicy) error {
	if !workflowstate.IntegrationCycleMode(policy.Mode).Valid() || !slices.Contains([]string{"merge", "rebase", "squash"}, policy.Method) || !slices.Contains([]string{"legacy_default", "version_default", "project", "invocation_restriction"}, policy.Source) || len(policy.Revision) != 64 {
		return fmt.Errorf("invalid current workflow review policy")
	}

	return nil
}

// WorkflowReviewDecisionState preserves the retained review input and decision contract.
func WorkflowReviewDecisionState(decision workflowstate.IntegrationReviewDecision) string {
	switch decision {
	case workflowstate.IntegrationReviewDecisionPass:
		return "authorized"
	case workflowstate.IntegrationReviewDecisionNeedsChanges:
		return "needs_repair"
	case workflowstate.IntegrationReviewDecisionHumanRequired:
		return "human_required"
	default:
		return "insufficient_evidence"
	}
}
