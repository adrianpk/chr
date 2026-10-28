package review

import (
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Admission is the review-mode decision applied by the guarded host transaction.
type Admission struct {
	State    workflowstate.ReviewState
	Outcome  string
	Decision workflowstate.ReviewDecision
}

// Prepare selects a wait, trusted decision or separate review role. The retained
// binding is read only for the trusted decision; no claim or write is performed.
func Prepare(mode workflowstate.ReviewMode, binding *workflowstate.ReviewBinding, command []string) (Admission, error) {
	switch mode {
	case workflowstate.ReviewHuman:
		return Admission{State: workflowstate.ReviewWaitingApproval, Outcome: "slice_review_waiting_approval"}, nil
	case workflowstate.ReviewTrustedAutoMerge:
		decision := workflowstate.ReviewDecision{BindingHash: binding.InputHash, Result: workflowstate.AgenticReviewResult{Decision: workflowstate.IntegrationReviewDecisionPass, Summary: "Required host validation accepted under current trusted policy."}}

		return Admission{State: workflowstate.ReviewAuthorized, Decision: decision}, nil
	case workflowstate.ReviewAgentic:
		if err := workflowstate.ValidateCommand(command); err != nil {
			return Admission{}, fmt.Errorf("separate workflow review command required: %w", err)
		}

		return Admission{State: workflowstate.ReviewRunning}, nil
	default:
		return Admission{}, workflowstate.ErrConflict
	}
}
