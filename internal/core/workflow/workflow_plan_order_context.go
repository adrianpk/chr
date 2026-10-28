package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowTrackerRegistration struct {
	Status TrackerStatus
	Found  bool
}

// OrderMapInput supplies declared order only. Live task inputs and
// prerequisite images remain the authority for current processing state.
type OrderMapInput struct {
	Binding    workflowstate.PlanRevisionBinding `json:"binding"`
	Reference  string                            `json:"reference"`
	SourceHash string                            `json:"source_hash"`
	Content    string                            `json:"content"`
}

func (s *Store) requiredWorkflowOrderMap(ctx context.Context, record Record) (*OrderMapInput, error) {
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

	trackerImage, err := s.workflowOrderTrackerImage(snapshot, set.ID)
	if err != nil {
		return nil, err
	}

	if err := s.checkRequiredOrderStatus(trackerImage.Status, trackerImage.Output); err != nil {
		return nil, err
	}

	return &OrderMapInput{Binding: *set.Set.PlanRevision, Reference: trackerImage.Status.Path, SourceHash: trackerImage.Status.SourceHash, Content: string(trackerImage.Output)}, nil
}

func (s *Store) checkRequiredOrderImages(source TrackerStatus, output []byte) error {
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

func (s *Store) repairRequiredOrderImages(snapshot Snapshot, source TrackerStatus, output []byte, opts TrackerOptions) error {
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

func (s *Store) retainedOrderRegistration(setID string) (workflowTrackerRegistration, error) {
	set, err := s.readWorkflowOrderSet(setID)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, workflowstate.ErrNotWorkflow) {
		return workflowTrackerRegistration{Status: TrackerStatus{}, Found: false}, nil
	}

	if err != nil {
		return workflowTrackerRegistration{Status: TrackerStatus{}, Found: false}, err
	}

	if set.Set.PlanRevision == nil {
		return workflowTrackerRegistration{Status: TrackerStatus{}, Found: false}, nil
	}

	reference := filepath.ToSlash(filepath.Join(filepath.Dir(set.Set.PlanRevision.OrderRef), "projection.json"))

	data, err := s.readWorkflowTrackerFile(reference)
	if err != nil {
		return workflowTrackerRegistration{Status: TrackerStatus{}, Found: false}, err
	}

	var status TrackerStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return workflowTrackerRegistration{Status: status, Found: false}, err
	}

	if err := s.validateWorkflowTracker(status, setID); err != nil {
		return workflowTrackerRegistration{Status: status, Found: false}, err
	}

	if !matchingOrderRegistration(status, *set.Set.PlanRevision) {
		return workflowTrackerRegistration{Status: status, Found: false}, workflowstate.ErrConflict
	}

	return workflowTrackerRegistration{Status: status, Found: true}, nil
}

func (s *Store) guardBoundWorkflowTaskOrder(old, next Record) error {
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

func (s *Store) checkRequiredOrderStatus(source TrackerStatus, output []byte) error {
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

func (s *Store) requiredOrderRepairImages(snapshot Snapshot, source TrackerStatus, output []byte) ([]workflowOrderRepairImage, error) {
	binding := source.PlanRevision
	if binding == nil || binding.OrderHash != source.SourceHash || binding.TrackerHash != contenthash.Sum(output) {
		return nil, workflowstate.ErrProjectionStale
	}

	trackerMembers, err := workflowTrackerMembers(snapshot, source.SetID)
	if err != nil {
		return nil, err
	}

	document, err := s.workflowOrderDocument(trackerMembers.Set, trackerMembers.Tasks)
	if err != nil {
		return nil, err
	}

	orderImages, err := authoringcore.EncodePlanOrder(document)
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

func workflowTaskOrderChanged(old, next Record) bool {
	changed := old.Task.Key != next.Task.Key || old.Task.SetID != next.Task.SetID || old.Task.SliceID != next.Task.SliceID || old.Task.Order != next.Task.Order

	return changed || string(old.Metadata["title"]) != string(next.Metadata["title"]) || string(old.Domain["plan_task_id"]) != string(next.Domain["plan_task_id"])
}

func matchingOrderRegistration(status TrackerStatus, binding workflowstate.PlanRevisionBinding) bool {
	return status.Schema == "chr-tracker/v2" && status.PlanRevision != nil && *status.PlanRevision == binding
}
