package core

import (
	"fmt"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidateReviewWorkflowConfig validates the corresponding package boundary value.
func ValidateReviewWorkflowConfig(config ReviewWorkflowConfig) error {
	if !IsValidReviewWorkflowMode(config.Mode) {
		return fmt.Errorf("invalid review workflow mode: %s", config.Mode)
	}

	if config.NonBlockingCommentLimit < 0 {
		return fmt.Errorf("review non_blocking_comment_limit cannot be negative")
	}

	return nil
}

// IsValidReviewWorkflowMode executes the corresponding package operation.
func IsValidReviewWorkflowMode(mode workflowstate.ReviewWorkflowMode) bool {
	switch workflowstate.ReviewWorkflowMode(strings.TrimSpace(string(mode))) {
	case workflowstate.ReviewWorkflowModeAgenticReview,
		workflowstate.ReviewWorkflowModeHumanReview,
		workflowstate.ReviewWorkflowModeTrustedAutoMerge:
		return true
	default:
		return false
	}
}

// ReviewWorkflowModeDescription executes the corresponding package operation.
func ReviewWorkflowModeDescription(mode workflowstate.ReviewWorkflowMode) string {
	switch mode {
	case workflowstate.ReviewWorkflowModeAgenticReview:
		return "agentic review inspects pull request evidence before merge"
	case workflowstate.ReviewWorkflowModeHumanReview:
		return "human review waits for a person to inspect and approve the diff"
	case workflowstate.ReviewWorkflowModeTrustedAutoMerge:
		return "trusted auto-merge relies on the configured process without a review stop"
	default:
		return "unknown review mode"
	}
}
