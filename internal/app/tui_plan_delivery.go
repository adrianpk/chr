package app

import (
	"fmt"

	"forge.adrianpk.com/chr/internal/app/operation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiPlanActionExecutor) generatePlanDelivery(request tui.PlanActionRequest) (tui.PlanActionResult, error) {
	delivered, err := e.operations.GeneratePlanDelivery(e.operations.ctx, PlanGenerateRequest{
		Scope:       quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		OperationID: request.OperationID, ExpectedRevision: request.Revision,
	})

	result := tui.PlanActionResult{Outcome: operation.Result{ChangedFiles: compactChangedFiles(delivered.Plan.RelativePath)}}
	if err != nil {
		return result, err
	}

	result.Message = fmt.Sprintf("Plan ready with %d tickets in %d slices. Work has not started.", len(delivered.Workflow.Tasks), len(delivered.Workflow.Set.Set.Slices))
	board, err := e.operations.ReadTicketBoard(0)
	result.Tickets = board.Tickets

	return result, err
}
