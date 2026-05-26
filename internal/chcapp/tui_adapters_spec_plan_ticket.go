package chcapp

import (
	"errors"
	"fmt"

	"forge.adrianpk.com/chc/internal/chc"
	"forge.adrianpk.com/chc/internal/chctui"
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
	request chctui.SpecActionRequest,
) (chctui.SpecActionResult, error) {
	switch request.Action {
	case chctui.SpecActionGenerate:
		_, err := e.operations.GenerateSpec(e.operations.ctx, SpecGenerateRequest{
			Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			return chctui.SpecActionResult{}, err
		}
	case chctui.SpecActionRefresh:
	default:
		return chctui.SpecActionResult{}, errors.New("unknown spec action")
	}

	documents, err := chctui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		return chctui.SpecActionResult{}, err
	}

	specs, err := loadTUISpecs(e.operations)
	if err != nil {
		return chctui.SpecActionResult{}, err
	}

	message := "Specs refreshed."
	if request.Action == chctui.SpecActionGenerate {
		message = "Spec generated."
	}

	return chctui.SpecActionResult{
		Message:   message,
		Documents: documents,
		Specs:     specs,
		Scope:     request.Scope,
	}, nil
}

//nolint:cyclop,funlen,wsl_v5 // Adapter dispatch keeps action-specific service calls in one place.
func (e tuiPlanActionExecutor) ExecutePlanAction(
	request chctui.PlanActionRequest,
) (chctui.PlanActionResult, error) {
	e.operations.appendPlanGenerationLog("tui_plan_action_started", map[string]string{
		"action":    string(request.Action),
		"case_slug": request.Scope.CaseSlug,
	})

	message := ""
	var tickets []chc.IndexedTicket

	switch request.Action {
	case chctui.PlanActionGenerate:
		result, err := e.operations.GeneratePlan(e.operations.ctx, PlanGenerateRequest{
			Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return chctui.PlanActionResult{}, err
		}

		message = result.Message
	case chctui.PlanActionRegenerate:
		result, err := e.operations.RegeneratePlan(e.operations.ctx, PlanRegenerateRequest{
			Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return chctui.PlanActionResult{}, err
		}

		message = result.Message
	case chctui.PlanActionMaterializeTracking:
		result, err := e.operations.MaterializePlanTracking(e.operations.ctx, PlanTrackingRequest{
			Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return chctui.PlanActionResult{}, err
		}

		message = result.Message
		tickets = result.Tickets
	case chctui.PlanActionRefresh:
	default:
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     "unknown plan action",
		})

		return chctui.PlanActionResult{}, errors.New("unknown plan action")
	}

	documents, err := chctui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     err.Error(),
		})

		return chctui.PlanActionResult{}, err
	}

	plans, err := loadTUIPlans(e.operations)
	if err != nil {
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     err.Error(),
		})

		return chctui.PlanActionResult{}, err
	}

	if message == "" {
		message = "Plans refreshed."
	}

	e.operations.appendPlanGenerationLog("tui_plan_action_finished", map[string]string{
		"action":     string(request.Action),
		"case_slug":  request.Scope.CaseSlug,
		"plan_count": fmt.Sprintf("%d", len(plans)),
	})

	return chctui.PlanActionResult{
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
	request chctui.TicketActionRequest,
) (chctui.TicketActionResult, error) {
	switch request.Action {
	case chctui.TicketActionAddComment:
		result, err := e.operations.AddSelectedTicketComment(request.Key)
		if err != nil {
			return chctui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case chctui.TicketActionMoveStatus:
		result, err := e.operations.MoveSelectedTicketStatus(request.Key, request.Status)
		if err != nil {
			return chctui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case chctui.TicketActionToggleChecklist:
		return chctui.TicketActionResult{}, errors.New(
			"checklist toggle requires a safe checklist item writer and is not available from this TUI action yet",
		)
	case chctui.TicketActionAddChecklistItem:
		return chctui.TicketActionResult{}, errors.New(
			"checklist item creation requires a safe checklist item writer and is not available from this TUI action yet",
		)
	default:
		return chctui.TicketActionResult{}, errors.New("unknown ticket action")
	}
}

func ticketActionResult(result TicketOperationResult) chctui.TicketActionResult {
	return chctui.TicketActionResult{
		Message: result.Message,
		Tickets: result.Tickets,
	}
}
