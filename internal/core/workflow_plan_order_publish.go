package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) materializedOrderChanges(result *WorkflowMaterializationResult, plan generationcore.GeneratedWorkflowPlan) ([]workflowstate.FileChange, error) {
	if plan.PlanRevision == nil {
		return nil, nil
	}

	previous := *plan.PlanRevision
	binding := previous
	result.Set.Set.PlanRevision = &binding

	document, err := s.workflowOrderDocument(result.Set, result.Tasks)
	if err != nil {
		return nil, err
	}

	images, err := authoringcore.EncodePlanOrder(document)
	if err != nil {
		return nil, err
	}

	order, tracker := images.Order, images.Tracker
	binding.OrderHash, binding.TrackerHash = contenthash.Sum(order), contenthash.Sum(tracker)

	changes, err := s.retainedBoundOrderChanges(previous, binding, order, tracker)
	if err != nil {
		return nil, err
	}

	registration := workflowBoundTrackerStatus(s, result.Set, binding, contenthash.Sum(order))

	registered, err := s.boundTrackerChanges(registration, tracker)
	if err != nil {
		return nil, err
	}

	current, err := s.boundCurrentPlanChanges(previous, binding, tracker)
	if err != nil {
		return nil, err
	}

	return append(append(changes, registered...), current...), nil
}

func workflowBoundTrackerStatus(s *WorkflowStore, set WorkflowRecord, binding workflowstate.PlanRevisionBinding, sourceHash string) WorkflowTrackerStatus {
	return WorkflowTrackerStatus{Schema: "chr-tracker/v2", ProjectID: s.projectID, Namespace: s.layout.Namespace(), SetID: set.ID, Path: "tracker/" + set.ID + ".md", PlanRevision: &binding, SourceHash: sourceHash, OutputHash: binding.TrackerHash, ObservedHash: binding.TrackerHash, GeneratedAt: time.Now().UTC(), Status: "current", Sources: []WorkflowTrackerSource{{ID: set.ID, Path: set.Path, Hash: sourceHash}, {Path: binding.BodyRef, Hash: binding.BodyHash}}}
}

func (s *WorkflowStore) retainedBoundOrderChanges(previous, next workflowstate.PlanRevisionBinding, order, tracker []byte) ([]workflowstate.FileChange, error) {
	data, err := json.Marshal(next)
	if err != nil {
		return nil, err
	}

	bindingRef := filepath.ToSlash(filepath.Join(filepath.Dir(next.OrderRef), "binding.json"))

	current, err := s.readPlanOrderBinding(previous)
	if err != nil || current != previous {
		return nil, errors.Join(workflowstate.ErrConflict, err)
	}

	currentData, err := s.readWorkflowTrackerFile(bindingRef)
	if err != nil {
		return nil, err
	}

	images := []struct {
		ref, expected string
		body          []byte
	}{{next.OrderRef, previous.OrderHash, order}, {next.TrackerRef, previous.TrackerHash, tracker}, {bindingRef, contenthash.Sum(currentData), data}}

	changes := make([]workflowstate.FileChange, 0, len(images))
	for _, image := range images {
		change, err := s.orderImageChange(image.ref, image.body, image.expected, "")
		if err != nil {
			return nil, err
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func (s *WorkflowStore) boundTrackerChanges(status WorkflowTrackerStatus, tracker []byte) ([]workflowstate.FileChange, error) {
	prior, err := s.workflowTrackerPrevious(status.SetID)
	if err != nil {
		return nil, err
	}

	output, err := s.orderImageChange(status.Path, tracker, prior.OutputHash, "")
	if err != nil {
		return nil, err
	}

	data, err := workflowTrackerMetadataBytes(status)
	if err != nil {
		return nil, err
	}

	metadataRef, err := workflowstate.TrackerMetadataPath(status.SetID)
	if err != nil {
		return nil, err
	}

	metadata, err := s.observedOrderImage(metadataRef, data)
	if err != nil {
		return nil, err
	}

	retainedRef := filepath.ToSlash(filepath.Join(filepath.Dir(status.PlanRevision.OrderRef), "projection.json"))

	retained, err := s.orderImageChange(retainedRef, data, "", "")
	if err != nil {
		return nil, err
	}

	return []workflowstate.FileChange{output, metadata, retained}, nil
}

func (s *WorkflowStore) orderImageChange(reference string, body []byte, expected, replace string) (workflowstate.FileChange, error) {
	change, err := s.observedOrderImage(reference, body)
	if err != nil {
		return change, err
	}

	if change.Before != "" && change.Before != expected && change.Before != replace {
		return change, fmt.Errorf("%w: edited order image %s requires its observed hash", workflowstate.ErrProjectionStale, reference)
	}

	return change, nil
}

func (s *WorkflowStore) observedOrderImage(reference string, body []byte) (workflowstate.FileChange, error) {
	change := workflowstate.FileChange{Path: reference, After: body}

	current, err := s.readWorkflowTrackerFile(reference)
	if errors.Is(err, os.ErrNotExist) {
		return change, nil
	}

	if err != nil {
		return change, err
	}

	change.Before = contenthash.Sum(current)

	return change, nil
}

func (s *WorkflowStore) currentPlanOrderMetadata(binding workflowstate.PlanRevisionBinding) (map[string]any, workflowstate.FileChange, bool, error) {
	reference := filepath.ToSlash(filepath.Join(filepath.Dir(binding.Reference), "metadata.json"))

	data, err := s.readWorkflowTrackerFile(reference)
	if err != nil {
		return nil, workflowstate.FileChange{}, false, err
	}

	var metadata map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, workflowstate.FileChange{}, false, err
	}

	if err := authoringcore.ValidatePlanTargetMetadata(metadata, filepath.ToSlash(filepath.Join(s.layout.RelativeRoot(), binding.Reference))); err != nil {
		return nil, workflowstate.FileChange{}, false, err
	}

	encoded, err := json.Marshal(metadata["plan_binding"])
	if err != nil {
		return nil, workflowstate.FileChange{}, false, err
	}

	var current workflowstate.PlanRevisionBinding
	if err := json.Unmarshal(encoded, &current); err != nil {
		return nil, workflowstate.FileChange{}, false, err
	}

	same := current.Reference == binding.Reference && current.Revision == binding.Revision && current.BodyHash == binding.BodyHash

	if same && current != binding {
		return nil, workflowstate.FileChange{}, false, workflowstate.ErrConflict
	}

	return metadata, workflowstate.FileChange{Path: reference, Before: contenthash.Sum(data)}, same, nil
}

func (s *WorkflowStore) checkCurrentOrderOutput(binding workflowstate.PlanRevisionBinding) error {
	_, _, same, err := s.currentPlanOrderMetadata(binding)
	if err != nil || !same {
		return err
	}

	body, err := s.readWorkflowTrackerFile(binding.Reference)
	if err != nil {
		return err
	}

	if contenthash.Sum(body) != binding.BodyHash {
		return workflowstate.ErrConflict
	}

	tracker, err := s.readWorkflowTrackerFile(binding.CurrentTrackerRef)
	if err != nil {
		return err
	}

	if contenthash.Sum(tracker) != binding.TrackerHash {
		return workflowstate.ErrProjectionStale
	}

	return nil
}

func (s *WorkflowStore) boundCurrentPlanChanges(previous, next workflowstate.PlanRevisionBinding, tracker []byte) ([]workflowstate.FileChange, error) {
	metadata, planImage, same, err := s.currentPlanOrderMetadata(previous)
	if err != nil {
		return nil, err
	}

	if same {
		metadata["plan_binding"] = next
	}

	if err := s.registerBoundProjection(metadata, next); err != nil {
		return nil, err
	}

	if err := authoringcore.AdvanceArtifactMetadata(metadata); err != nil {
		return nil, err
	}

	planImage.After, err = json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	changes := []workflowstate.FileChange{planImage}
	if !same {
		return changes, nil
	}

	trackerChanges, err := s.boundCurrentTrackerChanges(previous, next, tracker)

	return append(changes, trackerChanges...), err
}

func (s *WorkflowStore) boundCurrentTrackerChanges(previous, next workflowstate.PlanRevisionBinding, tracker []byte) ([]workflowstate.FileChange, error) {
	output, err := s.orderImageChange(next.CurrentTrackerRef, tracker, previous.TrackerHash, "")
	if err != nil {
		return nil, err
	}

	metadataRef := filepath.ToSlash(filepath.Join(filepath.Dir(next.CurrentTrackerRef), "metadata.json"))

	current, err := s.readWorkflowTrackerFile(metadataRef)
	if err != nil {
		return nil, err
	}

	var envelope map[string]any
	if err := json.Unmarshal(current, &envelope); err != nil {
		return nil, err
	}

	if err := authoringcore.ValidatePlanTrackerMetadata(envelope, filepath.ToSlash(filepath.Join(s.layout.RelativeRoot(), metadataRef))); err != nil {
		return nil, err
	}

	if err := authoringcore.AdvanceArtifactMetadata(envelope); err != nil {
		return nil, err
	}

	envelope["plan_binding"] = next

	encoded, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	return []workflowstate.FileChange{output, {Path: metadataRef, Before: contenthash.Sum(current), After: encoded}}, nil
}

func (s *WorkflowStore) registerBoundProjection(metadata map[string]any, binding workflowstate.PlanRevisionBinding) error {
	reference := filepath.ToSlash(filepath.Join(filepath.Dir(binding.OrderRef), "projection.json"))

	relative, err := filepath.Rel(filepath.Dir(binding.Reference), reference)
	if err != nil {
		return err
	}

	files, err := authoringcore.PlanManifestFiles(metadata["files"])
	if err != nil {
		return err
	}

	for _, file := range files {
		if file == filepath.ToSlash(relative) {
			return nil
		}
	}

	metadata["files"] = append(files, filepath.ToSlash(relative))

	return nil
}
