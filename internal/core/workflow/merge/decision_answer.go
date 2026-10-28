package merge

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// DecisionResolved uses the latest retained answer for the requested decision.
func DecisionResolved(answers []workflowstate.IntegrationAnswer, id string) bool {
	for i := len(answers) - 1; i >= 0; i-- {
		if answers[i].DecisionID == id {
			return answers[i].Action == "resolve"
		}
	}

	return false
}
