package closeout

import (
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// BudgetUsage counts retained attempts for the original invocation budget.
func BudgetUsage(snapshot workflowstate.Snapshot, id string, evidence Evidence) ([]string, error) {
	attempts := []string{}

	for _, record := range snapshot.Records {
		if record.Set == nil || record.Set.Closing == nil {
			continue
		}

		for i := range record.Set.Closing.Attempts {
			attempt, err := ReadAttempt(record.ID, record.Set.Closing, i, evidence)
			if err != nil {
				return nil, err
			}

			if attempt.BudgetID == id {
				attempts = append(attempts, record.ID+":"+fmt.Sprint(attempt.Generation))
			}
		}
	}

	return attempts, nil
}
