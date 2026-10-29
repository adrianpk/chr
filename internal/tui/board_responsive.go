package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func (m *Model) renderResponsiveBoard(width, height int, detail []string) string {
	detailWidth := max(32, width/4)
	boardWidth := width - detailWidth - panelVerticalGapSize
	attentionHeight := 7
	rowHeight := (height - attentionHeight) / 2

	rows := []string{}
	for start := 0; start < len(ticketcore.NormalBoardColumns); start += 3 {
		rows = append(rows, m.renderMaximizedBoardRow(ticketcore.NormalBoardColumns[start:start+3], boardWidth, rowHeight))
	}

	rows = append(rows, renderPanel("Attention", m.renderBoardAttentionFocusList(attentionHeight-2), boardWidth, attentionHeight, m.selectedPane == 0 && m.boardAttentionActive()))
	board := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return fitBlock(lipgloss.JoinHorizontal(lipgloss.Top, board, panelVerticalGap(height), m.renderDetailPanel("Detail", detail, detailWidth, height, m.selectedPane == 1)), width, height)
}

func (m *Model) visibleBoardTickets(column string) []TicketView {
	tickets := m.ticketsForBoardColumn(column)
	if column == boardAttentionColumn {
		tickets = m.attentionTickets()
	}

	result := []TicketView{}

	for _, ticket := range tickets {
		if ticket.Matches(strings.ToLower(m.ticketFilter)) {
			result = append(result, ticket)
		}
	}

	return result
}
