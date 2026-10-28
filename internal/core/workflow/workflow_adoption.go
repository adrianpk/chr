package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// AdoptionTask selects one observed legacy ticket for one approved task.
type AdoptionTask struct {
	TaskID     string `json:"task_id"`
	Path       string `json:"path"`
	SourceHash string `json:"source_hash"`
}

// PlanAdoptionOptions binds a selected canonical plan envelope and an
// exact task selection. CheckQuiescent must also check legacy runtime evidence.
type PlanAdoptionOptions struct {
	Plan           generationcore.GeneratedWorkflowPlan
	PlanPath       string
	SourceHash     string
	Tasks          []AdoptionTask
	CheckQuiescent func(context.Context, []string) error `json:"-"`
}

// AdoptPlanWorkflow explicitly adopts selected active records. Ordinary reads
// never invoke adoption, and inferred historical status creates no receipt.
func (s *Store) AdoptPlanWorkflow(ctx context.Context, opts PlanAdoptionOptions) (result MaterializationResult, err error) {
	if err := workspaceconfig.RequireNoWorkflowUpgrade(s.layout.ProjectDir()); err != nil {
		return result, err
	}

	if err := generationcore.ValidateGeneratedWorkflowPlan(opts.Plan); err != nil {
		return result, err
	}

	if opts.Plan.PlanRevision != nil {
		return result, fmt.Errorf("revision-bound plans require native materialization")
	}

	if opts.CheckQuiescent == nil {
		return result, fmt.Errorf("legacy ownership observation is required")
	}

	locks, err := s.acquireWorkflowAdoptionLocks(ctx)
	if err != nil {
		return result, err
	}

	defer func() { err = errors.Join(err, releaseWorkflowAdoptionLocks(locks)) }()

	err = s.withWorkflowPlanMutation(ctx, func(ctx context.Context) error {
		if err := s.requirePlanRevisionBinding(opts.Plan.PlanRef); err != nil {
			return err
		}

		var adoptErr error

		result, adoptErr = s.adoptPlanWorkflow(ctx, opts)

		return adoptErr
	})

	return result, err
}

func (s *Store) adoptPlanWorkflow(ctx context.Context, opts PlanAdoptionOptions) (MaterializationResult, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return MaterializationResult{}, err
	}

	hash, err := workflowAdoptionHash(opts)
	if err != nil {
		return MaterializationResult{}, err
	}

	if retained, err := s.resumePlanAdoption(snapshot, opts, hash); retained.Found || err != nil {
		return retained.Result, err
	}

	result, err := s.preparePlanAdoption(ctx, opts)
	if err != nil {
		return result, err
	}

	result.Set.Metadata["adoption_hash"] = json.RawMessage(fmt.Sprintf("%q", hash))
	snapshot.Records = append(snapshot.Records, result.Set)

	snapshot.Records = append(snapshot.Records, result.Tasks...)
	if err := validateWorkflowSnapshot(snapshot); err != nil {
		return MaterializationResult{}, err
	}

	coordinator, err := s.workflowCoordinatorFiles(ctx)
	if err != nil {
		return MaterializationResult{}, err
	}

	files, err := s.planAdoptionFiles(result, opts.Plan)
	if err != nil {
		return MaterializationResult{}, err
	}

	if err := s.commitFiles(append(files, coordinator...)); err != nil {
		return MaterializationResult{}, err
	}

	return s.readMaterializationReceipt(result.Set, opts.Plan, snapshot)
}

func workflowAdoptionHash(opts PlanAdoptionOptions) (string, error) {
	selected := append([]AdoptionTask(nil), opts.Tasks...)
	sort.Slice(selected, func(i, j int) bool { return selected[i].TaskID < selected[j].TaskID })

	data, err := json.Marshal(struct {
		PlanPath   string
		SourceHash string
		Tasks      []AdoptionTask
	}{opts.PlanPath, opts.SourceHash, selected})

	return contenthash.Sum(data), err
}

func (s *Store) preparePlanAdoption(ctx context.Context, opts PlanAdoptionOptions) (MaterializationResult, error) {
	set, err := s.prepareLegacyWorkflowPlan(opts)
	if err != nil {
		return MaterializationResult{}, err
	}

	plan := generationcore.OrderedWorkflowPlan(opts.Plan)

	selected, err := s.prepareLegacyWorkflowTasks(plan, opts.Tasks)
	if err != nil {
		return MaterializationResult{}, err
	}

	if err := s.checkLegacyWorkflowKeys(ctx, selected.Records); err != nil {
		return MaterializationResult{}, err
	}

	if err := opts.CheckQuiescent(ctx, selected.Keys); err != nil {
		return MaterializationResult{}, err
	}

	next := 0
	materialize := MaterializationOptions{Plan: plan, ReserveTicketKey: func(context.Context) (string, error) { key := selected.Keys[next]; next++; return key, nil }}

	result, err := s.buildMaterializationForSet(ctx, materialize, workflowMaterializationSetInput{Plan: plan, Body: set.Body, Set: set, Allocation: workflowArtifactAllocation{Number: 1, Time: time.Now().UTC()}})
	if err != nil {
		return result, err
	}

	result.Tasks, err = s.bindLegacyWorkflowTasks(result.Tasks, selected.Records)

	return result, err
}

func (s *Store) planAdoptionFiles(result MaterializationResult, plan generationcore.GeneratedWorkflowPlan) ([]workflowstate.FileChange, error) {
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

		observed := encoded[0].Before
		if len(encoded) == 2 {
			observed += ":" + encoded[1].Before
		}

		if observed != record.SourceHash {
			return nil, workflowstate.ErrConflict
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

func (s *Store) resumePlanAdoption(snapshot workflowstate.Snapshot, opts PlanAdoptionOptions, hash string) (workflowMaterializationReplay, error) {
	retained, err := s.resumeMaterialization(snapshot, opts.Plan)
	if err != nil || !retained.Found {
		return retained, err
	}

	if string(retained.Result.Set.Metadata["adoption_hash"]) != fmt.Sprintf("%q", hash) {
		return retained, workflowstate.ErrConflict
	}

	return retained, nil
}
