package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// MaterializationOptions supplies approved input and the existing native
// key allocator. The allocator is called only for a new materialization.
type MaterializationOptions struct {
	Plan             generationcore.GeneratedWorkflowPlan
	ReserveTicketKey func(context.Context) (string, error)
}

// MaterializationResult returns authoritative set and task records.
type MaterializationResult struct {
	ReceiptRef string
	Set        workflowstate.Record
	Tasks      []workflowstate.Record
}

// MaterializePlan creates one set and one ticket per task, with one descriptor
// per integration unit. All shared images are validated before the first write.
func (s *Store) MaterializePlan(ctx context.Context, opts MaterializationOptions) (result MaterializationResult, err error) {
	if err := workspaceconfig.RequireNoWorkflowUpgrade(s.layout.ProjectDir()); err != nil {
		return result, err
	}

	if err := generationcore.ValidateGeneratedWorkflowPlan(opts.Plan); err != nil {
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

func (s *Store) materializePlan(ctx context.Context, opts MaterializationOptions) (result MaterializationResult, err error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return result, err
	}

	if previous, err := s.resumeMaterialization(snapshot, opts.Plan); previous.Found || err != nil {
		return previous.Result, err
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

func (s *Store) materializeNewPlan(ctx context.Context, opts MaterializationOptions, snapshot workflowstate.Snapshot, body []byte, coordinator []workflowstate.FileChange) (result MaterializationResult, err error) {
	guard, err := s.acquireWorkflowAllocation()
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, s.releaseWorkflowAllocation(guard)) }()

	result, err = s.buildMaterialization(ctx, opts, body, opscore.NextArtifactNumber(s.layout.Root()))
	if err != nil {
		return result, err
	}

	orderFiles, err := s.materializedOrderChanges(&result, opts.Plan)
	if err != nil {
		return result, err
	}

	snapshot.Records = append(snapshot.Records, result.Set)

	snapshot.Records = append(snapshot.Records, result.Tasks...)
	if err := validateWorkflowSnapshot(snapshot); err != nil {
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

func (s *Store) readApprovedPlan(plan generationcore.GeneratedWorkflowPlan) ([]byte, error) {
	path, err := s.safePath(plan.PlanRef)
	if err != nil {
		return nil, err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return nil, err
	}

	if contenthash.Sum(data) != plan.PlanHash {
		return nil, workflowstate.ErrConflict
	}

	return data, nil
}

func (s *Store) materializationFiles(result MaterializationResult, plan generationcore.GeneratedWorkflowPlan) ([]workflowstate.FileChange, error) {
	receipt, err := materializationReceiptFile(plan, result)
	if err != nil {
		return nil, err
	}

	result.Set.Metadata["materialization_receipt_hash"] = json.RawMessage(fmt.Sprintf("%q", contenthash.Sum(receipt.After)))
	files := []workflowstate.FileChange{}

	for _, record := range append([]workflowstate.Record{result.Set}, result.Tasks...) {
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

	contracts, err := materializationContractFiles(result, plan)
	if err != nil {
		return nil, err
	}

	files = append(files, contracts...)

	files = append(files, receipt)

	return files, nil
}

func materializationContractFiles(result MaterializationResult, plan generationcore.GeneratedWorkflowPlan) ([]workflowstate.FileChange, error) {
	files := []workflowstate.FileChange{}

	for _, slice := range plan.Slices {
		data, err := json.Marshal(struct {
			Contract   string
			Validation workflowstate.Validation
		}{slice.Contract, slice.Validation})
		if err != nil {
			return nil, err
		}

		path := filepath.ToSlash(filepath.Join("plan", result.Set.ID, "extension", "chr", "slice-"+slice.ID+".json"))
		files = append(files, workflowstate.FileChange{Path: path, After: data})
	}

	next := 0

	for _, slice := range generationcore.OrderedWorkflowPlan(plan).Slices {
		for _, task := range slice.Tasks {
			next++

			data, err := json.Marshal(task)
			if err != nil {
				return nil, err
			}

			path := filepath.ToSlash(filepath.Join("plan", result.Set.ID, "extension", "chr", fmt.Sprintf("task-%d.json", next)))
			files = append(files, workflowstate.FileChange{Path: path, After: data})
		}
	}

	return files, nil
}
