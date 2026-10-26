package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chr/internal/core"
)

func (m *Model) renderCasesWorkspace(width, height int) string {
	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel("Cases", m.renderCaseListLines(maxInt(width-4, 1), height-2), width, height, true), width, height)
	case 1:
		return fitBlock(renderPanel("Case", m.renderCaseDocumentLines(maxInt(width-4, 1)), width, height, true), width, height)
	case 2:
		return fitBlock(renderPanel("Artifacts", m.renderCaseArtifactLines(), width, height, true), width, height)
	}

	horizontal := m.splitWidths(width)
	leftWidth, rightWidth := horizontal.Left, horizontal.Right
	vertical := m.splitHeights(height)
	rightEditableHeight, rightArtifactsHeight := vertical.Top, vertical.Bottom
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

func (m *Model) renderCaseDocumentLines(width int) []string {
	lines := m.renderCaseDetailLines(width)

	lines = append(lines, "")
	lines = append(lines, m.renderCaseContextLines(width)...)

	return lines
}

func (m *Model) renderCaseListLines(width, limit int) []string {
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

func (m *Model) renderUnresolvedCaseReferenceLines(width, limit int) []string {
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

func (m *Model) renderCaseDetailLines(width int) []string {
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

func (m *Model) renderCaseContextLines(width int) []string {
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

func (m *Model) renderCaseFieldLine(field int, label, value string, width int) string {
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

func (m *Model) renderCaseFieldBody(field int, value string, width int) []string {
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

func (m *Model) renderCaseInput(width int) string {
	input := m.caseEdit.Input
	input.Width = width

	return input.View()
}

func (m *Model) renderCaseArtifactLines() []string {
	metadata, ok := m.selectedCaseMetadata()
	if !ok {
		return []string{"No case selected."}
	}

	return []string{
		fmt.Sprintf("Dumps             %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, core.WorkspaceNodeDump)),
		fmt.Sprintf("Quizzes           %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, core.WorkspaceNodeQuiz)),
		fmt.Sprintf("Specs             %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, core.WorkspaceNodeSpec)),
		fmt.Sprintf("Plans             %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, core.WorkspaceNodePlan)),
		fmt.Sprintf("Board tickets     %d", workspaceIndexCaseDocumentCount(m.workspaceIndex, metadata.Slug, core.WorkspaceNodeTicket)),
		fmt.Sprintf("Unresolved refs   %d", workspaceGraphCaseReferenceCount(m.workspaceGraph, metadata.Slug)),
		"",
		"Source",
		"  " + emptyValue(projectMetadataRelativePath(m.projectDir, metadata.SourcePath)),
	}
}

func (m *Model) selectedCaseMetadata() (core.CaseMetadata, bool) {
	if len(m.cases) == 0 || m.selectedCase < 0 || m.selectedCase >= len(m.cases) {
		return core.CaseMetadata{}, false
	}

	return m.cases[m.selectedCase], true
}

func (m *Model) filteredCases() []core.CaseMetadata {
	filter := strings.ToLower(strings.TrimSpace(m.caseFilter))
	if filter == "" {
		return append([]core.CaseMetadata(nil), m.cases...)
	}

	if len(m.workspaceIndex.CasesBySlug) > 0 {
		return m.caseMetadataFromSuggestions(m.workspaceIndex.SuggestCases(filter))
	}

	cases := []core.CaseMetadata{}
	for _, metadata := range m.cases {
		haystack := strings.ToLower(metadata.Name + " " + metadata.Description)
		if strings.Contains(haystack, filter) {
			cases = append(cases, metadata)
		}
	}

	return cases
}

func (m *Model) caseMetadataFromSuggestions(
	suggestions []core.WorkspaceCaseSuggestion,
) []core.CaseMetadata {
	bySlug := make(map[string]core.CaseMetadata, len(m.cases))
	for _, metadata := range m.cases {
		bySlug[metadata.Slug] = metadata
	}

	cases := make([]core.CaseMetadata, 0, len(suggestions))
	for _, suggestion := range suggestions {
		metadata := bySlug[suggestion.Slug]
		metadata.SchemaVersion = core.CaseMetadataSchemaVersion
		metadata.Slug = suggestion.Slug
		metadata.Name = suggestion.Name
		metadata.Description = suggestion.Description
		metadata.SourcePath = suggestion.Path
		metadata.Exists = true
		cases = append(cases, metadata)
	}

	return cases
}
