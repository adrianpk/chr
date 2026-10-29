package tui

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/app/work"
)

func (m *Model) observedDashTickets() []TicketView {
	tickets := make([]TicketView, 0, len(m.workObservation.Queue))
	for _, item := range m.workObservation.Queue {
		ticket := TicketView{
			Key: item.Key, Title: item.Title, Revision: item.Revision,
			Priority: item.Priority, Severity: item.Severity, UseCase: item.Scope,
		}
		for _, candidate := range m.tickets {
			if candidate.Key == item.Key {
				ticket = candidate

				break
			}
		}
		// Display and action guards belong to the same observation, even if another
		// document read has already seen a newer revision.
		ticket.Key, ticket.Title, ticket.Revision = item.Key, item.Title, item.Revision
		ticket.Priority, ticket.Severity, ticket.UseCase = item.Priority, item.Severity, item.Scope
		tickets = append(tickets, ticket)
	}

	return tickets
}

func (m *Model) observedQueueItem(key string) (work.QueueItem, bool) {
	if m.workObservation != nil {
		for _, item := range m.workObservation.Queue {
			if item.Key == key {
				return item, true
			}
		}
	}

	return work.QueueItem{}, false
}

func (m *Model) dashQueueSummary() []string {
	if m.queuePending {
		return []string{"Loading queue observation..."}
	}

	if m.workObservation == nil {
		return []string{fmt.Sprintf("Ready queue: %d", len(m.readyAgentWorkTickets()))}
	}

	view := m.workObservation

	lines := []string{
		fmt.Sprintf("Eligible: %d / %d queued", view.Eligible, view.QueueTotal),
		fmt.Sprintf("Page: %d-%d; [ previous / ] next", min(m.queueRequest.Offset+1, view.QueueTotal), min(m.queueRequest.Offset+len(view.Queue), view.QueueTotal)),
	}
	for _, reason := range slices.Sorted(maps.Keys(view.Excluded)) {
		if count := view.Excluded[reason]; count > 0 {
			lines = append(lines, fmt.Sprintf("%d excluded: %s", count, strings.ReplaceAll(reason, "_", " ")))
		}
	}

	return lines
}

func (m *Model) dashActionBlocked(action WorkActionID) bool {
	if m.SelectedView() != ViewDash || m.workObservation == nil {
		return false
	}

	if m.queuePending && (action == WorkActionStartTicket || action == WorkActionRunAllReady) {
		m.statusMessage = "Wait for the current queue observation."

		return true
	}

	if action == WorkActionStartTicket {
		return m.dashSelectedActionBlocked()
	}

	if action == WorkActionRunAllReady && m.workObservation.Eligible == 0 {
		m.statusMessage = "No eligible work in the observed queue."

		return true
	}

	return false
}

func (m *Model) renderDashQueueSummary() string {
	if m.workObservation == nil {
		return ""
	}

	return "  " + strings.Join(m.dashQueueSummary(), "\n  ") + "\n"
}

func (m *Model) dashSelectedActionBlocked() bool {
	ticket, ok := m.selectedDashTicketView()

	item, observed := m.observedQueueItem(ticket.Key)
	if !ok || !observed || item.Readiness != "ready" {
		m.statusMessage = "Selected work is unavailable: " + item.Readiness + " " + item.Reason

		return true
	}

	return false
}
