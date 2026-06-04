package chcapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

// PlanView is the TUI-facing projection of a generated plan document.
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

// PlanListResult contains the current plan list and a status message.
type PlanListResult struct {
	Message string
	Plans   []PlanView
}

// PlanGenerateRequest selects the quiz scope used to generate a plan.
type PlanGenerateRequest struct {
	Scope chc.QuizScope
}

// PlanRegenerateRequest selects the quiz scope used to replace a plan.
type PlanRegenerateRequest struct {
	Scope chc.QuizScope
}

// PlanTrackingRequest selects the plan scope used to materialize tracking
// artifacts.
type PlanTrackingRequest struct {
	Scope chc.QuizScope
}

// PlanOperationResult is the TUI-facing result for plan generation operations.
type PlanOperationResult struct {
	Message string
	Plan    PlanView
}

// PlanTrackingResult reports generated tracking artifacts for a plan.
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

	plan, requestIDs, responseIDs, err := s.generatePlanWithConfiguredBackend(ctx, scope, state)
	if err != nil {
		s.appendPlanGenerationLog("generate_plan_model_generation_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return PlanOperationResult{}, err
	}

	plan.SourceRequestID = strings.Join(requestIDs, ",")
	plan.SourceResponseID = strings.Join(responseIDs, ",")

	paths, err := chc.WriteGeneratedPlan(ctx, s.projectContext.ProjectDir, plan)
	if err != nil {
		s.appendPlanGenerationLog("generate_plan_write_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return PlanOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_written", map[string]string{
		"case_slug":     scope.CaseSlug,
		"request_ids":   strings.Join(requestIDs, ","),
		"response_ids":  strings.Join(responseIDs, ","),
		"relative_path": paths.RelativePath,
		"path":          paths.Path,
	})

	view := s.planViewFromPath(scope, paths.RelativePath, paths.Path)

	return PlanOperationResult{
		Message: "Plan generated.",
		Plan:    view,
	}, nil
}

func (s service) generatePlanWithConfiguredBackend(
	ctx context.Context,
	scope chc.QuizScope,
	state chc.QuizRuntimeState,
) (chc.GeneratedPlan, []string, []string, error) {
	if codexGenerationBackend("CHC_PLAN_GENERATION_BACKEND") {
		return s.generateSinglePassPlan(ctx, scope, state)
	}

	return s.generateStagedPlan(ctx, scope, state)
}

func (s service) generateSinglePassPlan(
	ctx context.Context,
	scope chc.QuizScope,
	state chc.QuizRuntimeState,
) (chc.GeneratedPlan, []string, []string, error) {
	modelRequest, err := s.newPlanGenerationModelRequest(state)
	if err != nil {
		return chc.GeneratedPlan{}, nil, nil, err
	}

	s.appendPlanGenerationLog("generate_plan_single_pass_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := chc.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		return chc.GeneratedPlan{}, nil, nil, err
	}

	plan := response.Plan
	plan.Scope = scope
	plan.SourcePaths = specSourcePaths(modelRequest.ContextSections)
	plan.CreatedAt = response.CreatedAt.UTC()

	s.appendPlanGenerationLog("generate_plan_single_pass_finished", map[string]string{
		"case_slug":   scope.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"title":       plan.Title,
	})

	return plan, []string{modelRequest.ID}, []string{response.ID}, nil
}

func (s service) generateStagedPlan(
	ctx context.Context,
	scope chc.QuizScope,
	state chc.QuizRuntimeState,
) (chc.GeneratedPlan, []string, []string, error) {
	outline, sourcePaths, requestIDs, responseIDs, err := s.generatePlanOutline(ctx, scope, state)
	if err != nil {
		return chc.GeneratedPlan{}, nil, nil, err
	}

	details := make([]chc.GeneratedPlanSliceDetail, 0, len(outline.Slices))
	for _, slice := range outline.Slices {
		detail, requestID, responseID, err := s.generatePlanSliceDetail(ctx, state, outline, slice.ID)
		if err != nil {
			return chc.GeneratedPlan{}, nil, nil, err
		}

		details = append(details, detail)
		requestIDs = append(requestIDs, requestID)
		responseIDs = append(responseIDs, responseID)
	}

	body, err := chc.AssembleGeneratedPlanBody(outline, details, requestIDs, responseIDs)
	if err != nil {
		return chc.GeneratedPlan{}, nil, nil, err
	}

	plan := chc.GeneratedPlan{
		Scope:       scope,
		Title:       outline.Title,
		Body:        body,
		SourcePaths: sourcePaths,
		CreatedAt:   outline.CreatedAt.UTC(),
	}

	return plan, requestIDs, responseIDs, nil
}

func (s service) generatePlanOutline(
	ctx context.Context,
	scope chc.QuizScope,
	state chc.QuizRuntimeState,
) (chc.GeneratedPlanOutline, []string, []string, []string, error) {
	modelRequest, err := s.newPlanOutlineGenerationModelRequest(state)
	if err != nil {
		return chc.GeneratedPlanOutline{}, nil, nil, nil, err
	}

	s.appendPlanGenerationLog("generate_plan_outline_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := chc.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		return chc.GeneratedPlanOutline{}, nil, nil, nil, err
	}

	outline := response.PlanOutline
	outline.Scope = scope
	outline.CreatedAt = response.CreatedAt.UTC()

	s.appendPlanGenerationLog("generate_plan_outline_finished", map[string]string{
		"case_slug":   scope.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"title":       outline.Title,
		"slices":      fmt.Sprintf("%d", len(outline.Slices)),
	})

	return outline,
		specSourcePaths(modelRequest.ContextSections),
		[]string{modelRequest.ID},
		[]string{response.ID},
		nil
}

func (s service) generatePlanSliceDetail(
	ctx context.Context,
	state chc.QuizRuntimeState,
	outline chc.GeneratedPlanOutline,
	targetSliceID string,
) (chc.GeneratedPlanSliceDetail, string, string, error) {
	modelRequest, err := s.newPlanSliceDetailGenerationModelRequest(state, outline, targetSliceID)
	if err != nil {
		return chc.GeneratedPlanSliceDetail{}, "", "", err
	}

	s.appendPlanGenerationLog("generate_plan_slice_detail_started", map[string]string{
		"case_slug":  state.CaseSlug,
		"request_id": modelRequest.ID,
		"slice_id":   targetSliceID,
	})

	response, err := chc.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		return chc.GeneratedPlanSliceDetail{}, "", "", err
	}

	detail := response.SliceDetail
	detail.Scope = outline.Scope
	detail.CreatedAt = response.CreatedAt.UTC()

	if err := chc.ValidateGeneratedPlanSliceDetailAgainstOutline(detail, outline, targetSliceID); err != nil {
		return chc.GeneratedPlanSliceDetail{}, "", "", err
	}

	s.appendPlanGenerationLog("generate_plan_slice_detail_finished", map[string]string{
		"case_slug":   state.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"slice_id":    targetSliceID,
	})

	return detail, modelRequest.ID, response.ID, nil
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
	if err := validatePlanTrackingTicketPayloads(ticketPayloads); err != nil {
		return nil, err
	}

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

func validatePlanTrackingTicketPayloads(ticketPayloads []chc.GeneratedPlanTrackingTicket) error {
	if len(ticketPayloads) == 0 {
		return fmt.Errorf("plan tracking ticket payloads are required")
	}

	for index, payload := range ticketPayloads {
		if err := validatePlanTrackingTicketPayload(payload); err != nil {
			return fmt.Errorf("plan tracking ticket payload %d: %w", index+1, err)
		}
	}

	return nil
}

func validatePlanTrackingTicketPayload(payload chc.GeneratedPlanTrackingTicket) error {
	switch {
	case strings.TrimSpace(payload.Key) != "":
		return fmt.Errorf("generated ticket key must be empty")
	case strings.TrimSpace(payload.Path) != "":
		return fmt.Errorf("generated ticket path must be empty")
	case strings.TrimSpace(payload.SliceID) == "":
		return fmt.Errorf("slice id is required")
	case strings.TrimSpace(payload.Title) == "":
		return fmt.Errorf("ticket title is required")
	case strings.TrimSpace(payload.Synopsis) == "":
		return fmt.Errorf("ticket synopsis is required")
	case strings.TrimSpace(payload.Summary) == "":
		return fmt.Errorf("ticket summary is required")
	case strings.TrimSpace(payload.Body) == "":
		return fmt.Errorf("ticket body is required")
	}

	return nil
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

func (s service) newPlanOutlineGenerationModelRequest(state chc.QuizRuntimeState) (chc.QuizModelRequest, error) {
	return s.newPlanningModelRequest(
		state,
		chc.QuizModelActionGeneratePlanOutline,
		"plan_outline_generation",
		"plan-outline-generation",
	)
}

func (s service) newPlanGenerationModelRequest(state chc.QuizRuntimeState) (chc.QuizModelRequest, error) {
	return s.newPlanningModelRequest(
		state,
		chc.QuizModelActionGeneratePlan,
		"plan_generation",
		"plan-generation",
	)
}

func (s service) newPlanSliceDetailGenerationModelRequest(
	state chc.QuizRuntimeState,
	outline chc.GeneratedPlanOutline,
	targetSliceID string,
) (chc.QuizModelRequest, error) {
	scope := chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	if err := chc.ValidateGeneratedPlanOutline(outline); err != nil {
		return chc.QuizModelRequest{}, err
	}

	targetSlice, ok := outline.Slice(targetSliceID)
	if !ok {
		return chc.QuizModelRequest{}, fmt.Errorf("plan outline slice is unknown: %s", targetSliceID)
	}

	payload, err := s.newQuizModelPayload(scope, state, chc.QuizModelActionGeneratePlanSliceDetail)
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

	sections, err = appendPlanSliceDetailContextSections(sections, outline, targetSlice)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	return chc.QuizModelRequest{
		ID:              fmt.Sprintf("plan-slice-detail-generation-%s-%s", specScopeID(scope), planSliceIDSlug(targetSliceID)),
		Action:          chc.QuizModelActionGeneratePlanSliceDetail,
		Strategy:        "plan_slice_detail_generation",
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

func (s service) newPlanningModelRequest(
	state chc.QuizRuntimeState,
	action chc.QuizModelAction,
	strategy string,
	idPrefix string,
) (chc.QuizModelRequest, error) {
	scope := chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, action)
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
		ID:              fmt.Sprintf("%s-%s", idPrefix, specScopeID(scope)),
		Action:          action,
		Strategy:        strategy,
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

func appendPlanSliceDetailContextSections(
	sections []chc.QuizModelContextSection,
	outline chc.GeneratedPlanOutline,
	targetSlice chc.GeneratedPlanOutlineSlice,
) ([]chc.QuizModelContextSection, error) {
	outlineContent, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		return nil, err
	}

	targetContent, err := json.MarshalIndent(targetSlice, "", "  ")
	if err != nil {
		return nil, err
	}

	sections = append(sections, chc.QuizModelContextSection{
		Name:      "plan_outline",
		Layer:     "plan",
		Authority: "generated_plan_outline",
		Content:   string(outlineContent),
		Metadata: map[string]string{
			"title":  outline.Title,
			"status": outline.Status,
		},
	})
	sections = append(sections, chc.QuizModelContextSection{
		Name:      "target_slice",
		Layer:     "plan",
		Authority: targetSlice.ID,
		Content:   string(targetContent),
		Metadata: map[string]string{
			"slice_id":    targetSlice.ID,
			"short_name":  targetSlice.ShortName,
			"branch_name": targetSlice.BranchName,
		},
	})

	return sections, nil
}

func planSliceIDSlug(sliceID string) string {
	slug := strings.ToLower(strings.TrimSpace(sliceID))
	slug = strings.ReplaceAll(slug, " ", "-")

	return slug
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
