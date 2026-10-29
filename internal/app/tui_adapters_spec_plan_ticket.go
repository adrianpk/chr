package app

import (
	"crypto/rand"
	"errors"
	"fmt"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

type tuiTicketActionExecutor struct {
	operations service
}

type tuiSpecActionExecutor struct {
	operations service
}

type tuiPlanActionExecutor struct {
	operations service
}

func newTUISpecActionExecutor(
	operations service,
) tuiSpecActionExecutor {
	return tuiSpecActionExecutor{
		operations: operations,
	}
}

func newTUIPlanActionExecutor(
	operations service,
) tuiPlanActionExecutor {
	return tuiPlanActionExecutor{
		operations: operations,
	}
}

func (e tuiSpecActionExecutor) ExecuteSpecAction(
	request tui.SpecActionRequest,
) (tui.SpecActionResult, error) {
	result := tui.SpecActionResult{}

	switch request.Action {
	case tui.SpecActionGenerate:
		generated, err := e.operations.GenerateSpec(e.operations.ctx, SpecGenerateRequest{
			Scope: quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			return tui.SpecActionResult{}, err
		}

		result.Message = generated.Message
	case tui.SpecActionGenerateFeature:
		generated, err := e.operations.GenerateAcceptanceFeature(e.operations.ctx, AcceptanceFeatureGenerateRequest{
			Scope: quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			return tui.SpecActionResult{}, err
		}

		result.Message = generated.Message
		result.AcceptanceFeatureStatus = generated.Status
		result.AcceptanceFeaturePath = generated.Feature
		result.AcceptanceFeatureRound = generated.FeatureRound
		result.AcceptanceFeatureReason = generated.RejectionReason
	case tui.SpecActionRefresh:
	default:
		return tui.SpecActionResult{}, errors.New("unknown spec action")
	}

	documents, err := tui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		return tui.SpecActionResult{}, err
	}

	specs, err := loadTUISpecs(e.operations)
	if err != nil {
		return tui.SpecActionResult{}, err
	}

	if result.Message == "" {
		result.Message = "Specs refreshed."
	}

	result.Documents = documents
	result.Specs = specs
	result.Scope = request.Scope

	return result, nil
}

func (e tuiPlanActionExecutor) ExecutePlanAction(request tui.PlanActionRequest) (result tui.PlanActionResult, err error) {
	e.operations.appendPlanGenerationLog("tui_plan_action_started", map[string]string{"action": string(request.Action), "case_slug": request.Scope.CaseSlug})

	defer func() {
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{"action": string(request.Action), "case_slug": request.Scope.CaseSlug, "error": err.Error()})
		}
	}()

	result, err = e.runPlanAction(request)
	if err != nil {
		return tui.PlanActionResult{}, err
	}

	result.Documents, err = tui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		return tui.PlanActionResult{}, err
	}

	result.Plans, err = loadTUIPlans(e.operations)
	if err != nil {
		return tui.PlanActionResult{}, err
	}

	if result.Message == "" {
		result.Message = "Plans refreshed."
	}

	result.Scope = request.Scope
	e.operations.appendPlanGenerationLog("tui_plan_action_finished", map[string]string{"action": string(request.Action), "case_slug": request.Scope.CaseSlug, "plan_count": fmt.Sprintf("%d", len(result.Plans))})

	return result, nil
}

func (e tuiPlanActionExecutor) runPlanAction(request tui.PlanActionRequest) (tui.PlanActionResult, error) {
	switch request.Action {
	case tui.PlanActionGenerate:
		result, err := e.operations.GeneratePlan(e.operations.ctx, PlanGenerateRequest{Scope: quizcore.Scope{CaseSlug: request.Scope.CaseSlug}})

		return tui.PlanActionResult{Message: result.Message}, err
	case tui.PlanActionRegenerate:
		result, err := e.operations.RegeneratePlan(e.operations.ctx, PlanRegenerateRequest{Scope: quizcore.Scope{CaseSlug: request.Scope.CaseSlug}})

		return tui.PlanActionResult{Message: result.Message}, err
	case tui.PlanActionMaterializeTracking:
		result, err := e.operations.MaterializePlanTracking(e.operations.ctx, PlanTrackingRequest{Scope: quizcore.Scope{CaseSlug: request.Scope.CaseSlug}})

		return tui.PlanActionResult{Message: result.Message, Tickets: result.Tickets}, err
	case tui.PlanActionRefresh:
		return tui.PlanActionResult{}, nil
	default:
		return tui.PlanActionResult{}, errors.New("unknown plan action")
	}
}

func newTUITicketActionExecutor(
	operations service,
) tuiTicketActionExecutor {
	return tuiTicketActionExecutor{
		operations: operations,
	}
}

func (e tuiTicketActionExecutor) ExecuteTicketAction(
	request tui.TicketActionRequest,
) (tui.TicketActionResult, error) {
	switch request.Action {
	case tui.TicketActionAddComment:
		result, err := e.operations.AddSelectedTicketComment(request.Key)
		if err != nil {
			return tui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case tui.TicketActionMoveStatus:
		return e.moveTicketStatus(request)
	case tui.TicketActionAttachReviewRef:
		result, err := e.operations.AttachTicketReviewReference(TicketReviewReferenceRequest{
			Key:    request.Key,
			Branch: request.Branch,
			Ref:    request.ReviewRef,
		})
		if err != nil {
			return tui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case tui.TicketActionToggleChecklist:
		return tui.TicketActionResult{}, errors.New(
			"checklist toggle requires a safe checklist item writer and is not available from this TUI action yet",
		)
	case tui.TicketActionAddChecklistItem:
		return tui.TicketActionResult{}, errors.New(
			"checklist item creation requires a safe checklist item writer and is not available from this TUI action yet",
		)
	default:
		return tui.TicketActionResult{}, errors.New("unknown ticket action")
	}
}

func (e tuiTicketActionExecutor) moveTicketStatus(request tui.TicketActionRequest) (tui.TicketActionResult, error) {
	if !ticketcore.NormalBoardStatus(ticketcore.Status(request.Status)) {
		return tui.TicketActionResult{}, fmt.Errorf("select a canonical Board destination")
	}

	_, err := e.operations.SaveTicketDocument(workflowcore.TicketEditRequest{
		Key: request.Key, Revision: request.Revision, Fields: map[string]string{"status": request.Status},
		Binding: workflowstate.MutationBinding{RequestID: rand.Text(), Operator: "tui"},
	})
	if err != nil {
		return tui.TicketActionResult{}, err
	}

	board, err := e.operations.ReadTicketBoard(0)

	return tui.TicketActionResult{Message: "Ticket status saved.", Tickets: board.Tickets}, err
}

func ticketActionResult(result TicketOperationResult) tui.TicketActionResult {
	return tui.TicketActionResult{
		Message: result.Message,
		Tickets: result.Tickets,
	}
}
