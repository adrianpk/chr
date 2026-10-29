package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderWorkflowWorkspace(width, height int) string {
	lines := []string{fmt.Sprintf("Page offset %d; [ previous / ] next", m.workflowPage.Offset), "up/down select; Enter refresh; Esc queue", "n/m member pages; r reload", ""}

	start := m.workflowListStart(max(1, (height-8)/3))
	for _, item := range m.workflowPage.Items[start:] {
		prefix := "  "
		if item.ID == m.workflow.ID {
			prefix = "> "
		}

		lines = append(lines, prefix+emptyValue(item.Title)+"  ["+item.State+"]", "  "+item.Extent+" "+item.Identity, "  "+item.ID)
	}

	if len(m.workflowPage.Items) == 0 {
		lines = append(lines, "No retained workflows.")
	}

	detail := append([]string{"c continue; p pause; x cancel; u resume; a approve; R reconcile", "g extend allowance; d answer decision; t ticket; C case", "b preview bounded work"}, m.nativeWorkflowLines()...)

	detail = append(detail, "Linked ticket (t): "+emptyValue(m.workflowLinkedTicket()), "Linked case (C): open from extent or linked ticket")
	for _, action := range m.workflow.Actions {
		line := action.Label
		if action.Reason != "" {
			line += " unavailable: " + action.Reason
		}

		detail = append(detail, line)
	}

	left := max(30, width/3)
	for index := range lines {
		lines[index] = fitPlainLine(lines[index], max(1, left-4))
	}

	if m.maximizedPane == 1 {
		return m.renderDetailPanel("Workflow", detail, width, height, true)
	}

	return fitBlock(lipgloss.JoinHorizontal(lipgloss.Top, renderPanel("Retained workflows", lines, left, height, m.selectedPane == 0), panelVerticalGap(height), m.renderDetailPanel("Workflow", detail, width-left-panelVerticalGapSize, height, m.selectedPane != 0)), width, height)
}

func (m *Model) handleWorkflowBrowseKey(key string) (tea.Cmd, bool) {
	if command, ok := m.handleWorkflowControlKey(key); ok {
		return command, true
	}

	if command, ok := m.handleWorkflowPageKey(key); ok {
		return command, true
	}

	switch key {
	case "w", "esc":
		m.dashWorkflowDetails = false

		return nil, true
	case "down", "j":
		return m.moveWorkflowSelection(1), true
	case "up", "k":
		return m.moveWorkflowSelection(-1), true
	case "r":
		if m.selectedPane == 0 {
			return m.readWorkflowPage(m.workflowPage.Offset), true
		}

		return m.readWorkflowMembers(0), true
	case "enter":
		return m.readWorkflowMembers(0), true
	default:
		return nil, false
	}
}

func (m *Model) moveWorkflowSelection(delta int) tea.Cmd {
	if m.selectedPane != 0 {
		m.scrollDetail(delta)

		return nil
	}

	items := m.workflowPage.Items
	for index, item := range items {
		if item.ID == m.workflow.ID {
			m.workflow = items[(index+delta+len(items))%len(items)]
			m.detailScroll = 0

			return m.readWorkflowMembers(0)
		}
	}

	return nil
}

func (m *Model) readWorkflowPage(offset int) tea.Cmd {
	return m.startWorkAction(WorkActionRequest{Action: WorkActionWorkflowPage, Workflow: &WorkflowActionRequest{Offset: offset}})
}

func (m *Model) readWorkflowMembers(offset int) tea.Cmd {
	if m.workflow.ID == "" {
		return m.readWorkflowPage(m.workflowPage.Offset)
	}

	if offset < 0 {
		return nil
	}

	request := WorkflowActionRequest{ID: m.workflow.ID, Offset: offset}
	if offset > 0 {
		request.ExpectedRevision = m.workflow.Revision
	}

	return m.startWorkAction(WorkActionRequest{Action: WorkActionWorkflowStatus, Workflow: &request})
}

func (m *Model) workflowListStart(visible int) int {
	for index, item := range m.workflowPage.Items {
		if item.ID == m.workflow.ID {
			return max(0, index-visible+1)
		}
	}

	return 0
}

func (m *Model) handleWorkflowPageKey(key string) (tea.Cmd, bool) {
	switch key {
	case "]":
		if m.workflowPage.NextOffset >= 0 {
			return m.readWorkflowPage(m.workflowPage.NextOffset), true
		}

		return nil, true
	case "[":
		return m.readWorkflowPage(max(0, m.workflowPage.Offset-16)), true
	case "n":
		return m.readWorkflowMembers(m.workflow.NextOffset), true
	case "m":
		return m.readWorkflowMembers(max(0, m.workflow.Offset-8)), true
	default:
		return nil, false
	}
}
