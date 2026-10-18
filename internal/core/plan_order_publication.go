package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func preparePlanOrderChanges(projectDir string, paths GeneratedPlanPaths, target planTargetObservation, metadata map[string]any, plan GeneratedPlan, body []byte) ([]workflowFileChange, error) {
	if !plan.RequireOrderMap && len(plan.Order) == 0 {
		if metadata["plan_binding"] != nil {
			return nil, fmt.Errorf("replacement requires its revision-bound order map")
		}

		return nil, nil
	}

	if err := ValidatePlannedOrder(plan.Order); err != nil {
		return nil, err
	}

	binding, document, err := newPlanRevision(paths, metadata, plan, body)
	if err != nil {
		return nil, err
	}

	encoded, tracker, err := planOrderBytes(document)
	if err != nil {
		return nil, err
	}

	binding.OrderHash, binding.TrackerHash = workflowHash(encoded), workflowHash(tracker)
	binding.PlannedOrderHash, binding.PlannedTrackerHash = binding.OrderHash, binding.TrackerHash
	r := &fileAuthoringSessionRepository{projectDir: projectDir}

	changes, err := r.planRetainedImages(document.Namespace, binding, body, encoded, tracker)
	if err != nil {
		return nil, err
	}

	current, err := r.planCurrentTrackerChanges(document.Namespace, target, binding, plan, tracker)
	if err != nil {
		return nil, err
	}

	metadata["plan_binding"] = binding
	metadata["plan_revision"] = binding.Revision

	if err := appendPlanRevisionManifest(paths, metadata, changes); err != nil {
		return nil, err
	}

	return append(changes, current...), nil
}

func newPlanRevision(paths GeneratedPlanPaths, metadata map[string]any, plan GeneratedPlan, body []byte) (PlanRevisionBinding, PlanOrderDocument, error) {
	revision := uint64(1)

	if value, ok := metadata["plan_revision"]; ok {
		number, ok := value.(float64)
		if !ok || number < 1 || number >= 4096 || number != float64(uint64(number)) {
			return PlanRevisionBinding{}, PlanOrderDocument{}, fmt.Errorf("invalid plan revision")
		}

		revision = uint64(number) + 1
	}

	parts := strings.Split(paths.RelativePath, "/")
	reference := strings.Join(parts[2:], "/")
	hash := workflowHash(body)
	retained := filepath.ToSlash(filepath.Join(filepath.Dir(reference), "files", "plan-revisions", fmt.Sprintf("%d-%s", revision, hash)))
	binding := PlanRevisionBinding{Reference: reference, Revision: revision, BodyHash: hash, BodyRef: retained + "/plan.md", OrderRef: retained + "/order.json", TrackerRef: retained + "/tracker.md", CurrentTrackerRef: "tracker/00000-" + paths.ScopeID + "-tracking/artifact.md"}
	document := PlanOrderDocument{Schema: "chr-tracker/v2", ProjectID: plan.Scope.ProjectID, Namespace: parts[1], PlanRef: reference, PlanRevision: revision, PlanBodyHash: hash, SourceObservationHash: plan.SourceObservationHash, SourceInputs: plan.SourceInputs, Slices: orderedPlanSlices(plan.Order)}

	return binding, document, nil
}

func (r *fileAuthoringSessionRepository) planRetainedImages(namespace string, binding PlanRevisionBinding, body, order, tracker []byte) ([]workflowFileChange, error) {
	var changes []workflowFileChange

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: r.projectDir, Namespace: namespace})
	if err != nil {
		return nil, err
	}

	root := layout.RelativeRoot()

	bindingBytes, err := json.Marshal(binding)
	if err != nil {
		return nil, err
	}

	rootRef := filepath.Dir(binding.OrderRef)
	refs := []string{binding.BodyRef, binding.OrderRef, binding.TrackerRef, filepath.ToSlash(filepath.Join(rootRef, "planned-order.json")), filepath.ToSlash(filepath.Join(rootRef, "planned-tracker.md")), filepath.ToSlash(filepath.Join(rootRef, "binding.json")), filepath.ToSlash(filepath.Join(rootRef, "planned-binding.json"))}
	images := [][]byte{body, order, tracker, order, tracker, bindingBytes, bindingBytes}

	for i, reference := range refs {
		path := filepath.ToSlash(filepath.Join(root, reference))
		if _, err := r.authoringRead(path); !errors.Is(err, os.ErrNotExist) {
			return nil, errors.Join(ErrWorkflowConflict, err)
		}

		changes = append(changes, workflowFileChange{Path: path, After: images[i]})
	}

	return changes, nil
}

func (r *fileAuthoringSessionRepository) planCurrentTrackerChanges(namespace string, target planTargetObservation, binding PlanRevisionBinding, plan GeneratedPlan, tracker []byte) ([]workflowFileChange, error) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: r.projectDir, Namespace: namespace})
	if err != nil {
		return nil, err
	}

	reference := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), binding.CurrentTrackerRef))
	images := []workflowFileChange{}

	for _, path := range []string{reference, filepath.ToSlash(filepath.Join(filepath.Dir(reference), "metadata.json"))} {
		image, err := r.quizImage(path)
		if err != nil {
			return nil, err
		}

		images = append(images, image)
	}

	if (images[0].After == nil) != (images[1].After == nil) {
		return nil, ErrWorkflowConflict
	}

	if err := checkPreviousPlanTracker(target, images[0], plan.TrackerReplaceHash); err != nil {
		return nil, err
	}

	metadata, err := planCurrentTrackerMetadata(layout, reference, plan, images[1].After)
	if err != nil {
		return nil, err
	}

	metadata["plan_binding"], metadata["workflow_projection"] = binding, "chr-tracker/v2"

	encoded, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	images[0].After, images[1].After = tracker, encoded

	return images, nil
}

func checkPreviousPlanTracker(target planTargetObservation, image workflowFileChange, replace string) error {
	if image.After == nil {
		return nil
	}

	if replace != "" && image.Before == replace {
		return nil
	}

	var previous PlanRevisionBinding

	encoded, err := json.Marshal(target.metadata["plan_binding"])
	if err != nil {
		return err
	}

	if err := json.Unmarshal(encoded, &previous); err != nil {
		return err
	}

	if previous.TrackerHash == "" || previous.TrackerHash != image.Before {
		return fmt.Errorf("%w: tracker edited; replacement requires its observed hash", ErrWorkflowProjectionStale)
	}

	return nil
}

func validPlanRevisionJournalPath(reference string) bool {
	parts := strings.Split(filepath.ToSlash(reference), "/")
	if !validPlanRevisionDirectory(parts) || !validPlanRevisionKey(parts[6]) {
		return false
	}

	switch parts[7] {
	case "plan.md", "order.json", "tracker.md", "planned-order.json", "planned-tracker.md", "binding.json", "planned-binding.json", "projection.json":
		return true
	default:
		return false
	}
}

func validPlanTrackerJournalPath(reference string) bool {
	parts := strings.Split(filepath.ToSlash(reference), "/")

	return len(parts) == 5 && parts[0] == "ops" && IsValidOpsNamespace(parts[1]) && parts[2] == "tracker" && opsArtifactIDPattern.MatchString(parts[3]) && (parts[4] == "artifact.md" || parts[4] == "metadata.json")
}

func appendPlanRevisionManifest(paths GeneratedPlanPaths, metadata map[string]any, changes []workflowFileChange) error {
	files, err := planManifestFiles(metadata["files"])
	if err != nil {
		return err
	}

	root := filepath.Dir(paths.RelativePath)
	for _, change := range changes {
		relative, err := filepath.Rel(root, change.Path)
		if err != nil {
			return err
		}

		files = append(files, filepath.ToSlash(relative))
	}

	metadata["files"] = files

	return nil
}

func planCurrentTrackerMetadata(layout OpsLayout, reference string, plan GeneratedPlan, previous []byte) (map[string]any, error) {
	parts := strings.Split(reference, "/")
	if previous == nil {
		return newOpsArtifactMetadata(CreateOpsArtifactOptions{Layout: layout, Type: "tracker", Title: plan.Title + " tracking", Producer: "chr"}, parts[3], plan.CreatedAt), nil
	}

	var metadata map[string]any
	if err := decodeAuthoringState(previous, &metadata); err != nil {
		return nil, err
	}

	if err := validatePlanTrackerMetadata(metadata, reference); err != nil {
		return nil, err
	}

	if err := advanceAuthoringMetadata(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func validPlanRevisionDirectory(parts []string) bool {
	return len(parts) == 8 && parts[0] == "ops" && IsValidOpsNamespace(parts[1]) && parts[2] == "plan" && opsArtifactIDPattern.MatchString(parts[3]) && parts[4] == "files" && parts[5] == "plan-revisions"
}

func validPlanRevisionKey(value string) bool {
	key := strings.SplitN(value, "-", 2)
	if len(key) != 2 || len(key[1]) != 64 {
		return false
	}

	revision, err := strconv.ParseUint(key[0], 10, 64)
	if err != nil || revision < 1 || revision > 4096 {
		return false
	}

	_, err = hex.DecodeString(key[1])

	return err == nil
}

func planManifestFiles(value any) ([]any, error) {
	switch files := value.(type) {
	case nil:
		return nil, nil
	case []any:
		return files, nil
	case []string:
		result := make([]any, 0, len(files))
		for _, file := range files {
			result = append(result, file)
		}

		return result, nil
	default:
		return nil, ErrWorkflowConflict
	}
}

func validatePlanTrackerMetadata(metadata map[string]any, reference string) error {
	parts := strings.Split(reference, "/")
	if metadata["schema_version"] != "ops-layout/v1" || metadata["type"] != "tracker" || metadata["namespace"] != parts[1] || metadata["id"] != parts[3] {
		return ErrWorkflowConflict
	}

	return nil
}
