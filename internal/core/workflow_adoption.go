package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"
)

// WorkflowAdoptionTask selects one observed legacy ticket for one approved task.
type WorkflowAdoptionTask struct {
	TaskID     string `json:"task_id"`
	Path       string `json:"path"`
	SourceHash string `json:"source_hash"`
}

// WorkflowPlanAdoptionOptions binds a selected canonical plan envelope and an
// exact task selection. CheckQuiescent must also check legacy runtime evidence.
type WorkflowPlanAdoptionOptions struct {
	Plan           GeneratedWorkflowPlan
	PlanPath       string
	SourceHash     string
	Tasks          []WorkflowAdoptionTask
	CheckQuiescent func(context.Context, []string) error `json:"-"`
}

// AdoptPlanWorkflow explicitly adopts selected active records. Ordinary reads
// never invoke adoption, and inferred historical status creates no receipt.
func (s *WorkflowStore) AdoptPlanWorkflow(ctx context.Context, opts WorkflowPlanAdoptionOptions) (result WorkflowMaterializationResult, err error) {
	if err := requireNoWorkflowUpgrade(s.layout.ProjectDir()); err != nil {
		return result, err
	}

	if err := ValidateGeneratedWorkflowPlan(opts.Plan); err != nil {
		return result, err
	}

	if opts.CheckQuiescent == nil {
		return result, fmt.Errorf("legacy ownership observation is required")
	}

	locks, err := s.acquireWorkflowAdoptionLocks(ctx)
	if err != nil {
		return result, err
	}

	defer func() { err = errors.Join(err, releaseWorkflowAdoptionLocks(locks)) }()

	err = s.withLock(ctx, func() error {
		var adoptErr error

		result, adoptErr = s.adoptPlanWorkflow(ctx, opts)

		return adoptErr
	})

	return result, err
}

func (s *WorkflowStore) adoptPlanWorkflow(ctx context.Context, opts WorkflowPlanAdoptionOptions) (WorkflowMaterializationResult, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	hash, err := workflowAdoptionHash(opts)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	if result, found, err := s.resumePlanAdoption(snapshot, opts, hash); found || err != nil {
		return result, err
	}

	result, err := s.preparePlanAdoption(ctx, opts)
	if err != nil {
		return result, err
	}

	result.Set.Metadata["adoption_hash"] = json.RawMessage(fmt.Sprintf("%q", hash))
	snapshot.Records = append(snapshot.Records, result.Set)

	snapshot.Records = append(snapshot.Records, result.Tasks...)
	if err := snapshot.Validate(); err != nil {
		return WorkflowMaterializationResult{}, err
	}

	coordinator, err := s.workflowCoordinatorFiles(ctx)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	files, err := s.planAdoptionFiles(result, opts.Plan)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	if err := s.commitFiles(append(files, coordinator...)); err != nil {
		return WorkflowMaterializationResult{}, err
	}

	return s.readMaterializationReceipt(result.Set, opts.Plan, snapshot)
}

func workflowAdoptionHash(opts WorkflowPlanAdoptionOptions) (string, error) {
	selected := append([]WorkflowAdoptionTask(nil), opts.Tasks...)
	sort.Slice(selected, func(i, j int) bool { return selected[i].TaskID < selected[j].TaskID })

	data, err := json.Marshal(struct {
		PlanPath   string
		SourceHash string
		Tasks      []WorkflowAdoptionTask
	}{opts.PlanPath, opts.SourceHash, selected})

	return workflowHash(data), err
}

func (s *WorkflowStore) preparePlanAdoption(ctx context.Context, opts WorkflowPlanAdoptionOptions) (WorkflowMaterializationResult, error) {
	set, err := s.prepareLegacyWorkflowPlan(opts)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	plan := orderedWorkflowPlan(opts.Plan)

	selected, keys, err := s.prepareLegacyWorkflowTasks(plan, opts.Tasks)
	if err != nil {
		return WorkflowMaterializationResult{}, err
	}

	if err := s.checkLegacyWorkflowKeys(ctx, selected); err != nil {
		return WorkflowMaterializationResult{}, err
	}

	if err := opts.CheckQuiescent(ctx, keys); err != nil {
		return WorkflowMaterializationResult{}, err
	}

	next := 0
	materialize := WorkflowMaterializationOptions{Plan: plan, ReserveTicketKey: func(context.Context) (string, error) { key := keys[next]; next++; return key, nil }}

	result, err := s.buildMaterializationForSet(ctx, materialize, plan, set.Body, 1, time.Now().UTC(), set)
	if err != nil {
		return result, err
	}

	result.Tasks, err = s.bindLegacyWorkflowTasks(result.Tasks, selected)

	return result, err
}

func (s *WorkflowStore) planAdoptionFiles(result WorkflowMaterializationResult, plan GeneratedWorkflowPlan) ([]workflowFileChange, error) {
	receipt, err := materializationReceiptFile(plan, result)
	if err != nil {
		return nil, err
	}

	result.Set.Metadata["materialization_receipt_hash"] = json.RawMessage(fmt.Sprintf("%q", workflowHash(receipt.After)))
	files := []workflowFileChange{}

	for _, record := range append([]WorkflowRecord{result.Set}, result.Tasks...) {
		encoded, err := s.encodeChanges(record)
		if err != nil {
			return nil, err
		}

		observed := encoded[0].Before
		if len(encoded) == 2 {
			observed += ":" + encoded[1].Before
		}

		if observed != record.SourceHash {
			return nil, ErrWorkflowConflict
		}

		files = append(files, encoded...)
	}

	contracts, err := s.materializationContractFiles(result, plan)
	if err != nil {
		return nil, err
	}

	files = append(files, contracts...)

	return append(files, receipt), nil
}

func workflowAdoptionPlanBody(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(path), "artifact.md"))
}

func (s *WorkflowStore) resumePlanAdoption(snapshot WorkflowSnapshot, opts WorkflowPlanAdoptionOptions, hash string) (WorkflowMaterializationResult, bool, error) {
	result, found, err := s.resumeMaterialization(snapshot, opts.Plan)
	if err != nil || !found {
		return result, found, err
	}

	if string(result.Set.Metadata["adoption_hash"]) != fmt.Sprintf("%q", hash) {
		return result, true, ErrWorkflowConflict
	}

	return result, true, nil
}
