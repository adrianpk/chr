package integrate

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SourceInput removes prepared-only evidence before source observation.
func SourceInput(input workflowstate.ReviewInput) workflowstate.ReviewInput {
	input.PreparationRef, input.PreparationHash, input.PreparedGit = "", "", nil
	input.Native = nil

	return input
}

// GateFailed recognizes a complete command failure eligible for repair.
func GateFailed(result workflowstate.TaskResult) bool {
	for _, command := range result.Validation {
		if command.ExitCode != 0 && !command.Truncated {
			return true
		}
	}

	return false
}

// CheckGate binds exact commands or the explicit waiver.
func CheckGate(gate workflowstate.Validation, result workflowstate.TaskResult) error {
	if !gate.Required {
		if result.WaiverReason != gate.WaiverReason || len(result.Validation) != 0 {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if len(result.Validation) != len(gate.Commands) {
		return workflowstate.ErrConflict
	}

	for i, command := range gate.Commands {
		observed := result.Validation[i]
		if !slices.Equal(command, observed.Command) || observed.ExitCode != 0 || observed.Truncated {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
