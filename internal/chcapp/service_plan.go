package chcapp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type PlanView struct {
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        chc.QuizScope
	RelativePath string
	UpdatedAt    time.Time
	Content      string
}

type PlanListResult struct {
	Message string
	Plans   []PlanView
}

type PlanGenerateRequest struct {
	Scope chc.QuizScope
}

type PlanOperationResult struct {
	Message string
	Plan    PlanView
}

func (s service) ListPlans() (PlanListResult, error) {
	scopes, err := s.quizReadScopes()
	if err != nil {
		return PlanListResult{}, err
	}

	index, err := s.workspaceIndex()
	if err != nil {
		return PlanListResult{}, err
	}

	plans := make([]PlanView, 0, len(scopes))
	for _, scope := range scopes {
		plans = append(plans, s.planViewFromScope(scope, index))
	}

	for index := range plans {
		plans[index].Index = index
	}

	return PlanListResult{Message: "Plans loaded.", Plans: plans}, nil
}

func (s service) GeneratePlan(request PlanGenerateRequest) (PlanOperationResult, error) {
	scope := s.quizScope(request.Scope)
	s.appendPlanGenerationLog("generate_plan_started", map[string]string{
		"case_slug": scope.CaseSlug,
	})

	state, err := s.specGenerationRuntimeState(scope)
	if err != nil {
		s.appendPlanGenerationLog("generate_plan_runtime_state_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return PlanOperationResult{}, err
	}

	modelRequest, err := s.newPlanGenerationModelRequest(state)
	if err != nil {
		s.appendPlanGenerationLog("generate_plan_request_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return PlanOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_model_invocation_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := chc.InvokeQuizModel(s.ctx, s.quizModel, modelRequest)
	if err != nil {
		s.appendPlanGenerationLog("generate_plan_model_invocation_failed", map[string]string{
			"case_slug":  scope.CaseSlug,
			"request_id": modelRequest.ID,
			"error":      err.Error(),
		})

		return PlanOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_model_invocation_finished", map[string]string{
		"case_slug":   scope.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"title":       response.Plan.Title,
	})

	plan := chc.GeneratedPlan{
		Scope:            scope,
		Title:            response.Plan.Title,
		Body:             response.Plan.Body,
		SourcePaths:      specSourcePaths(modelRequest.ContextSections),
		SourceRequestID:  modelRequest.ID,
		SourceResponseID: response.ID,
		CreatedAt:        response.CreatedAt.UTC(),
	}

	paths, err := chc.WriteGeneratedPlan(s.ctx, s.projectContext.ProjectDir, plan)
	if err != nil {
		s.appendPlanGenerationLog("generate_plan_write_failed", map[string]string{
			"case_slug":   scope.CaseSlug,
			"request_id":  modelRequest.ID,
			"response_id": response.ID,
			"error":       err.Error(),
		})

		return PlanOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_written", map[string]string{
		"case_slug":     scope.CaseSlug,
		"request_id":    modelRequest.ID,
		"response_id":   response.ID,
		"relative_path": paths.RelativePath,
		"path":          paths.Path,
	})

	view := s.planViewFromPath(scope, paths.RelativePath, paths.Path)

	return PlanOperationResult{
		Message: "Plan generated.",
		Plan:    view,
	}, nil
}

func (s service) planViewFromScope(scope quizReadScope, index chc.WorkspaceIndex) PlanView {
	node, ok := s.latestPlanNodeForScope(scope.scope, index)
	if !ok {
		return PlanView{
			UseCaseTitle: scope.title,
			Title:        scope.title,
			Status:       "missing",
			Scope:        scope.scope,
		}
	}

	return s.planViewFromNode(scope.scope, scope.title, node)
}

func (s service) latestPlanNodeForScope(scope chc.QuizScope, index chc.WorkspaceIndex) (chc.WorkspaceNode, bool) {
	nodes := []chc.WorkspaceNode{}
	for _, nodeID := range index.ByKind[chc.WorkspaceNodePlan] {
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
		return chc.WorkspaceNode{}, false
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].ModifiedAt.After(nodes[j].ModifiedAt)
	})

	return nodes[0], true
}

func (s service) planViewFromNode(scope chc.QuizScope, fallbackTitle string, node chc.WorkspaceNode) PlanView {
	title := node.Title
	if strings.TrimSpace(title) == "" {
		title = fallbackTitle
	}

	return s.planViewFromPath(scope, node.RelativePath, filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(node.RelativePath))).withMetadata(
		fallbackTitle,
		title,
		node.Status,
		node.ModifiedAt,
	)
}

func (s service) planViewFromPath(scope chc.QuizScope, relativePath string, path string) PlanView {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		content = nil
	} else if err != nil {
		content = []byte("Unable to read plan: " + err.Error())
	}

	title, status, _ := summarizeWorkspaceMarkdownDocument(string(content), filepath.Base(path))

	return PlanView{
		UseCaseTitle: specUseCaseFallbackTitle(scope),
		Title:        title,
		Status:       status,
		Scope:        scope,
		RelativePath: filepath.ToSlash(relativePath),
		UpdatedAt:    fileModifiedAt(path),
		Content:      string(content),
	}
}

func (v PlanView) withMetadata(useCaseTitle, title, status string, updatedAt time.Time) PlanView {
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

func (s service) newPlanGenerationModelRequest(state chc.QuizRuntimeState) (chc.QuizModelRequest, error) {
	scope := chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, chc.QuizModelActionGeneratePlan)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err := chc.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err = s.appendSpecGenerationContextSections(sections, scope)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	if scope.CaseSlug == "" {
		planningSections, err := s.projectWidePlanningSections()
		if err != nil {
			return chc.QuizModelRequest{}, err
		}
		sections = append(sections, planningSections...)
	}

	return chc.QuizModelRequest{
		ID:              fmt.Sprintf("plan-generation-%s", specScopeID(scope)),
		Action:          chc.QuizModelActionGeneratePlan,
		Strategy:        "plan_generation",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: chc.QuizModelOutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}
