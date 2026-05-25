package chc

import (
	"bytes"
	"context"
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
