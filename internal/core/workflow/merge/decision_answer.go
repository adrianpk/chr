package merge

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowDecisionResolved uses the latest retained answer for the requested decision.
func WorkflowDecisionResolved(answers []workflowstate.IntegrationAnswer, id string) bool {
	for i := len(answers) - 1; i >= 0; i-- {
		if answers[i].DecisionID == id {
			return answers[i].Action == "resolve"
		}
	}

	return false
}
