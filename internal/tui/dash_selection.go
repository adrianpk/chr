package tui

func (m *Model) selectedDashTicketView() (TicketView, bool) {
	queue := m.readyAgentWorkTickets()
	for _, ticket := range queue {
		if ticket.Key == m.dashTicketKey {
			return ticket, true
		}
	}

	if len(queue) > 0 {
		return queue[0], true
	}

	return TicketView{}, false
}

func (m *Model) reconcileDashSelection() {
	if m.queuePending {
		return
	}

	ticket, _ := m.selectedDashTicketView()
	m.dashTicketKey = ticket.Key
}

func (m *Model) moveDashSelection(delta int) {
	queue := m.readyAgentWorkTickets()
	if len(queue) == 0 {
		m.dashTicketKey = ""

		return
	}

	selected, _ := m.selectedDashTicketView()
	for position, ticket := range queue {
		if ticket.Key == selected.Key {
			m.dashTicketKey = queue[(position+delta+len(queue))%len(queue)].Key

			return
		}
	}
}
