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
		observed, err := workflowCapacityProcess(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		if !observed.Alive {
			continue
		}
		// Old serial observations have no slot binding and conservatively exclude all
		// new workers until their actual group is quiescent.
		if observed.Process.CapacitySlot == 0 {
			return nil, ErrWorkflowCapacity
		}

		occupied[observed.Process.CapacitySlot] = true
	}

	return occupied, nil
}

type workflowCapacityObservation struct {
	Process workflowstate.Process
	Alive   bool
}

func workflowCapacityProcess(path string) (workflowCapacityObservation, error) {
	observed, err := observeWorkflowCapacityProcess(path)
	if err != nil || observed.Alive {
		return observed, err
	}

	return observed, os.Remove(path)
}

func observeWorkflowCapacityProcess(path string) (workflowCapacityObservation, error) {
	data, err := opsfile.Read(path)
	if err != nil {
		return workflowCapacityObservation{}, err
	}

	var process workflowstate.Process
	if err := json.Unmarshal(data, &process); err != nil {
		return workflowCapacityObservation{Process: process}, err
	}

	name, err := workflowCapacityProcessName(process)
	if err != nil {
		return workflowCapacityObservation{Process: process}, err
	}

	if name != filepath.Base(path) || !validWorkflowCapacityProcess(process) {
		return workflowCapacityObservation{Process: process}, workflowstate.ErrConflict
	}

	alive, err := workflowProcessGroupAlive(process)
	if err != nil || alive {
		return workflowCapacityObservation{Process: process, Alive: alive}, err
	}

	return workflowCapacityObservation{Process: process}, nil
}

func validWorkflowCapacityProcess(process workflowstate.Process) bool {
	return workflowstate.ValidProcessIdentity(process.Identity) && process.GroupID == process.Identity.PID && process.CapacitySlot >= 0 && process.CapacitySlot <= workflowstate.WorkerSlotLimit
}
