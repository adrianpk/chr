package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowMaterializationSchema = "chr-materialization/v1"

type workflowMaterializationBinding struct {
	Path         string `json:"path"`
	Key          string `json:"key"`
	SliceID      string `json:"slice_id"`
	Order        int    `json:"task_order"`
	ContractHash string `json:"execution_contract_hash"`
}

type workflowMaterializationReceipt struct {
	Schema    string                           `json:"materialization_schema"`
	RequestID string                           `json:"request_id"`
	InputHash string                           `json:"input_hash"`
	SetID     string                           `json:"set_id"`
	SetPath   string                           `json:"set_path"`
	Tasks     []workflowMaterializationBinding `json:"tasks"`
	Slices    []workflowstate.WorkflowSlice    `json:"slices"`
}

func workflowMaterializationHash(plan generationcore.GeneratedWorkflowPlan) (string, error) {
	data, err := json.Marshal(generationcore.OrderedWorkflowPlan(plan))

	return contenthash.Sum(data), err
}

func materializationReceiptFile(plan generationcore.GeneratedWorkflowPlan, result MaterializationResult) (workflowstate.FileChange, error) {
	hash, err := workflowMaterializationHash(plan)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	receipt := workflowMaterializationReceipt{Schema: workflowMaterializationSchema, RequestID: plan.MaterializationID, InputHash: hash, SetID: result.Set.ID, SetPath: result.Set.Path, Slices: result.Set.Set.Slices}
	for _, record := range result.Tasks {
		task := record.Task
		receipt.Tasks = append(receipt.Tasks, workflowMaterializationBinding{Path: record.Path, Key: task.Key, SliceID: task.SliceID, Order: task.Order, ContractHash: task.ContractHash})
	}

	data, err := json.Marshal(receipt)

	return workflowstate.FileChange{Path: filepath.ToSlash(filepath.Join("plan", result.Set.ID, "evidence", "materialization.json")), After: data}, err
}

type workflowMaterializationReplay struct {
	Result MaterializationResult
	Found  bool
}

func (s *Store) resumeMaterialization(snapshot workflowstate.Snapshot, plan generationcore.GeneratedWorkflowPlan) (workflowMaterializationReplay, error) {
	var set *workflowstate.Record

	for i, record := range snapshot.Records {
		if record.Set == nil {
			continue
		}

		var requestID string
		if raw, ok := record.Metadata["materialization_id"]; ok {
			if err := json.Unmarshal(raw, &requestID); err != nil {
				return workflowMaterializationReplay{}, err
			}
		}

		if requestID != plan.MaterializationID {
			continue
		}

		if set != nil {
			return workflowMaterializationReplay{}, fmt.Errorf("duplicate materialization identity")
		}

		set = &snapshot.Records[i]
	}

	if set == nil {
		return workflowMaterializationReplay{}, nil
	}

	result, err := s.readMaterializationReceipt(*set, plan, snapshot)

	return workflowMaterializationReplay{Result: result, Found: true}, err
}

func (s *Store) readMaterializationReceipt(set workflowstate.Record, plan generationcore.GeneratedWorkflowPlan, snapshot workflowstate.Snapshot) (MaterializationResult, error) {
	observed, err := s.readRecord(set.Path)
	if err != nil {
		return MaterializationResult{}, err
	}

	set = observed

	path, err := s.safePath(filepath.ToSlash(filepath.Join("plan", set.ID, "evidence", "materialization.json")))
	if err != nil {
		return MaterializationResult{}, err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return MaterializationResult{}, err
	}

	if string(set.Metadata["materialization_receipt_hash"]) != fmt.Sprintf("%q", contenthash.Sum(data)) {
		return MaterializationResult{}, workflowstate.ErrConflict
	}

	var receipt workflowMaterializationReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return MaterializationResult{}, err
	}

	hash, err := workflowMaterializationHash(plan)
	if err != nil {
		return MaterializationResult{}, err
	}

	if !matchingMaterializationReceipt(receipt, set, plan, hash) {
		return MaterializationResult{}, workflowstate.ErrConflict
	}

	tasks, err := s.readMaterializationBindings(set, receipt)
	if err != nil {
		return MaterializationResult{}, err
	}

	if err := s.checkMaterializationSlices(set, receipt, snapshot); err != nil {
		return MaterializationResult{}, err
	}

	return MaterializationResult{Set: set, Tasks: tasks, ReceiptRef: filepath.ToSlash(filepath.Join("plan", set.ID, "evidence", "materialization.json"))}, nil
}

func (s *Store) readMaterializationBindings(set workflowstate.Record, receipt workflowMaterializationReceipt) ([]workflowstate.Record, error) {
	if len(receipt.Tasks) == 0 || len(receipt.Tasks) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("invalid materialization receipt task bound")
	}

	tasks := []workflowstate.Record{}
	seen := map[string]bool{}

	for _, binding := range receipt.Tasks {
		record, err := s.readRecord(binding.Path)
		if err != nil {
			return nil, err
		}

		if seen[binding.Path] || !matchingMaterializationTask(record, set.ID, binding) {
			return nil, workflowstate.ErrConflict
		}

		seen[binding.Path] = true

		tasks = append(tasks, record)
	}

	return tasks, nil
}

func matchingMaterializationTask(record workflowstate.Record, setID string, binding workflowMaterializationBinding) bool {
	task := record.Task

	return task != nil && task.SetID == setID && task.Key == binding.Key && task.SliceID == binding.SliceID && task.Order == binding.Order && task.ContractHash == binding.ContractHash
}

func (s *Store) checkMaterializationSlices(set workflowstate.Record, receipt workflowMaterializationReceipt, snapshot workflowstate.Snapshot) error {
	if len(receipt.Slices) != len(set.Set.Slices) {
		return workflowstate.ErrConflict
	}

	for i, slice := range receipt.Slices {
		current := set.Set.Slices[i]
		if current.ID != slice.ID || current.Order != slice.Order || current.ContractHash != slice.ContractHash || current.TaskInputHash != TaskInputHash(snapshot.Records, slice.ID) {
			return workflowstate.ErrConflict
		}

		path, err := s.safePath(current.ContractRef)
		if err != nil {
			return err
		}

		data, err := opsfile.Read(path)
		if err != nil {
			return err
		}

		if contenthash.Sum(data) != slice.ContractHash {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func matchingMaterializationReceipt(receipt workflowMaterializationReceipt, set workflowstate.Record, plan generationcore.GeneratedWorkflowPlan, hash string) bool {
	return receipt.Schema == workflowMaterializationSchema && receipt.RequestID == plan.MaterializationID && receipt.InputHash == hash && receipt.SetID == set.ID && receipt.SetPath == set.Path
}
