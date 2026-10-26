package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chr/internal/core"
)

func (m *Model) renderDash() string {
	var output strings.Builder

	output.WriteString("Dash\n")
	output.WriteString("Agent Work\n\n")
	output.WriteString("Readiness\n")

	for _, line := range m.renderDashReadinessLines(4) {
		output.WriteString("  ")
		output.WriteString(line)
		output.WriteString("\n")
	}

	output.WriteString("\n")
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

	output.WriteString(m.renderNativeWorkflow())
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

	for _, line := range m.renderDashActionLines() {
		output.WriteString("  ")
		output.WriteString(line)
		output.WriteString("\n")
	}

	return output.String()
}

func (m *Model) renderDashWorkspace(width, height int) string {
	queueLines := append([]string{
		fmt.Sprintf("Ready queue: %d", len(m.readyAgentWorkTickets())),
		"Priority: " + displayDashPriorityFilter(m.agentWorkPriorityFilter),
		"Order: " + displayDashAgentWorkOrder(m.agentWorkOrder),
		"Action: a run all ready tickets",
		"",
	}, m.renderAgentWorkQueueLines(maxInt(height-5, 1))...)

	currentLines := append(m.renderDashReadinessLines(maxInt(height/3, 3)), "")
	currentLines = append(currentLines, m.nativeWorkflowLines()...)
	currentLines = append(currentLines, m.renderAgentWorkCurrentLines()...)

	recentLines := append([]string{
		"Actions",
	}, indentDashActionLines(m.renderDashActionLines())...)
	recentLines = append(recentLines, "",
		"Recent",
	)
	recentLines = append(recentLines, m.renderAgentWorkRecentLines()...)

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

func (m *Model) renderDashReadinessLines(limit int) []string {
	view := m.workReadiness
	lines := []string{"Status: " + displayWorkReadinessStatus(view.Status)}

	lines = appendWorkReadinessContextLines(lines, view)
	lines = appendWorkReadinessAppliedRepairLines(lines, view.RepairsApplied)

	diagnosticLines := renderDashReadinessDiagnosticLines(view.Diagnostics, limit)
	if len(diagnosticLines) == 0 {
		if view.ready() {
			return append(lines, "No blocking diagnostics.")
		}

		return lines
	}

	lines = append(lines, diagnosticLines...)

	return appendDashReadinessRepairGuidance(lines, view)
}

func appendWorkReadinessContextLines(lines []string, view WorkReadinessView) []string {
	if strings.TrimSpace(view.WorkspaceMode) != "" {
		lines = append(lines, "Workspace mode: "+view.WorkspaceMode)
	}

	if strings.TrimSpace(view.CurrentBranch) != "" {
		lines = append(lines, "Current branch: "+view.CurrentBranch)
	}

	if strings.TrimSpace(view.IntegrationBranch) != "" {
		lines = append(lines, "Expected branch: "+view.IntegrationBranch)
	}

	if strings.TrimSpace(view.TicketWorktree.State) != "" {
		lines = append(lines, "Ticket worktree: "+view.TicketWorktree.State)
	}

	return lines
}

func appendWorkReadinessAppliedRepairLines(lines []string, repairs []WorkReadinessAppliedRepairView) []string {
	for _, repair := range repairs {
		lines = append(lines, renderAppliedReadinessRepairLine(repair))
	}

	return lines
}

func renderDashReadinessDiagnosticLines(diagnostics []WorkReadinessDiagnosticView, limit int) []string {
	lines := []string{}

	remaining := 0

	for index, diagnostic := range diagnostics {
		if index >= limit {
			remaining = len(diagnostics) - index

			break
		}

		lines = append(lines, renderWorkReadinessDiagnosticLine(diagnostic))
	}

	if remaining > 0 {
		lines = append(lines, fmt.Sprintf("+%d more diagnostics", remaining))
	}

	return lines
}

func appendDashReadinessRepairGuidance(lines []string, view WorkReadinessView) []string {
	if view.safeRepairAvailable() {
		return append(lines, "Repair: x apply safe repair")
	}

	if view.blocked() {
		return append(lines, "Repair: manual action required")
	}

	return lines
}

func (m *Model) renderDashActionLines() []string {
	lines := []string{
		"a Run all ready tickets",
		"enter Start selected ticket",
		"p Cycle priority filter",
		"o Cycle queue order",
	}
	if m.workReadiness.safeRepairAvailable() {
		lines = append(lines, "x Apply safe readiness repair")
	}

	return append(lines, "s Request agent stop")
}

func indentDashActionLines(lines []string) []string {
	indented := make([]string, 0, len(lines))
	for _, line := range lines {
		indented = append(indented, "  "+line)
	}

	return indented
}

func displayWorkReadinessStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "unknown"
	}

	return status
}

func renderWorkReadinessDiagnosticLine(diagnostic WorkReadinessDiagnosticView) string {
	line := diagnostic.Code
	if strings.TrimSpace(diagnostic.Detail) != "" {
		line += ": " + diagnostic.Detail
	}

	return line
}

func renderAppliedReadinessRepairLine(repair WorkReadinessAppliedRepairView) string {
	line := "Applied: " + repair.Code
	if strings.TrimSpace(repair.Detail) != "" {
		line += ": " + repair.Detail
	}

	return line
}

func (m *Model) readyAgentWorkTickets() []TicketView {
	tickets := []TicketView{}
	for _, ticket := range m.tickets {
		if m.matchesReadyAgentWorkFilter(ticket) {
			tickets = append(tickets, ticket)
		}
	}

	sortReadyAgentWorkTickets(tickets, m.agentWorkOrder)

	return tickets
}

func (m *Model) matchesReadyAgentWorkFilter(ticket TicketView) bool {
	if ticket.Status != "ready_for_dev" {
		return false
	}

	return m.agentWorkPriorityFilter == "" || ticket.Priority == m.agentWorkPriorityFilter
}

func (m *Model) renderAgentWorkQueueLines(limit int) []string {
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
	orders := []core.TicketLoopOrder{
		core.TicketLoopOrderCreatedAt,
		core.TicketLoopOrderPriority,
		core.TicketLoopOrderSeverity,
		core.TicketLoopOrderPrioritySeverity,
		core.TicketLoopOrderKey,
	}
	m.agentWorkOrder = nextTicketLoopOrder(orders, m.agentWorkOrder)
}

func sortReadyAgentWorkTickets(tickets []TicketView, order core.TicketLoopOrder) {
	switch order {
	case core.TicketLoopOrderPriority:
		slices.SortFunc(tickets, compareTicketViewsByPriority)
	case core.TicketLoopOrderSeverity:
		slices.SortFunc(tickets, compareTicketViewsBySeverity)
	case core.TicketLoopOrderPrioritySeverity:
		slices.SortFunc(tickets, func(left, right TicketView) int {
			return firstNonZeroInt(
				compareTicketViewsByPriority(left, right),
				compareTicketViewsBySeverity(left, right),
				strings.Compare(left.Key, right.Key),
			)
		})
	case core.TicketLoopOrderKey:
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

func nextTicketLoopOrder(values []core.TicketLoopOrder, current core.TicketLoopOrder) core.TicketLoopOrder {
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

func displayDashAgentWorkOrder(order core.TicketLoopOrder) string {
	switch order {
	case "", core.TicketLoopOrderCreatedAt:
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

func (m *Model) renderAgentWorkCurrentLines() []string {
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

	if m.observability.LatestEvent.Phase != "" {
		lines = append(lines,
			fmt.Sprintf("Phase: %s", m.observability.LatestEvent.Phase),
			fmt.Sprintf("Last event: %s", m.observability.LatestEvent.Message),
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

func (m *Model) renderAgentWorkRecentLines() []string {
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
