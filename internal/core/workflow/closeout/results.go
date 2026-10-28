package closeout

import (
	"reflect"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidateWorkflowClosingResults preserves the existing closing contract.
func ValidateWorkflowClosingResults(gate workflowstate.Validation, receipt workflowstate.ClosingReceipt) error {
	if len(receipt.Validation) > len(gate.Commands) || receipt.Outcome == "completed" && len(receipt.Validation) != len(gate.Commands) {
		return workflowstate.ErrConflict
	}

	for i, result := range receipt.Validation {
		if !reflect.DeepEqual(result.Command, gate.Commands[i]) || receipt.Outcome == "completed" && result.ExitCode != 0 {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
