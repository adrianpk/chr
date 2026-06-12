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

	"forge.adrianpk.com/chc/internal/core"
)

// SpecView is the TUI-facing projection of a spec document.
type SpecView struct {
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        core.QuizScope
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
	Scope core.QuizScope
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
		var err error

		result, err = s.generateSpec(ctx, scope)

		return err
	})

	return result, err
}

//nolint:funlen,wsl_v5 // Generation flow is intentionally linear for logging and failure attribution.
func (s service) generateSpec(ctx context.Context, scope core.QuizScope) (SpecOperationResult, error) {
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

	response, err := core.InvokeQuizModel(ctx, s.quizModel, modelRequest)
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

	spec := core.GeneratedSpec{
		Scope:            scope,
		Title:            response.DraftSpec.Title,
		Body:             response.DraftSpec.Body,
		SourcePaths:      specSourcePaths(modelRequest.ContextSections),
		SourceRequestID:  modelRequest.ID,
		SourceResponseID: response.ID,
		CreatedAt:        response.CreatedAt.UTC(),
	}

	paths, err := core.WriteGeneratedSpec(ctx, s.projectContext.ProjectDir, spec)
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

//nolint:wsl_v5 // Tight loop keeps filtering and candidate assignment together.
func (s service) latestSpecNodeForScope(scope core.QuizScope, index core.WorkspaceIndex) (core.WorkspaceNode, bool) {
	nodes := []core.WorkspaceNode{}
	for _, nodeID := range index.ByKind[core.WorkspaceNodeSpec] {
		node, ok := index.NodesByID[nodeID]
		if !ok {
			continue
		}

		if node.CaseSlug != scope.CaseSlug {
			continue
		}

		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return core.WorkspaceNode{}, false
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].ModifiedAt.After(nodes[j].ModifiedAt)
	})

	return nodes[0], true
}

func (s service) specViewFromNode(scope core.QuizScope, fallbackTitle string, node core.WorkspaceNode) SpecView {
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

func (s service) specViewFromPath(scope core.QuizScope, relativePath, path string) SpecView {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		content = nil
	} else if err != nil {
		content = []byte("Unable to read spec: " + err.Error())
	}

	title, status, _ := summarizeWorkspaceMarkdownDocument(string(content), filepath.Base(path))

	return SpecView{
		UseCaseTitle: specUseCaseFallbackTitle(scope),
		Title:        title,
		Status:       status,
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

func (s service) specGenerationRuntimeState(scope core.QuizScope) (core.QuizRuntimeState, error) {
	state, err := s.quizRepository().LoadRuntimeState(s.ctx, scope)
	if err == nil {
		return state, nil
	}

	state, err = s.newQuizRuntimeState(scope, core.QuizModeDeep)
	if err != nil {
		return core.QuizRuntimeState{}, err
	}

	state.Status = core.QuizRuntimeStatusDone
	state.ActiveQuestion = nil
	state.UpdatedAt = s.utcNow()

	return state, nil
}

func (s service) newSpecGenerationModelRequest(state core.QuizRuntimeState) (core.QuizModelRequest, error) {
	scope := core.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, core.QuizModelActionGenerateSpec)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	sections, err := core.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	sections, err = s.appendSpecGenerationContextSections(sections, scope)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	return core.QuizModelRequest{
		ID:              fmt.Sprintf("spec-generation-%s", specScopeID(scope)),
		Action:          core.QuizModelActionGenerateSpec,
		Strategy:        "spec_generation",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: core.QuizModelOutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) appendSpecGenerationContextSections(
	sections []core.QuizModelContextSection,
	scope core.QuizScope,
) ([]core.QuizModelContextSection, error) {
	return s.appendGenerationContextSections(sections, scope, sortedSpecContextNodeIDs)
}

func (s service) appendGenerationContextSections(
	sections []core.QuizModelContextSection,
	scope core.QuizScope,
	nodeIDs func(core.WorkspaceIndex) []string,
) ([]core.QuizModelContextSection, error) {
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "agent_protocol", "protocol", "AGENTS.md")
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "project_document", "project", core.ProjectMetadataPath(""))

	if scope.CaseSlug != "" {
		sections = appendQuizFileContextSection(
			sections,
			s.projectContext.ProjectDir,
			"use_case_document",
			"use_case",
			core.CaseMetadataPath("", scope.CaseSlug),
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

func generationContextNodeAllowed(scope core.QuizScope, node core.WorkspaceNode) bool {
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

func (s service) appendPlanGenerationContextSections(
	sections []core.QuizModelContextSection,
	scope core.QuizScope,
) ([]core.QuizModelContextSection, error) {
	return s.appendGenerationContextSections(sections, scope, sortedPlanContextNodeIDs)
}

func sortedPlanContextNodeIDs(index core.WorkspaceIndex) []string {
	return sortedGenerationContextNodeIDs(index, []core.WorkspaceNodeKind{
		core.WorkspaceNodeDump,
		core.WorkspaceNodeQuiz,
		core.WorkspaceNodeSpec,
		core.WorkspaceNodePlan,
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
	sections []core.QuizModelContextSection,
	projectDir string,
	node core.WorkspaceNode,
) []core.QuizModelContextSection {
	content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(node.RelativePath)))
	if err != nil {
		return sections
	}

	return append(sections, core.QuizModelContextSection{
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
func specSourcePaths(sections []core.QuizModelContextSection) []string {
	paths := []string{}
	for _, section := range sections {
		if strings.TrimSpace(section.Authority) == "" {
			continue
		}

		paths = append(paths, section.Authority)
	}

	return paths
}

func specScopeID(scope core.QuizScope) string {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "project"
	}

	return scope.CaseSlug
}

func specUseCaseFallbackTitle(scope core.QuizScope) string {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "Project-wide"
	}

	return scope.CaseSlug
}

func summarizeWorkspaceMarkdownDocument(content, fallbackName string) (string, string, string) {
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

	return title, status, preview
}

func fileModifiedAt(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}

	return info.ModTime().UTC()
}
