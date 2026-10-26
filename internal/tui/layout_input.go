package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type layoutDrag int

const (
	layoutDragNone layoutDrag = iota
	layoutDragVertical
	layoutDragHorizontal
)

const (
	defaultPanelSplitPercent = 50
	minPanelSplitPercent     = 25
	maxPanelSplitPercent     = 75
	boardAttentionColumn     = "attention"
)

func (m *Model) updateMouse(msg tea.MouseMsg) tea.Cmd {
	if m.guided.Active {
		return nil
	}

	if m.harnessSelection.Active {
		return nil
	}

	event := tea.MouseEvent(msg)

	if event.Action == tea.MouseActionRelease {
		m.layoutDrag = layoutDragNone

		return nil
	}

	if event.Action == tea.MouseActionMotion {
		return m.updateMouseMotion(event)
	}

	if event.Action != tea.MouseActionPress || event.Button != tea.MouseButtonLeft {
		return nil
	}

	return m.updateMousePress(event)
}

func (m *Model) updateMousePress(event tea.MouseEvent) tea.Cmd {
	if m.workBusy {
		m.statusMessage = "Work action already running."

		return nil
	}

	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)

	if drag, ok := m.layoutDragAt(event.X, event.Y, width, height); ok {
		m.layoutDrag = drag
		m.applyLayoutDrag(event)

		return nil
	}

	button, ok := shellFooterButtonAt(event.X, event.Y, width, height)
	if !ok {
		if m.SelectedView() == ViewDumps {
			m.selectDumpPaneAt(event.X, width)
		}

		return nil
	}

	if button.View != "" {
		m.selectView(button.View)

		return nil
	}

	if button.Key == "F10" {
		return tea.Quit
	}

	return nil
}

func (m *Model) selectDumpPaneAt(x, width int) {
	if m.maximizedPane >= 0 {
		m.selectedPane = m.maximizedPane

		return
	}

	leftWidth, _ := m.splitWidths(width)
	if x <= leftWidth {
		m.selectedPane = 0

		return
	}

	m.selectedPane = 1
}

func (m *Model) updateMouseMotion(event tea.MouseEvent) tea.Cmd {
	if m.layoutDrag != layoutDragNone {
		m.applyLayoutDrag(event)

		return nil
	}

	if event.Button != tea.MouseButtonLeft {
		return nil
	}

	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)

	if drag, ok := m.layoutDragAt(event.X, event.Y, width, height); ok {
		m.layoutDrag = drag
		m.applyLayoutDrag(event)
	}

	return nil
}

func (m *Model) toggleMaximizedPane() {
	if m.maximizedPane == m.selectedPane {
		m.maximizedPane = -1

		return
	}

	m.maximizedPane = m.selectedPane
}

func (m *Model) layoutDragAt(x, y, width, height int) (layoutDrag, bool) {
	if m.maximizedPane >= 0 {
		return layoutDragNone, false
	}

	bodyTop, bodyHeight := m.shellBodyGeometry(width, height)
	if y < bodyTop || y >= bodyTop+bodyHeight {
		return layoutDragNone, false
	}

	leftWidth, _ := m.splitWidths(width)
	if nearInt(x, leftWidth, 4) {
		return layoutDragVertical, true
	}

	if !m.hasStackedRightPanels() || x < leftWidth {
		return layoutDragNone, false
	}

	topHeight, _ := m.splitHeights(bodyHeight)
	if nearInt(y-bodyTop, topHeight, 1) {
		return layoutDragHorizontal, true
	}

	return layoutDragNone, false
}

func (m *Model) applyLayoutDrag(event tea.MouseEvent) {
	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)
	bodyTop, bodyHeight := m.shellBodyGeometry(width, height)

	switch m.layoutDrag {
	case layoutDragVertical:
		m.verticalSplitPercent = percentFromPosition(event.X, width)
	case layoutDragHorizontal:
		m.horizontalSplitPercent = percentFromPosition(event.Y-bodyTop, bodyHeight)
	}
}

func (m *Model) shellBodyGeometry(width, height int) (int, int) {
	topbar := m.renderShellTopbar(width)
	footer := m.renderShellFooter(width)
	status := m.renderShellStatus(width)

	bodyTop := lineCount(topbar)
	bodyHeight := height - lineCount(topbar) - lineCount(status) - lineCount(footer)

	if bodyHeight < 1 {
		bodyHeight = 1
	}

	return bodyTop, bodyHeight
}

func (m *Model) hasStackedRightPanels() bool {
	switch m.SelectedView() {
	case ViewProject, ViewDashboard, ViewCases:
		return true
	default:
		return false
	}
}

func (m *Model) handlePaneNavigationKey(key string) bool {
	switch key {
	case "shift+tab":
		m.selectPreviousPane()
	case "tab":
		m.selectNextPane()
	default:
		return false
	}

	return true
}

func (m *Model) selectNextPane() {
	count := m.paneCount()
	if count <= 1 {
		m.selectedPane = 0

		return
	}

	m.selectedPane++
	if m.selectedPane >= count {
		m.selectedPane = 0
	}
}

func (m *Model) selectPreviousPane() {
	count := m.paneCount()
	if count <= 1 {
		m.selectedPane = 0

		return
	}

	if m.selectedPane == 0 {
		m.selectedPane = count - 1

		return
	}

	m.selectedPane--
}

func (m *Model) selectPrimaryPane() {
	m.selectedPane = 0
	m.detailScroll = 0
}

func (m *Model) paneCount() int {
	switch m.SelectedView() {
	case ViewProject, ViewCases, ViewDash:
		return 3
	case ViewBoard, ViewTickets:
		return 2
	case ViewDumps, ViewQuizzes, ViewSpecs, ViewPlans, ViewDocuments:
		return 2
	case ViewAuthoring:
		return 2
	case ViewAgent, ViewRuntime:
		return 2
	default:
		return 1
	}
}
