package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Durable group observations retain the occupied slot if both the host and
// supervisor die while a descendant that closed inherited descriptors survives.
func workflowOccupiedProcessSlots(common string) (map[int]bool, error) {
	dir := filepath.Join(common, "chr-workflow-processes")
	entries, err := os.ReadDir(dir)

	occupied := map[int]bool{}
	if errors.Is(err, os.ErrNotExist) {
		return occupied, nil
	}

	if err != nil {
		return nil, err
	}

	if len(entries) > workflowstate.RecordLimit {
		return nil, workflowstate.ErrConflict
	}

	for _, entry := range entries {
		process, alive, err := workflowCapacityProcess(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		if !alive {
			continue
		}
		// Old serial observations have no slot binding and conservatively exclude all
		// new workers until their actual group is quiescent.
		if process.CapacitySlot == 0 {
			return nil, ErrWorkflowCapacity
		}

		occupied[process.CapacitySlot] = true
	}

	return occupied, nil
}

func workflowCapacityProcess(path string) (Process, bool, error) {
	process, alive, err := observeWorkflowCapacityProcess(path)
	if err != nil || alive {
		return process, alive, err
	}

	return process, false, os.Remove(path)
}

func observeWorkflowCapacityProcess(path string) (Process, bool, error) {
	data, err := opsfile.Read(path)
	if err != nil {
		return Process{}, false, err
	}

	var process Process
	if err := json.Unmarshal(data, &process); err != nil {
		return process, false, err
	}

	name, err := workflowCapacityProcessName(process)
	if err != nil {
		return process, false, err
	}

	if name != filepath.Base(path) || !validWorkflowCapacityProcess(process) {
		return process, false, workflowstate.ErrConflict
	}

	alive, err := workflowProcessGroupAlive(process)
	if err != nil || alive {
		return process, alive, err
	}

	return process, false, nil
}

func validWorkflowCapacityProcess(process Process) bool {
	return validWorkflowProcessIdentity(process.Identity) && process.GroupID == process.Identity.PID && process.CapacitySlot >= 0 && process.CapacitySlot <= workflowWorkerSlotLimit
}
