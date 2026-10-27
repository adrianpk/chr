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

// WorkflowOrderMapInput supplies declared order only. Live task inputs and
// prerequisite images remain the authority for current processing state.
type WorkflowOrderMapInput struct {
	Binding    workflowstate.PlanRevisionBinding `json:"binding"`
	Reference  string                            `json:"reference"`
	SourceHash string                            `json:"source_hash"`
	Content    string                            `json:"content"`
}

func (s *WorkflowStore) requiredWorkflowOrderMap(ctx context.Context, record WorkflowRecord) (*WorkflowOrderMapInput, error) {
	if record.Task.SetID == "" {
		return nil, nil
	}

	set, err := s.readWorkflowOrderSet(record.Task.SetID)
	if err != nil {
		return nil, err
	}

	if set.Set.PlanRevision == nil {
		return nil, nil
	}

	snapshot, err := s.readWorkflowOrderSources(ctx, set.ID)
	if err != nil {
		return nil, err
	}

	source, output, err := s.workflowOrderTrackerImage(snapshot, set.ID)
	if err != nil {
		return nil, err
	}

	if err := s.checkRequiredOrderStatus(source, output); err != nil {
		return nil, err
	}

	return &WorkflowOrderMapInput{Binding: *set.Set.PlanRevision, Reference: source.Path, SourceHash: source.SourceHash, Content: string(output)}, nil
}

func (s *WorkflowStore) checkRequiredOrderImages(source WorkflowTrackerStatus, output []byte) error {
	binding := source.PlanRevision
	if binding == nil || binding.OrderHash != source.SourceHash || binding.TrackerHash != contenthash.Sum(output) {
		return workflowstate.ErrProjectionStale
	}

	current, err := s.readPlanOrderBinding(*binding)
	if err != nil || current != *binding {
		return errors.Join(workflowstate.ErrProjectionStale, err)
	}

	images := map[string]string{source.Path: binding.TrackerHash, binding.TrackerRef: binding.TrackerHash, binding.OrderRef: binding.OrderHash}
	for reference, hash := range images {
		data, err := s.readWorkflowTrackerFile(reference)
		if err != nil || contenthash.Sum(data) != hash {
			return fmt.Errorf("%w: %s observed hash %s: %w", workflowstate.ErrProjectionStale, reference, contenthash.Sum(data), err)
		}
	}

	return nil
}

func (s *WorkflowStore) repairRequiredOrderImages(snapshot WorkflowSnapshot, source WorkflowTrackerStatus, output []byte, opts WorkflowTrackerOptions) error {
	images, err := s.requiredOrderRepairImages(snapshot, source, output)
	if err != nil {
		return err
	}

	changes := []workflowstate.FileChange{}

	for _, image := range images {
		change, err := s.orderImageChange(image.ref, image.body, image.hash, opts.ReplaceHash)
		if err != nil {
			return err
		}

		if change.Before != contenthash.Sum(change.After) {
			changes = append(changes, change)
		}
	}

	if len(changes) == 0 {
		return nil
	}

	return s.commitFiles(changes)
}

func (s *WorkflowStore) retainedOrderRegistration(setID string) (WorkflowTrackerStatus, bool, error) {
	set, err := s.readWorkflowOrderSet(setID)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, workflowstate.ErrNotWorkflow) {
		return WorkflowTrackerStatus{}, false, nil
	}

	if err != nil {
		return WorkflowTrackerStatus{}, false, err
	}

	if set.Set.PlanRevision == nil {
		return WorkflowTrackerStatus{}, false, nil
	}

	reference := filepath.ToSlash(filepath.Join(filepath.Dir(set.Set.PlanRevision.OrderRef), "projection.json"))

	data, err := s.readWorkflowTrackerFile(reference)
	if err != nil {
		return WorkflowTrackerStatus{}, false, err
	}

	var status WorkflowTrackerStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return status, false, err
	}

	if err := s.validateWorkflowTracker(status, setID); err != nil {
		return status, false, err
	}

	if !matchingOrderRegistration(status, *set.Set.PlanRevision) {
		return status, false, workflowstate.ErrConflict
	}

	return status, true, nil
}

func (s *WorkflowStore) guardBoundWorkflowTaskOrder(old, next WorkflowRecord) error {
	if old.Task == nil || old.Task.SetID == "" {
		return nil
	}

	if next.Task == nil {
		return workflowstate.ErrConflict
	}

	set, err := s.readWorkflowOrderSet(old.Task.SetID)
	if err != nil {
		return err
	}

	if set.Set.PlanRevision == nil {
		return nil
	}

	if workflowTaskOrderChanged(old, next) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) checkRequiredOrderStatus(source WorkflowTrackerStatus, output []byte) error {
	status, err := s.readWorkflowTracker(source.SetID)
	if err != nil {
		return errors.Join(workflowstate.ErrProjectionStale, err)
	}

	if status.Schema != "chr-tracker/v2" || status.SourceHash != source.SourceHash || status.OutputHash != contenthash.Sum(output) {
		return workflowstate.ErrProjectionStale
	}

	if err := s.checkRequiredOrderImages(source, output); err != nil {
		return err
	}

	return nil
}

type workflowOrderRepairImage struct {
	ref, hash string
	body      []byte
}

func (s *WorkflowStore) requiredOrderRepairImages(snapshot WorkflowSnapshot, source WorkflowTrackerStatus, output []byte) ([]workflowOrderRepairImage, error) {
	binding := source.PlanRevision
	if binding == nil || binding.OrderHash != source.SourceHash || binding.TrackerHash != contenthash.Sum(output) {
		return nil, workflowstate.ErrProjectionStale
	}

	set, tasks, err := workflowTrackerMembers(snapshot, source.SetID)
	if err != nil {
		return nil, err
	}

	document, err := s.workflowOrderDocument(set, tasks)
	if err != nil {
		return nil, err
	}

	orderImages, err := planOrderBytes(document)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(binding)
	if err != nil {
		return nil, err
	}

	images := []workflowOrderRepairImage{{binding.OrderRef, binding.OrderHash, orderImages.Order}, {binding.TrackerRef, binding.TrackerHash, output}, {filepath.ToSlash(filepath.Join(filepath.Dir(binding.OrderRef), "binding.json")), contenthash.Sum(encoded), encoded}}

	return images, nil
}

func workflowTaskOrderChanged(old, next WorkflowRecord) bool {
	changed := old.Task.Key != next.Task.Key || old.Task.SetID != next.Task.SetID || old.Task.SliceID != next.Task.SliceID || old.Task.Order != next.Task.Order

	return changed || string(old.Metadata["title"]) != string(next.Metadata["title"]) || string(old.Domain["plan_task_id"]) != string(next.Domain["plan_task_id"])
}

func matchingOrderRegistration(status WorkflowTrackerStatus, binding workflowstate.PlanRevisionBinding) bool {
	return status.Schema == "chr-tracker/v2" && status.PlanRevision != nil && *status.PlanRevision == binding
}
