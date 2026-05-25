package chc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const PlanDirectory = "plans"

type GeneratedPlan struct {
	Scope            QuizScope
	Title            string
	Body             string
	SourcePaths      []string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

type GeneratedPlanPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

type ArchivedPlanPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

func GeneratedPlanLayout(projectDir string, scope QuizScope) (GeneratedPlanPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return GeneratedPlanPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return GeneratedPlanPaths{}, err
	}

	relativePath := filepath.ToSlash(filepath.Join(DirectoryName, PlanDirectory, scopeID+"-plan.md"))

	return GeneratedPlanPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

func ArchivedPlanLayout(projectDir string, scope QuizScope, archivedAt time.Time) (ArchivedPlanPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return ArchivedPlanPaths{}, err
	}

	if archivedAt.IsZero() {
		return ArchivedPlanPaths{}, fmt.Errorf("plan archive time is required")
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return ArchivedPlanPaths{}, err
	}

	archivedAt = archivedAt.UTC()
	timestamp := archivedAt.Format("20060102-150405") + fmt.Sprintf("%09d", archivedAt.Nanosecond())
	relativePath := filepath.ToSlash(filepath.Join(
		DirectoryName,
		PlanDirectory,
		"archive",
		scopeID+"-plan-"+timestamp+".md",
	))

	return ArchivedPlanPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

func WriteGeneratedPlan(ctx context.Context, projectDir string, plan GeneratedPlan) (GeneratedPlanPaths, error) {
	err := ctx.Err()
	if err != nil {
		return GeneratedPlanPaths{}, err
	}

	err = ValidateGeneratedPlan(plan)
	if err != nil {
		return GeneratedPlanPaths{}, err
	}

	paths, err := GeneratedPlanLayout(projectDir, plan.Scope)
	if err != nil {
		return GeneratedPlanPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return GeneratedPlanPaths{}, fmt.Errorf("create plan directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderGeneratedPlan(plan)), 0o644)
	if err != nil {
		return GeneratedPlanPaths{}, fmt.Errorf("write generated plan: %w", err)
	}

	return paths, nil
}

func ArchiveGeneratedPlan(
	ctx context.Context,
	projectDir string,
	scope QuizScope,
	archivedAt time.Time,
) (ArchivedPlanPaths, bool, error) {
	err := ctx.Err()
	if err != nil {
		return ArchivedPlanPaths{}, false, err
	}

	canonicalPaths, err := GeneratedPlanLayout(projectDir, scope)
	if err != nil {
		return ArchivedPlanPaths{}, false, err
	}

	content, err := os.ReadFile(canonicalPaths.Path)
	if errors.Is(err, os.ErrNotExist) {
		return ArchivedPlanPaths{}, false, nil
	}

	if err != nil {
		return ArchivedPlanPaths{}, false, fmt.Errorf("read generated plan before archive: %w", err)
	}

	archivePaths, err := ArchivedPlanLayout(projectDir, scope, archivedAt)
	if err != nil {
		return ArchivedPlanPaths{}, false, err
	}

	err = os.MkdirAll(filepath.Dir(archivePaths.Path), 0o755)
	if err != nil {
		return ArchivedPlanPaths{}, false, fmt.Errorf("create plan archive directory: %w", err)
	}

	archiveContent := RenderArchivedPlan(string(content), canonicalPaths.RelativePath, archivedAt)

	err = os.WriteFile(archivePaths.Path, []byte(archiveContent), 0o644)
	if err != nil {
		return ArchivedPlanPaths{}, false, fmt.Errorf("write archived plan: %w", err)
	}

	return archivePaths, true, nil
}

func RenderArchivedPlan(content string, sourceRelativePath string, archivedAt time.Time) string {
	archiveMetadata := []string{
		"archived_from: " + yamlString(filepath.ToSlash(sourceRelativePath)),
		"archived_at: " + yamlString(archivedAt.UTC().Format(time.RFC3339Nano)),
	}

	trimmed := strings.TrimLeft(content, "\ufeff\r\n\t ")
	if !strings.HasPrefix(trimmed, "---\n") {
		return "---\n" + strings.Join(archiveMetadata, "\n") + "\n---\n\n" + strings.TrimSpace(content) + "\n"
	}

	frontmatterEnd := strings.Index(trimmed[len("---\n"):], "\n---")
	if frontmatterEnd < 0 {
		return "---\n" + strings.Join(archiveMetadata, "\n") + "\n---\n\n" + strings.TrimSpace(content) + "\n"
	}

	insertAt := len("---\n") + frontmatterEnd

	return trimmed[:insertAt] + "\n" + strings.Join(archiveMetadata, "\n") + trimmed[insertAt:]
}

func RenderGeneratedPlan(plan GeneratedPlan) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("title: " + yamlString(plan.Title) + "\n")
	output.WriteString("project_id: " + yamlString(plan.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + yamlString(plan.Scope.CaseSlug) + "\n")
	output.WriteString("status: " + yamlString("planned") + "\n")
	output.WriteString("source_request_id: " + yamlString(plan.SourceRequestID) + "\n")
	output.WriteString("source_response_id: " + yamlString(plan.SourceResponseID) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(plan.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# " + plan.Title + "\n\n")
	output.WriteString(strings.TrimSpace(plan.Body) + "\n\n")
	output.WriteString("## Provenance\n\n")

	for _, path := range plan.SourcePaths {
		output.WriteString("- source: " + path + "\n")
	}

	return output.String()
}

func ValidateGeneratedPlan(plan GeneratedPlan) error {
	if strings.TrimSpace(plan.Title) == "" {
		return fmt.Errorf("generated plan title is required")
	}

	if strings.TrimSpace(plan.Body) == "" {
		return fmt.Errorf("generated plan body is required")
	}

	err := ValidateQuizScope(plan.Scope)
	if err != nil {
		return err
	}

	if plan.CreatedAt.IsZero() {
		return fmt.Errorf("generated plan created time is required")
	}

	return nil
}

func ValidateGeneratedPlanModelContent(plan GeneratedPlan) error {
	if strings.TrimSpace(plan.Title) == "" {
		return fmt.Errorf("generated plan title is required")
	}

	if strings.TrimSpace(plan.Body) == "" {
		return fmt.Errorf("generated plan body is required")
	}

	return ValidateQuizScope(plan.Scope)
}

func (p GeneratedPlan) HasContent() bool {
	return strings.TrimSpace(p.Title) != "" || strings.TrimSpace(p.Body) != ""
}
