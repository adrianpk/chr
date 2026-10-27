package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type planTargetObservation struct {
	revision string
	images   []workflowstate.FileChange
	metadata map[string]any
}

// PlanTargetRevision binds a plan operation to the current body and envelope, or
// their complete absence. Reading a pending or partially present target fails.
func PlanTargetRevision(ctx context.Context, projectDir string, scope quizcore.Scope) (string, error) {
	paths, err := GeneratedPlanLayout(projectDir, scope)
	if err != nil {
		return "", err
	}

	target, err := observePlanTarget(ctx, projectDir, paths.RelativePath)

	return target.revision, err
}

func observePlanTarget(ctx context.Context, projectDir, reference string) (planTargetObservation, error) {
	r := &fileSessionRepository{projectDir: projectDir}

	target := planTargetObservation{}
	if err := r.checkPlanWorkflowPending(reference); err != nil {
		return target, err
	}

	if err := r.checkProjectDumpInspection(ctx); err != nil {
		return target, err
	}

	for _, path := range []string{reference, filepath.ToSlash(filepath.Join(filepath.Dir(reference), "metadata.json"))} {
		image, err := r.quizImage(path)
		if err != nil {
			return target, err
		}

		target.images = append(target.images, image)
	}

	if (target.images[0].After == nil) != (target.images[1].After == nil) {
		return target, workflowstate.ErrConflict
	}

	if target.images[1].After != nil {
		if err := authoringstate.Decode(target.images[1].After, &target.metadata); err != nil {
			return target, err
		}

		if err := ValidatePlanTargetMetadata(target.metadata, reference); err != nil {
			return target, err
		}
	}

	root, err := filepath.Abs(projectDir)
	if err != nil {
		return target, err
	}

	target.revision = contenthash.Sum([]byte(strings.Join([]string{root, reference, target.images[0].Before, target.images[1].Before}, "\x00")))

	return target, r.checkQuizImages(target.images)
}

func (r *fileSessionRepository) publishGeneratedPlan(ctx context.Context, paths GeneratedPlanPaths, plan generationcore.GeneratedPlan) error {
	projectDir := r.projectDir

	if found, err := r.completedPlanOperation(plan); err != nil || found {
		return err
	}

	target, err := observePlanTarget(ctx, projectDir, paths.RelativePath)
	if err != nil {
		return err
	}

	if err := validatePlanPublicationTarget(target, plan); err != nil {
		return err
	}

	changes, err := generatedPlanChanges(projectDir, paths, target, plan)
	if err != nil {
		return err
	}

	receipt, err := r.planPublicationReceiptChange(paths, plan, changes)
	if err != nil {
		return err
	}

	changes = append(changes, receipt...)

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, changes); err != nil {
		return err
	}

	return r.recoverAuthoringMutation()
}

func generatedPlanChanges(projectDir string, paths GeneratedPlanPaths, target planTargetObservation, plan generationcore.GeneratedPlan) ([]workflowstate.FileChange, error) {
	metadata := target.metadata
	if metadata == nil {
		var err error

		metadata, err = generatedPlanMetadata(projectDir, paths.RelativePath, plan.Title, plan)
		if err != nil {
			return nil, err
		}
	} else if err := AdvanceArtifactMetadata(metadata); err != nil {
		return nil, err
	}

	metadata["title"], metadata["status"] = plan.Title, "planned"
	metadata["source_observation_hash"], metadata["source_inputs"] = plan.SourceObservationHash, plan.SourceInputs
	metadata["updated_at"] = plan.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")

	body := []byte(RenderGeneratedPlan(plan))

	companions, err := preparePlanOrderChanges(projectDir, target, planRevisionInput{Paths: paths, Metadata: metadata, Plan: plan, Body: body})
	if err != nil {
		return nil, err
	}

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, err
	}

	changes := append([]workflowstate.FileChange(nil), target.images...)

	changes[0].After, changes[1].After = body, append(encoded, '\n')
	if target.images[0].After != nil && !plan.ArchiveAt.IsZero() {
		archive, err := generatedPlanArchiveChanges(projectDir, paths, target, plan)
		if err != nil {
			return nil, err
		}

		changes = append(archive, changes...)
	}

	return append(companions, changes...), nil
}

func generatedPlanMetadata(projectDir, reference, title string, plan generationcore.GeneratedPlan) (map[string]any, error) {
	parts := strings.Split(reference, "/")

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir, Namespace: parts[1]})
	if err != nil {
		return nil, err
	}

	return opscore.NewArtifactMetadata(opscore.ArtifactOptions{Layout: layout, Type: "plan", Title: title, Producer: "chr"}, parts[3], plan.CreatedAt), nil
}

func generatedPlanArchiveChanges(projectDir string, current GeneratedPlanPaths, target planTargetObservation, plan generationcore.GeneratedPlan) ([]workflowstate.FileChange, error) {
	archive, err := ArchivedPlanLayout(projectDir, plan.Scope, plan.ArchiveAt)
	if err != nil {
		return nil, err
	}

	r := &fileSessionRepository{projectDir: projectDir}
	for _, reference := range []string{archive.RelativePath, filepath.Join(filepath.Dir(archive.RelativePath), "metadata.json")} {
		if _, err := r.authoringRead(reference); !errors.Is(err, os.ErrNotExist) {
			return nil, errors.Join(workflowstate.ErrConflict, err)
		}
	}

	var metadata map[string]any
	if err := authoringstate.Decode(target.images[1].After, &metadata); err != nil {
		return nil, err
	}

	metadata["id"] = strings.Split(archive.RelativePath, "/")[3]
	metadata["status"], metadata["archived_from"] = "archived", current.RelativePath

	metadata["source_metadata_hash"] = target.images[1].Before
	if metadata["plan_binding"] != nil {
		// Retained files stay below the canonical plan. The archive keeps their
		// source manifest and explicit namespace-relative binding, not local copies.
		metadata["source_files"], metadata["files"] = metadata["files"], []string{"artifact.md"}
	}

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, err
	}

	return []workflowstate.FileChange{
		{Path: archive.RelativePath, After: []byte(RenderArchivedPlan(string(target.images[0].After), current.RelativePath, plan.ArchiveAt))},
		{Path: filepath.ToSlash(filepath.Join(filepath.Dir(archive.RelativePath), "metadata.json")), After: append(encoded, '\n')},
	}, nil
}

// ValidatePlanTargetMetadata checks the existing canonical plan envelope identity.
func ValidatePlanTargetMetadata(metadata map[string]any, reference string) error {
	parts := strings.Split(reference, "/")
	if metadata["schema_version"] != "ops-layout/v1" || metadata["type"] != "plan" || metadata["namespace"] != parts[1] || metadata["id"] != parts[3] {
		return workflowstate.ErrConflict
	}

	return nil
}

func validatePlanTargetScope(target planTargetObservation, scope quizcore.Scope) error {
	parsedMarkdown, err := workflowstate.RequireMarkdown(target.images[0].After)
	front := parsedMarkdown.Front

	if err != nil {
		return err
	}

	fields, err := ProjectDocumentFields(front)
	if err != nil {
		return err
	}

	declared := fields["case_slug"]
	if declared == "" {
		declared = fields["case"]
	}

	if declared != scope.CaseSlug || fields["project_id"] != "" && fields["project_id"] != scope.ProjectID {
		return workflowstate.ErrConflict
	}

	return nil
}

func validatePlanPublicationTarget(target planTargetObservation, plan generationcore.GeneratedPlan) error {
	if plan.CreateOnly && target.metadata != nil || plan.ExpectedRevision != "" && target.revision != plan.ExpectedRevision {
		return workflowstate.ErrConflict
	}

	if target.metadata != nil {
		return validatePlanTargetScope(target, plan.Scope)
	}

	return nil
}
