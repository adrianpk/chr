package chctui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chc/internal/chc"
)

func (m Model) renderDash() string {
	var output strings.Builder

	output.WriteString("Dash\n")
	output.WriteString("Agent Work\n\n")
	output.WriteString("Queue\n")
	output.WriteString("  Priority: ")
	output.WriteString(displayDashPriorityFilter(m.agentWorkPriorityFilter))
	output.WriteString("\n")
	output.WriteString("  Order: ")
	output.WriteString(displayDashAgentWorkOrder(m.agentWorkOrder))
	output.WriteString("\n")

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
	output.WriteString("  a Run all ready tickets\n")
	output.WriteString("  enter Start selected ticket\n")
	output.WriteString("  s Request agent stop")

	return output.String()
}

func (m Model) renderDashWorkspace(width, height int) string {
	queueLines := append([]string{
		fmt.Sprintf("Ready queue: %d", len(m.readyAgentWorkTickets())),
		"Priority: " + displayDashPriorityFilter(m.agentWorkPriorityFilter),
		"Order: " + displayDashAgentWorkOrder(m.agentWorkOrder),
		"Action: a run all ready tickets",
		"",
	}, m.renderAgentWorkQueueLines(maxInt(height-5, 1))...)

	currentLines := m.renderAgentWorkCurrentLines()
	recentLines := append([]string{
		"Actions",
		"  a Run all ready tickets",
		"  enter Start selected ticket",
		"  p Cycle priority filter",
		"  o Cycle queue order",
		"  s Request agent stop",
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
		if m.matchesReadyAgentWorkFilter(ticket) {
			tickets = append(tickets, ticket)
		}
	}

	sortReadyAgentWorkTickets(tickets, m.agentWorkOrder)

	return tickets
}

func (m Model) matchesReadyAgentWorkFilter(ticket TicketView) bool {
	if ticket.Status != "ready_for_dev" {
		return false
	}

	return m.agentWorkPriorityFilter == "" || ticket.Priority == m.agentWorkPriorityFilter
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

func (m *Model) cycleAgentWorkPriorityFilter() {
	filters := []string{"", "critical", "high", "normal", "low"}
	m.agentWorkPriorityFilter = nextStringValue(filters, m.agentWorkPriorityFilter)
}

func (m *Model) cycleAgentWorkOrder() {
	orders := []chc.TicketLoopOrder{
		chc.TicketLoopOrderCreatedAt,
		chc.TicketLoopOrderPriority,
		chc.TicketLoopOrderSeverity,
		chc.TicketLoopOrderPrioritySeverity,
		chc.TicketLoopOrderKey,
	}
	m.agentWorkOrder = nextTicketLoopOrder(orders, m.agentWorkOrder)
}

func sortReadyAgentWorkTickets(tickets []TicketView, order chc.TicketLoopOrder) {
	switch order {
	case chc.TicketLoopOrderPriority:
		slices.SortFunc(tickets, compareTicketViewsByPriority)
	case chc.TicketLoopOrderSeverity:
		slices.SortFunc(tickets, compareTicketViewsBySeverity)
	case chc.TicketLoopOrderPrioritySeverity:
		slices.SortFunc(tickets, func(left, right TicketView) int {
			return firstNonZeroInt(
				compareTicketViewsByPriority(left, right),
				compareTicketViewsBySeverity(left, right),
				strings.Compare(left.Key, right.Key),
			)
		})
	case chc.TicketLoopOrderKey:
		slices.SortFunc(tickets, func(left, right TicketView) int {
			return strings.Compare(left.Key, right.Key)
		})
	}
}

func compareTicketViewsByPriority(left, right TicketView) int {
	return firstNonZeroInt(
		compareInt(ticketPriorityRank(left.Priority), ticketPriorityRank(right.Priority)),
		strings.Compare(left.Key, right.Key),
	)
}

func compareTicketViewsBySeverity(left, right TicketView) int {
	return firstNonZeroInt(
		compareInt(ticketSeverityRank(left.Severity), ticketSeverityRank(right.Severity)),
		strings.Compare(left.Key, right.Key),
	)
}

func nextStringValue(values []string, current string) string {
	for index, value := range values {
		if value == current {
			return values[(index+1)%len(values)]
		}
	}

	return values[0]
}

func nextTicketLoopOrder(values []chc.TicketLoopOrder, current chc.TicketLoopOrder) chc.TicketLoopOrder {
	for index, value := range values {
		if value == current {
			return values[(index+1)%len(values)]
		}
	}

	return values[0]
}

func displayDashPriorityFilter(priority string) string {
	if priority == "" {
		return "all"
	}

	return priority
}

func displayDashAgentWorkOrder(order chc.TicketLoopOrder) string {
	switch order {
	case "", chc.TicketLoopOrderCreatedAt:
		return "oldest first"
	default:
		return string(order)
	}
}

func ticketPriorityRank(priority string) int {
	switch priority {
	case "critical":
		return 0
	case "high":
		return 1
	case "normal":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func ticketSeverityRank(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "high":
		return 1
	case "normal":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func firstNonZeroInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}

func compareInt(left, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
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
