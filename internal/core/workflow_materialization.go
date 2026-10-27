package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowMaterializationOptions supplies approved input and the existing native
// key allocator. The allocator is called only for a new materialization.
type WorkflowMaterializationOptions struct {
	Plan             GeneratedWorkflowPlan
	ReserveTicketKey func(context.Context) (string, error)
}

// WorkflowMaterializationResult returns authoritative set and task records.
type WorkflowMaterializationResult struct {
	ReceiptRef string
	Set        WorkflowRecord
	Tasks      []WorkflowRecord
}

// MaterializePlan creates one set and one ticket per task, with one descriptor
// per integration unit. All shared images are validated before the first write.
func (s *WorkflowStore) MaterializePlan(ctx context.Context, opts WorkflowMaterializationOptions) (result WorkflowMaterializationResult, err error) {
	if err := requireNoWorkflowUpgrade(s.layout.ProjectDir()); err != nil {
		return result, err
	}

	if err := ValidateGeneratedWorkflowPlan(opts.Plan); err != nil {
		return result, err
	}

	if opts.ReserveTicketKey == nil {
		return result, fmt.Errorf("native ticket key allocator is required")
	}

	err = s.withWorkflowPlanMutation(ctx, func(ctx context.Context) error {
		var materializeErr error

		result, materializeErr = s.materializePlan(ctx, opts)

		return materializeErr
	})

	return result, err
}

func (s *WorkflowStore) materializePlan(ctx context.Context, opts WorkflowMaterializationOptions) (result WorkflowMaterializationResult, err error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return result, err
	}

	if previous, found, err := s.resumeMaterialization(snapshot, opts.Plan); found || err != nil {
		return previous, err
	}

	body, err := s.readApprovedPlan(opts.Plan)
	if err != nil {
		return result, err
	}

	if err := s.validateApprovedOrder(opts.Plan, ""); err != nil {
		return result, err
	}

	coordinator, err := s.workflowCoordinatorFiles(ctx)
	if err != nil {
		return result, err
	}

	return s.materializeNewPlan(ctx, opts, snapshot, body, coordinator)
}

func (s *WorkflowStore) materializeNewPlan(ctx context.Context, opts WorkflowMaterializationOptions, snapshot WorkflowSnapshot, body []byte, coordinator []workflowFileChange) (result WorkflowMaterializationResult, err error) {
	guard, err := s.acquireWorkflowAllocation()
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, s.releaseWorkflowAllocation(guard)) }()

	result, err = s.buildMaterialization(ctx, opts, body, nextOpsArtifactNumber(s.layout.Root()))
	if err != nil {
		return result, err
	}

	orderFiles, err := s.materializedOrderChanges(&result, opts.Plan)
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

	if err := s.commitAllocatedFiles(append(append(files, orderFiles...), coordinator...), guard); err != nil {
		return result, err
	}

	return s.readMaterializationReceipt(result.Set, opts.Plan, snapshot)
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

	if contenthash.Sum(data) != plan.PlanHash {
		return nil, workflowstate.ErrConflict
	}

	return data, nil
}

func (s *WorkflowStore) materializationFiles(result WorkflowMaterializationResult, plan GeneratedWorkflowPlan) ([]workflowFileChange, error) {
	receipt, err := materializationReceiptFile(plan, result)
	if err != nil {
		return nil, err
	}

	result.Set.Metadata["materialization_receipt_hash"] = json.RawMessage(fmt.Sprintf("%q", contenthash.Sum(receipt.After)))
	files := []workflowFileChange{}

	for _, record := range append([]WorkflowRecord{result.Set}, result.Tasks...) {
		encoded, err := s.encodeChanges(record)
		if err != nil {
			return nil, err
		}

		for _, change := range encoded {
			if change.Before != "" {
				return nil, fmt.Errorf("%w: allocated path already exists", workflowstate.ErrConflict)
			}
		}

		files = append(files, encoded...)
	}

	contracts, err := s.materializationContractFiles(result, plan)
	if err != nil {
		return nil, err
	}

	files = append(files, contracts...)

	files = append(files, receipt)

	return files, nil
}

func (s *WorkflowStore) materializationContractFiles(result WorkflowMaterializationResult, plan GeneratedWorkflowPlan) ([]workflowFileChange, error) {
	files := []workflowFileChange{}

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

	next := 0

	for _, slice := range orderedWorkflowPlan(plan).Slices {
		for _, task := range slice.Tasks {
			next++

			data, err := json.Marshal(task)
			if err != nil {
				return nil, err
			}

			path := filepath.ToSlash(filepath.Join("plan", result.Set.ID, "extension", "chr", fmt.Sprintf("task-%d.json", next)))
			files = append(files, workflowFileChange{Path: path, After: data})
		}
	}

	return files, nil
}
