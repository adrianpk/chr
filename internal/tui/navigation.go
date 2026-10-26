package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SelectedView returns the active TUI view.
func (m *Model) SelectedView() View {
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
	if m.handleDashActionKey(key) {
		return nil
	}

	if m.handleDetailScrollKey(key) {
		return nil
	}

	switch key {
	case "e":
		if m.selectedPane == 1 {
			return m.openSelectedSource()
		}
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
	case "p":
		m.cycleAgentWorkPriorityFilter()
	case "o":
		m.cycleAgentWorkOrder()
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
	viewsByType := map[tea.KeyType]View{
		tea.KeyF1: ViewProject,
		tea.KeyF2: ViewCases,
		tea.KeyF3: ViewDumps,
		tea.KeyF4: ViewQuizzes,
		tea.KeyF5: ViewSpecs,
		tea.KeyF6: ViewPlans,
		tea.KeyF7: ViewBoard,
		tea.KeyF8: ViewDash,
		tea.KeyF9: ViewHelp,
	}

	if view, ok := viewsByType[msg.Type]; ok {
		return view, true
	}

	viewsByString := map[string]View{
		"f1": ViewProject,
		"f2": ViewCases,
		"f3": ViewDumps,
		"f4": ViewQuizzes,
		"f5": ViewSpecs,
		"f6": ViewPlans,
		"f7": ViewBoard,
		"f8": ViewDash,
		"f9": ViewHelp,
	}

	view, ok := viewsByString[msg.String()]

	return view, ok
}

func (m *Model) selectPreviousRecord() {
	if m.selectedPane != 0 {
		m.scrollDetail(-1)

		return
	}

	selectedPath := m.selectedSourcePath()
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		m.selectPreviousTicket()
	case ViewDumps:
		m.selectPreviousDump()
	case ViewQuizzes:
		m.selectPreviousQuiz()
	case ViewSpecs, ViewPlans, ViewDocuments:
		if m.SelectedView() == ViewSpecs {
			m.selectPreviousSpec()
		} else if m.SelectedView() == ViewPlans {
			m.selectPreviousPlan()
		} else {
			m.selectPreviousDocument()
		}
	case ViewAuthoring:
		m.selectPreviousAuthoringSession()
	}

	m.resetDetailScrollIfSourceChanged(selectedPath)
}

func (m *Model) selectNextRecord() {
	if m.selectedPane != 0 {
		m.scrollDetail(1)

		return
	}

	selectedPath := m.selectedSourcePath()
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		m.selectNextTicket()
	case ViewDumps:
		m.selectNextDump()
	case ViewQuizzes:
		m.selectNextQuiz()
	case ViewSpecs, ViewPlans, ViewDocuments:
		if m.SelectedView() == ViewSpecs {
			m.selectNextSpec()
		} else if m.SelectedView() == ViewPlans {
			m.selectNextPlan()
		} else {
			m.selectNextDocument()
		}
	case ViewAuthoring:
		m.selectNextAuthoringSession()
	}

	m.resetDetailScrollIfSourceChanged(selectedPath)
}

func lastRune(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)

	return string(runes[len(runes)-1])
}
