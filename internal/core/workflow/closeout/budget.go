package closeout

import (
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// BudgetID reuses one standalone budget until the set is explicitly reopened.
func BudgetID(set workflowstate.WorkflowSet, invocationID string) string {
	if invocationID != "" {
		return invocationID
	}

	// A standalone decision has one budget per explicitly reopened cycle;
	// changed gate contracts and ordinary retries cannot allocate another one.
	return fmt.Sprintf("standalone-%d", len(set.ClosureHistory))
}

// CheckBudget preserves the bounded retry budget for one invocation or reopened cycle.
func CheckBudget(setID string, set workflowstate.WorkflowSet, invocationID string, evidence Evidence) error {
	if set.Closing == nil {
		return nil
	}

	if len(set.Closing.Attempts) >= workflowstate.RecordLimit {
		return fmt.Errorf("set closing history bound exhausted")
	}

	budget := BudgetID(set, invocationID)
	count := 0

	for index := range set.Closing.Attempts {
		state, err := ReadAttempt(setID, set.Closing, index, evidence)
		if err != nil {
			return err
		}

		if state.BudgetID == budget {
			count++
		}
	}

	if count >= workflowstate.PageLimit {
		return fmt.Errorf("set invocation closing attempt budget exhausted")
	}

	return nil
}
