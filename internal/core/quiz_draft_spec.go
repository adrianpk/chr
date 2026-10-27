package core

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// QuizSpecDirectory defines a package boundary constant.
const QuizSpecDirectory = "specs/drafts"

// QuizDraftSpec describes a package boundary value.
type QuizDraftSpec struct {
	ExpectedRevision string `json:"-"`
	Scope            quizcore.Scope
	Title            string
	Body             string
	Feature          AcceptanceFeatureGeneration
	SourceQuizPath   string
	SourceRawPath    string
	SourceDumpPaths  []string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// QuizDraftSpecPaths describes a package boundary value.
type QuizDraftSpecPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// QuizDraftSpecLayout executes the corresponding package operation.
func QuizDraftSpecLayout(projectDir string, scope quizcore.Scope) (QuizDraftSpecPaths, error) {
	err := quizcore.ValidateScope(scope)
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	body, err := NamedOpsArtifactBody(projectDir, "spec", scopeID+"-quiz-spec")
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	return QuizDraftSpecPaths{
		ScopeID:      scopeID,
		RelativePath: body.RelativePath,
		Path:         body.Path,
	}, nil
}

func writeQuizDraftSpec(ctx context.Context, projectDir string, spec QuizDraftSpec) (QuizDraftSpecPaths, error) {
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

	publication := GeneratedSpec{ExpectedRevision: spec.ExpectedRevision, Scope: spec.Scope, Title: spec.Title, CreatedAt: spec.CreatedAt}
	if err := publishSpecContent(ctx, projectDir, GeneratedSpecPaths(paths), publication, RenderQuizDraftSpec(spec)); err != nil {
		return QuizDraftSpecPaths{}, err
	}

	return paths, nil
}

// RenderQuizDraftSpec renders the corresponding package data.
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

// ValidateQuizDraftSpec validates the corresponding package boundary value.
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

	if spec.Feature.HasContent() {
		_, err = ValidateAcceptanceFeatureGeneration(spec.Feature)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateQuizDraftSpecModelContent validates the corresponding package boundary value.
func ValidateQuizDraftSpecModelContent(spec QuizDraftSpec) error {
	if !spec.HasContent() {
		return fmt.Errorf("quiz draft spec content is required")
	}

	err := quizcore.ValidateScope(spec.Scope)
	if err != nil {
		return err
	}

	if spec.CreatedAt.IsZero() {
		return fmt.Errorf("quiz draft spec created time is required")
	}

	if spec.Feature.HasContent() {
		_, err = ValidateAcceptanceFeatureGeneration(spec.Feature)
		if err != nil {
			return err
		}
	}

	return nil
}

// HasContent executes the corresponding package operation.
func (s QuizDraftSpec) HasContent() bool {
	return strings.TrimSpace(s.Title) != "" && strings.TrimSpace(s.Body) != ""
}
