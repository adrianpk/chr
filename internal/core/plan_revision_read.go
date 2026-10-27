package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ObservePlanRevision reads the current retained order binding without repairing
// or rewriting artifacts. Missing or edited companions require explicit repair.
// A legacy or absent plan returns an empty binding without migration.
func ObservePlanRevision(ctx context.Context, projectDir string, scope quizcore.Scope) (generationcore.PlanRevisionBinding, error) {
	paths, err := GeneratedPlanLayout(projectDir, scope)
	if err != nil {
		return generationcore.PlanRevisionBinding{}, err
	}

	target, err := observePlanTarget(ctx, projectDir, paths.RelativePath)
	if err != nil {
		return generationcore.PlanRevisionBinding{}, err
	}

	var binding generationcore.PlanRevisionBinding
	if target.metadata["plan_binding"] == nil {
		return binding, nil
	}

	encoded, err := json.Marshal(target.metadata["plan_binding"])
	if err != nil {
		return binding, err
	}

	if err := json.Unmarshal(encoded, &binding); err != nil {
		return binding, err
	}

	r := &fileAuthoringSessionRepository{projectDir: projectDir}
	namespace := strings.Split(paths.RelativePath, "/")[1]

	if binding.BodyHash != target.images[0].Before {
		return binding, ErrWorkflowProjectionStale
	}

	if err := r.validatePlanRevisionImages(namespace, binding); err != nil {
		return binding, err
	}

	return binding, r.checkProjectDumpInspection(ctx)
}

func (r *fileAuthoringSessionRepository) validatePlanReceiptResult(scope quizcore.Scope, result PlanPublicationResult) error {
	paths, err := GeneratedPlanLayout(r.projectDir, scope)
	if err != nil {
		return err
	}

	if result.Reference != paths.RelativePath {
		return workflowstate.ErrConflict
	}

	namespace := strings.Split(result.Reference, "/")[1]

	binding := result.Binding
	if !validPlanBindingPaths(namespace, binding) {
		return workflowstate.ErrConflict
	}

	root := filepath.Join(OpsDirectoryName, namespace)
	images := map[string]string{binding.BodyRef: binding.BodyHash, filepath.ToSlash(filepath.Join(filepath.Dir(binding.OrderRef), "planned-order.json")): binding.PlannedOrderHash, filepath.ToSlash(filepath.Join(filepath.Dir(binding.TrackerRef), "planned-tracker.md")): binding.PlannedTrackerHash}

	return r.checkPlanImages(root, images)
}

func (r *fileAuthoringSessionRepository) validatePlanRevisionImages(namespace string, binding generationcore.PlanRevisionBinding) error {
	if !validPlanBindingPaths(namespace, binding) {
		return workflowstate.ErrConflict
	}

	root := filepath.Join(OpsDirectoryName, namespace)
	images := map[string]string{binding.BodyRef: binding.BodyHash, binding.OrderRef: binding.OrderHash, binding.TrackerRef: binding.TrackerHash, binding.CurrentTrackerRef: binding.TrackerHash}

	return r.checkPlanImages(root, images)
}

func (r *fileAuthoringSessionRepository) checkPlanImages(root string, images map[string]string) error {
	for reference, hash := range images {
		data, err := r.authoringRead(filepath.ToSlash(filepath.Join(root, reference)))
		if err != nil {
			return fmt.Errorf("plan companion unavailable: %w", err)
		}

		if hash == "" || contenthash.Sum(data) != hash {
			return ErrWorkflowProjectionStale
		}
	}

	return nil
}

func validPlanBindingPaths(namespace string, binding generationcore.PlanRevisionBinding) bool {
	parts := strings.Split(binding.Reference, "/")
	if len(parts) != 3 || parts[0] != "plan" || !opsArtifactIDPattern.MatchString(parts[1]) || parts[2] != "artifact.md" {
		return false
	}

	retained := filepath.ToSlash(filepath.Join(filepath.Dir(binding.Reference), "files", "plan-revisions", fmt.Sprintf("%d-%s", binding.Revision, binding.BodyHash)))
	if binding.BodyRef != retained+"/plan.md" || binding.OrderRef != retained+"/order.json" || binding.TrackerRef != retained+"/tracker.md" {
		return false
	}

	root := filepath.ToSlash(filepath.Join(OpsDirectoryName, namespace)) + "/"

	return validPlanRevisionJournalPath(root+binding.BodyRef) && validPlanTrackerJournalPath(root+binding.CurrentTrackerRef)
}
