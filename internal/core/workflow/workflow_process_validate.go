package workflow

import (
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowAttemptProcesses(invocationID string, attempt workflowstate.Attempt) error {
	if err := validateWorkflowAttemptOwner(attempt); err != nil {
		return err
	}

	if len(attempt.Processes) > 17 {
		return fmt.Errorf("workflow command process limit exceeded")
	}

	for index, process := range attempt.Processes {
		if process.CapacitySlot != attempt.CapacitySlot {
			return workflowstate.ErrConflict
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

func validateWorkflowAttemptOwner(attempt workflowstate.Attempt) error {
	if attempt.CapacitySlot < 0 || attempt.CapacitySlot > workflowstate.WorkerSlotLimit {
		return workflowstate.ErrConflict
	}

	if attempt.OwnerProcess != nil && !workflowstate.ValidProcessIdentity(*attempt.OwnerProcess) {
		return fmt.Errorf("invalid workflow owner process identity")
	}

	return nil
}

func validateWorkflowProcess(invocationID, attemptID string, index int, process workflowstate.Process) error {
	if !validWorkflowCapacityProcess(process) || !slices.Contains([]workflowstate.ProcessState{workflowstate.ProcessRunning, workflowstate.ProcessStopped}, process.State) {
		return fmt.Errorf("invalid workflow command process identity or state")
	}

	if err := workflowstate.ValidateCommand(process.Command); err != nil {
		return err
	}

	base := workflowAttemptFile(invocationID, attemptID, fmt.Sprintf("command-%d", index+1))
	if process.StdoutRef != base+".stdout" || process.StderrRef != base+".stderr" {
		return fmt.Errorf("invalid workflow process output references")
	}

	if process.State == workflowstate.ProcessStopped && (len(process.StdoutHash) != 64 || len(process.StderrHash) != 64) {
		return fmt.Errorf("stopped process requires output hashes")
	}

	return nil
}
