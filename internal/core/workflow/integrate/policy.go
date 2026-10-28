package integrate

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowPreparationSource removes prepared-only evidence before source observation.
func WorkflowPreparationSource(input workflowstate.ReviewInput) workflowstate.ReviewInput {
	input.PreparationRef, input.PreparationHash, input.PreparedGit = "", "", nil
	input.Native = nil

	return input
}

// WorkflowCandidateGateFailed recognizes a complete command failure eligible for repair.
func WorkflowCandidateGateFailed(result workflowstate.TaskResult) bool {
	for _, command := range result.Validation {
		if command.ExitCode != 0 && !command.Truncated {
			return true
		}
	}

	return false
}

// ValidateWorkflowPreparationGate binds exact commands or the explicit waiver.
func ValidateWorkflowPreparationGate(gate workflowstate.Validation, result workflowstate.TaskResult) error {
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
