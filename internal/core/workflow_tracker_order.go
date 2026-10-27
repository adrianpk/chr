package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validWorkflowTrackerSchema(schema string) bool {
	return schema == "chr-tracker/v1" || schema == "chr-tracker/v2"
}

func workflowTrackerSchema(snapshot WorkflowSnapshot, setID, requested, previous string) string {
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

func (s *WorkflowStore) versionedWorkflowTrackerImage(snapshot WorkflowSnapshot, setID, schema string) (WorkflowTrackerStatus, []byte, error) {
	switch schema {
	case "chr-tracker/v1":
		return s.workflowTrackerImage(snapshot, setID)
	case "chr-tracker/v2":
		return s.workflowOrderTrackerImage(snapshot, setID)
	default:
		return WorkflowTrackerStatus{}, nil, fmt.Errorf("unsupported tracker schema")
	}
}

func (s *WorkflowStore) workflowOrderTrackerImage(snapshot WorkflowSnapshot, setID string) (WorkflowTrackerStatus, []byte, error) {
	set, tasks, err := workflowTrackerMembers(snapshot, setID)
	if err != nil {
		return WorkflowTrackerStatus{}, nil, err
	}

	result := WorkflowTrackerStatus{Schema: "chr-tracker/v2", ProjectID: s.projectID, Namespace: s.layout.Namespace(), SetID: setID, Path: "tracker/" + setID + ".md", PlanRevision: set.Set.PlanRevision}

	document, err := s.workflowOrderDocument(set, tasks)
	if err != nil {
		return result, nil, err
	}

	encoded, output, err := planOrderBytes(document)
	if err != nil {
		return result, nil, err
	}

	result.SourceHash = contenthash.Sum(encoded)
	result.Sources = []WorkflowTrackerSource{{ID: set.ID, Path: set.Path, Hash: result.SourceHash}, {Path: set.Set.PlanRevision.BodyRef, Hash: set.Set.PlanRevision.BodyHash}}

	return result, output, nil
}

func (s *WorkflowStore) workflowOrderDocument(set WorkflowRecord, tasks []WorkflowRecord) (PlanOrderDocument, error) {
	binding := set.Set.PlanRevision
	if binding == nil || !validPlanBindingPaths(s.layout.Namespace(), *binding) || binding.BodyHash != set.Set.PlanHash {
		return PlanOrderDocument{}, fmt.Errorf("order-only tracker requires its retained plan revision")
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

func (s *WorkflowStore) validateWorkflowOrderSource(document PlanOrderDocument, binding PlanRevisionBinding) error {
	if document.Schema != "chr-tracker/v2" || document.ProjectID != s.projectID || document.Namespace != s.layout.Namespace() || document.PlanRef != binding.Reference || document.PlanRevision != binding.Revision || document.PlanBodyHash != binding.BodyHash {
		return workflowstate.ErrConflict
	}

	return ValidatePlannedOrder(document.Slices)
}

func workflowNativeOrder(set WorkflowRecord, tasks []WorkflowRecord) ([]PlanOrderSlice, error) {
	slices := make([]PlanOrderSlice, 0, len(set.Set.Slices))

	indexes := map[string]int{}
	for _, descriptor := range set.Set.Slices {
		indexes[descriptor.ID] = len(slices)
		slices = append(slices, PlanOrderSlice{ID: descriptor.ID, Order: descriptor.Order, Title: descriptor.Title})
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

	return orderedPlanSlices(slices), nil
}

func workflowOrderTask(record WorkflowRecord) (PlanOrderTask, error) {
	var id, title string
	if err := json.Unmarshal(record.Domain["plan_task_id"], &id); err != nil {
		return PlanOrderTask{}, err
	}

	if err := json.Unmarshal(record.Metadata["title"], &title); err != nil {
		return PlanOrderTask{}, err
	}

	return PlanOrderTask{ID: id, Order: record.Task.Order, Title: title, TicketKey: record.Task.Key, TicketRef: workflowRecordBodyRef(record)}, nil
}

func validateBoundWorkflowOrder(slices []PlanOrderSlice) error {
	planned := orderedPlanSlices(slices)
	for i := range planned {
		for j := range planned[i].Tasks {
			planned[i].Tasks[j].TicketKey, planned[i].Tasks[j].TicketRef = "", ""
		}
	}

	return ValidatePlannedOrder(planned)
}

func workflowRecordBodyRef(record WorkflowRecord) string {
	if filepath.Base(record.Path) == "metadata.json" {
		return filepath.ToSlash(filepath.Join(filepath.Dir(record.Path), "artifact.md"))
	}

	return record.Path
}

func (s *WorkflowStore) workflowPlannedOrder(binding PlanRevisionBinding) (PlanOrderDocument, error) {
	body, err := s.readWorkflowTrackerFile(binding.BodyRef)
	if err != nil {
		return PlanOrderDocument{}, err
	}

	if contenthash.Sum(body) != binding.BodyHash {
		return PlanOrderDocument{}, ErrWorkflowProjectionStale
	}

	planned, err := s.readWorkflowTrackerFile(filepath.ToSlash(filepath.Join(filepath.Dir(binding.OrderRef), "planned-order.json")))
	if err != nil {
		return PlanOrderDocument{}, err
	}

	if contenthash.Sum(planned) != binding.PlannedOrderHash {
		return PlanOrderDocument{}, ErrWorkflowProjectionStale
	}

	var document PlanOrderDocument
	if err := decodeAuthoringState(planned, &document); err != nil {
		return document, err
	}

	if err := s.validateWorkflowOrderSource(document, binding); err != nil {
		return document, err
	}

	return document, nil
}

func validateWorkflowPlanRevision(namespace string, set WorkflowSet) error {
	if set.PlanRevision == nil {
		return nil
	}

	binding := *set.PlanRevision
	if !validPlanBindingPaths(namespace, binding) || binding.BodyHash != set.PlanHash {
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
