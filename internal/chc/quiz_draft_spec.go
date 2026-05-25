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

const QuizSpecDirectory = "specs/drafts"

type QuizDraftSpec struct {
	Scope            QuizScope
	Title            string
	Body             string
	SourceQuizPath   string
	SourceRawPath    string
	SourceDumpPaths  []string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

type QuizDraftSpecPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

func QuizDraftSpecLayout(projectDir string, scope QuizScope) (QuizDraftSpecPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	relativePath := filepath.ToSlash(filepath.Join(DirectoryName, filepath.FromSlash(QuizSpecDirectory), scopeID+"-quiz-spec.md"))

	return QuizDraftSpecPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

func WriteQuizDraftSpec(ctx context.Context, projectDir string, spec QuizDraftSpec) (QuizDraftSpecPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	err = ValidateQuizDraftSpec(spec)
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	paths, err := QuizDraftSpecLayout(projectDir, spec.Scope)
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return QuizDraftSpecPaths{}, fmt.Errorf("create quiz draft spec directory: %w", err)
	}

	content := RenderQuizDraftSpec(spec)

	err = os.WriteFile(paths.Path, []byte(content), 0o644)
	if err != nil {
		return QuizDraftSpecPaths{}, fmt.Errorf("write quiz draft spec: %w", err)
	}

	return paths, nil
}

func RenderQuizDraftSpec(spec QuizDraftSpec) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("title: " + yamlString(spec.Title) + "\n")
	output.WriteString("project_id: " + yamlString(spec.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + yamlString(spec.Scope.CaseSlug) + "\n")
	output.WriteString("status: " + yamlString("draft") + "\n")
	output.WriteString("source_quiz_path: " + yamlString(spec.SourceQuizPath) + "\n")
	output.WriteString("source_raw_path: " + yamlString(spec.SourceRawPath) + "\n")
	output.WriteString("source_request_id: " + yamlString(spec.SourceRequestID) + "\n")
	output.WriteString("source_response_id: " + yamlString(spec.SourceResponseID) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(spec.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# " + spec.Title + "\n\n")
	output.WriteString(strings.TrimSpace(spec.Body) + "\n\n")
	output.WriteString("## Provenance\n\n")
	output.WriteString("- curated quiz: " + spec.SourceQuizPath + "\n")
	output.WriteString("- raw quiz: " + spec.SourceRawPath + "\n")

	for _, path := range spec.SourceDumpPaths {
		output.WriteString("- dump: " + path + "\n")
	}

	return output.String()
}

func ValidateQuizDraftSpec(spec QuizDraftSpec) error {
	err := ValidateQuizDraftSpecModelContent(spec)
	if err != nil {
		return err
	}

	if strings.TrimSpace(spec.SourceQuizPath) == "" {
		return fmt.Errorf("quiz draft spec source quiz path is required")
	}

	if strings.TrimSpace(spec.SourceRawPath) == "" {
		return fmt.Errorf("quiz draft spec source raw path is required")
	}

	if spec.CreatedAt.IsZero() {
		return fmt.Errorf("quiz draft spec created time is required")
	}

	return nil
}

func ValidateQuizDraftSpecModelContent(spec QuizDraftSpec) error {
	if !spec.HasContent() {
		return fmt.Errorf("quiz draft spec content is required")
	}

	err := ValidateQuizScope(spec.Scope)
	if err != nil {
		return err
	}

	if spec.CreatedAt.IsZero() {
		return fmt.Errorf("quiz draft spec created time is required")
	}

	return nil
}

func (s QuizDraftSpec) HasContent() bool {
	return strings.TrimSpace(s.Title) != "" && strings.TrimSpace(s.Body) != ""
}
