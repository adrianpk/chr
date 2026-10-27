package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// PlanTrackingRequest selects the plan scope used to materialize tracking
// artifacts.
type PlanTrackingRequest struct {
	OperationID string
	Workflow    *generationcore.GeneratedWorkflowPlan
	Scope       quizcore.Scope
}

// PlanTrackingResult reports generated tracking artifacts for a plan.
type PlanTrackingResult struct {
	Workflow     *core.WorkflowMaterializationResult
	Message      string
	Plan         PlanView
	TrackingPath string
	Tickets      []core.IndexedTicket
}

func (s service) MaterializePlanTracking(ctx context.Context, request PlanTrackingRequest) (PlanTrackingResult, error) {
	if request.Workflow != nil {
		return s.materializeApprovedWorkflow(ctx, request)
	}

	scope := s.quizScope(request.Scope)

	existing, err := s.existingGeneratedPlan(scope)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	if !existing.Exists {
		return PlanTrackingResult{}, fmt.Errorf("plan is required before tracking can be created")
	}

	plan := existing.Plan

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

func (s service) materializePlanTracking(ctx context.Context, scope quizcore.Scope, plan PlanView, trackingRelativePath string) (PlanTrackingResult, error) {
	state, err := s.specGenerationRuntimeState(scope)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	modelRequest, err := s.newPlanTrackingModelRequest(state, plan)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
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

func (s service) createPlanTrackingTickets(scope quizcore.Scope, planPath, trackingPath string, ticketPayloads []generationcore.GeneratedPlanTrackingTicket) ([]generationcore.GeneratedPlanTrackingTicket, error) {
	if err := validatePlanTrackingTicketPayloads(ticketPayloads); err != nil {
		return nil, err
	}

	created := make([]generationcore.GeneratedPlanTrackingTicket, 0, len(ticketPayloads))

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

func validatePlanTrackingTicketPayloads(ticketPayloads []generationcore.GeneratedPlanTrackingTicket) error {
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

func validatePlanTrackingTicketPayload(payload generationcore.GeneratedPlanTrackingTicket) error {
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
func renderPlanTrackingTicketBody(planPath, trackingPath string, ticket generationcore.GeneratedPlanTrackingTicket) string {
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

func (s service) newPlanTrackingModelRequest(state quizcore.RuntimeState, plan PlanView) (quizmodel.Request, error) {
	scope := quizcore.Scope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, quizmodel.ActionMaterializePlanTracking)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections = append(sections, quizmodel.ContextSection{
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

	return quizmodel.Request{
		ID:              core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("plan-tracking-%s", specScopeID(scope))),
		Action:          quizmodel.ActionMaterializePlanTracking,
		Strategy:        "plan_tracking_materialization",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: quizmodel.OutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}
