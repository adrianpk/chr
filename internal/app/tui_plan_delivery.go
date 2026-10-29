package app

import (
	"fmt"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiPlanActionExecutor) generatePlanDelivery(request tui.PlanActionRequest) (tui.PlanActionResult, error) {
	delivered, err := e.operations.GeneratePlanDelivery(e.operations.ctx, PlanGenerateRequest{
		Scope:       quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		OperationID: request.OperationID, ExpectedRevision: request.Revision,
	})

	result := tui.PlanActionResult{DocumentPath: delivered.Plan.RelativePath, Committed: delivered.Plan.RelativePath != "", Outcome: operation.Result{ChangedFiles: compactChangedFiles(delivered.Plan.RelativePath)}}
	if err != nil {
		return result, err
	}

	layout, err := authoringcore.GeneratedPlanLayout(e.operations.projectContext.ProjectDir, e.operations.quizScope(quizcore.Scope{CaseSlug: request.Scope.CaseSlug}))
	if err != nil {
		return result, err
	}

	result.DocumentPath = layout.RelativePath
	result.Prepared = true
	result.Message = fmt.Sprintf("Plan ready with %d tickets in %d slices. Work has not started.", len(delivered.Workflow.Tasks), len(delivered.Workflow.Set.Set.Slices))
	board, err := e.operations.ReadTicketBoard(0)
	result.Tickets = board.Tickets

	return result, err
}

func tuiPlanRecoverySummary(request tui.PlanActionRequest, result tui.PlanActionResult) string {
	if request.Action == tui.PlanActionRegenerate {
		return "Plan saved; refresh failed. Retry the original operation."
	}

	if result.Prepared {
		return "Plan and tickets saved; refresh failed. Retry the original operation."
	}

	return "Plan saved; ticket preparation is incomplete. Retry the original operation."
}
