package app

import (
	"slices"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func observedQueueItem(ticket ticketcore.Indexed) WorkQueueItem {
	return WorkQueueItem{
		Key: ticket.Key, Title: ticket.Title, Scope: ticket.UseCase,
		Revision: ticket.ContentHash, Priority: ticket.Priority, Severity: ticket.Severity, CreatedAt: ticket.CreatedAt,
	}
}

func selectObservedQueue(result *WorkObservation, priority string, order workflowstate.TicketOrder) {
	result.Queue = slices.DeleteFunc(result.Queue, func(item WorkQueueItem) bool { return priority != "" && item.Priority != priority })
	slices.SortStableFunc(result.Queue, func(a, b WorkQueueItem) int {
		return ticketcore.CompareLoopOrder(order, observedQueueTicket(a), observedQueueTicket(b))
	})

	result.Eligible = 0
	for _, item := range result.Queue {
		if item.Readiness == "ready" {
			result.Eligible++
		}
	}
}

func observedQueueTicket(item WorkQueueItem) ticketcore.Indexed {
	return ticketcore.Indexed{
		IndexedIdentity:       ticketcore.IndexedIdentity{Key: item.Key},
		IndexedClassification: ticketcore.IndexedClassification{Priority: item.Priority, Severity: item.Severity},
		IndexedTimes:          ticketcore.IndexedTimes{CreatedAt: item.CreatedAt},
	}
}
