package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// A durable group ledger complements flock when the supervisor is killed before
// its children actually stop. All work-start entrypoints consult it under flock.
func registerWorkflowProcessCapacity(ctx context.Context, claim *WorkflowClaim, process WorkflowProcess) error {
	common, err := workflowRepositoryDirectory(ctx, claim.store.layout.ProjectDir())
	if err != nil {
		return err
	}

	data, err := json.Marshal(process)
	if err != nil {
		return err
	}

	name, err := workflowCapacityProcessName(process)
	if err != nil {
		return err
	}

	return atomicWorkflowFile(filepath.Join(common, "chr-workflow-processes", name), data)
}

func workflowCapacityProcessName(process WorkflowProcess) (string, error) {
	data, err := json.Marshal(process.Identity)

	return contenthash.Sum(data) + ".json", err
}

func guardWorkflowProcessCapacity(common string) error {
	dir := filepath.Join(common, "chr-workflow-processes")

	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	if len(entries) > workflowstate.WorkflowRecordLimit {
		return fmt.Errorf("workflow process observation limit exceeded")
	}

	for _, entry := range entries {
		if err := guardWorkflowCapacityProcess(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func guardWorkflowCapacityProcess(path string) error {
	_, alive, err := workflowCapacityProcess(path)
	if err != nil {
		return err
	}

	if alive {
		return ErrWorkflowCapacity
	}

	return nil
}
