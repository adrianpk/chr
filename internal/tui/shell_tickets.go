package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func (m *Model) renderTicketWorkspace(width, height int) string {
	detail := m.renderSelectedTicketDetail()
	detailLines := strings.Split(detail, "\n")[1:]

	switch m.maximizedPane {
	case 0:
		if m.SelectedView() == ViewBoard {
			return m.renderMaximizedBoardWorkspace(width, height)
		}

		list := m.renderTicketFocusList(maxInt(width-4, 1), height-2)

		return fitBlock(renderPanel("Tickets", list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel("Detail", detailLines, width, height, true), width, height)
	}

	horizontal := m.splitWidths(width)
	leftWidth, rightWidth := horizontal.Left, horizontal.Right
	list := m.renderTicketFocusList(maxInt(leftWidth-4, 1), height-2)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Tickets", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel("Detail", detailLines, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m *Model) renderTicketFocusList(width, limit int) []string {
	lines := m.renderBoardFocusHeaderLines(limit)

	tickets := m.filteredTickets()
	if len(tickets) == 0 {
		return appendLimitedLines(lines, []string{"No tickets match."}, limit)
	}

	lines = appendLimitedLines(lines, []string{shellMutedStyle.Render(renderCompactTicketHeaderLine(width))}, limit)

	for _, ticket := range tickets {
		prefix := " "
		if ticket.Index == m.selectedTicket {
			prefix = ">"
		}

		lines = append(lines, renderCompactTicketFocusLine(prefix, ticket, width))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m *Model) renderMaximizedBoardWorkspace(width, height int) string {
	if m.boardAttentionActive() {
		return fitBlock(renderPanel("Attention", m.renderBoardAttentionFocusList(height-2), width, height, true), width, height)
	}

	columnsPerRow := boardMaximizedColumnsPerRow(width)
	rows := boardMaximizedRows(len(core.NormalBoardColumns), columnsPerRow)
	rowHeight := maxInt(height/rows, 1)
	renderedRows := make([]string, 0, rows)

	for row := range rows {
		start := row * columnsPerRow

		end := minInt(start+columnsPerRow, len(core.NormalBoardColumns))
		if start >= end {
			break
		}

		renderedRows = append(renderedRows, m.renderMaximizedBoardRow(
			core.NormalBoardColumns[start:end],
			width,
			rowHeight,
		))
	}

	return fitBlock(lipgloss.JoinVertical(lipgloss.Left, renderedRows...), width, height)
}

func (m *Model) renderMaximizedBoardRow(columns []string, width, height int) string {
	gapWidth := panelVerticalGapSize * maxInt(len(columns)-1, 0)
	columnWidth := maxInt((width-gapWidth)/maxInt(len(columns), 1), 1)
	panels := make([]string, 0, len(columns)*2-1)

	for index, column := range columns {
		if index > 0 {
			panels = append(panels, panelVerticalGap(height))
		}

		tickets := m.ticketsForBoardColumn(column)
		lines := m.renderMaximizedBoardColumnLines(tickets, maxInt(columnWidth-4, 1), height-2)
		title := fmt.Sprintf("%s (%d)", core.BoardColumnTitle(column), len(tickets))

		panels = append(panels, renderPanel(title, lines, columnWidth, height, m.activeBoardColumn() == column))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

func (m *Model) renderMaximizedBoardColumnLines(tickets []TicketView, width, limit int) []string {
	if len(tickets) == 0 {
		return []string{"No tickets."}
	}

	lines := make([]string, 0, maxInt(limit, 1))
	start := m.maximizedBoardColumnStart(tickets, width, limit)

	for _, ticket := range tickets[start:] {
		card := renderMaximizedBoardTicketCard(m.maximizedBoardTicketPrefix(ticket), ticket, width)
		if len(lines) > 0 && len(lines)+len(card) > limit {
			break
		}

		lines = appendBoardCardLines(lines, card, limit)

		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m *Model) maximizedBoardTicketPrefix(ticket TicketView) string {
	switch {
	case len(ticket.AttentionReasons) > 0:
		return "!"
	case ticket.Index == m.selectedTicket:
		return ">"
	default:
		return " "
	}
}

func appendBoardCardLines(lines, card []string, limit int) []string {
	for _, line := range card {
		if len(lines) >= limit {
			break
		}

		lines = append(lines, line)
	}

	return lines
}

func (m *Model) maximizedBoardColumnStart(tickets []TicketView, width, limit int) int {
	selectedPosition := m.selectedTicketPosition(tickets)
	if selectedPosition <= 0 {
		return 0
	}

	cardHeight := len(renderMaximizedBoardTicketCard(" ", tickets[selectedPosition], width))
	if cardHeight <= 0 {
		return 0
	}

	visibleCards := maxInt(limit/cardHeight, 1)
	if selectedPosition < visibleCards {
		return 0
	}

	return selectedPosition - visibleCards + 1
}

func renderMaximizedBoardTicketCard(prefix string, ticket TicketView, width int) []string {
	if width < 8 {
		metadata := fmt.Sprintf(
			"%s %s  %s  %s",
			prefix,
			ticket.Key,
			compactTicketStatusLabel(ticket.Status),
			compactTicketPriorityLabel(ticket.Priority),
		)

		return []string{
			fitPlainLine(metadata, width),
			fitPlainLine("  "+truncatePlainLine(ticket.Title, maxInt(width-2, 1)), width),
		}
	}

	innerWidth := maxInt(width-2, 1)
	metadata := fmt.Sprintf(
		"%s %s  %s  %s",
		prefix,
		ticket.Key,
		compactTicketStatusLabel(ticket.Status),
		compactTicketPriorityLabel(ticket.Priority),
	)
	metadata = truncatePlainLine(metadata, innerWidth)
	titleLines := maximizedBoardTicketTitleLines(ticket.Title, innerWidth, 3)

	return []string{
		"┌" + strings.Repeat("─", innerWidth) + "┐",
		"│" + fitPlainLine(metadata, innerWidth) + "│",
		"├" + strings.Repeat("─", innerWidth) + "┤",
		"│" + fitPlainLine(titleLines[0], innerWidth) + "│",
		"│" + fitPlainLine(titleLines[1], innerWidth) + "│",
		"│" + fitPlainLine(titleLines[2], innerWidth) + "│",
		"└" + strings.Repeat("─", innerWidth) + "┘",
	}
}

func maximizedBoardTicketTitleLines(title string, width, lineCount int) []string {
	lines := wrapPlainText(title, width)
	fitted := make([]string, 0, lineCount)

	for _, line := range lines {
		if len(fitted) >= lineCount {
			break
		}

		fitted = append(fitted, truncatePlainLine(line, width))
	}

	for len(fitted) < lineCount {
		fitted = append(fitted, "")
	}

	return fitted
}

func boardMaximizedColumnsPerRow(width int) int {
	switch {
	case width >= 132:
		return 6
	case width >= 90:
		return 3
	case width >= 64:
		return 2
	default:
		return 1
	}
}

func boardMaximizedRows(columnCount, columnsPerRow int) int {
	return maxInt((columnCount+columnsPerRow-1)/columnsPerRow, 1)
}

func (m *Model) renderBoardAttentionFocusList(limit int) []string {
	tickets := m.attentionTickets()
	if len(tickets) == 0 {
		return []string{"No attention tickets."}
	}

	lines := []string{shellMutedStyle.Render("Key     Status       Priority  Title")}

	for _, ticket := range tickets {
		prefix := " "
		if ticket.Index == m.selectedTicket {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %-7s %-12s %-9s %s", prefix, ticket.Key, ticket.Status, ticket.Priority, ticket.Title))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func renderCompactTicketHeaderLine(width int) string {
	return renderCompactTicketColumns(compactTicketRow{Prefix: " ", Key: "Key", Status: "Sts", Priority: "Pri", Title: "Title", Header: true}, width)
}

func renderCompactTicketFocusLine(prefix string, ticket TicketView, width int) string {
	return renderCompactTicketColumns(compactTicketRow{Prefix: prefix, Key: ticket.Key, Status: compactTicketStatusLabel(ticket.Status), Priority: compactTicketPriorityLabel(ticket.Priority), Title: ticket.Title, Header: false}, width)
}

func renderCompactTicketColumns(row compactTicketRow, width int) string {
	line := strings.Builder{}
	line.WriteString(compactTicketCell(row.Prefix, 1, false))
	line.WriteString(" ")
	line.WriteString(compactTicketCell(row.Key, 7, row.Header))
	line.WriteString("  ")
	line.WriteString(compactTicketCell(row.Status, 5, row.Header))
	line.WriteString("  ")
	line.WriteString(compactTicketCell(row.Priority, 4, row.Header))
	line.WriteString("  ")

	titleWidth := maxInt(width-lipgloss.Width(line.String()), 1)
	line.WriteString(truncatePlainLine(row.Title, titleWidth))

	return fitPlainLine(line.String(), width)
}

func compactTicketCell(value string, width int, center bool) string {
	if width <= 0 {
		return ""
	}

	value = truncatePlainLine(value, width)
	if !center {
		return fitPlainLine(value, width)
	}

	valueWidth := lipgloss.Width(value)
	left := maxInt((width-valueWidth)/2, 0)
	right := maxInt(width-valueWidth-left, 0)

	return strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
}

func compactTicketStatusLabel(status string) string {
	labels := map[string]string{
		string(ticketcore.StatusBacklog):         "BL",
		string(ticketcore.StatusReadyForDev):     "RFD",
		string(ticketcore.StatusInProgress):      "IP",
		string(ticketcore.StatusReadyForReview):  "RFR",
		string(ticketcore.StatusReadyForMerge):   "RFM",
		string(ticketcore.StatusDone):            "DONE",
		string(ticketcore.StatusOpenLegacy):      "OPN",
		string(ticketcore.StatusReadyLegacy):     "RDY",
		string(ticketcore.StatusReviewingLegacy): "REV",
		string(ticketcore.StatusBlockedLegacy):   "BLK",
		string(ticketcore.StatusSolvedLegacy):    "SLV",
		string(ticketcore.StatusMergedLegacy):    "DONE",
		string(ticketcore.StatusClosedLegacy):    "CLS",
	}

	label, ok := labels[strings.TrimSpace(status)]
	if ok {
		return label
	}

	return truncatePlainLine(humanizeCompactToken(status), 5)
}

func compactTicketPriorityLabel(priority string) string {
	switch strings.TrimSpace(priority) {
	case "urgent":
		return "Urg"
	case "high":
		return "High"
	case "normal":
		return "Norm"
	case "low":
		return "Low"
	default:
		return truncatePlainLine(humanizeCompactToken(priority), 4)
	}
}

func compactBoardColumnLabel(column string) string {
	switch strings.TrimSpace(column) {
	case boardAttentionColumn:
		return "ATT"
	case string(ticketcore.StatusBacklog):
		return "BL"
	case string(ticketcore.StatusReadyForDev):
		return "RFD"
	case string(ticketcore.StatusInProgress):
		return "IP"
	case string(ticketcore.StatusReadyForReview):
		return "RFR"
	case string(ticketcore.StatusReadyForMerge):
		return "RFM"
	case string(ticketcore.StatusDone):
		return "Done"
	default:
		return truncatePlainLine(humanizeCompactToken(column), 5)
	}
}

func humanizeCompactToken(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "_", " "))
	if value == "" {
		return "-"
	}

	words := strings.Fields(value)
	for index, word := range words {
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}

func (m *Model) renderBoardFocusHeaderLines(limit int) []string {
	if m.SelectedView() != ViewBoard {
		return []string{}
	}

	if m.boardAttentionActive() {
		return appendLimitedLines([]string{}, []string{
			"Attention",
			"",
		}, limit)
	}

	return appendLimitedLines([]string{}, []string{
		fmt.Sprintf("%s (%d/%d)",
			compactBoardColumnLabel(m.activeBoardColumn()),
			m.activeBoardColumnPosition()+1,
			len(core.NormalBoardColumns),
		),
		"",
	}, limit)
}

func appendLimitedLines(lines, additions []string, limit int) []string {
	for _, line := range additions {
		if len(lines) >= limit {
			return lines
		}

		lines = append(lines, line)
	}

	return lines
}

type compactTicketRow struct {
	Prefix, Key, Status, Priority, Title string
	Header                               bool
}
