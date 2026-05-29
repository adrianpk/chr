package chctui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderDash() string {
	var output strings.Builder

	output.WriteString("Dash\n")
	output.WriteString("Agent Work\n\n")
	output.WriteString("Queue\n")

	queue := m.readyAgentWorkTickets()
	if len(queue) == 0 {
		output.WriteString("  No ready tickets queued.\n")
	} else {
		for _, line := range m.renderAgentWorkQueueLines(6) {
			output.WriteString("  ")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	output.WriteString("\nCurrent\n")

	for _, line := range m.renderAgentWorkCurrentLines() {
		output.WriteString("  ")
		output.WriteString(line)
		output.WriteString("\n")
	}

	output.WriteString("\nRecent\n")

	for _, line := range m.renderAgentWorkRecentLines() {
		output.WriteString("  ")
		output.WriteString(line)
		output.WriteString("\n")
	}

	output.WriteString("\nActions\n")
	output.WriteString("  Start selected ticket\n")
	output.WriteString("  Run all ready tickets\n")
	output.WriteString("  Request agent stop")

	return output.String()
}

func (m Model) renderDashWorkspace(width, height int) string {
	queueLines := append([]string{
		fmt.Sprintf("Ready queue: %d", len(m.readyAgentWorkTickets())),
		"Order: oldest first",
		"",
	}, m.renderAgentWorkQueueLines(maxInt(height-5, 1))...)

	currentLines := m.renderAgentWorkCurrentLines()
	recentLines := append([]string{
		"Actions",
		"  Start selected ticket",
		"  Run all ready tickets",
		"  Request agent stop",
		"",
		"Recent",
	}, m.renderAgentWorkRecentLines()...)

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Queue", queueLines, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel("Current", currentLines, width, height, true), width, height)
	case 2:
		return fitBlock(m.renderDetailPanel("Recent", recentLines, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	rightTopHeight, rightBottomHeight := m.splitHeights(height)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Queue", queueLines, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.renderDetailPanel("Current", currentLines, rightWidth, rightTopHeight, m.selectedPane == 1),
			m.renderDetailPanel("Recent", recentLines, rightWidth, rightBottomHeight, m.selectedPane == 2),
		),
	), width, height)
}

func (m Model) readyAgentWorkTickets() []TicketView {
	tickets := []TicketView{}
	for _, ticket := range m.tickets {
		if ticket.Status == "ready_for_dev" {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

func (m Model) renderAgentWorkQueueLines(limit int) []string {
	queue := m.readyAgentWorkTickets()
	if len(queue) == 0 {
		return []string{"No ready tickets queued."}
	}

	lines := []string{}

	for index, ticket := range queue {
		if index >= limit {
			lines = append(lines, fmt.Sprintf("+%d more", len(queue)-index))

			break
		}

		prefix := " "
		if ticket.Index == m.selectedTicket {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %s  %s  %s", prefix, ticket.Key, ticket.Priority, ticket.Title))
	}

	return lines
}

func (m Model) renderAgentWorkCurrentLines() []string {
	loop := m.observability.LatestLoop
	execution := m.observability.LatestExecution

	if loop.ID == "" && execution.ID == "" {
		return []string{"No active or recent agent work."}
	}

	lines := []string{}
	if loop.ID != "" {
		lines = append(lines,
			fmt.Sprintf("Loop: %s", loop.ID),
			fmt.Sprintf("Status: %s", loop.Status),
			fmt.Sprintf("Counts: %d eligible / %d processed / %d blocked / %d failed",
				loop.EligibleCount,
				loop.ProcessedCount,
				loop.BlockedCount,
				loop.FailedCount,
			),
		)
	}

	if execution.ID != "" {
		lines = append(lines,
			fmt.Sprintf("Ticket: %s", execution.TicketKey),
			fmt.Sprintf("Execution: %s", execution.ID),
			fmt.Sprintf("Execution status: %s", execution.Status),
		)
	}

	return lines
}

func (m Model) renderAgentWorkRecentLines() []string {
	lines := []string{}

	if m.observability.LatestExecution.ID != "" {
		lines = append(lines, fmt.Sprintf(
			"Execution %s  %s  %s",
			m.observability.LatestExecution.ID,
			m.observability.LatestExecution.TicketKey,
			m.observability.LatestExecution.Status,
		))
	}

	if m.observability.LatestLoop.ID != "" {
		lines = append(lines, fmt.Sprintf(
			"Loop %s  %s",
			m.observability.LatestLoop.ID,
			m.observability.LatestLoop.Status,
		))
	}

	if len(lines) == 0 {
		return []string{"No completed work recorded."}
	}

	return lines
}
