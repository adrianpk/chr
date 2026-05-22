package chctui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const minShellHeight = 12

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

func (m Model) renderShell() string {
	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)
	topbar := m.renderShellTopbar(width)
	tabs := m.renderShellTabs(width)
	footer := m.renderShellFooter(width)
	status := m.renderShellStatus(width)

	bodyHeight := height - lineCount(topbar) - lineCount(tabs) - lineCount(status) - lineCount(footer)
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	body := m.renderWorkspace(width, bodyHeight)

	sections := []string{topbar, tabs, body}
	if status != "" {
		sections = append(sections, status)
	}

	sections = append(sections, footer)

	return normalizeBlock(strings.Join(sections, "\n"), width, height)
}

func (m Model) renderShellTopbar(width int) string {
	project := m.projectID
	if project == "" {
		project = "no project"
	}

	dir := m.projectDir
	if dir == "" {
		dir = "."
	}

	dir = filepath.Base(dir)
	if dir == "." || dir == string(filepath.Separator) {
		dir = m.projectDir
	}

	left := fmt.Sprintf(" ChrysoCoord  %s ", project)
	right := fmt.Sprintf(" %s  %s ", string(m.SelectedView()), dir)
	spacer := strings.Repeat(" ", maxInt(width-lipgloss.Width(left)-lipgloss.Width(right), 1))

	return shellTopbarStyle.Render(fitPlainLine(left+spacer+right, width))
}

func (m Model) renderShellTabs(width int) string {
	var output strings.Builder

	for index, view := range m.views {
		label := fmt.Sprintf("%d %s", index+1, view)
		if view == m.SelectedView() {
			label = "[" + label + "]"
		}

		if output.Len() > 0 {
			output.WriteString("  ")
		}

		output.WriteString(label)
	}

	return fitPlainLine(output.String(), width)
}

func (m Model) renderWorkspace(width, height int) string {
	overlay := m.renderShellOverlay()
	if overlay != "" {
		return m.renderOverlayWorkspace(width, height, overlay)
	}

	switch m.SelectedView() {
	case ViewDashboard:
		return m.renderDashboardWorkspace(width, height)
	case ViewBoard, ViewTickets:
		return m.renderTicketWorkspace(width, height)
	case ViewDocuments:
		return m.renderDocumentWorkspace(width, height)
	case ViewAuthoring:
		return m.renderAuthoringWorkspace(width, height)
	case ViewAgent, ViewRuntime:
		return m.renderOperationsWorkspace(width, height)
	default:
		return fitBlock(m.renderSelectedView(), width, height)
	}
}

func (m Model) renderDashboardWorkspace(width, height int) string {
	leftWidth, rightWidth := splitWidths(width)
	topHeight := maxInt(height/2, 1)
	bottomHeight := maxInt(height-topHeight, 1)
	project := []string{
		"Project",
		"ID: " + emptyValue(m.projectID),
		"Dir: " + emptyValue(m.projectDir),
	}
	tickets := []string{
		"Tickets",
		fmt.Sprintf("Total: %d", len(m.tickets)),
		fmt.Sprintf("Filtered: %d", len(m.filteredTickets())),
	}
	agent := []string{
		"Agent",
		"Daemon: " + emptyValue(m.observability.DaemonStatus),
		"Latest loop: " + emptyValue(m.observability.LatestLoop.ID),
	}
	runtime := []string{
		"Runtime",
		"Image: " + emptyValue(m.observability.RuntimeImage),
		"Validation: " + emptyValue(m.observability.RuntimeValidationCommand),
	}
	top := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Project", project[1:], leftWidth, topHeight, true),
		renderPanel("Tickets", tickets[1:], rightWidth, topHeight, false),
	)
	bottom := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Agent", agent[1:], leftWidth, bottomHeight, false),
		renderPanel("Runtime", runtime[1:], rightWidth, bottomHeight, false),
	)

	return fitBlock(top+"\n"+bottom, width, height)
}

func (m Model) renderTicketWorkspace(width, height int) string {
	leftWidth, rightWidth := splitWidths(width)
	list := m.renderTicketFocusList(height - 2)
	detail := m.renderSelectedTicketDetail()

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Tickets", list, leftWidth, height, true),
		renderPanel("Detail", strings.Split(detail, "\n")[1:], rightWidth, height, false),
	), width, height)
}

func (m Model) renderDocumentWorkspace(width, height int) string {
	leftWidth, rightWidth := splitWidths(width)
	list := m.renderDocumentFocusList(height - 2)
	detail := m.renderSelectedDocumentDetail()

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Documents", list, leftWidth, height, true),
		renderPanel("Preview", strings.Split(detail, "\n")[1:], rightWidth, height, false),
	), width, height)
}

func (m Model) renderAuthoringWorkspace(width, height int) string {
	leftWidth, rightWidth := splitWidths(width)
	list := m.renderAuthoringFocusList(height - 2)
	detail := m.renderSelectedAuthoringSessionDetail()

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Authoring", list, leftWidth, height, true),
		renderPanel("Session", strings.Split(detail, "\n")[1:], rightWidth, height, false),
	), width, height)
}

func (m Model) renderOperationsWorkspace(width, height int) string {
	leftWidth, rightWidth := splitWidths(width)
	agent := strings.Split(m.renderAgent(), "\n")
	runtime := strings.Split(m.renderRuntime(), "\n")

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Agent", agent[1:], leftWidth, height, m.SelectedView() == ViewAgent),
		renderPanel("Runtime", runtime[1:], rightWidth, height, m.SelectedView() == ViewRuntime),
	), width, height)
}

func (m Model) renderOverlayWorkspace(width, height int, overlay string) string {
	content := []string{
		shellTitleStyle.Render("Active interaction"),
		"",
	}
	content = append(content, strings.Split(overlay, "\n")...)

	return renderPanel("Action", content, width, height, true)
}

func (m Model) renderTicketFocusList(limit int) []string {
	tickets := m.filteredTickets()
	if len(tickets) == 0 {
		return []string{"No tickets match."}
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

func (m Model) renderDocumentFocusList(limit int) []string {
	if len(m.documents) == 0 {
		return []string{"No documents found."}
	}

	lines := []string{shellMutedStyle.Render("Type       Status      Title")}

	for _, document := range m.documents {
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

func (m Model) renderAuthoringFocusList(limit int) []string {
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

func (m Model) renderShellOverlay() string {
	var overlays []string

	for _, overlay := range []string{
		m.renderArtifactAction(),
		m.renderTicketAction(),
		m.renderWorkAction(),
		m.renderCommandPalette(),
	} {
		if strings.TrimSpace(overlay) != "" {
			overlays = append(overlays, strings.TrimSpace(overlay))
		}
	}

	return strings.Join(overlays, "\n\n")
}

func (m Model) renderShellStatus(width int) string {
	if m.statusMessage == "" {
		return ""
	}

	return shellStatusStyle.Render(fitPlainLine(" "+m.statusMessage, width))
}

func (m Model) renderShellFooter(width int) string {
	buttons := []shellButton{
		{Key: "F1", Label: "Project", View: ViewDashboard},
		{Key: "F2", Label: "Dumps"},
		{Key: "F3", Label: "Quizzes"},
		{Key: "F4", Label: "Drafts"},
		{Key: "F5", Label: "Specs", View: ViewDocuments},
		{Key: "F6", Label: "Plans", View: ViewReports},
		{Key: "F7", Label: "Board", View: ViewBoard},
		{Key: "F8", Label: "Dash", View: ViewAgent},
		{Key: "F9", Label: "Help", View: ViewHelp},
		{Key: "F10", Label: "Quit"},
	}

	return m.renderShellButtons(buttons, width)
}

func (m Model) renderShellButtons(buttons []shellButton, width int) string {
	widths := distributedWidths(width, len(buttons))
	keyLine := make([]string, 0, len(buttons))
	labelLine := make([]string, 0, len(buttons))

	for index, button := range buttons {
		style := shellButtonInactiveStyle
		if button.View != "" && button.View == m.SelectedView() {
			style = shellButtonStyle
		}

		cellWidth := widths[index]
		style = style.Width(cellWidth).Align(lipgloss.Center)
		keyLine = append(keyLine, style.Render(button.Key))
		labelLine = append(labelLine, style.Render(button.Label))
	}

	return strings.Join(keyLine, "") + "\n" + strings.Join(labelLine, "")
}

func renderPanel(title string, lines []string, width, height int, focused bool) string {
	style := shellPanelStyle
	if focused {
		style = shellFocusedPanelStyle
	}

	contentHeight := maxInt(height-2, 1)

	content := append([]string{shellTitleStyle.Render(title)}, lines...)

	return style.Width(maxInt(width-2, 1)).Height(contentHeight).Render(fitBlock(strings.Join(content, "\n"), maxInt(width-4, 1), contentHeight))
}

func normalizeBlock(content string, width, height int) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) > height {
		lines = lines[:height]
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	for index, line := range lines {
		lines[index] = fitStyledLine(line, width)
	}

	return strings.Join(lines, "\n")
}

func fitBlock(content string, width, height int) string {
	return normalizeBlock(content, width, height)
}

func fitStyledLine(line string, width int) string {
	visible := lipgloss.Width(line)
	if visible >= width {
		return line
	}

	return line + strings.Repeat(" ", width-visible)
}

func fitPlainLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(line)
	if len(runes) > width {
		if width == 1 {
			return "…"
		}

		return string(runes[:width-1]) + "…"
	}

	return line + strings.Repeat(" ", width-len(runes))
}

func splitWidths(width int) (int, int) {
	left := width * 55 / 100
	right := width - left

	return left, right
}

func distributedWidths(width, count int) []int {
	if count <= 0 {
		return nil
	}

	base := width / count
	remainder := width % count
	widths := make([]int, count)

	for index := range widths {
		widths[index] = base
		if index < remainder {
			widths[index]++
		}
	}

	return widths
}

func lineCount(value string) int {
	if value == "" {
		return 0
	}

	return strings.Count(value, "\n") + 1
}

func emptyValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}

	return value
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}

	return right
}
