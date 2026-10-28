package closeout

import (
	"fmt"
	"path/filepath"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// checkAttempt validates bounded process and outcome records without observing live processes.
func checkAttempt(setID string, closing *workflowstate.WorkflowClosingBinding, index int, state Attempt) error {
	if !attemptIdentity(setID, index, state) {
		return workflowstate.ErrConflict
	}

	if err := checkOutcome(state, index < len(closing.Attempts)-1); err != nil {
		return err
	}

	if state.Input.Ref != File(setID, index+1, "input.json") || len(state.Input.Hash) != 64 {
		return workflowstate.ErrConflict
	}

	if state.Result.Ref != "" && (state.Result.Ref != File(setID, index+1, "result.json") || len(state.Result.Hash) != 64) {
		return workflowstate.ErrConflict
	}

	for i, process := range state.Processes {
		if err := checkProcess(setID, state, i, process); err != nil {
			return err
		}
	}

	return nil
}

func attemptIdentity(setID string, index int, state Attempt) bool {
	return state.Schema == "chr-set-closing-attempt/v1" && state.SetID == setID && state.Generation == index+1 && validOwner(state)
}

func validOwner(state Attempt) bool {
	return validSlot(state.CapacitySlot) && filepath.IsAbs(state.OwnerRoot) && workflowstate.ValidProcessIdentity(state.Owner) && len(state.TokenHash) == 64 && !state.LeaseUntil.IsZero() && len(state.Processes) <= 16 && len(state.BudgetID) >= 1 && len(state.BudgetID) <= 128
}

func checkOutcome(state Attempt, historical bool) error {
	if state.Outcome == "running" {
		if !runningOutcome(state, historical) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if state.FinishedAt.IsZero() {
		return workflowstate.ErrConflict
	}

	switch state.Outcome {
	case "completed", "closing_validation_failed":
		if state.Result.Ref == "" {
			return workflowstate.ErrConflict
		}
	case "interrupted":
		if state.Result.Ref != "" {
			return workflowstate.ErrConflict
		}
	default:
		return workflowstate.ErrConflict
	}

	return nil
}

func checkProcess(id string, state Attempt, i int, process workflowstate.Process) error {
	if !validProcess(state, process) {
		return workflowstate.ErrConflict
	}

	base := File(id, state.Generation, fmt.Sprintf("command-%d", i+1))
	if process.StdoutRef != base+".stdout" || process.StderrRef != base+".stderr" {
		return workflowstate.ErrConflict
	}

	if process.State == workflowstate.ProcessStopped {
		if len(process.StdoutHash) != 64 || len(process.StderrHash) != 64 {
			return workflowstate.ErrConflict
		}
	} else if state.Outcome != "running" {
		return workflowstate.ErrConflict
	}

	return nil
}

func validSlot(slot int) bool {
	return slot >= 0 && slot <= workflowstate.WorkerSlotLimit
}

func validProcess(state Attempt, process workflowstate.Process) bool {
	if !validSlot(process.CapacitySlot) || state.CapacitySlot != 0 && process.CapacitySlot != state.CapacitySlot {
		return false
	}

	return workflowstate.ValidProcessIdentity(process.Identity) && process.GroupID == process.Identity.PID && slices.Contains([]workflowstate.ProcessState{workflowstate.ProcessRunning, workflowstate.ProcessStopped}, process.State)
}

func runningOutcome(state Attempt, historical bool) bool {
	return !historical && state.FinishedAt.IsZero() && state.Result.Ref == ""
}
