package closeout

import (
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckHistory checks each retained attempt and its evidence in journal order.
func CheckHistory(setID string, closing *workflowstate.WorkflowClosingBinding, evidence Evidence) error {
	if closing == nil {
		return nil
	}

	budgets := map[string]int{}

	for index := range closing.Attempts {
		state, err := ReadAttempt(setID, closing, index, evidence)
		if err != nil {
			return err
		}

		budgets[state.BudgetID]++
		if budgets[state.BudgetID] > workflowstate.PageLimit {
			return fmt.Errorf("set invocation closing attempt budget exceeded")
		}

		if err := checkAttemptEvidence(state, evidence); err != nil {
			return err
		}
	}

	return nil
}

func checkAttemptEvidence(state Attempt, evidence Evidence) error {
	if err := checkResult(state, evidence); err != nil {
		return err
	}

	refs := map[string]string{state.Input.Ref: state.Input.Hash}
	if state.Result.Ref != "" {
		refs[state.Result.Ref] = state.Result.Hash
	}

	for _, process := range state.Processes {
		if process.State == workflowstate.ProcessStopped {
			refs[process.StdoutRef], refs[process.StderrRef] = process.StdoutHash, process.StderrHash
		}
	}

	return evidence.Check(refs)
}
