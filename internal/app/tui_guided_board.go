package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiGuidedExecutor) projectGuidedBoard(ctx context.Context, observed GuidedUseCaseInspection, view *tui.GuidedView) error {
	if observed.Session == nil {
		return nil
	}

	stage := observed.Session.Guided.Stage
	if stage != "execution" && stage != "results" {
		return nil
	}

	operations := e.operations
	operations.ctx, operations.workflowNamespace = ctx, observed.Namespace

	tickets, err := loadTUITicketSnapshot(operations)
	if err != nil {
		return err
	}

	view.Tickets = tickets

	return nil
}
