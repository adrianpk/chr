package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderDocumentWorkspace(width, height int, title string, documents []DocumentView) string {
	list := m.renderDocumentFocusList(documents, height-2)
	detail := m.renderSelectedDocumentDetailFor(documents)

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel(title, list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel("Preview", strings.Split(detail, "\n")[1:], width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel(title, list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel("Preview", strings.Split(detail, "\n")[1:], rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m *Model) renderDumpWorkspace(width, height int) string {
	list := m.renderDumpFocusList(height - 2)
	detailTitle, detailLines := m.renderDumpDetailPanel(maxInt(width-4, 1), height-2)

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Dumps", list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel(detailTitle, detailLines, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	detailTitle, detailLines = m.renderDumpDetailPanel(maxInt(rightWidth-4, 1), height-2)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Dumps", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel(detailTitle, detailLines, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

//nolint:unused // Kept for document-family workspaces planned after shell consolidation.
func (m *Model) renderDocumentFamilyWorkspace(width, height int, title string, families map[string]bool) string {
	return m.renderDocumentWorkspace(width, height, title, m.documentsByFamily(families))
}

func (m *Model) renderQuizWorkspace(width, height int) string {
	list := m.renderQuizFocusList(height - 2)
	detailTitle := m.quizDetailPanelTitle()
	detail := m.renderSelectedQuizDetailLines(maxInt(width-4, 1))

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Quizzes", list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel(detailTitle, detail, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	detail = m.renderSelectedQuizDetailLines(maxInt(rightWidth-4, 1))

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Quizzes", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel(detailTitle, detail, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m *Model) renderSpecWorkspace(width, height int) string {
	list := m.renderSpecFocusList(height - 2)
	detail := m.renderSelectedSpecDetailLines(maxInt(width-4, 1))

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Specs", list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel("Spec", detail, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	detail = m.renderSelectedSpecDetailLines(maxInt(rightWidth-4, 1))

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Specs", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel("Spec", detail, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m *Model) renderPlanWorkspace(width, height int) string {
	list := m.renderPlanFocusList(height - 2)
	detail := m.renderSelectedPlanDetailLines(maxInt(width-4, 1))

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Plans", list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel("Plan", detail, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	detail = m.renderSelectedPlanDetailLines(maxInt(rightWidth-4, 1))

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Plans", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel("Plan", detail, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m *Model) quizDetailPanelTitle() string {
	quiz, ok := m.selectedQuizView()
	if ok && m.quizAnswer.Active && quizScopeLabel(m.quizAnswer.Scope) == quizScopeLabel(quiz.Scope) {
		return "Answer"
	}

	return "Quiz"
}

func (m *Model) renderAuthoringWorkspace(width, height int) string {
	list := m.renderAuthoringFocusList(height - 2)
	detail := m.renderSelectedAuthoringSessionDetail()

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Authoring", list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel("Session", strings.Split(detail, "\n")[1:], width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Authoring", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel("Session", strings.Split(detail, "\n")[1:], rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m *Model) renderOperationsWorkspace(width, height int) string {
	agent := strings.Split(m.renderAgent(), "\n")
	runtime := strings.Split(m.renderRuntime()+"\n\n"+m.renderWorkspaceIndexHealth(), "\n")

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Agent", agent[1:], width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel("Runtime", runtime[1:], width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Agent", agent[1:], leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel("Runtime", runtime[1:], rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m *Model) renderDocumentFocusList(documents []DocumentView, limit int) []string {
	if len(documents) == 0 {
		return []string{"No documents found."}
	}

	lines := []string{shellMutedStyle.Render("Type       Status      Title")}

	for _, document := range documents {
		prefix := " "
		if document.Index == m.selectedDocument {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %-10s %-11s %s", prefix, document.Family, document.Status, document.Title))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m *Model) renderAuthoringFocusList(limit int) []string {
	if len(m.authoringSessions) == 0 {
		return []string{"No authoring sessions found."}
	}

	lines := []string{shellMutedStyle.Render("Status      Target     Topic")}

	for _, session := range m.authoringSessions {
		prefix := " "
		if session.Index == m.selectedAuthoringSession {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %-11s %-10s %s", prefix, session.Status, session.Target, session.Topic))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}
