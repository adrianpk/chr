package app

import (
	"context"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Read the independent Board catalog through bounded canonical pages. Queue
// pagination must not remove a Board selection on another page.
func loadTUITicketSnapshot(operations service) ([]ticketcore.Indexed, error) {
	ctx, cancel := context.WithTimeout(operations.ctx, 15*time.Second)
	defer cancel()

	operations.ctx = ctx
	tickets := []ticketcore.Indexed{}
	revision := ""

	offset := 0
	for offset < workflowstate.RecordLimit {
		page, err := operations.ReadTicketBoard(offset)
		if err != nil {
			return nil, err
		}

		if offset != 0 && revision != page.Revision {
			return nil, workflowstate.ErrConflict
		}

		revision = page.Revision

		tickets = append(tickets, page.Tickets...)
		if page.NextOffset < 0 {
			return tickets, nil
		}

		if page.NextOffset <= offset {
			return nil, workflowstate.ErrConflict
		}

		offset = page.NextOffset
	}

	return nil, workflowstate.ErrConflict
}
