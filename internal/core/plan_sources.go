package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// PlanSourceInput binds one selected source to its exact retained image.
type PlanSourceInput struct {
	Kind, Reference, Revision, Content string
}

// PlanSourceObservation is a bounded, read-only generation snapshot. Its hash
// includes absent quiz state and the selected source inventory, so newly added
// discovery invalidates an in-flight generation just like an edited document.
type PlanSourceObservation struct {
	Scope        QuizScope
	Project      ProjectMetadata
	Quiz         QuizObservation
	Inputs       []PlanSourceInput
	Hash         string
	HasDiscovery bool
}

// ObservePlanSources checks two complete observations without acquiring a
// mutation owner, allocating artifacts, rebuilding indexes or recovering writes.
func ObservePlanSources(ctx context.Context, projectDir string, scope QuizScope) (PlanSourceObservation, error) {
	if err := ValidateQuizScope(scope); err != nil {
		return PlanSourceObservation{}, err
	}

	first, err := observePlanSources(ctx, projectDir, scope)
	if err != nil {
		return first, err
	}

	second, err := observePlanSources(ctx, projectDir, scope)
	if err == nil && second.Hash != first.Hash {
		err = ErrWorkflowConflict
	}

	return first, err
}

func observePlanSources(ctx context.Context, projectDir string, scope QuizScope) (PlanSourceObservation, error) {
	dumps, err := ObserveProjectDump(ctx, projectDir)
	if err != nil {
		return PlanSourceObservation{}, err
	}

	if dumps.ReconciliationReason != "" {
		return PlanSourceObservation{}, ErrProjectDumpReconciliation
	}

	observed := PlanSourceObservation{Scope: scope, Project: dumps.Project}

	r := &fileAuthoringSessionRepository{projectDir: projectDir}
	if err := r.appendPlanMetadata(ctx, &observed); err != nil {
		return observed, err
	}

	if err := appendPlanDumps(ctx, projectDir, dumps, &observed); err != nil {
		return observed, err
	}

	if err := appendPlanQuiz(ctx, projectDir, &observed); err != nil {
		return observed, err
	}

	if err := appendPlanSpecs(ctx, projectDir, &observed); err != nil {
		return observed, err
	}

	root, err := filepath.Abs(projectDir)
	if err != nil {
		return observed, err
	}

	encoded, err := json.Marshal(struct {
		Root, Policy string
		Scope        QuizScope
		Inputs       []PlanSourceInput
	}{root, "chr-plan-sources/v1", scope, observed.Inputs})
	if len(encoded) > 96*1024 {
		return observed, errors.New("plan source context exceeds 96 KiB; narrow the selected context")
	}

	observed.Hash = workflowHash(encoded)

	return observed, errors.Join(err, r.checkProjectDumpInspection(ctx))
}

func (r *fileAuthoringSessionRepository) appendPlanMetadata(ctx context.Context, observed *PlanSourceObservation) error {
	paths := []PlanSourceInput{
		{Kind: "agent_protocol", Reference: "AGENTS.md"},
		{Kind: "project_document", Reference: ProjectMetadataPath("")},
	}

	if observed.Scope.CaseSlug != "" {
		if _, err := NewCaseMetadataService(r.projectDir).ReadBounded(ctx, observed.Scope.CaseSlug); err != nil {
			return err
		}

		paths = append(paths, PlanSourceInput{Kind: "use_case_document", Reference: CaseMetadataPath("", observed.Scope.CaseSlug)})
	}

	for _, input := range paths {
		data, err := r.readAuthoringDocument(input.Reference)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if input.Kind == "project_document" && observed.Project.Exists && workflowHash(data) != observed.Project.SourceHash {
			return ErrWorkflowConflict
		}

		input.Reference = filepath.ToSlash(input.Reference)
		input.Revision, input.Content = workflowHash(data), string(data)
		observed.Inputs = append(observed.Inputs, input)
	}

	return nil
}

func appendPlanDumps(ctx context.Context, projectDir string, dumps ProjectDumpObservation, observed *PlanSourceObservation) error {
	references := []string{}
	if dumps.Destination != "" {
		references = append(references, dumps.Destination)
	}

	for _, dump := range dumps.Scoped {
		if observed.Scope.CaseSlug != "" && dump.CaseSlug == observed.Scope.CaseSlug {
			references = append(references, dump.Reference)
		}
	}

	for _, reference := range references {
		document, err := ReadDumpDocument(ctx, projectDir, reference)
		if err != nil {
			return err
		}

		observed.Inputs = append(observed.Inputs, PlanSourceInput{"dump", reference, document.Revision, document.Body})
		observed.HasDiscovery = observed.HasDiscovery || usablePlanDiscovery(document.Body)
	}

	return nil
}

func usablePlanDiscovery(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "<!--") {
			continue
		}

		switch strings.ToLower(strings.Trim(line, "-* .!\t")) {
		case "", "todo", "tbd", "pending", "placeholder", "no context yet":
			continue
		default:
			return true
		}
	}

	return false
}
