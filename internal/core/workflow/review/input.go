package review

import (
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowReviewInputBytes preserves the retained review input and decision contract.
func WorkflowReviewInputBytes(input workflowstate.ReviewInput) ([]byte, error) {
	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return nil, fmt.Errorf("review input exceeds file bounds")
	}

	return append(data, '\n'), nil
}

// WorkflowReviewInputHash preserves the retained review input and decision contract.
func WorkflowReviewInputHash(input workflowstate.ReviewInput) (string, error) {
	data, err := WorkflowReviewInputBytes(input)

	return contenthash.Sum(data), err
}

// CheckWorkflowReviewInputHash preserves the retained review input and decision contract.
func CheckWorkflowReviewInputHash(input workflowstate.ReviewInput, expected string) error {
	hash, err := WorkflowReviewInputHash(input)
	if err != nil || hash != expected {
		return fmt.Errorf("%w: review inputs changed during execution", workflowstate.ErrConflict)
	}

	return nil
}
