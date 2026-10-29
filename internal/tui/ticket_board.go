package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func (m *Model) filteredTickets() []TicketView {
	tickets := m.caseTickets()
	if m.SelectedView() == ViewBoard {
		if m.boardAttentionActive() {
			tickets = m.attentionTickets()
		} else {
			tickets = m.activeBoardColumnTickets()
		}
	}

	if m.ticketFilter == "" {
		return append([]TicketView(nil), tickets...)
	}

	filter := strings.ToLower(m.ticketFilter)
	filtered := []TicketView{}

	for _, ticket := range tickets {
		if ticket.Matches(filter) {
			filtered = append(filtered, ticket)
		}
	}

	return filtered
}

func (m *Model) updateFilterKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.filterEntryMode = false
	case tea.KeyEnter:
		m.filterEntryMode = false
		m.selectFirstFilteredTicket()
	case tea.KeyBackspace:
		m.ticketFilter = strings.TrimSuffix(m.ticketFilter, lastRune(m.ticketFilter))
	default:
		if msg.Type == tea.KeyRunes {
			m.ticketFilter += msg.String()
		}
	}

	return nil
}

func (m *Model) selectFirstFilteredTicket() {
	tickets := m.filteredTickets()
	if len(tickets) == 0 {
		m.selectedTicket = 0

		return
	}

	m.selectedTicket = tickets[0].Index
}

func (m *Model) activeBoardColumn() string {
	if m.boardAttentionActive() {
		return m.previousNormalBoardColumn()
	}

	if ticketcore.NormalBoardStatus(ticketcore.Status(m.boardColumn)) {
		return m.boardColumn
	}

	return string(ticketcore.StatusBacklog)
}

func (m *Model) activeBoardColumnPosition() int {
	for index, column := range ticketcore.NormalBoardColumns {
		if column == m.activeBoardColumn() {
			return index
		}
	}

	return 0
}

func (m *Model) activeBoardColumnTickets() []TicketView {
	return m.ticketsForBoardColumn(m.activeBoardColumn())
}

func (m *Model) ticketsForBoardColumn(column string) []TicketView {
	tickets := []TicketView{}

	for _, ticket := range m.caseTickets() {
		if ticket.BoardColumn == column {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

func (m *Model) attentionTickets() []TicketView {
	tickets := []TicketView{}

	for _, ticket := range m.caseTickets() {
		if len(ticket.AttentionReasons) > 0 {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

func (m *Model) boardAttentionActive() bool {
	return m.SelectedView() == ViewBoard && m.boardColumn == boardAttentionColumn
}

func (m *Model) toggleBoardAttention() {
	if !m.boardColumnNavigationActive() {
		return
	}

	if m.boardAttentionActive() {
		m.boardColumn = m.previousNormalBoardColumn()
	} else {
		m.previousBoardColumn = m.activeBoardColumn()
		m.boardColumn = boardAttentionColumn
	}

	m.selectFirstFilteredTicket()
}

func (m *Model) previousNormalBoardColumn() string {
	if ticketcore.NormalBoardStatus(ticketcore.Status(m.previousBoardColumn)) {
		return m.previousBoardColumn
	}

	return string(ticketcore.StatusBacklog)
}

func (m *Model) selectPreviousBoardColumn() {
	if !m.boardColumnNavigationActive() {
		return
	}

	if m.boardAttentionActive() {
		return
	}

	position := m.activeBoardColumnPosition()
	if position == 0 {
		position = len(ticketcore.NormalBoardColumns) - 1
	} else {
		position--
	}

	m.selectBoardColumn(ticketcore.NormalBoardColumns[position])
}

func (m *Model) selectNextBoardColumn() {
	if !m.boardColumnNavigationActive() {
		return
	}

	if m.boardAttentionActive() {
		return
	}

	position := m.activeBoardColumnPosition() + 1
	if position >= len(ticketcore.NormalBoardColumns) {
		position = 0
	}

	m.selectBoardColumn(ticketcore.NormalBoardColumns[position])
}

func (m *Model) boardColumnNavigationActive() bool {
	return m.SelectedView() == ViewBoard && m.selectedPane == 0
}

func (m *Model) selectBoardColumn(column string) {
	if !ticketcore.NormalBoardStatus(ticketcore.Status(column)) {
		return
	}

	m.boardColumn = column
	m.previousBoardColumn = column
	m.selectFirstFilteredTicket()
}

func defaultBoardColumn(tickets []TicketView) string {
	for _, column := range ticketcore.NormalBoardColumns {
		for _, ticket := range tickets {
			if ticket.BoardColumn == column {
				return column
			}
		}
	}

	return string(ticketcore.StatusBacklog)
}
