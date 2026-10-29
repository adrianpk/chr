package tui

import ticketcore "forge.adrianpk.com/chr/internal/core/ticket"

// Reconcile the independently selected Board key before replacing row indices.
// While Dash is visible, Board eligibility still uses its own column and filter.
func (m *Model) replaceTickets(tickets []ticketcore.Indexed) {
	key := ""

	for _, ticket := range m.tickets {
		if ticket.Index == m.selectedTicket {
			key = ticket.Key

			break
		}
	}

	m.tickets = NewTicketViews(tickets)
	visible := m.filteredTicketRows(m.SelectedView() != ViewTickets)

	m.selectedTicket = 0
	if len(visible) > 0 {
		m.selectedTicket = visible[0].Index
	}

	for _, ticket := range visible {
		if ticket.Key == key {
			m.selectedTicket = ticket.Index

			break
		}
	}

	m.reconcileDashSelection()
}
