package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// SpecView is the TUI-facing projection of a spec document.
type SpecView struct {
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        quizcore.Scope
	RelativePath string
	UpdatedAt    time.Time
	Content      string
}

// SpecListResult contains the current spec list and a status message.
type SpecListResult struct {
	Message string
	Specs   []SpecView
}

// SpecGenerateRequest selects the quiz scope used to generate a spec.
type SpecGenerateRequest struct {
	OperationID string
	Scope       quizcore.Scope
}

// SpecOperationResult is the TUI-facing result for spec generation.
type SpecOperationResult struct {
	Message string
	Spec    SpecView
}

func (s service) ListSpecs() (SpecListResult, error) {
	scopes, err := s.quizReadScopes()
	if err != nil {
		return SpecListResult{}, err
	}

	index, err := s.workspaceIndex()
	if err != nil {
		return SpecListResult{}, err
	}

	specs := make([]SpecView, 0, len(scopes))
	for _, scope := range scopes {
		specs = append(specs, s.specViewFromScope(scope, index))
	}

	for index := range specs {
		specs[index].Index = index
	}

	return SpecListResult{Message: "Specs loaded.", Specs: specs}, nil
}

//nolint:wsl_v5 // Keeps result capture adjacent to the guarded runtime call.
func (s service) GenerateSpec(ctx context.Context, request SpecGenerateRequest) (SpecOperationResult, error) {
	scope := s.quizScope(request.Scope)
	var result SpecOperationResult

	err := s.withGenerationRuntime(ctx, "spec", scope, func(ctx context.Context) error {
		bound, ctx, err := s.bindGeneration(ctx, "spec-"+dumpRequestID(request.OperationID), string(quizmodel.ActionGenerateSpec))
		if err != nil {
			return err
		}

		result, err = bound.generateSpec(ctx, scope)

		return err
	})

	return result, err
}

//nolint:funlen,wsl_v5 // Generation flow is intentionally linear for logging and failure attribution.
func (s service) generateSpec(ctx context.Context, scope quizcore.Scope) (SpecOperationResult, error) {
	s.appendPlanGenerationLog("generate_spec_started", map[string]string{
		"case_slug": scope.CaseSlug,
		"timeout":   longGenerationTimeout.String(),
	})

	state, err := s.specGenerationRuntimeState(scope)
	if err != nil {
		s.appendPlanGenerationLog("generate_spec_runtime_state_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return SpecOperationResult{}, err
	}

	target, err := authoringcore.GeneratedSpecLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return SpecOperationResult{}, err
	}
	targetRevision, err := authoringcore.SpecTargetRevision(ctx, s.projectContext.ProjectDir, target.RelativePath)
	if err != nil {
		return SpecOperationResult{}, err
	}

	modelRequest, err := s.newSpecGenerationModelRequest(state)
	if err != nil {
		s.appendPlanGenerationLog("generate_spec_request_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return SpecOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_spec_model_invocation_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		s.appendPlanGenerationLog("generate_spec_model_invocation_failed", map[string]string{
			"case_slug":  scope.CaseSlug,
			"request_id": modelRequest.ID,
			"error":      err.Error(),
		})

		return SpecOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_spec_model_invocation_finished", map[string]string{
		"case_slug":   scope.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"title":       response.DraftSpec.Title,
	})

	spec := authoringcore.GeneratedSpec{
		ExpectedRevision: targetRevision,
		Scope:            scope,
		Title:            response.DraftSpec.Title,
		Body:             response.DraftSpec.Body,
		SourcePaths:      specSourcePaths(modelRequest.ContextSections),
		SourceRequestID:  modelRequest.ID,
		SourceResponseID: response.ID,
		CreatedAt:        response.CreatedAt.UTC(),
	}

	paths, err := authoringcore.WriteGeneratedSpec(ctx, s.projectContext.ProjectDir, spec)
	if err != nil {
		s.appendPlanGenerationLog("generate_spec_write_failed", map[string]string{
			"case_slug":   scope.CaseSlug,
			"request_id":  modelRequest.ID,
			"response_id": response.ID,
			"error":       err.Error(),
		})

		return SpecOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_spec_written", map[string]string{
		"case_slug":     scope.CaseSlug,
		"request_id":    modelRequest.ID,
		"response_id":   response.ID,
		"relative_path": paths.RelativePath,
		"path":          paths.Path,
	})

	view := s.specViewFromPath(scope, paths.RelativePath, paths.Path)

	return SpecOperationResult{
		Message: "Spec generated.",
		Spec:    view,
	}, nil
}

func (s service) specViewFromScope(scope quizReadScope, index core.WorkspaceIndex) SpecView {
	node, ok := s.latestSpecNodeForScope(scope.scope, index)
	if !ok {
		return SpecView{
			UseCaseTitle: scope.title,
			Title:        scope.title,
			Status:       "missing",
			Scope:        scope.scope,
		}
	}

	return s.specViewFromNode(scope.scope, scope.title, node)
}

func (s service) latestSpecNodeForScope(scope quizcore.Scope, index core.WorkspaceIndex) (core.WorkspaceNode, bool) {
	return latestDocumentNodeForScope(scope, index, core.WorkspaceNodeSpec)
}

func (s service) specViewFromNode(scope quizcore.Scope, fallbackTitle string, node core.WorkspaceNode) SpecView {
	title := node.Title
	if strings.TrimSpace(title) == "" {
		title = fallbackTitle
	}

	return s.specViewFromPath(scope, node.RelativePath, filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(node.RelativePath))).withMetadata(
		fallbackTitle,
		title,
		node.Status,
		node.ModifiedAt,
	)
}

func (s service) specViewFromPath(scope quizcore.Scope, relativePath, path string) SpecView {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		content = nil
	} else if err != nil {
		content = []byte("Unable to read spec: " + err.Error())
	}

	summary := summarizeWorkspaceMarkdownDocument(string(content), filepath.Base(path))

	return SpecView{
		UseCaseTitle: specUseCaseFallbackTitle(scope),
		Title:        summary.Title,
		Status:       summary.Status,
		Scope:        scope,
		RelativePath: filepath.ToSlash(relativePath),
		UpdatedAt:    fileModifiedAt(path),
		Content:      string(content),
	}
}

func (v SpecView) withMetadata(useCaseTitle, title, status string, updatedAt time.Time) SpecView {
	if strings.TrimSpace(useCaseTitle) != "" {
		v.UseCaseTitle = useCaseTitle
	}

	if strings.TrimSpace(title) != "" {
		v.Title = title
	}

	if strings.TrimSpace(status) != "" {
		v.Status = status
	}

	if !updatedAt.IsZero() {
		v.UpdatedAt = updatedAt
	}

	return v
}

func (s service) specGenerationRuntimeState(scope quizcore.Scope) (quizcore.RuntimeState, error) {
	state, err := s.quizRepository().LoadRuntimeState(s.ctx, scope)
	if err == nil {
		return state, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return quizcore.RuntimeState{}, err
	}

	state, err = s.newQuizRuntimeState(scope, quizcore.ModeDeep)
	if err != nil {
		return quizcore.RuntimeState{}, err
	}

	state.Status = quizcore.RuntimeStatusDone
	state.ActiveQuestion = nil
	state.UpdatedAt = s.utcNow()

	return state, nil
}

func (s service) newSpecGenerationModelRequest(state quizcore.RuntimeState) (quizmodel.Request, error) {
	return s.newDocumentGenerationModelRequest(state, documentModelOperation{
		action: quizmodel.ActionGenerateSpec, strategy: "spec_generation", requestIDPrefix: "spec-generation",
	}, s.appendSpecGenerationContextSections)
}

func (s service) appendSpecGenerationContextSections(sections []quizmodel.ContextSection, scope quizcore.Scope) ([]quizmodel.ContextSection, error) {
	dumps, err := s.specGenerationSources()
	if err != nil {
		return nil, err
	}

	return s.appendGenerationContextSections(sections, scope, func(index core.WorkspaceIndex) []string { return specContextNodeIDs(index, scope, dumps) })
}

func (s service) appendGenerationContextSections(
	sections []quizmodel.ContextSection,
	scope quizcore.Scope,
	nodeIDs func(core.WorkspaceIndex) []string,
) ([]quizmodel.ContextSection, error) {
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "agent_protocol", "protocol", "AGENTS.md")
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "project_document", "project", authoringcore.ProjectMetadataPath(""))

	if scope.CaseSlug != "" {
		sections = appendQuizFileContextSection(
			sections,
			s.projectContext.ProjectDir,
			"use_case_document",
			"use_case",
			authoringcore.CaseMetadataPath("", scope.CaseSlug),
		)
	}

	index, err := s.workspaceIndex()
	if err != nil {
		return nil, err
	}

	for _, nodeID := range nodeIDs(index) {
		node := index.NodesByID[nodeID]
		if core.IsRawDumpNode(node) || core.IsRawQuizNode(node) {
			continue
		}

		if !generationContextNodeAllowed(scope, node) {
			continue
		}

		sections = appendSpecNodeContextSection(sections, s.projectContext.ProjectDir, node)
	}

	return sections, nil
}

func generationContextNodeAllowed(scope quizcore.Scope, node core.WorkspaceNode) bool {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return true
	}

	if node.CaseSlug == scope.CaseSlug {
		return true
	}

	if strings.TrimSpace(node.CaseSlug) != "" {
		return false
	}

	return node.Kind == core.WorkspaceNodeDump
}

func sortedSpecContextNodeIDs(index core.WorkspaceIndex) []string {
	return sortedGenerationContextNodeIDs(index, []core.WorkspaceNodeKind{
		core.WorkspaceNodeDump,
		core.WorkspaceNodeQuiz,
	})
}

func sortedGenerationContextNodeIDs(index core.WorkspaceIndex, kinds []core.WorkspaceNodeKind) []string {
	ids := []string{}

	for _, kind := range kinds {
		ids = append(ids, index.ByKind[kind]...)
	}

	sort.Strings(ids)

	return ids
}

func appendSpecNodeContextSection(
	sections []quizmodel.ContextSection,
	projectDir string,
	node core.WorkspaceNode,
) []quizmodel.ContextSection {
	content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(node.RelativePath)))
	if err != nil {
		return sections
	}

	return append(sections, quizmodel.ContextSection{
		Name:      string(node.Kind),
		Layer:     string(node.Kind),
		Authority: node.RelativePath,
		Content:   string(content),
		Metadata: map[string]string{
			"case_slug": node.CaseSlug,
			"title":     node.Title,
			"status":    node.Status,
		},
	})
}

//nolint:wsl_v5 // Straight extraction loop is easier to read without extra helpers.
func specSourcePaths(sections []quizmodel.ContextSection) []string {
	paths := []string{}
	for _, section := range sections {
		if strings.TrimSpace(section.Authority) == "" {
			continue
		}

		paths = append(paths, section.Authority)
	}

	return paths
}

func specScopeID(scope quizcore.Scope) string {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "project"
	}

	return scope.CaseSlug
}

func specUseCaseFallbackTitle(scope quizcore.Scope) string {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "Project-wide"
	}

	return scope.CaseSlug
}

type workspaceMarkdownSummary struct {
	Title   string
	Status  string
	Preview string
}

func summarizeWorkspaceMarkdownDocument(content, fallbackName string) workspaceMarkdownSummary {
	title := strings.TrimSuffix(fallbackName, filepath.Ext(fallbackName))
	status := "unknown"
	preview := ""

	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || line == "---" {
			continue
		}

		if strings.HasPrefix(line, "# ") && title == strings.TrimSuffix(fallbackName, filepath.Ext(fallbackName)) {
			title = strings.TrimSpace(strings.TrimPrefix(line, "# "))

			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "status") {
			status = strings.Trim(strings.TrimSpace(value), `"`)

			continue
		}

		if preview == "" {
			preview = line
		}
	}

	return workspaceMarkdownSummary{Title: title, Status: status, Preview: preview}
}

func fileModifiedAt(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}

	return info.ModTime().UTC()
}
