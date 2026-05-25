package chctui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chc/internal/chc"
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
	case ViewDumps:
		return m.renderDumpWorkspace(width, height), true
	case ViewBoard, ViewTickets:
		return m.renderTicketWorkspace(width, height), true
	case ViewQuizzes:
		return m.renderQuizWorkspace(width, height), true
	case ViewSpecs:
		return m.renderSpecWorkspace(width, height), true
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
	leftWidth, rightWidth := m.splitWidths(width)
	project := m.renderProjectContextLines(maxInt(leftWidth-4, 1))
	artifacts := m.renderProjectArtifactLines()
	location := m.renderProjectLocationLines()

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Project", m.renderProjectContextLines(maxInt(width-4, 1)), width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Location", location, width, height, true), width, height)
	case 2:
		return fitBlock(renderPanel("Artifacts", artifacts, width, height, true), width, height)
	}

	rightTopHeight, rightBottomHeight := m.splitHeights(height)
	right := lipgloss.JoinVertical(
		lipgloss.Left,
		renderPanel("Location", location, rightWidth, rightTopHeight, m.selectedPane == 1),
		renderPanel("Artifacts", artifacts, rightWidth, rightBottomHeight, m.selectedPane == 2),
	)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Project", project, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
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
	if m.selectedPane == 0 && field == m.selectedProjectField {
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
		fmt.Sprintf("Board tickets     %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, chc.WorkspaceNodeTicket)),
		fmt.Sprintf("Dumps             %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, chc.WorkspaceNodeDump)),
		fmt.Sprintf("Quizzes           %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, chc.WorkspaceNodeQuiz)),
		fmt.Sprintf("Specs             %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, chc.WorkspaceNodeSpec)),
		fmt.Sprintf("Plans             %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, chc.WorkspaceNodePlan)),
		fmt.Sprintf("Diagnostics       %d", len(m.workspaceGraph.Diagnostics)),
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
	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Cases", m.renderCaseListLines(maxInt(width-4, 1), height-2), width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Case", m.renderCaseDocumentLines(maxInt(width-4, 1)), width, height, true), width, height)
	case 2:
		return fitBlock(renderPanel("Artifacts", m.renderCaseArtifactLines(), width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	rightEditableHeight, rightArtifactsHeight := m.splitHeights(height)
	right := lipgloss.JoinVertical(
		lipgloss.Left,
		renderPanel("Case", m.renderCaseDocumentLines(maxInt(rightWidth-4, 1)), rightWidth, rightEditableHeight, m.selectedPane == 1),
		renderPanel("Artifacts", m.renderCaseArtifactLines(), rightWidth, rightArtifactsHeight, m.selectedPane == 2),
	)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Cases", m.renderCaseListLines(maxInt(leftWidth-4, 1), height-2), leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		right,
	), width, height)
}

func (m Model) renderCaseDocumentLines(width int) []string {
	lines := m.renderCaseDetailLines(width)

	lines = append(lines, "")
	lines = append(lines, m.renderCaseContextLines(width)...)

	return lines
}

func (m Model) renderCaseListLines(width, limit int) []string {
	filter := strings.TrimSpace(m.caseFilter)

	filterPrefix := "Filter"
	if m.caseFilterEditing {
		filterPrefix = "> Filter"
	}

	lines := []string{filterPrefix + ": " + filter}
	if filter == "" {
		lines[0] = filterPrefix
	}

	cases := m.filteredCases()
	if len(cases) == 0 {
		lines = append(lines, "", "No cases found.", "Press n to create one.")
		lines = append(lines, m.renderUnresolvedCaseReferenceLines(width, limit-len(lines))...)

		return lines
	}

	lines = append(lines, "")

	for _, metadata := range cases {
		prefix := " "

		selected, ok := m.selectedCaseMetadata()
		if ok && metadata.Slug == selected.Slug {
			prefix = ">"
		}

		name := strings.TrimSpace(metadata.Name)
		if name == "" {
			name = "Untitled case"
		}

		lines = append(lines, fitPlainLine(prefix+" "+name, width))
		if len(lines) >= limit {
			break
		}
	}

	lines = append(lines, m.renderUnresolvedCaseReferenceLines(width, limit-len(lines))...)

	return lines
}

func (m Model) renderUnresolvedCaseReferenceLines(width, limit int) []string {
	if limit <= 0 {
		return nil
	}

	references := workspaceGraphUnresolvedCaseReferences(m.workspaceGraph)
	if len(references) == 0 {
		return nil
	}

	lines := []string{}
	for _, line := range []string{"", "Unresolved references"} {
		if len(lines) >= limit {
			return lines
		}

		lines = append(lines, line)
	}

	for _, reference := range references {
		if len(lines) >= limit {
			break
		}

		line := fmt.Sprintf("  %s (%d)", reference.CaseSlug, len(reference.References))
		lines = append(lines, fitPlainLine(line, width))
	}

	return lines
}

func (m Model) renderCaseDetailLines(width int) []string {
	metadata, ok := m.selectedCaseMetadata()
	if !ok {
		return []string{"No case selected.", "", "Press n to create a case."}
	}

	lines := []string{
		m.renderCaseFieldLine(caseFieldName, "Name", metadata.Name, width),
		"",
		m.renderCaseFieldLine(caseFieldDescription, "Description", metadata.Description, width),
	}

	for _, line := range m.renderCaseFieldBody(caseFieldDescription, metadata.Description, width) {
		lines = append(lines, "  "+line)
	}

	if m.caseDirty {
		lines = append(lines, "", shellMutedStyle.Render("Unsaved changes. Press s to save."))
	}

	return lines
}

func (m Model) renderCaseContextLines(width int) []string {
	metadata, ok := m.selectedCaseMetadata()
	if !ok {
		return []string{"No case selected."}
	}

	lines := []string{m.renderCaseFieldLine(caseFieldContext, "Context", metadata.Context, width)}

	for _, line := range m.renderCaseFieldBody(caseFieldContext, metadata.Context, width) {
		lines = append(lines, "  "+line)
	}

	return lines
}

func (m Model) renderCaseFieldLine(field int, label, value string, width int) string {
	prefix := "  "
	if m.selectedPane == 1 && field == m.selectedCaseField {
		prefix = "> "
	}

	if m.caseEdit.Active && field == m.caseEdit.Field && field == caseFieldName {
		return prefix + label + ": " + m.renderCaseInput(maxInt(width-lipgloss.Width(prefix)-lipgloss.Width(label)-2, 1))
	}

	if field != caseFieldName {
		return prefix + label
	}

	value = strings.TrimSpace(value)
	if value == "" {
		value = "Untitled case"
	}

	return prefix + label + ": " + value
}

func (m Model) renderCaseFieldBody(field int, value string, width int) []string {
	if m.caseEdit.Active && field == m.caseEdit.Field {
		if field == caseFieldName {
			return nil
		}

		return []string{m.renderCaseInput(maxInt(width-2, 1))}
	}

	value = strings.TrimSpace(value)
	if value == "" {
		value = "Not set."
	}

	return wrapPlainText(value, maxInt(width-2, 1))
}

func (m Model) renderCaseInput(width int) string {
	input := m.caseEdit.Input
	input.Width = width

	return input.View()
}

func (m Model) renderCaseArtifactLines() []string {
	metadata, ok := m.selectedCaseMetadata()
	if !ok {
		return []string{"No case selected."}
	}

	return []string{
		fmt.Sprintf("Dumps             %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, chc.WorkspaceNodeDump)),
		fmt.Sprintf("Quizzes           %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, chc.WorkspaceNodeQuiz)),
		fmt.Sprintf("Specs             %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, chc.WorkspaceNodeSpec)),
		fmt.Sprintf("Plans             %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, chc.WorkspaceNodePlan)),
		fmt.Sprintf("Board tickets     %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, chc.WorkspaceNodeTicket)),
		fmt.Sprintf("Unresolved refs   %d", workspaceGraphCaseReferenceCount(m.workspaceGraph, metadata.Slug)),
		"",
		"Source",
		"  " + emptyValue(projectMetadataRelativePath(m.projectDir, metadata.SourcePath)),
	}
}

func (m Model) selectedCaseMetadata() (chc.CaseMetadata, bool) {
	if len(m.cases) == 0 || m.selectedCase < 0 || m.selectedCase >= len(m.cases) {
		return chc.CaseMetadata{}, false
	}

	return m.cases[m.selectedCase], true
}

func (m Model) filteredCases() []chc.CaseMetadata {
	filter := strings.ToLower(strings.TrimSpace(m.caseFilter))
	if filter == "" {
		return append([]chc.CaseMetadata(nil), m.cases...)
	}

	if len(m.workspaceIndex.CasesBySlug) > 0 {
		return m.caseMetadataFromSuggestions(m.workspaceIndex.SuggestCases(filter))
	}

	cases := []chc.CaseMetadata{}
	for _, metadata := range m.cases {
		haystack := strings.ToLower(metadata.Name + " " + metadata.Description)
		if strings.Contains(haystack, filter) {
			cases = append(cases, metadata)
		}
	}

	return cases
}

func (m Model) caseMetadataFromSuggestions(
	suggestions []chc.WorkspaceCaseSuggestion,
) []chc.CaseMetadata {
	bySlug := make(map[string]chc.CaseMetadata, len(m.cases))
	for _, metadata := range m.cases {
		bySlug[metadata.Slug] = metadata
	}

	cases := make([]chc.CaseMetadata, 0, len(suggestions))
	for _, suggestion := range suggestions {
		metadata := bySlug[suggestion.Slug]
		metadata.SchemaVersion = chc.CaseMetadataSchemaVersion
		metadata.Slug = suggestion.Slug
		metadata.Name = suggestion.Name
		metadata.Description = suggestion.Description
		metadata.SourcePath = suggestion.Path
		metadata.Exists = true
		cases = append(cases, metadata)
	}

	return cases
}

func (m Model) renderTicketWorkspace(width, height int) string {
	list := m.renderTicketFocusList(height - 2)
	detail := m.renderSelectedTicketDetail()

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Tickets", list, width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Detail", strings.Split(detail, "\n")[1:], width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Tickets", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		renderPanel("Detail", strings.Split(detail, "\n")[1:], rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m Model) renderDocumentWorkspace(width, height int, title string, documents []DocumentView) string {
	list := m.renderDocumentFocusList(documents, height-2)
	detail := m.renderSelectedDocumentDetailFor(documents)

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel(title, list, width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Preview", strings.Split(detail, "\n")[1:], width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel(title, list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		renderPanel("Preview", strings.Split(detail, "\n")[1:], rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m Model) renderDumpWorkspace(width, height int) string {
	list := m.renderDumpFocusList(height - 2)
	detailTitle, detailLines := m.renderDumpDetailPanel(maxInt(width-4, 1), height-2)

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Dumps", list, width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel(detailTitle, detailLines, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	detailTitle, detailLines = m.renderDumpDetailPanel(maxInt(rightWidth-4, 1), height-2)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Dumps", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		renderPanel(detailTitle, detailLines, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m Model) renderDocumentFamilyWorkspace(width, height int, title string, families map[string]bool) string {
	return m.renderDocumentWorkspace(width, height, title, m.documentsByFamily(families))
}

func (m Model) renderQuizWorkspace(width, height int) string {
	list := m.renderQuizFocusList(height - 2)
	detailTitle := m.quizDetailPanelTitle()
	detail := m.renderSelectedQuizDetailLines(maxInt(width-4, 1))

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Quizzes", list, width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel(detailTitle, detail, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	detail = m.renderSelectedQuizDetailLines(maxInt(rightWidth-4, 1))

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Quizzes", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		renderPanel(detailTitle, detail, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m Model) renderSpecWorkspace(width, height int) string {
	list := m.renderSpecFocusList(height - 2)
	detail := m.renderSelectedSpecDetailLines(maxInt(width-4, 1))

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Specs", list, width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Spec", detail, width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)
	detail = m.renderSelectedSpecDetailLines(maxInt(rightWidth-4, 1))

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Specs", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		renderPanel("Spec", detail, rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m Model) quizDetailPanelTitle() string {
	quiz, ok := m.selectedQuizView()
	if ok && m.quizAnswer.Active && quizScopeLabel(m.quizAnswer.Scope) == quizScopeLabel(quiz.Scope) {
		return "Answer"
	}

	return "Quiz"
}

func (m Model) renderAuthoringWorkspace(width, height int) string {
	list := m.renderAuthoringFocusList(height - 2)
	detail := m.renderSelectedAuthoringSessionDetail()

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Authoring", list, width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Session", strings.Split(detail, "\n")[1:], width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Authoring", list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		renderPanel("Session", strings.Split(detail, "\n")[1:], rightWidth, height, m.selectedPane == 1),
	), width, height)
}

func (m Model) renderOperationsWorkspace(width, height int) string {
	agent := strings.Split(m.renderAgent(), "\n")
	runtime := strings.Split(m.renderRuntime()+"\n\n"+m.renderWorkspaceIndexHealth(), "\n")

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Agent", agent[1:], width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Runtime", runtime[1:], width, height, true), width, height)
	}

	leftWidth, rightWidth := m.splitWidths(width)

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel("Agent", agent[1:], leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		renderPanel("Runtime", runtime[1:], rightWidth, height, m.selectedPane == 1),
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
		m.renderDumpAction(),
		m.renderTicketAction(),
		m.renderWorkAction(),
		m.renderQuizAction(),
		m.renderSpecAction(),
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
	return m.renderShellButtons(shellFooterButtons(), width)
}

func shellFooterButtons() []shellButton {
	return []shellButton{
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
}

func (m Model) renderShellButtons(buttons []shellButton, width int) string {
	widths := distributedWidths(width, len(buttons))
	cells := make([]string, 0, len(buttons))

	for index, button := range buttons {
		style := shellButtonInactiveStyle
		if button.View != "" && button.View == m.SelectedView() {
			style = shellButtonStyle
		}

		cellWidth := widths[index]
		style = style.Width(cellWidth)
		cells = append(cells, style.Render(centerShellButtonLines(button.Key, button.Label, cellWidth)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func shellFooterButtonAt(x, y, width, height int) (shellButton, bool) {
	if y < height-2 || y >= height || x < 0 || x >= width {
		return shellButton{}, false
	}

	buttons := shellFooterButtons()
	offset := 0

	for index, cellWidth := range distributedWidths(width, len(buttons)) {
		if x >= offset && x < offset+cellWidth {
			return buttons[index], true
		}

		offset += cellWidth
	}

	return shellButton{}, false
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

	if !filepath.IsAbs(sourcePath) {
		return filepath.ToSlash(sourcePath)
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

func panelVerticalGap(height int) string {
	lines := make([]string, maxInt(height, 1))
	for index := range lines {
		lines[index] = strings.Repeat(" ", panelVerticalGapSize)
	}

	return strings.Join(lines, "\n")
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

func centerShellButtonLines(key, label string, width int) string {
	if width <= 0 {
		return ""
	}

	keyRunes := []rune(key)
	labelRunes := []rune(label)
	keyWidth := minInt(len(keyRunes), width)
	labelWidth := minInt(len(labelRunes), width)

	anchorWidth := maxInt(keyWidth, labelWidth)
	anchorLeft := maxInt((width-anchorWidth)/2, 0)
	keyLeft := anchorLeft + maxInt((anchorWidth-keyWidth)/2, 0)
	labelLeft := anchorLeft + maxInt((anchorWidth-labelWidth)/2, 0)

	return placePlainLine(key, width, keyLeft) + "\n" +
		placePlainLine(label, width, labelLeft)
}

func placePlainLine(line string, width, left int) string {
	if width <= 0 {
		return ""
	}

	fitted := truncatePlainLine(line, width)
	lineWidth := lipgloss.Width(fitted)

	if lineWidth >= width {
		return fitted
	}

	left = minInt(maxInt(left, 0), width-lineWidth)
	right := width - lineWidth - left

	return strings.Repeat(" ", left) + fitted + strings.Repeat(" ", right)
}

func truncatePlainLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(line)
	if len(runes) <= width {
		return line
	}

	if width == 1 {
		return "…"
	}

	return string(runes[:width-1]) + "…"
}

func (m Model) splitWidths(width int) (int, int) {
	available := maxInt(width-panelVerticalGapSize, 1)
	left := available * m.verticalSplitPercent / 100
	right := available - left

	return left, right
}

func (m Model) splitHeights(height int) (int, int) {
	available := maxInt(height, 1)
	top := maxInt(available*m.horizontalSplitPercent/100, 1)
	bottom := maxInt(available-top, 1)

	return top, bottom
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

func percentFromPosition(position, total int) int {
	if total <= 0 {
		return defaultPanelSplitPercent
	}

	percent := position * 100 / total

	return minInt(maxInt(percent, minPanelSplitPercent), maxPanelSplitPercent)
}

func nearInt(left, right, distance int) bool {
	delta := left - right
	if delta < 0 {
		delta = -delta
	}

	return delta <= distance
}

func minInt(left, right int) int {
	if left < right {
		return left
	}

	return right
}
