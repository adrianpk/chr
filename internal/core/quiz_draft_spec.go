package core

import (
	"bytes"
	"context"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// QuizSpecDirectory defines a package boundary constant.
const QuizSpecDirectory = "specs/drafts"

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

	body, err := opscore.NamedArtifactBody(projectDir, "spec", scopeID+"-quiz-spec")
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	return QuizDraftSpecPaths{
		ScopeID:      scopeID,
		RelativePath: body.RelativePath,
		Path:         body.Path,
	}, nil
}

func writeQuizDraftSpec(ctx context.Context, projectDir string, spec generationcore.QuizDraftSpec) (QuizDraftSpecPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizDraftSpecPaths{}, err
	}

	err = generationcore.ValidateQuizDraftSpec(spec)
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
func RenderQuizDraftSpec(spec generationcore.QuizDraftSpec) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("title: " + opsmetadata.Quote(spec.Title) + "\n")
	output.WriteString("project_id: " + opsmetadata.Quote(spec.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + opsmetadata.Quote(spec.Scope.CaseSlug) + "\n")
	output.WriteString("status: " + opsmetadata.Quote("draft") + "\n")
	output.WriteString("source_quiz_path: " + opsmetadata.Quote(spec.SourceQuizPath) + "\n")
	output.WriteString("source_raw_path: " + opsmetadata.Quote(spec.SourceRawPath) + "\n")
	output.WriteString("source_request_id: " + opsmetadata.Quote(spec.SourceRequestID) + "\n")
	output.WriteString("source_response_id: " + opsmetadata.Quote(spec.SourceResponseID) + "\n")
	output.WriteString("created_at: " + opsmetadata.Timestamp(spec.CreatedAt) + "\n")
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
