package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	minShellHeight       = 12
	panelVerticalGapSize = 1
)

var (
	shellTopbarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("24")).
				Bold(true)
	shellPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)
	shellFocusedPanelStyle = shellPanelStyle.
				BorderForeground(lipgloss.Color("80"))
	shellTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("80")).
			Bold(true)
	shellMutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	shellStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("80")).
				Bold(true)
	shellButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("80")).
				Bold(true)
	shellButtonInactiveStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("24"))
)

type shellButton struct {
	Key   string
	Label string
	View  View
}

func (m *Model) renderShell() string {
	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)
	topbar := m.renderShellTopbar(width)
	footer := m.renderShellFooter(width)
	status := m.renderShellStatus(width)

	bodyHeight := height - lineCount(topbar) - lineCount(status) - lineCount(footer)
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	body := m.renderWorkspace(width, bodyHeight)

	sections := []string{topbar, body}
	if status != "" {
		sections = append(sections, status)
	}

	sections = append(sections, footer)

	return normalizeBlock(strings.Join(sections, "\n"), width, height)
}

func (m *Model) renderShellTopbar(width int) string {
	project := m.projectDisplayName()

	left := fmt.Sprintf(" %s [alt+p Project] ", project)
	if context := m.caseContextLabel(); context != "" {
		left += context + " "
	}

	right := fmt.Sprintf(" %s ", string(m.SelectedView()))
	if m.guided.Active {
		right = " Guided use cases "
	}

	spacer := strings.Repeat(" ", maxInt(width-lipgloss.Width(left)-lipgloss.Width(right), 1))

	return shellTopbarStyle.Render(fitPlainLine(left+spacer+right, width))
}

func (m *Model) renderWorkspace(width, height int) string {
	if m.guided.Active && !m.textForm.Active {
		return m.renderGuidedWorkspace(width, height)
	}

	if m.harnessSelection.Active {
		return m.renderHarnessWorkspace(width, height)
	}

	overlay := m.renderShellOverlay()
	if overlay != "" {
		return m.renderOverlayWorkspace(width, height, overlay)
	}

	if m.textForm.Active {
		return fitBlock(renderPanel(m.textForm.Title, m.textFormLines(width-4, height-2), width, height, true), width, height)
	}

	if workspace, ok := m.renderProductWorkspace(width, height); ok {
		return workspace
	}

	return m.renderLegacyWorkspace(width, height)
}

func (m *Model) renderProductWorkspace(width, height int) (string, bool) {
	switch m.SelectedView() {
	case ViewProject, ViewDashboard:
		return m.renderProjectWorkspace(width, height), true
	case ViewCases:
		return m.renderCasesWorkspace(width, height), true
	case ViewDumps:
		return m.renderDumpWorkspace(width, height), true
	case ViewBoard, ViewTickets:
		return m.renderTicketWorkspace(width, height), true
	case ViewQuizzes:
		return m.renderQuizWorkspace(width, height), true
	case ViewSpecs:
		return m.renderSpecWorkspace(width, height), true
	case ViewPlans:
		return m.renderPlanWorkspace(width, height), true
	case ViewDash:
		return m.renderDashWorkspace(width, height), true
	default:
		return "", false
	}
}

func (m *Model) renderLegacyWorkspace(width, height int) string {
	switch m.SelectedView() {
	case ViewHelp:
		return m.renderHelpWorkspace(width, height)
	case ViewDocuments:
		return m.renderDocumentWorkspace(width, height, "Documents", m.documents)
	case ViewAuthoring:
		return m.renderAuthoringWorkspace(width, height)
	case ViewAgent, ViewRuntime:
		return m.renderOperationsWorkspace(width, height)
	default:
		return fitBlock(m.renderSelectedView(), width, height)
	}
}
