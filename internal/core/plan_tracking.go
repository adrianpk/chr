package core

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// GeneratedPlanTrackingPaths describes a package boundary value.
type GeneratedPlanTrackingPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// GeneratedPlanTrackingLayout executes the corresponding package operation.
func GeneratedPlanTrackingLayout(projectDir string, scope quizcore.Scope) (GeneratedPlanTrackingPaths, error) {
	err := quizcore.ValidateScope(scope)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	body, err := opscore.NamedArtifactBody(projectDir, "tracker", scopeID+"-tracking")
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	return GeneratedPlanTrackingPaths{
		ScopeID:      scopeID,
		RelativePath: body.RelativePath,
		Path:         body.Path,
	}, nil
}

// WriteGeneratedPlanTracking executes the corresponding package operation.
func WriteGeneratedPlanTracking(
	ctx context.Context,
	projectDir string,
	tracking generationcore.GeneratedPlanTracking,
) (GeneratedPlanTrackingPaths, error) {
	err := ctx.Err()
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	err = generationcore.ValidateGeneratedPlanTracking(tracking)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	paths, err := GeneratedPlanTrackingLayout(projectDir, tracking.Scope)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	scopeID, err := quizScopeID(tracking.Scope)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	if err := opscore.EnsureNamedArtifactMetadata(projectDir, "tracker", scopeID+"-tracking"); err != nil {
		return GeneratedPlanTrackingPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, fmt.Errorf("create plan tracking directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderGeneratedPlanTracking(tracking)), 0o644)
	if err != nil {
		return GeneratedPlanTrackingPaths{}, fmt.Errorf("write plan tracking: %w", err)
	}

	return paths, nil
}

// RenderGeneratedPlanTracking renders generated plan tracking as Markdown.
//
//nolint:wsl_v5 // Markdown rendering keeps adjacent writes together.
func RenderGeneratedPlanTracking(tracking generationcore.GeneratedPlanTracking) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("title: " + opsmetadata.Quote(tracking.Title) + "\n")
	output.WriteString("project_id: " + opsmetadata.Quote(tracking.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + opsmetadata.Quote(tracking.Scope.CaseSlug) + "\n")
	output.WriteString("status: " + opsmetadata.Quote("tracking") + "\n")
	output.WriteString("plan_path: " + opsmetadata.Quote(filepath.ToSlash(tracking.PlanPath)) + "\n")
	output.WriteString("created_at: " + opsmetadata.Timestamp(tracking.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# " + tracking.Title + "\n\n")
	output.WriteString(strings.TrimSpace(tracking.Body) + "\n\n")
	output.WriteString("## Tickets\n\n")

	for _, ticket := range tracking.Tickets {
		output.WriteString("- ")
		output.WriteString(ticket.SliceID)
		output.WriteString(": ")
		output.WriteString(ticket.Key)
		output.WriteString(" - ")
		output.WriteString(ticket.Title)
		if strings.TrimSpace(ticket.Path) != "" {
			output.WriteString(" (`")
			output.WriteString(filepath.ToSlash(ticket.Path))
			output.WriteString("`)")
		}
		output.WriteString("\n")
	}

	output.WriteString("\n## Provenance\n\n")
	output.WriteString("- plan: " + filepath.ToSlash(tracking.PlanPath) + "\n")
	for _, path := range tracking.SourcePaths {
		output.WriteString("- source: " + filepath.ToSlash(path) + "\n")
	}

	return output.String()
}
