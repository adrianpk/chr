package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// WorkflowMaterializationOptions supplies approved input and the existing native
// key allocator. The allocator is called only for a new materialization.
type WorkflowMaterializationOptions struct {
	Plan             GeneratedWorkflowPlan
	ReserveTicketKey func(context.Context) (string, error)
}

// WorkflowMaterializationResult returns authoritative set and task records.
type WorkflowMaterializationResult struct {
	Set   WorkflowRecord
	Tasks []WorkflowRecord
}

// MaterializePlan creates one set and one ticket per task, with one descriptor
// per integration unit. All shared images are validated before the first write.
func (s *WorkflowStore) MaterializePlan(ctx context.Context, opts WorkflowMaterializationOptions) (result WorkflowMaterializationResult, err error) {
	if err := ValidateGeneratedWorkflowPlan(opts.Plan); err != nil {
		return result, err
	}

	if opts.ReserveTicketKey == nil {
		return result, fmt.Errorf("native ticket key allocator is required")
	}

	err = s.withLock(ctx, func() error {
		var materializeErr error

		result, materializeErr = s.materializePlan(ctx, opts)

		return materializeErr
	})

	return result, err
}

func (s *WorkflowStore) materializePlan(ctx context.Context, opts WorkflowMaterializationOptions) (result WorkflowMaterializationResult, err error) {
	body, err := s.readApprovedPlan(opts.Plan)
	if err != nil {
		return result, err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return result, err
	}

	lock := filepath.Join(s.layout.Root(), "lock", "id.lock")
	if err := acquireOpsArtifactLock(lock); err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, os.Remove(lock)) }()

	result, err = s.buildMaterialization(ctx, opts, body, nextOpsArtifactNumber(s.layout.Root()))
	if err != nil {
		return result, err
	}

	snapshot.Records = append(snapshot.Records, result.Set)

	snapshot.Records = append(snapshot.Records, result.Tasks...)
	if err := snapshot.Validate(); err != nil {
		return result, err
	}

	files, err := s.materializationFiles(result, opts.Plan)
	if err != nil {
		return result, err
	}

	if err := s.commitFiles(files); err != nil {
		return result, err
	}

	return s.readMaterialization(result)
}

func (s *WorkflowStore) readApprovedPlan(plan GeneratedWorkflowPlan) ([]byte, error) {
	path, err := s.safePath(plan.PlanRef)
	if err != nil {
		return nil, err
	}

	data, err := readWorkflowFile(path)
	if err != nil {
		return nil, err
	}

	if workflowHash(data) != plan.PlanHash {
		return nil, ErrWorkflowConflict
	}

	return data, nil
}

func (s *WorkflowStore) materializationFiles(result WorkflowMaterializationResult, plan GeneratedWorkflowPlan) ([]workflowFileChange, error) {
	files := []workflowFileChange{}

	for _, record := range append([]WorkflowRecord{result.Set}, result.Tasks...) {
		encoded, err := s.encodeChanges(record)
		if err != nil {
			return nil, err
		}

		for _, change := range encoded {
			if change.Before != "" {
				return nil, fmt.Errorf("%w: allocated path already exists", ErrWorkflowConflict)
			}
		}

		files = append(files, encoded...)
	}

	for _, slice := range plan.Slices {
		data, err := json.Marshal(struct {
			Contract   string
			Validation WorkflowValidation
		}{slice.Contract, slice.Validation})
		if err != nil {
			return nil, err
		}

		path := filepath.ToSlash(filepath.Join("plan", result.Set.ID, "extension", "chr", "slice-"+slice.ID+".json"))
		files = append(files, workflowFileChange{Path: path, After: data})
	}

	return files, nil
}

func (s *WorkflowStore) readMaterialization(result WorkflowMaterializationResult) (WorkflowMaterializationResult, error) {
	set, err := s.readRecord(result.Set.Path)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	tasks := make([]WorkflowRecord, 0, len(result.Tasks))
	for _, task := range result.Tasks {
		record, err := s.readRecord(task.Path)
		if err != nil {
			return WorkflowMaterializationResult{}, err
		}

		tasks = append(tasks, record)
	}

	return WorkflowMaterializationResult{Set: set, Tasks: tasks}, nil
}
