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

const SpecDirectory = "specs"

type GeneratedSpec struct {
	Scope            QuizScope
	Title            string
	Body             string
	SourcePaths      []string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

type GeneratedSpecPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

func GeneratedSpecLayout(projectDir string, scope QuizScope) (GeneratedSpecPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	relativePath := filepath.ToSlash(filepath.Join(DirectoryName, SpecDirectory, scopeID+"-spec.md"))

	return GeneratedSpecPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

func WriteGeneratedSpec(ctx context.Context, projectDir string, spec GeneratedSpec) (GeneratedSpecPaths, error) {
	err := ctx.Err()
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	err = ValidateGeneratedSpec(spec)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	paths, err := GeneratedSpecLayout(projectDir, spec.Scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return GeneratedSpecPaths{}, fmt.Errorf("create spec directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderGeneratedSpec(spec)), 0o644)
	if err != nil {
		return GeneratedSpecPaths{}, fmt.Errorf("write generated spec: %w", err)
	}

	return paths, nil
}

func RenderGeneratedSpec(spec GeneratedSpec) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("title: " + yamlString(spec.Title) + "\n")
	output.WriteString("project_id: " + yamlString(spec.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + yamlString(spec.Scope.CaseSlug) + "\n")
	output.WriteString("status: " + yamlString("draft") + "\n")
	output.WriteString("source_request_id: " + yamlString(spec.SourceRequestID) + "\n")
	output.WriteString("source_response_id: " + yamlString(spec.SourceResponseID) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(spec.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# " + spec.Title + "\n\n")
	output.WriteString(strings.TrimSpace(spec.Body) + "\n\n")
	output.WriteString("## Provenance\n\n")

	for _, path := range spec.SourcePaths {
		output.WriteString("- source: " + path + "\n")
	}

	return output.String()
}

func ValidateGeneratedSpec(spec GeneratedSpec) error {
	if strings.TrimSpace(spec.Title) == "" {
		return fmt.Errorf("generated spec title is required")
	}

	if strings.TrimSpace(spec.Body) == "" {
		return fmt.Errorf("generated spec body is required")
	}

	err := ValidateQuizScope(spec.Scope)
	if err != nil {
		return err
	}

	if spec.CreatedAt.IsZero() {
		return fmt.Errorf("generated spec created time is required")
	}

	return nil
}
