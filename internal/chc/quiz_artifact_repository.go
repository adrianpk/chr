package chc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	QuizArtifactsDirectory = "quizzes"
	QuizRuntimeDirectory   = "tmp/quiz-runtime"
)

type QuizArtifactPaths struct {
	ScopeID             string
	RawRelativePath     string
	RawPath             string
	CuratedRelativePath string
	CuratedPath         string
	RuntimeRelativePath string
	RuntimePath         string
}

type QuizArtifactRepository interface {
	Paths(scope QuizScope) (QuizArtifactPaths, error)
	AppendRawCapture(ctx context.Context, state QuizRuntimeState, entry QuizEntry) (QuizArtifactPaths, error)
	WriteCuratedNotes(ctx context.Context, notes QuizCuratedNotes) (QuizArtifactPaths, error)
	SaveRuntimeState(ctx context.Context, state QuizRuntimeState) (QuizArtifactPaths, error)
	LoadRuntimeState(ctx context.Context, scope QuizScope) (QuizRuntimeState, error)
}

type QuizCuratedNotes struct {
	Scope                      QuizScope
	Status                     string
	Decisions                  []string
	Constraints                []string
	Preferences                []string
	OpenQuestions              []string
	ImplementationImplications []string
	Entries                    []QuizEntry
	UpdatedAt                  string
}

type fileQuizArtifactRepository struct {
	projectDir string
}

func NewQuizArtifactRepository(projectDir string) QuizArtifactRepository {
	return &fileQuizArtifactRepository{projectDir: projectDir}
}

func QuizArtifactLayout(projectDir string, scope QuizScope) (QuizArtifactPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	rawRelativePath := filepath.ToSlash(filepath.Join(DirectoryName, QuizArtifactsDirectory, scopeID+"-quiz-raw.md"))
	curatedRelativePath := filepath.ToSlash(filepath.Join(DirectoryName, QuizArtifactsDirectory, scopeID+"-quiz.md"))
	runtimeRelativePath := filepath.ToSlash(filepath.Join(DirectoryName, filepath.FromSlash(QuizRuntimeDirectory), scopeID+".json"))

	return QuizArtifactPaths{
		ScopeID:             scopeID,
		RawRelativePath:     rawRelativePath,
		RawPath:             filepath.Join(projectDir, filepath.FromSlash(rawRelativePath)),
		CuratedRelativePath: curatedRelativePath,
		CuratedPath:         filepath.Join(projectDir, filepath.FromSlash(curatedRelativePath)),
		RuntimeRelativePath: runtimeRelativePath,
		RuntimePath:         filepath.Join(projectDir, filepath.FromSlash(runtimeRelativePath)),
	}, nil
}

func (r *fileQuizArtifactRepository) Paths(scope QuizScope) (QuizArtifactPaths, error) {
	return QuizArtifactLayout(r.projectDir, scope)
}

func (r *fileQuizArtifactRepository) AppendRawCapture(
	ctx context.Context,
	state QuizRuntimeState,
	entry QuizEntry,
) (QuizArtifactPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = ValidateQuizRuntimeState(state)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = ValidateQuizEntry(entry)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	paths, err := r.Paths(quizScopeFromRuntime(state))
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.RawPath), 0o755)
	if err != nil {
		return QuizArtifactPaths{}, fmt.Errorf("create quiz raw directory: %w", err)
	}

	content := RenderQuizRawCapture(state, entry)

	file, err := os.OpenFile(paths.RawPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return QuizArtifactPaths{}, fmt.Errorf("open quiz raw capture: %w", err)
	}

	_, writeErr := file.WriteString(content)
	closeErr := file.Close()

	if writeErr != nil {
		return QuizArtifactPaths{}, fmt.Errorf("write quiz raw capture: %w", writeErr)
	}

	if closeErr != nil {
		return QuizArtifactPaths{}, fmt.Errorf("close quiz raw capture: %w", closeErr)
	}

	return paths, nil
}

func (r *fileQuizArtifactRepository) WriteCuratedNotes(
	ctx context.Context,
	notes QuizCuratedNotes,
) (QuizArtifactPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = ValidateQuizScope(notes.Scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	for _, entry := range notes.Entries {
		err = ValidateQuizEntry(entry)
		if err != nil {
			return QuizArtifactPaths{}, err
		}
	}

	paths, err := r.Paths(notes.Scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.CuratedPath), 0o755)
	if err != nil {
		return QuizArtifactPaths{}, fmt.Errorf("create quiz curated directory: %w", err)
	}

	content := RenderQuizCuratedNotes(paths, notes)

	err = os.WriteFile(paths.CuratedPath, []byte(content), 0o644)
	if err != nil {
		return QuizArtifactPaths{}, fmt.Errorf("write quiz curated notes: %w", err)
	}

	return paths, nil
}

func (r *fileQuizArtifactRepository) SaveRuntimeState(
	ctx context.Context,
	state QuizRuntimeState,
) (QuizArtifactPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = ValidateQuizRuntimeState(state)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	paths, err := r.Paths(quizScopeFromRuntime(state))
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.RuntimePath), 0o755)
	if err != nil {
		return QuizArtifactPaths{}, fmt.Errorf("create quiz runtime directory: %w", err)
	}

	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return QuizArtifactPaths{}, fmt.Errorf("encode quiz runtime state: %w", err)
	}

	err = os.WriteFile(paths.RuntimePath, append(content, '\n'), 0o644)
	if err != nil {
		return QuizArtifactPaths{}, fmt.Errorf("write quiz runtime state: %w", err)
	}

	return paths, nil
}

func (r *fileQuizArtifactRepository) LoadRuntimeState(ctx context.Context, scope QuizScope) (QuizRuntimeState, error) {
	err := ctx.Err()
	if err != nil {
		return QuizRuntimeState{}, err
	}

	paths, err := r.Paths(scope)
	if err != nil {
		return QuizRuntimeState{}, err
	}

	content, err := os.ReadFile(paths.RuntimePath)
	if err != nil {
		return QuizRuntimeState{}, fmt.Errorf("read quiz runtime state: %w", err)
	}

	var state QuizRuntimeState

	err = json.Unmarshal(content, &state)
	if err != nil {
		return QuizRuntimeState{}, fmt.Errorf("decode quiz runtime state: %w", err)
	}

	err = ValidateQuizRuntimeState(state)
	if err != nil {
		return QuizRuntimeState{}, err
	}

	return state, nil
}

func RenderQuizRawCapture(state QuizRuntimeState, entry QuizEntry) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("project_id: " + yamlString(state.ProjectID) + "\n")
	output.WriteString("case_slug: " + yamlString(state.CaseSlug) + "\n")
	output.WriteString("scope: " + yamlString(quizScopeName(state.CaseSlug)) + "\n")
	output.WriteString("status: " + yamlString("raw") + "\n")
	output.WriteString("runtime_id: " + yamlString(state.ID) + "\n")
	output.WriteString("curated_path: " + yamlString(state.CuratedPath) + "\n")
	output.WriteString("updated_at: " + yamlTimestamp(state.UpdatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("## Turn\n\n")
	output.WriteString("- question_id: " + entry.QuestionID + "\n")
	output.WriteString("- status: " + string(entry.Status) + "\n")
	output.WriteString("- mode: " + string(entry.Mode) + "\n")
	output.WriteString("- topic: " + entry.Topic + "\n")
	output.WriteString("- depth: " + fmt.Sprintf("%d", entry.Depth) + "\n\n")
	output.WriteString("Question:\n\n")
	output.WriteString(strings.TrimSpace(entry.Question) + "\n\n")

	if entry.Status == QuizEntryStatusSkipped {
		output.WriteString("Skipped: true\n\n")

		return output.String()
	}

	output.WriteString("Answer:\n\n")
	output.WriteString(strings.TrimSpace(entry.Answer) + "\n\n")

	return output.String()
}

func RenderQuizCuratedNotes(paths QuizArtifactPaths, notes QuizCuratedNotes) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("project_id: " + yamlString(notes.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + yamlString(notes.Scope.CaseSlug) + "\n")
	output.WriteString("scope: " + yamlString(quizScopeName(notes.Scope.CaseSlug)) + "\n")
	output.WriteString("status: " + yamlString("curated") + "\n")
	output.WriteString("raw_path: " + yamlString(paths.RawRelativePath) + "\n")
	output.WriteString("updated_at: " + yamlString(notes.UpdatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Quiz Notes\n\n")
	renderQuizCuratedSection(&output, "Decisions", notes.Decisions)
	renderQuizCuratedSection(&output, "Constraints", notes.Constraints)
	renderQuizCuratedSection(&output, "Preferences", notes.Preferences)
	renderQuizCuratedSection(&output, "Open Questions", notes.OpenQuestions)
	renderQuizCuratedSection(&output, "Implementation Implications", notes.ImplementationImplications)
	output.WriteString("## Provenance\n\n")
	output.WriteString("- raw quiz: " + paths.RawRelativePath + "\n")

	return output.String()
}

func renderQuizCuratedSection(output *bytes.Buffer, title string, values []string) {
	output.WriteString("## " + title + "\n\n")

	if len(values) == 0 {
		output.WriteString("- None recorded.\n\n")

		return
	}

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		output.WriteString("- " + value + "\n")
	}

	output.WriteString("\n")
}

func quizScopeFromRuntime(state QuizRuntimeState) QuizScope {
	return QuizScope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}
}

func quizScopeID(scope QuizScope) (string, error) {
	caseSlug := strings.TrimSpace(scope.CaseSlug)
	if caseSlug == "" {
		return "project", nil
	}

	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(caseSlug)))
	if clean == "." || clean == ".." || strings.Contains(clean, "/") {
		return "", fmt.Errorf("quiz scope case slug must be a single path segment")
	}

	return clean, nil
}

func quizScopeName(caseSlug string) string {
	if strings.TrimSpace(caseSlug) == "" {
		return "project"
	}

	return "use_case"
}
