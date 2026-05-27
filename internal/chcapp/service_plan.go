package chcapp

import (
	"context"
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

type PlanRegenerateRequest struct {
	Scope chc.QuizScope
}

type PlanTrackingRequest struct {
	Scope chc.QuizScope
}

type PlanOperationResult struct {
	Message string
	Plan    PlanView
}

type PlanTrackingResult struct {
	Message      string
	Plan         PlanView
	TrackingPath string
	Tickets      []chc.IndexedTicket
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

//nolint:wsl_v5 // Guard clauses read clearer in this service flow.
func (s service) GeneratePlan(ctx context.Context, request PlanGenerateRequest) (PlanOperationResult, error) {
	scope := s.quizScope(request.Scope)
	existing, exists, err := s.existingGeneratedPlan(scope)
	if err != nil {
		return PlanOperationResult{}, err
	}

	if exists {
		return PlanOperationResult{
			Message: "Plan already exists. Use regenerate plan to replace it.",
			Plan:    existing,
		}, nil
	}

	var result PlanOperationResult

	err = s.withGenerationRuntime(ctx, "plan", scope, func(ctx context.Context) error {
		var generateErr error

		result, generateErr = s.generatePlan(ctx, scope)

		return generateErr
	})

	return result, err
}

func (s service) existingGeneratedPlan(scope chc.QuizScope) (PlanView, bool, error) {
	paths, err := chc.GeneratedPlanLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return PlanView{}, false, err
	}

	_, err = os.Stat(paths.Path)
	if errors.Is(err, os.ErrNotExist) {
		return PlanView{}, false, nil
	}

	if err != nil {
		return PlanView{}, false, fmt.Errorf("stat generated plan: %w", err)
	}

	return s.planViewFromPath(scope, paths.RelativePath, paths.Path), true, nil
}

//nolint:funlen // Generation flow is intentionally linear for logging and failure attribution.
func (s service) generatePlan(ctx context.Context, scope chc.QuizScope) (PlanOperationResult, error) {
	s.appendPlanGenerationLog("generate_plan_started", map[string]string{
		"case_slug": scope.CaseSlug,
		"timeout":   longGenerationTimeout.String(),
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

	response, err := chc.InvokeQuizModel(ctx, s.quizModel, modelRequest)
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

	paths, err := chc.WriteGeneratedPlan(ctx, s.projectContext.ProjectDir, plan)
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

//nolint:wsl_v5 // Keeps result capture adjacent to the guarded runtime call.
func (s service) RegeneratePlan(ctx context.Context, request PlanRegenerateRequest) (PlanOperationResult, error) {
	scope := s.quizScope(request.Scope)
	var result PlanOperationResult

	err := s.withGenerationRuntime(ctx, "plan", scope, func(ctx context.Context) error {
		var err error

		result, err = s.regeneratePlan(ctx, scope)

		return err
	})

	return result, err
}

func (s service) regeneratePlan(ctx context.Context, scope chc.QuizScope) (PlanOperationResult, error) {
	archivedAt := s.utcNow()

	s.appendPlanGenerationLog("regenerate_plan_started", map[string]string{
		"case_slug": scope.CaseSlug,
		"timeout":   longGenerationTimeout.String(),
	})

	archivePaths, archived, err := chc.ArchiveGeneratedPlan(
		ctx,
		s.projectContext.ProjectDir,
		scope,
		archivedAt,
	)
	if err != nil {
		s.appendPlanGenerationLog("regenerate_plan_archive_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return PlanOperationResult{}, err
	}

	if archived {
		s.appendPlanGenerationLog("regenerate_plan_archived", map[string]string{
			"case_slug":     scope.CaseSlug,
			"relative_path": archivePaths.RelativePath,
			"path":          archivePaths.Path,
		})
	}

	result, err := s.generatePlan(ctx, scope)
	if err != nil {
		return PlanOperationResult{}, err
	}

	result.Message = "Plan regenerated."

	return result, nil
}

func (s service) MaterializePlanTracking(ctx context.Context, request PlanTrackingRequest) (PlanTrackingResult, error) {
	scope := s.quizScope(request.Scope)

	plan, exists, err := s.existingGeneratedPlan(scope)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	if !exists {
		return PlanTrackingResult{}, fmt.Errorf("plan is required before tracking can be created")
	}

	trackingPaths, err := chc.GeneratedPlanTrackingLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	_, err = os.Stat(trackingPaths.Path)
	if err == nil {
		return PlanTrackingResult{
			Message:      "Plan tracking already exists.",
			Plan:         plan,
			TrackingPath: trackingPaths.RelativePath,
		}, nil
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return PlanTrackingResult{}, fmt.Errorf("stat plan tracking: %w", err)
	}

	var result PlanTrackingResult

	err = s.withGenerationRuntime(ctx, "plan_tracking", scope, func(ctx context.Context) error {
		var materializeErr error

		result, materializeErr = s.materializePlanTracking(ctx, scope, plan, trackingPaths.RelativePath)

		return materializeErr
	})

	return result, err
}

func (s service) materializePlanTracking(
	ctx context.Context,
	scope chc.QuizScope,
	plan PlanView,
	trackingRelativePath string,
) (PlanTrackingResult, error) {
	state, err := s.specGenerationRuntimeState(scope)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	modelRequest, err := s.newPlanTrackingModelRequest(state, plan)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	response, err := chc.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	tracking := response.PlanTracking
	tracking.Scope = scope
	tracking.PlanPath = plan.RelativePath
	tracking.SourcePaths = specSourcePaths(modelRequest.ContextSections)
	tracking.CreatedAt = modelRequest.CreatedAt

	createdTickets, err := s.createPlanTrackingTickets(scope, plan.RelativePath, trackingRelativePath, tracking.Tickets)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	tracking.Tickets = createdTickets

	paths, err := chc.WriteGeneratedPlanTracking(ctx, s.projectContext.ProjectDir, tracking)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	tickets, err := s.listProjectTickets()
	if err != nil {
		return PlanTrackingResult{}, err
	}

	return PlanTrackingResult{
		Message:      fmt.Sprintf("Created plan tracking and %d tickets.", len(createdTickets)),
		Plan:         s.planViewFromPath(scope, plan.RelativePath, filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(plan.RelativePath))),
		TrackingPath: paths.RelativePath,
		Tickets:      tickets,
	}, nil
}

func (s service) createPlanTrackingTickets(
	scope chc.QuizScope,
	planPath string,
	trackingPath string,
	ticketPayloads []chc.GeneratedPlanTrackingTicket,
) ([]chc.GeneratedPlanTrackingTicket, error) {
	created := make([]chc.GeneratedPlanTrackingTicket, 0, len(ticketPayloads))

	for _, payload := range ticketPayloads {
		result, err := s.CreateTicket(CreateTicketRequest{
			Title:       payload.Title,
			Synopsis:    payload.Synopsis,
			Status:      chc.TicketStatusReadyForDev,
			Priority:    payload.Priority,
			Severity:    payload.Severity,
			Type:        payload.Type,
			Tags:        payload.Tags,
			UseCase:     scope.CaseSlug,
			ProjectArea: "plan",
			SourceKind:  "plan_tracking",
			SourceRef:   planPath,
			Slice:       payload.SliceID,
			Body:        renderPlanTrackingTicketBody(planPath, trackingPath, payload),
		})
		if err != nil {
			return nil, err
		}

		payload.Key = result.Key
		payload.Path = relativeProjectPath(s.projectContext.ProjectDir, result.Path)
		created = append(created, payload)
	}

	return created, nil
}

//nolint:wsl_v5 // Markdown body construction benefits from local grouping.
func renderPlanTrackingTicketBody(planPath, trackingPath string, ticket chc.GeneratedPlanTrackingTicket) string {
	var output strings.Builder

	output.WriteString("# Description\n\n")
	output.WriteString(strings.TrimSpace(ticket.Summary))
	output.WriteString("\n\n")
	output.WriteString("Source plan: `")
	output.WriteString(filepath.ToSlash(planPath))
	output.WriteString("`\n")
	output.WriteString("Tracking: `")
	output.WriteString(filepath.ToSlash(trackingPath))
	output.WriteString("`\n")
	output.WriteString("Slice: ")
	output.WriteString(ticket.SliceID)
	output.WriteString("\n\n")
	output.WriteString(strings.TrimSpace(ticket.Body))
	output.WriteString("\n\n")
	output.WriteString("# Dependencies\n\n")
	if len(ticket.Dependencies) == 0 {
		output.WriteString("None.\n\n")
	} else {
		for _, dependency := range ticket.Dependencies {
			output.WriteString("- ")
			output.WriteString(dependency)
			output.WriteString("\n")
		}
		output.WriteString("\n")
	}
	output.WriteString("# Comments\n\n")
	output.WriteString("# Blocker\n\n")
	output.WriteString("# Validation\n")

	return output.String()
}

func relativeProjectPath(projectDir, path string) string {
	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(relativePath)
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

//nolint:wsl_v5 // Tight loop keeps filtering and candidate assignment together.
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

func (s service) planViewFromPath(scope chc.QuizScope, relativePath, path string) PlanView {
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

//nolint:wsl_v5 // Request assembly keeps optional project-wide sections adjacent.
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

	sections, err = s.appendPlanGenerationContextSections(sections, scope)
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

func (s service) newPlanTrackingModelRequest(
	state chc.QuizRuntimeState,
	plan PlanView,
) (chc.QuizModelRequest, error) {
	scope := chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, chc.QuizModelActionMaterializePlanTracking)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err := chc.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections = append(sections, chc.QuizModelContextSection{
		Name:      "implementation_plan",
		Layer:     "plan",
		Authority: plan.RelativePath,
		Content:   plan.Content,
		Metadata: map[string]string{
			"case_slug": scope.CaseSlug,
			"title":     plan.Title,
			"status":    plan.Status,
		},
	})

	return chc.QuizModelRequest{
		ID:              fmt.Sprintf("plan-tracking-%s", specScopeID(scope)),
		Action:          chc.QuizModelActionMaterializePlanTracking,
		Strategy:        "plan_tracking_materialization",
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
