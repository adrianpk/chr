package core

import (
	"fmt"
	"slices"
	"strconv"
)

func validateWorkflowAttemptProcesses(invocationID string, attempt WorkflowAttempt) error {
	if err := validateWorkflowAttemptOwner(attempt); err != nil {
		return err
	}

	if len(attempt.Processes) > 17 {
		return fmt.Errorf("workflow command process limit exceeded")
	}

	for index, process := range attempt.Processes {
		if process.CapacitySlot != attempt.CapacitySlot {
			return ErrWorkflowConflict
		}

		if err := validateWorkflowProcess(invocationID, attempt.ID, index, process); err != nil {
			return err
		}
	}

	if attempt.CheckpointRef != "" && (attempt.CheckpointRef != workflowAttemptFile(invocationID, attempt.ID, "checkpoint.json") || len(attempt.CheckpointHash) != 64 || attempt.FinishedAt.IsZero()) {
		return fmt.Errorf("invalid workflow attempt checkpoint binding")
	}

	return nil
}

func validateWorkflowAttemptOwner(attempt WorkflowAttempt) error {
	if attempt.CapacitySlot < 0 || attempt.CapacitySlot > workflowWorkerSlotLimit {
		return ErrWorkflowConflict
	}

	if attempt.OwnerProcess != nil && !validWorkflowProcessIdentity(*attempt.OwnerProcess) {
		return fmt.Errorf("invalid workflow owner process identity")
	}

	return nil
}

func validWorkflowProcessIdentity(identity WorkflowProcessIdentity) bool {
	birth, err := strconv.ParseUint(identity.StartTicks, 10, 64)

	return identity.PID > 0 && len(identity.BootID) == 36 && err == nil && birth > 0
}

func validateWorkflowProcess(invocationID, attemptID string, index int, process WorkflowProcess) error {
	if !validWorkflowCapacityProcess(process) || !slices.Contains([]string{"running", "stopped"}, process.State) {
		return fmt.Errorf("invalid workflow command process identity or state")
	}

	if err := validateWorkflowCommand(process.Command); err != nil {
		return err
	}

	base := workflowAttemptFile(invocationID, attemptID, fmt.Sprintf("command-%d", index+1))
	if process.StdoutRef != base+".stdout" || process.StderrRef != base+".stderr" {
		return fmt.Errorf("invalid workflow process output references")
	}

	if process.State == "stopped" && (len(process.StdoutHash) != 64 || len(process.StderrHash) != 64) {
		return fmt.Errorf("stopped process requires output hashes")
	}

	return nil
}
