package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

const (
	// QuizArtifactsDirectory defines a package boundary constant.
	QuizArtifactsDirectory = "quizzes"
	// QuizRuntimeDirectory defines a package boundary constant.
	QuizRuntimeDirectory = "tmp/quiz-runtime"
)

// QuizArtifactPaths describes a package boundary value.
type QuizArtifactPaths struct {
	ScopeID             string
	RawRelativePath     string
	RawPath             string
	CuratedRelativePath string
	CuratedPath         string
	RuntimeRelativePath string
	RuntimePath         string
}

// QuizArtifactRepository defines a package boundary interface.
type QuizArtifactRepository interface {
	Paths(scope quizcore.Scope) (QuizArtifactPaths, error)
	AppendRawCapture(ctx context.Context, state quizcore.RuntimeState, entry quizcore.Entry) (QuizArtifactPaths, error)
	LoadRawCaptureEntries(ctx context.Context, scope quizcore.Scope) ([]quizcore.Entry, error)
	WriteCuratedNotes(ctx context.Context, notes quizcore.CuratedNotes) (QuizArtifactPaths, error)
	SaveRuntimeState(ctx context.Context, state quizcore.RuntimeState) (QuizArtifactPaths, error)
	LoadRuntimeState(ctx context.Context, scope quizcore.Scope) (quizcore.RuntimeState, error)
	Observe(ctx context.Context, scope quizcore.Scope) (quizcore.Observation, error)
	ApplyAction(ctx context.Context, request quizcore.ActionRequest, transition quizcore.TransitionFunc) (quizcore.ActionResult, error)
}

type fileQuizArtifactRepository struct {
	projectDir           string
	afterMutationReplace func(int) error
}

// NewQuizArtifactRepository constructs the corresponding package service or value.
func NewQuizArtifactRepository(projectDir string) QuizArtifactRepository {
	return &fileQuizArtifactRepository{projectDir: projectDir}
}

// QuizArtifactLayout executes the corresponding package operation.
func QuizArtifactLayout(projectDir string, scope quizcore.Scope) (QuizArtifactPaths, error) {
	err := quizcore.ValidateScope(scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	if scope.Namespace != "" {
		return scopedQuizArtifactLayout(projectDir, scope)
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	raw, err := opscore.NamedArtifactBody(projectDir, "quiz", scopeID+"-quiz-raw")
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	curated, err := opscore.NamedArtifactBody(projectDir, "quiz", scopeID+"-quiz")
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	runtimeRelativePath := UserRelativePath(filepath.FromSlash(QuizRuntimeDirectory), scopeID+".json")

	return QuizArtifactPaths{
		ScopeID:             scopeID,
		RawRelativePath:     raw.RelativePath,
		RawPath:             raw.Path,
		CuratedRelativePath: curated.RelativePath,
		CuratedPath:         curated.Path,
		RuntimeRelativePath: runtimeRelativePath,
		RuntimePath:         filepath.Join(projectDir, filepath.FromSlash(runtimeRelativePath)),
	}, nil
}

// Paths executes the corresponding package operation.
func (r *fileQuizArtifactRepository) Paths(scope quizcore.Scope) (QuizArtifactPaths, error) {
	return QuizArtifactLayout(r.projectDir, scope)
}

// AppendRawCapture executes the corresponding package operation.
func (r *fileQuizArtifactRepository) appendRawCapture(ctx context.Context, state quizcore.RuntimeState, entry quizcore.Entry) (QuizArtifactPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = quizcore.ValidateRuntimeState(state)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = quizcore.ValidateEntry(entry)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	paths, err := r.prepareRawCaptureArtifact(state)
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

func (r *fileQuizArtifactRepository) prepareRawCaptureArtifact(state quizcore.RuntimeState) (QuizArtifactPaths, error) {
	scope := quizScopeFromRuntime(state)

	paths, err := r.Paths(scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	if err := opscore.EnsureNamedArtifactMetadata(r.projectDir, "quiz", scopeID+"-quiz-raw"); err != nil {
		return QuizArtifactPaths{}, err
	}

	return paths, nil
}

// WriteCuratedNotes writes the requested output.
func (r *fileQuizArtifactRepository) writeCuratedNotes(ctx context.Context, notes quizcore.CuratedNotes) (QuizArtifactPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = quizcore.ValidateScope(notes.Scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	for _, entry := range notes.Entries {
		err = quizcore.ValidateEntry(entry)
		if err != nil {
			return QuizArtifactPaths{}, err
		}
	}

	paths, err := r.Paths(notes.Scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	scopeID, err := quizScopeID(notes.Scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	if err := opscore.EnsureNamedArtifactMetadata(r.projectDir, "quiz", scopeID+"-quiz"); err != nil {
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

// LoadRawCaptureEntries loads the requested data.
func (r *fileQuizArtifactRepository) LoadRawCaptureEntries(
	ctx context.Context,
	scope quizcore.Scope,
) ([]quizcore.Entry, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	paths, err := r.Paths(scope)
	if err != nil {
		return nil, err
	}

	owner := &fileAuthoringSessionRepository{projectDir: r.projectDir}

	content, err := owner.authoringRead(paths.RawRelativePath)
	if os.IsNotExist(err) {
		return []quizcore.Entry{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read quiz raw capture: %w", err)
	}

	return quizcore.ParseRawCaptureEntries(string(content))
}

// SaveRuntimeState persists the requested data.
func (r *fileQuizArtifactRepository) saveRuntimeState(ctx context.Context, state quizcore.RuntimeState) (QuizArtifactPaths, error) {
	err := ctx.Err()
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	err = quizcore.ValidateRuntimeState(state)
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

// LoadRuntimeState loads the requested data.
func (r *fileQuizArtifactRepository) LoadRuntimeState(ctx context.Context, scope quizcore.Scope) (quizcore.RuntimeState, error) {
	observation, err := r.Observe(ctx, scope)
	if err == nil && !observation.Exists {
		err = fmt.Errorf("read quiz runtime state: %w", os.ErrNotExist)
	}

	return observation.State, err
}

// RenderQuizRawCapture renders the corresponding package data.
func RenderQuizRawCapture(state quizcore.RuntimeState, entry quizcore.Entry) string {
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

	if encoded, err := json.Marshal(entry); err == nil {
		output.WriteString("- entry_json: " + base64.RawStdEncoding.EncodeToString(encoded) + "\n")
	}

	output.WriteString("- question_id: " + entry.QuestionID + "\n")
	output.WriteString("- status: " + string(entry.Status) + "\n")
	output.WriteString("- mode: " + string(entry.Mode) + "\n")
	output.WriteString("- topic: " + entry.Topic + "\n")
	output.WriteString("- depth: " + fmt.Sprintf("%d", entry.Depth) + "\n")

	if strings.TrimSpace(entry.ParentID) != "" {
		output.WriteString("- parent_id: " + entry.ParentID + "\n")
	}

	if len(entry.Tags) > 0 {
		output.WriteString("- tags: " + strings.Join(entry.Tags, ", ") + "\n")
	}

	if strings.TrimSpace(entry.PromptIntent) != "" {
		output.WriteString("\n")
		output.WriteString("Prompt intent:\n\n")
		output.WriteString(quizRawText(entry.PromptIntent) + "\n")
	}

	output.WriteString("\n")
	output.WriteString("Question:\n\n")
	output.WriteString(quizRawText(entry.Question) + "\n\n")

	if entry.Status == quizcore.EntryStatusSkipped {
		output.WriteString("Skipped: true\n\n")

		return output.String()
	}

	output.WriteString("Answer:\n\n")
	output.WriteString(quizRawText(entry.Answer) + "\n\n")

	return output.String()
}

// RenderQuizCuratedNotes renders the corresponding package data.
func RenderQuizCuratedNotes(paths QuizArtifactPaths, notes quizcore.CuratedNotes) string {
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
	renderQuizCuratedSection(&output, "Acceptance Readiness", notes.AcceptanceReadiness)
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

func quizScopeFromRuntime(state quizcore.RuntimeState) quizcore.Scope {
	return quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}
}

func quizScopeID(scope quizcore.Scope) (string, error) {
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

func quizRawText(value string) string {
	lines := strings.Split(strings.TrimSpace(value), "\n")
	for index, line := range lines {
		if line == "## Turn" {
			lines[index] = "\\" + line
		}
	}

	return strings.Join(lines, "\n")
}

func scopedQuizArtifactLayout(projectDir string, scope quizcore.Scope) (QuizArtifactPaths, error) {
	id, err := quizScopeID(scope)
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir, Namespace: scope.Namespace})
	if err != nil {
		return QuizArtifactPaths{}, err
	}

	paths := QuizArtifactPaths{ScopeID: id}
	for _, target := range []struct {
		name               string
		relative, absolute *string
	}{
		{id + "-quiz-raw", &paths.RawRelativePath, &paths.RawPath},
		{id + "-quiz", &paths.CuratedRelativePath, &paths.CuratedPath},
	} {
		dir, err := layout.ArtifactDirectory("quiz", "00000-"+opscore.ArtifactSlugOrDefault(target.name))
		if err != nil {
			return paths, err
		}

		*target.absolute = filepath.Join(dir, opscore.ArtifactBodyFile())

		relative, err := filepath.Rel(projectDir, *target.absolute)
		if err != nil {
			return paths, err
		}

		*target.relative = filepath.ToSlash(relative)
	}

	paths.RuntimeRelativePath = filepath.ToSlash(filepath.Join(workspacecore.DirectoryName, scope.Namespace, QuizRuntimeDirectory, id+".json"))
	paths.RuntimePath = filepath.Join(projectDir, filepath.FromSlash(paths.RuntimeRelativePath))

	return paths, nil
}
