package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SelectedView returns the active TUI view.
func (m *Model) SelectedView() View {
	if m.auxiliaryView != "" {
		return m.auxiliaryView
	}

	return m.views[m.selected]
}

func (m *Model) updateNavigationKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if view, ok := functionKeyView(msg); ok {
		m.selectView(view)

		return nil
	}

	if cmd, ok := m.handleNavigationCommandKey(key); ok {
		return cmd
	}

	if m.handleNavigationPageKey(msg.Type) {
		return nil
	}

	return m.handleNavigationActionKey(key)
}

func (m *Model) handleNavigationPageKey(keyType tea.KeyType) bool {
	switch keyType {
	case tea.KeyPgUp:
		m.scrollDetailPage(-1)
	case tea.KeyPgDown:
		m.scrollDetailPage(1)
	default:
		return false
	}

	return true
}

func (m *Model) handleNavigationActionKey(key string) tea.Cmd {
	if command, handled := m.handleDashKey(key); handled {
		return command
	}

	if m.handleDetailScrollKey(key) {
		return nil
	}

	switch key {
	case "e":
		return m.openSelectedPaneSource()
	case "up", "k":
		m.selectPreviousRecord()
	case "down", "j":
		m.selectNextRecord()
	case "esc":
		m.selectPrimaryPane()
	case ":":
		m.openCommandPalette()
	case "/":
		m.filterEntryMode = true
	}

	return nil
}

func (m *Model) handleDetailScrollKey(key string) bool {
	switch key {
	case "ctrl+u":
		m.scrollDetailPage(-1)
	case "ctrl+d":
		m.scrollDetailPage(1)
	default:
		return false
	}

	return true
}

func (m *Model) handleDashActionKey(key string) bool {
	if m.SelectedView() != ViewDash {
		return false
	}

	switch key {
	case "enter":
		m.planWorkAction(WorkActionStartTicket)
	case "a":
		m.planWorkAction(WorkActionRunAllReady)
	case "R":
		m.planWorkAction(WorkActionRunAllReady)
	case "x":
		m.handleDashRepairReadinessKey()
	case "s":
		m.planWorkAction(WorkActionRequestStop)
	default:
		return false
	}

	return true
}

func (m *Model) handleDashRepairReadinessKey() {
	if !m.workReadiness.safeRepairAvailable() {
		m.statusMessage = "No safe readiness repair is available."

		return
	}

	m.planWorkAction(WorkActionRepairReadiness)
}

func (m *Model) handleNavigationCommandKey(key string) (tea.Cmd, bool) {
	switch {
	case key == "esc" && m.SelectedView() == ViewHelp:
		view, pane := m.helpReturnView, m.helpReturnPane
		m.selectView(view)

		m.selectedPane = pane

		return nil, true
	case isQuitKey(key):
		return tea.Quit, true
	case key == "alt+m":
		m.toggleMaximizedPane()

		return nil, true
	case key == "ctrl+e":
		return m.openSelectedSource(), true
	case m.handlePaneNavigationKey(key):
		return nil, true
	case m.handleBoardColumnNavigationKey(key):
		return nil, true
	default:
		return nil, false
	}
}

func (m *Model) handleBoardColumnNavigationKey(key string) bool {
	switch key {
	case "alt+a":
		m.toggleBoardAttention()
	case "left", "h":
		m.selectPreviousBoardColumn()
	case "right", "l":
		m.selectNextBoardColumn()
	default:
		return false
	}

	return true
}

func isQuitKey(key string) bool {
	return key == "ctrl+c" || key == "q" || key == "f10"
}

func functionKeyView(msg tea.KeyMsg) (View, bool) {
	for _, button := range shellFooterButtons() {
		if button.View != "" && strings.EqualFold(button.Key, msg.String()) {
			return button.View, true
		}
	}

	return "", false
}

func (m *Model) selectPreviousRecord() {
	if m.selectedPane != 0 {
		m.scrollDetail(-1)

		return
	}

	selectedPath := m.selectedSourcePath()
	switch m.SelectedView() {
	case ViewHelp:
		m.scrollDetail(-1)
	case ViewDash:
		m.moveDashSelection(-1)
	case ViewBoard, ViewTickets:
		m.selectPreviousTicket()
	default:
		m.selectPreviousArtifact()
	}

	m.resetDetailScrollIfSourceChanged(selectedPath)
}

func (m *Model) selectPreviousArtifact() {
	switch m.SelectedView() {
	case ViewDumps:
		m.selectPreviousDump()
	case ViewQuizzes:
		m.selectPreviousQuiz()
	case ViewSpecs:
		m.selectPreviousSpec()
	case ViewPlans:
		m.selectPreviousPlan()
	case ViewDocuments:
		m.selectPreviousDocument()
	case ViewAuthoring:
		m.selectPreviousAuthoringSession()
	}
}

func (m *Model) selectNextRecord() {
	if m.selectedPane != 0 {
		m.scrollDetail(1)

		return
	}

	selectedPath := m.selectedSourcePath()
	switch m.SelectedView() {
	case ViewHelp:
		m.scrollDetail(1)
	case ViewDash:
		m.moveDashSelection(1)
	case ViewBoard, ViewTickets:
		m.selectNextTicket()
	default:
		m.selectNextArtifact()
	}

	m.resetDetailScrollIfSourceChanged(selectedPath)
}

func (m *Model) selectNextArtifact() {
	switch m.SelectedView() {
	case ViewDumps:
		m.selectNextDump()
	case ViewQuizzes:
		m.selectNextQuiz()
	case ViewSpecs:
		m.selectNextSpec()
	case ViewPlans:
		m.selectNextPlan()
	case ViewDocuments:
		m.selectNextDocument()
	case ViewAuthoring:
		m.selectNextAuthoringSession()
	}
}

func lastRune(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)

	return string(runes[len(runes)-1])
}

func (m *Model) openSelectedPaneSource() tea.Cmd {
	if m.selectedPane == 1 {
		return m.openSelectedSource()
	}

	return nil
}
