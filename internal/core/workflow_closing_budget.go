package core

import (
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowClosingBudgetID(set WorkflowRecord, invocationID string) string {
	if invocationID != "" {
		return invocationID
	}

	// A standalone decision has one budget per explicitly reopened cycle;
	// changed gate contracts and ordinary retries cannot allocate another one.
	return fmt.Sprintf("standalone-%d", len(set.Set.ClosureHistory))
}

func (s *WorkflowStore) workflowClosingBudget(set WorkflowRecord, invocationID string) error {
	if set.Set.Closing == nil {
		return nil
	}

	if len(set.Set.Closing.Attempts) >= workflowstate.RecordLimit {
		return fmt.Errorf("set closing history bound exhausted")
	}

	budget := workflowClosingBudgetID(set, invocationID)
	count := 0

	for index := range set.Set.Closing.Attempts {
		state, err := s.readWorkflowClosingAttempt(set, index)
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
