package app

import (
	"errors"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
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
			Scope: core.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			return tui.SpecActionResult{}, err
		}

		result.Message = generated.Message
	case tui.SpecActionGenerateFeature:
		generated, err := e.operations.GenerateAcceptanceFeature(e.operations.ctx, AcceptanceFeatureGenerateRequest{
			Scope: core.QuizScope{CaseSlug: request.Scope.CaseSlug},
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

//nolint:cyclop,funlen,wsl_v5 // Adapter dispatch keeps action-specific service calls in one place.
func (e tuiPlanActionExecutor) ExecutePlanAction(
	request tui.PlanActionRequest,
) (tui.PlanActionResult, error) {
	e.operations.appendPlanGenerationLog("tui_plan_action_started", map[string]string{
		"action":    string(request.Action),
		"case_slug": request.Scope.CaseSlug,
	})

	message := ""
	var tickets []core.IndexedTicket

	switch request.Action {
	case tui.PlanActionGenerate:
		result, err := e.operations.GeneratePlan(e.operations.ctx, PlanGenerateRequest{
			Scope: core.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return tui.PlanActionResult{}, err
		}

		message = result.Message
	case tui.PlanActionRegenerate:
		result, err := e.operations.RegeneratePlan(e.operations.ctx, PlanRegenerateRequest{
			Scope: core.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return tui.PlanActionResult{}, err
		}

		message = result.Message
	case tui.PlanActionMaterializeTracking:
		result, err := e.operations.MaterializePlanTracking(e.operations.ctx, PlanTrackingRequest{
			Scope: core.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return tui.PlanActionResult{}, err
		}

		message = result.Message
		tickets = result.Tickets
	case tui.PlanActionRefresh:
	default:
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     "unknown plan action",
		})

		return tui.PlanActionResult{}, errors.New("unknown plan action")
	}

	documents, err := tui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     err.Error(),
		})

		return tui.PlanActionResult{}, err
	}

	plans, err := loadTUIPlans(e.operations)
	if err != nil {
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     err.Error(),
		})

		return tui.PlanActionResult{}, err
	}

	if message == "" {
		message = "Plans refreshed."
	}

	e.operations.appendPlanGenerationLog("tui_plan_action_finished", map[string]string{
		"action":     string(request.Action),
		"case_slug":  request.Scope.CaseSlug,
		"plan_count": fmt.Sprintf("%d", len(plans)),
	})

	return tui.PlanActionResult{
		Message:   message,
		Documents: documents,
		Plans:     plans,
		Tickets:   tickets,
		Scope:     request.Scope,
	}, nil
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
		result, err := e.operations.MoveSelectedTicketStatus(request.Key, request.Status)
		if err != nil {
			return tui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
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

func ticketActionResult(result TicketOperationResult) tui.TicketActionResult {
	return tui.TicketActionResult{
		Message: result.Message,
		Tickets: result.Tickets,
	}
}
