package workflow

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validWorkflowTrackerSchema(schema string) bool {
	return schema == "chr-tracker/v1" || schema == "chr-tracker/v2"
}

func workflowTrackerSchema(snapshot workflowstate.Snapshot, setID, requested, previous string) string {
	if requested != "" {
		return requested
	}

	if previous != "" {
		return previous
	}

	for _, record := range snapshot.Records {
		if record.ID == setID && record.Set != nil && record.Set.PlanRevision != nil {
			return "chr-tracker/v2"
		}
	}

	return "chr-tracker/v1"
}

func (s *Store) versionedWorkflowTrackerImage(snapshot workflowstate.Snapshot, setID, schema string) (workflowTrackerImageResult, error) {
	switch schema {
	case "chr-tracker/v1":
		return (workflowTrackerObserver{store: s}).image(snapshot, setID)
	case "chr-tracker/v2":
		return s.workflowOrderTrackerImage(snapshot, setID)
	default:
		return workflowTrackerImageResult{Status: TrackerStatus{}, Output: nil}, fmt.Errorf("unsupported tracker schema")
	}
}

func (s *Store) workflowOrderTrackerImage(snapshot workflowstate.Snapshot, setID string) (workflowTrackerImageResult, error) {
	trackerMembers, err := workflowTrackerMembers(snapshot, setID)
	if err != nil {
		return workflowTrackerImageResult{Status: TrackerStatus{}, Output: nil}, err
	}

	result := TrackerStatus{Schema: "chr-tracker/v2", ProjectID: s.projectID, Namespace: s.layout.Namespace(), SetID: setID, Path: "tracker/" + setID + ".md", PlanRevision: trackerMembers.Set.Set.PlanRevision}

	document, err := s.workflowOrderDocument(trackerMembers.Set, trackerMembers.Tasks)
	if err != nil {
		return workflowTrackerImageResult{Status: result, Output: nil}, err
	}

	images, err := authoringcore.EncodePlanOrder(document)
	if err != nil {
		return workflowTrackerImageResult{Status: result, Output: nil}, err
	}

	result.SourceHash = contenthash.Sum(images.Order)
	result.Sources = []TrackerSource{{ID: trackerMembers.Set.ID, Path: trackerMembers.Set.Path, Hash: result.SourceHash}, {Path: trackerMembers.Set.Set.PlanRevision.BodyRef, Hash: trackerMembers.Set.Set.PlanRevision.BodyHash}}

	return workflowTrackerImageResult{Status: result, Output: images.Tracker}, nil
}

func (s *Store) workflowOrderDocument(set workflowstate.Record, tasks []workflowstate.Record) (generationcore.PlanOrderDocument, error) {
	binding := set.Set.PlanRevision
	if binding == nil || !authoringcore.ValidPlanBindingPaths(s.layout.Namespace(), *binding) || binding.BodyHash != set.Set.PlanHash {
		return generationcore.PlanOrderDocument{}, fmt.Errorf("order-only tracker requires its retained plan revision")
	}

	document, err := s.workflowPlannedOrder(*binding)
	if err != nil {
		return document, err
	}

	if err := json.Unmarshal(set.Metadata["materialization_id"], &document.MaterializationID); err != nil {
		return document, err
	}

	if document.MaterializationID == "" {
		return document, workflowstate.ErrConflict
	}

	document.SetID = set.ID
	document.Slices, err = workflowNativeOrder(set, tasks)

	return document, err
}

func (s *Store) validateWorkflowOrderSource(document generationcore.PlanOrderDocument, binding workflowstate.PlanRevisionBinding) error {
	if document.Schema != "chr-tracker/v2" || document.ProjectID != s.projectID || document.Namespace != s.layout.Namespace() || document.PlanRef != binding.Reference || document.PlanRevision != binding.Revision || document.PlanBodyHash != binding.BodyHash {
		return workflowstate.ErrConflict
	}

	return generationcore.ValidatePlannedOrder(document.Slices)
}

func workflowNativeOrder(set workflowstate.Record, tasks []workflowstate.Record) ([]generationcore.PlanOrderSlice, error) {
	slices := make([]generationcore.PlanOrderSlice, 0, len(set.Set.Slices))

	indexes := map[string]int{}
	for _, descriptor := range set.Set.Slices {
		indexes[descriptor.ID] = len(slices)
		slices = append(slices, generationcore.PlanOrderSlice{ID: descriptor.ID, Order: descriptor.Order, Title: descriptor.Title})
	}

	for _, record := range tasks {
		task, err := workflowOrderTask(record)
		if err != nil {
			return nil, err
		}

		index, ok := indexes[record.Task.SliceID]
		if !ok {
			return nil, workflowstate.ErrConflict
		}

		slices[index].Tasks = append(slices[index].Tasks, task)
	}

	if err := validateBoundWorkflowOrder(slices); err != nil {
		return nil, err
	}

	return generationcore.OrderedPlanSlices(slices), nil
}

func workflowOrderTask(record workflowstate.Record) (generationcore.PlanOrderTask, error) {
	var id, title string
	if err := json.Unmarshal(record.Domain["plan_task_id"], &id); err != nil {
		return generationcore.PlanOrderTask{}, err
	}

	if err := json.Unmarshal(record.Metadata["title"], &title); err != nil {
		return generationcore.PlanOrderTask{}, err
	}

	return generationcore.PlanOrderTask{ID: id, Order: record.Task.Order, Title: title, TicketKey: record.Task.Key, TicketRef: workflowRecordBodyRef(record)}, nil
}

func validateBoundWorkflowOrder(slices []generationcore.PlanOrderSlice) error {
	planned := generationcore.OrderedPlanSlices(slices)
	for i := range planned {
		for j := range planned[i].Tasks {
			planned[i].Tasks[j].TicketKey, planned[i].Tasks[j].TicketRef = "", ""
		}
	}

	return generationcore.ValidatePlannedOrder(planned)
}

func workflowRecordBodyRef(record workflowstate.Record) string {
	if filepath.Base(record.Path) == "metadata.json" {
		return filepath.ToSlash(filepath.Join(filepath.Dir(record.Path), "artifact.md"))
	}

	return record.Path
}

func (s *Store) workflowPlannedOrder(binding workflowstate.PlanRevisionBinding) (generationcore.PlanOrderDocument, error) {
	body, err := s.readWorkflowTrackerFile(binding.BodyRef)
	if err != nil {
		return generationcore.PlanOrderDocument{}, err
	}

	if contenthash.Sum(body) != binding.BodyHash {
		return generationcore.PlanOrderDocument{}, workflowstate.ErrProjectionStale
	}

	planned, err := s.readWorkflowTrackerFile(filepath.ToSlash(filepath.Join(filepath.Dir(binding.OrderRef), "planned-order.json")))
	if err != nil {
		return generationcore.PlanOrderDocument{}, err
	}

	if contenthash.Sum(planned) != binding.PlannedOrderHash {
		return generationcore.PlanOrderDocument{}, workflowstate.ErrProjectionStale
	}

	var document generationcore.PlanOrderDocument
	if err := authoringstate.Decode(planned, &document); err != nil {
		return document, err
	}

	if err := s.validateWorkflowOrderSource(document, binding); err != nil {
		return document, err
	}

	return document, nil
}

func validateWorkflowPlanRevision(namespace string, set workflowstate.WorkflowSet) error {
	if set.PlanRevision == nil {
		return nil
	}

	binding := *set.PlanRevision
	if !authoringcore.ValidPlanBindingPaths(namespace, binding) || binding.BodyHash != set.PlanHash {
		return workflowstate.ErrConflict
	}

	for _, hash := range []string{binding.OrderHash, binding.TrackerHash, binding.PlannedOrderHash, binding.PlannedTrackerHash} {
		if len(hash) != 64 {
			return workflowstate.ErrConflict
		}

		if _, err := hex.DecodeString(hash); err != nil {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
