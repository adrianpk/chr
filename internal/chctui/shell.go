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

func (m Model) renderShellTopbar(width int) string {
	project := m.projectDisplayName()
	left := fmt.Sprintf(" %s ", project)
	right := fmt.Sprintf(" %s ", string(m.SelectedView()))
	spacer := strings.Repeat(" ", maxInt(width-lipgloss.Width(left)-lipgloss.Width(right), 1))

	return shellTopbarStyle.Render(fitPlainLine(left+spacer+right, width))
}

func (m Model) renderWorkspace(width, height int) string {
	overlay := m.renderShellOverlay()
	if overlay != "" {
		return m.renderOverlayWorkspace(width, height, overlay)
	}

	if workspace, ok := m.renderProductWorkspace(width, height); ok {
		return workspace
	}

	return m.renderLegacyWorkspace(width, height)
}

func (m Model) renderProductWorkspace(width, height int) (string, bool) {
	switch m.SelectedView() {
	case ViewProject, ViewDashboard:
		return m.renderProjectWorkspace(width, height), true
	case ViewCases:
		return m.renderCasesWorkspace(width, height), true
	case ViewBoard, ViewTickets:
		return m.renderTicketWorkspace(width, height), true
	case ViewDumps:
		return m.renderDocumentFamilyWorkspace(width, height, "Dumps", map[string]bool{"dump": true}), true
	case ViewQuizzes:
		return m.renderQuizWorkspace(width, height), true
	case ViewSpecs:
		return m.renderDocumentFamilyWorkspace(width, height, "Specs", map[string]bool{"spec": true}), true
	case ViewPlans:
		return m.renderDocumentFamilyWorkspace(width, height, "Plans", map[string]bool{"plan": true, "tracker": true}), true
	case ViewDash:
		return m.renderOperationsWorkspace(width, height), true
	default:
		return "", false
	}
}

func (m Model) renderLegacyWorkspace(width, height int) string {
	switch m.SelectedView() {
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

func (m Model) renderProjectWorkspace(width, height int) string {
	leftWidth := width * 62 / 100
	rightWidth := width - leftWidth
	rightTopHeight := maxInt(height/2, 1)
	rightBottomHeight := maxInt(height-rightTopHeight, 1)
	project := m.renderProjectContextLines(maxInt(leftWidth-4, 1))
	artifacts := m.renderProjectArtifactLines()
	location := m.renderProjectLocationLines()
	right := lipgloss.JoinVertical(
		lipgloss.Left,
		renderPanel("Artifacts", artifacts, rightWidth, rightTopHeight, false),
		renderPanel("Location", location, rightWidth, rightBottomHeight, false),
	)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("", project, leftWidth, height, true),
		right,
	), width, height)
}

func (m Model) renderProjectContextLines(width int) []string {
	name := m.projectDisplayName()

	lines := []string{
		m.renderProjectFieldLine(projectFieldName, "Name", name, width),
		"",
		m.renderProjectFieldLine(projectFieldDescription, "Description", m.projectMetadata.Description, width),
	}

	for _, line := range m.renderProjectFieldBody(projectFieldDescription, m.projectMetadata.Description, width) {
		lines = append(lines, "  "+line)
	}

	lines = append(lines, "", m.renderProjectFieldLine(projectFieldContext, "Context", m.projectMetadata.Context, width))

	for _, line := range m.renderProjectFieldBody(projectFieldContext, m.projectMetadata.Context, width) {
		lines = append(lines, "  "+line)
	}

	if m.projectDirty {
		lines = append(lines, "", shellMutedStyle.Render("Unsaved changes. Press s to save."))
	}

	return lines
}

func (m Model) renderProjectFieldLine(field int, label, value string, width int) string {
	prefix := "  "
	if field == m.selectedProjectField {
		prefix = "> "
	}

	if m.projectEdit.Active && field == m.projectEdit.Field && field == projectFieldName {
		return prefix + label + ": " + m.renderProjectInput(maxInt(width-lipgloss.Width(prefix)-lipgloss.Width(label)-2, 1))
	}

	if field != projectFieldName {
		return prefix + label
	}

	return prefix + label + ": " + strings.TrimSpace(value)
}

func (m Model) renderProjectFieldBody(field int, value string, width int) []string {
	if m.projectEdit.Active && field == m.projectEdit.Field {
		if field == projectFieldName {
			return nil
		}

		return []string{m.renderProjectInput(maxInt(width-2, 1))}
	}

	value = strings.TrimSpace(value)
	if value == "" {
		value = "Not set."
	}

	return wrapPlainText(value, maxInt(width-2, 1))
}

func (m Model) renderProjectInput(width int) string {
	input := m.projectEdit.Input
	input.Width = width

	return input.View()
}

func (m Model) renderProjectArtifactLines() []string {
	return []string{
		fmt.Sprintf("Board tickets     %d", len(m.tickets)),
		fmt.Sprintf("Dumps             %d", len(m.documentsByFamily(map[string]bool{"dump": true}))),
		fmt.Sprintf("Quizzes           %d", len(m.documentsByFamily(map[string]bool{"quiz": true}))),
		fmt.Sprintf("Specs             %d", len(m.documentsByFamily(map[string]bool{"spec": true}))),
		fmt.Sprintf("Plans             %d", len(m.documentsByFamily(map[string]bool{"plan": true}))),
	}
}

func (m Model) renderProjectLocationLines() []string {
	return []string{
		"Path",
		"  " + emptyValue(m.projectDir),
		"",
		"File",
		"  " + emptyValue(projectMetadataRelativePath(m.projectDir, m.projectMetadata.SourcePath)),
	}
}

func (m Model) renderCasesWorkspace(width, height int) string {
	leftWidth, rightWidth := splitWidths(width)
	list := []string{
		"No use cases found.",
		"",
		"Use cases are intentionally small:",
		"name",
		"description",
	}
	detail := []string{
		"Scope",
		"Single-user workspace.",
		"No ownership fields.",
		"Artifacts can later attach to a case or stay project-wide.",
	}

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Cases", list, leftWidth, height, true),
		renderPanel("Case Detail", detail, rightWidth, height, false),
	), width, height)
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

func (m Model) renderDocumentWorkspace(width, height int, title string, documents []DocumentView) string {
	leftWidth, rightWidth := splitWidths(width)
	list := m.renderDocumentFocusList(documents, height-2)
	detail := m.renderSelectedDocumentDetailFor(documents)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel(title, list, leftWidth, height, true),
		renderPanel("Preview", strings.Split(detail, "\n")[1:], rightWidth, height, false),
	), width, height)
}

func (m Model) renderDocumentFamilyWorkspace(width, height int, title string, families map[string]bool) string {
	return m.renderDocumentWorkspace(width, height, title, m.documentsByFamily(families))
}

func (m Model) renderQuizWorkspace(width, height int) string {
	leftWidth, rightWidth := splitWidths(width)
	quizzes := m.documentsByFamily(map[string]bool{"quiz": true})
	list := m.renderDocumentFocusList(quizzes, height-2)
	interaction := []string{
		"Question/answer workspace",
		"Agent prompt: not active.",
		"User answer: not active.",
		"",
		"Quizzes can be project-wide or case-scoped.",
	}

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Quizzes", list, leftWidth, height, true),
		renderPanel("Interaction", interaction, rightWidth, height, false),
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
		renderPanel("Agent", agent[1:], leftWidth, height, m.SelectedView() == ViewDash || m.SelectedView() == ViewAgent),
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

func (m Model) renderDocumentFocusList(documents []DocumentView, limit int) []string {
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
		{Key: "F1", Label: "Project", View: ViewProject},
		{Key: "F2", Label: "Cases", View: ViewCases},
		{Key: "F3", Label: "Dumps", View: ViewDumps},
		{Key: "F4", Label: "Quizzes", View: ViewQuizzes},
		{Key: "F5", Label: "Specs", View: ViewSpecs},
		{Key: "F6", Label: "Plans", View: ViewPlans},
		{Key: "F7", Label: "Board", View: ViewBoard},
		{Key: "F8", Label: "Dash", View: ViewDash},
		{Key: "F9", Label: "Help", View: ViewHelp},
		{Key: "F10", Label: "Exit"},
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

func projectDisplayName(projectDir string) string {
	if strings.TrimSpace(projectDir) == "" {
		return "."
	}

	name := filepath.Base(projectDir)
	if name == "." || name == string(filepath.Separator) {
		return projectDir
	}

	return name
}

func (m Model) projectDisplayName() string {
	if strings.TrimSpace(m.projectMetadata.Name) != "" {
		return m.projectMetadata.Name
	}

	return projectDisplayName(m.projectDir)
}

func projectMetadataRelativePath(projectDir, sourcePath string) string {
	if strings.TrimSpace(sourcePath) == "" {
		return ""
	}

	relativePath, err := filepath.Rel(projectDir, sourcePath)
	if err != nil {
		return sourcePath
	}

	return filepath.ToSlash(relativePath)
}

func wrapPlainText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{}
	line := ""

	for _, word := range words {
		if line == "" {
			line = word

			continue
		}

		next := line + " " + word
		if lipgloss.Width(next) <= width {
			line = next

			continue
		}

		lines = append(lines, line)
		line = word
	}

	if line != "" {
		lines = append(lines, line)
	}

	return lines
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
