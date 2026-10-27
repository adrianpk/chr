package app

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

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlanView is the TUI-facing projection of a generated plan document.
type PlanView struct {
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        quizcore.QuizScope
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
	TicketDelivery     bool
	OperationID        string
	TrackerReplaceHash string
	ExpectedRevision   string
	Scope              quizcore.QuizScope
}

// PlanRegenerateRequest selects the quiz scope used to replace a plan.
type PlanRegenerateRequest struct {
	OperationID        string
	TrackerReplaceHash string
	ExpectedRevision   string
	Scope              quizcore.QuizScope
}

// PlanTrackingRequest selects the plan scope used to materialize tracking
// artifacts.
type PlanTrackingRequest struct {
	OperationID string
	Workflow    *core.GeneratedWorkflowPlan
	Scope       quizcore.QuizScope
}

// PlanOperationResult is the TUI-facing result for plan generation operations.
type PlanOperationResult struct {
	Existing bool
	Message  string
	Plan     PlanView
}

// PlanTrackingResult reports generated tracking artifacts for a plan.
type PlanTrackingResult struct {
	Workflow     *core.WorkflowMaterializationResult
	Message      string
	Plan         PlanView
	TrackingPath string
	Tickets      []core.IndexedTicket
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
	operation := core.PlanPublicationOperation{ID: request.OperationID, Scope: scope, ExpectedRevision: request.ExpectedRevision, TrackerReplaceHash: request.TrackerReplaceHash, TicketDelivery: request.TicketDelivery}
	if result, found, err := s.replayPlanPublication(ctx, operation); found || err != nil {
		return result, err
	}
	existing, exists, err := s.existingGeneratedPlan(scope)
	if err != nil {
		return PlanOperationResult{}, err
	}

	if exists {
		return PlanOperationResult{
			Existing: true,
			Message:  "Plan already exists. Use regenerate plan to replace it.",
			Plan:     existing,
		}, nil
	}

	var result PlanOperationResult

	err = s.withGenerationRuntime(ctx, "plan", scope, func(ctx context.Context) error {
		var generateErr error

		result, generateErr = s.prepareAndPublishPlan(ctx, operation)

		return generateErr
	})

	return result, err
}

func (s service) existingGeneratedPlan(scope quizcore.QuizScope) (PlanView, bool, error) {
	paths, err := core.GeneratedPlanLayout(s.projectContext.ProjectDir, scope)
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

func (s service) prepareAndPublishPlan(ctx context.Context, operation core.PlanPublicationOperation) (PlanOperationResult, error) {
	bound, ctx, err := s.bindGeneration(ctx, "plan-"+dumpRequestID(operation.ID), string(core.QuizModelActionGeneratePlan))
	if err != nil {
		return PlanOperationResult{}, err
	}

	s = bound
	scope, expected, replace := operation.Scope, operation.ExpectedRevision, operation.Replace
	s.appendPlanGenerationLog("generate_plan_started", map[string]string{"case_slug": scope.CaseSlug, "timeout": longGenerationTimeout.String()})

	sources, state, target, err := s.preparePlanGeneration(ctx, scope, expected)
	if err != nil {
		return PlanOperationResult{}, err
	}

	plan, requests, responses, err := s.generatePlanWithConfiguredBackend(ctx, scope, state, sources)
	if err != nil {
		return PlanOperationResult{}, err
	}

	plan.SourceRequestID, plan.SourceResponseID = strings.Join(requests, ","), strings.Join(responses, ",")
	plan.ExpectedRevision, plan.SourceObservationHash = target, sources.Hash

	plan.RequireOrderMap = true
	plan.Order = core.ScopePlanDeliveryOrder(plan.Order, operation)

	plan.TrackerReplaceHash = operation.TrackerReplaceHash
	if operation.ID != "" {
		plan.Operation = &operation
	}

	plan.CreateOnly = !replace
	for _, input := range sources.Inputs {
		plan.SourceInputs = append(plan.SourceInputs, core.PlanSourceBinding{Kind: input.Kind, Reference: input.Reference, Revision: input.Revision})
	}

	if replace {
		plan.ArchiveAt = s.utcNow()
	}

	if err := core.ValidateGeneratedPlan(plan); err != nil {
		return PlanOperationResult{}, err
	}

	paths, err := s.publishObservedPlan(ctx, plan, sources)
	if err != nil {
		return PlanOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_written", map[string]string{"case_slug": scope.CaseSlug, "source_observation_hash": sources.Hash, "relative_path": paths.RelativePath})

	message := "Plan generated."
	if replace {
		message = "Plan regenerated."
	}

	return PlanOperationResult{Message: message, Plan: s.planViewFromPath(scope, paths.RelativePath, paths.Path)}, nil
}

func (s service) preparePlanGeneration(ctx context.Context, scope quizcore.QuizScope, expected string) (core.PlanSourceObservation, quizcore.QuizRuntimeState, string, error) {
	var (
		sources core.PlanSourceObservation
		state   quizcore.QuizRuntimeState
	)

	target, err := core.PlanTargetRevision(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return sources, state, target, err
	}

	if expected != "" && expected != target {
		return sources, state, target, workflowstate.ErrConflict
	}

	sources, err = s.observePlanContext(ctx, scope)
	if err != nil {
		return sources, state, target, err
	}

	if _, err := s.assessPlanReadiness(ctx, sources); err != nil {
		return sources, state, target, err
	}

	state, err = s.planObservedRuntimeState(sources)

	return sources, state, target, err
}

func (s service) planObservedRuntimeState(sources core.PlanSourceObservation) (quizcore.QuizRuntimeState, error) {
	if sources.Quiz.Exists {
		return sources.Quiz.State, nil
	}

	state, err := s.newQuizRuntimeState(sources.Scope, quizcore.QuizModeDeep)
	state.Status = quizcore.QuizRuntimeStatusDone

	return state, err
}

func (s service) publishObservedPlan(ctx context.Context, plan core.GeneratedPlan, sources core.PlanSourceObservation) (paths core.GeneratedPlanPaths, err error) {
	err = core.WithAuthoringMutation(ctx, s.projectContext.ProjectDir, func(ctx context.Context) error {
		current, readErr := s.observePlanContext(ctx, sources.Scope)
		if readErr != nil {
			return readErr
		}

		if current.Hash != sources.Hash {
			return fmt.Errorf("plan sources changed; reload context and generate again: %w", workflowstate.ErrConflict)
		}

		paths, readErr = core.WriteGeneratedPlan(ctx, s.projectContext.ProjectDir, plan)

		return readErr
	})

	return paths, err
}

func (s service) generatePlanWithConfiguredBackend(ctx context.Context, scope quizcore.QuizScope, state quizcore.QuizRuntimeState, sources core.PlanSourceObservation) (core.GeneratedPlan, []string, []string, error) {
	if s.singlePassPlan(ctx) {
		return s.generateSinglePassPlan(ctx, scope, state, sources)
	}

	return s.generateStagedPlan(ctx, scope, state, sources)
}

func (s service) generateSinglePassPlan(ctx context.Context, scope quizcore.QuizScope, state quizcore.QuizRuntimeState, sources core.PlanSourceObservation) (core.GeneratedPlan, []string, []string, error) {
	modelRequest, err := s.newPlanGenerationModelRequest(state, sources)
	if err != nil {
		return core.GeneratedPlan{}, nil, nil, err
	}

	s.appendPlanGenerationLog("generate_plan_single_pass_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := core.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		return core.GeneratedPlan{}, nil, nil, err
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

func (s service) generateStagedPlan(ctx context.Context, scope quizcore.QuizScope, state quizcore.QuizRuntimeState, sources core.PlanSourceObservation) (core.GeneratedPlan, []string, []string, error) {
	outline, sourcePaths, requestIDs, responseIDs, err := s.generatePlanOutline(ctx, scope, state, sources)
	if err != nil {
		return core.GeneratedPlan{}, nil, nil, err
	}

	details := make([]core.GeneratedPlanSliceDetail, 0, len(outline.Slices))
	for _, slice := range outline.Slices {
		detail, requestID, responseID, err := s.generatePlanSliceDetail(ctx, state, sources, outline, slice.ID)
		if err != nil {
			return core.GeneratedPlan{}, nil, nil, err
		}

		details = append(details, detail)
		requestIDs = append(requestIDs, requestID)
		responseIDs = append(responseIDs, responseID)
	}

	body, err := core.AssembleGeneratedPlanBody(outline, details, requestIDs, responseIDs)
	if err != nil {
		return core.GeneratedPlan{}, nil, nil, err
	}

	plan := core.GeneratedPlan{
		Scope:       scope,
		Title:       outline.Title,
		Body:        body,
		Order:       core.PlannedOrderFromDetails(details),
		SourcePaths: sourcePaths,
		CreatedAt:   outline.CreatedAt.UTC(),
	}

	return plan, requestIDs, responseIDs, nil
}

func (s service) generatePlanOutline(ctx context.Context, scope quizcore.QuizScope, state quizcore.QuizRuntimeState, sources core.PlanSourceObservation) (core.GeneratedPlanOutline, []string, []string, []string, error) {
	modelRequest, err := s.newPlanOutlineGenerationModelRequest(state, sources)
	if err != nil {
		return core.GeneratedPlanOutline{}, nil, nil, nil, err
	}

	s.appendPlanGenerationLog("generate_plan_outline_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := core.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		return core.GeneratedPlanOutline{}, nil, nil, nil, err
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

func (s service) generatePlanSliceDetail(ctx context.Context, state quizcore.QuizRuntimeState, sources core.PlanSourceObservation, outline core.GeneratedPlanOutline, targetSliceID string) (core.GeneratedPlanSliceDetail, string, string, error) {
	modelRequest, err := s.newPlanSliceDetailGenerationModelRequest(state, sources, outline, targetSliceID)
	if err != nil {
		return core.GeneratedPlanSliceDetail{}, "", "", err
	}

	s.appendPlanGenerationLog("generate_plan_slice_detail_started", map[string]string{
		"case_slug":  state.CaseSlug,
		"request_id": modelRequest.ID,
		"slice_id":   targetSliceID,
	})

	response, err := core.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		return core.GeneratedPlanSliceDetail{}, "", "", err
	}

	detail := response.SliceDetail
	detail.Scope = outline.Scope
	detail.CreatedAt = response.CreatedAt.UTC()

	if err := core.ValidateGeneratedPlanSliceDetailAgainstOutline(detail, outline, targetSliceID); err != nil {
		return core.GeneratedPlanSliceDetail{}, "", "", err
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
	operation := core.PlanPublicationOperation{ID: request.OperationID, Scope: scope, ExpectedRevision: request.ExpectedRevision, Replace: true, TrackerReplaceHash: request.TrackerReplaceHash}
	if result, found, err := s.replayPlanPublication(ctx, operation); found || err != nil {
		return result, err
	}
	var result PlanOperationResult

	err := s.withGenerationRuntime(ctx, "plan", scope, func(ctx context.Context) error {
		var err error

		result, err = s.prepareAndPublishPlan(ctx, operation)

		return err
	})

	return result, err
}

func (s service) MaterializePlanTracking(ctx context.Context, request PlanTrackingRequest) (PlanTrackingResult, error) {
	if request.Workflow != nil {
		return s.materializeApprovedWorkflow(ctx, request)
	}

	scope := s.quizScope(request.Scope)

	plan, exists, err := s.existingGeneratedPlan(scope)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	if !exists {
		return PlanTrackingResult{}, fmt.Errorf("plan is required before tracking can be created")
	}

	binding, err := core.ObservePlanRevision(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	if binding.Reference != "" {
		return s.retainedPlanTracking(plan, binding)
	}

	trackingPaths, err := core.GeneratedPlanTrackingLayout(s.projectContext.ProjectDir, scope)
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

		result, materializeErr = s.materializeBoundPlanTracking(ctx, request.OperationID, scope, plan, trackingPaths.RelativePath)

		return materializeErr
	})

	return result, err
}

func (s service) materializePlanTracking(
	ctx context.Context,
	scope quizcore.QuizScope,
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

	response, err := core.InvokeQuizModel(ctx, s.quizModel, modelRequest)
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

	paths, err := core.WriteGeneratedPlanTracking(ctx, s.projectContext.ProjectDir, tracking)
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
	scope quizcore.QuizScope,
	planPath string,
	trackingPath string,
	ticketPayloads []core.GeneratedPlanTrackingTicket,
) ([]core.GeneratedPlanTrackingTicket, error) {
	if err := validatePlanTrackingTicketPayloads(ticketPayloads); err != nil {
		return nil, err
	}

	created := make([]core.GeneratedPlanTrackingTicket, 0, len(ticketPayloads))

	for _, payload := range ticketPayloads {
		result, err := s.CreateTicket(CreateTicketRequest{
			Title:       payload.Title,
			Synopsis:    payload.Synopsis,
			Status:      core.TicketStatusReadyForDev,
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

func validatePlanTrackingTicketPayloads(ticketPayloads []core.GeneratedPlanTrackingTicket) error {
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

func validatePlanTrackingTicketPayload(payload core.GeneratedPlanTrackingTicket) error {
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
func renderPlanTrackingTicketBody(planPath, trackingPath string, ticket core.GeneratedPlanTrackingTicket) string {
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

func (s service) planViewFromScope(scope quizReadScope, index core.WorkspaceIndex) PlanView {
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
func (s service) latestPlanNodeForScope(scope quizcore.QuizScope, index core.WorkspaceIndex) (core.WorkspaceNode, bool) {
	nodes := []core.WorkspaceNode{}
	for _, nodeID := range index.ByKind[core.WorkspaceNodePlan] {
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

func (s service) planViewFromNode(scope quizcore.QuizScope, fallbackTitle string, node core.WorkspaceNode) PlanView {
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

func (s service) planViewFromPath(scope quizcore.QuizScope, relativePath, path string) PlanView {
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

func (s service) newPlanOutlineGenerationModelRequest(state quizcore.QuizRuntimeState, sources core.PlanSourceObservation) (core.QuizModelRequest, error) {
	return s.newPlanningModelRequest(
		state, sources,
		core.QuizModelActionGeneratePlanOutline,
		"plan_outline_generation",
		"plan-outline-generation",
	)
}

func (s service) newPlanGenerationModelRequest(state quizcore.QuizRuntimeState, sources core.PlanSourceObservation) (core.QuizModelRequest, error) {
	return s.newPlanningModelRequest(
		state, sources,
		core.QuizModelActionGeneratePlan,
		"plan_generation",
		"plan-generation",
	)
}

func (s service) newPlanSliceDetailGenerationModelRequest(state quizcore.QuizRuntimeState, sources core.PlanSourceObservation, outline core.GeneratedPlanOutline, targetSliceID string) (core.QuizModelRequest, error) {
	if err := core.ValidateGeneratedPlanOutline(outline); err != nil {
		return core.QuizModelRequest{}, err
	}

	target, ok := outline.Slice(targetSliceID)
	if !ok {
		return core.QuizModelRequest{}, fmt.Errorf("plan outline slice is unknown: %s", targetSliceID)
	}

	request, err := s.newPlanningModelRequest(state, sources, core.QuizModelActionGeneratePlanSliceDetail, "plan_slice_detail_generation", "plan-slice-detail-generation")
	if err != nil {
		return request, err
	}

	request.ID += "-" + planSliceIDSlug(targetSliceID)
	request.ContextSections, err = appendPlanSliceDetailContextSections(request.ContextSections, outline, target)

	return request, err
}

func (s service) newPlanningModelRequest(state quizcore.QuizRuntimeState, sources core.PlanSourceObservation, action core.QuizModelAction, strategy, idPrefix string) (core.QuizModelRequest, error) {
	scope := sources.Scope
	payload := core.QuizModelPayload{
		Project: core.QuizModelProjectPayload{ID: scope.ProjectID, Name: sources.Project.Name, Description: sources.Project.Description},
		Quiz:    quizModelQuizPayload(state), RecentEntries: quizModelRecentEntryPayloads(sources.Quiz.Entries, 5),
		CurrentQuestion: quizModelCurrentQuestionPayload(state, sources.Quiz.Entries), OutputContract: core.NewQuizModelOutputContract(action, state.Mode),
	}

	sections, err := core.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	for _, input := range sources.Inputs {
		sections = append(sections, core.QuizModelContextSection{Name: input.Kind, Layer: input.Kind, Authority: input.Reference, Content: input.Content, Metadata: map[string]string{"source_revision": input.Revision, "source_observation_hash": sources.Hash}})
	}

	return core.QuizModelRequest{
		ID: core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("%s-%s", idPrefix, specScopeID(scope))), Action: action, Strategy: strategy, Scope: scope, Mode: state.Mode, RuntimeState: state, Payload: payload, ContextSections: sections,
		OutputContract: core.QuizModelOutputContract{RequiredFields: payload.OutputContract.RequiredFields}, CreatedAt: s.utcNow(),
	}, nil
}

func appendPlanSliceDetailContextSections(
	sections []core.QuizModelContextSection,
	outline core.GeneratedPlanOutline,
	targetSlice core.GeneratedPlanOutlineSlice,
) ([]core.QuizModelContextSection, error) {
	outlineContent, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		return nil, err
	}

	targetContent, err := json.MarshalIndent(targetSlice, "", "  ")
	if err != nil {
		return nil, err
	}

	sections = append(sections, core.QuizModelContextSection{
		Name:      "plan_outline",
		Layer:     "plan",
		Authority: "generated_plan_outline",
		Content:   string(outlineContent),
		Metadata: map[string]string{
			"title":  outline.Title,
			"status": outline.Status,
		},
	})
	sections = append(sections, core.QuizModelContextSection{
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
	state quizcore.QuizRuntimeState,
	plan PlanView,
) (core.QuizModelRequest, error) {
	scope := quizcore.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, core.QuizModelActionMaterializePlanTracking)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	sections, err := core.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	sections = append(sections, core.QuizModelContextSection{
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

	return core.QuizModelRequest{
		ID:              core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("plan-tracking-%s", specScopeID(scope))),
		Action:          core.QuizModelActionMaterializePlanTracking,
		Strategy:        "plan_tracking_materialization",
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
