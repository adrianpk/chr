package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chr/internal/core"
)

func (m *Model) renderProjectWorkspace(width, height int) string {
	horizontal := m.splitWidths(width)
	leftWidth, rightWidth := horizontal.Left, horizontal.Right
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

	vertical := m.splitHeights(height)
	rightTopHeight, rightBottomHeight := vertical.Top, vertical.Bottom
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

func (m *Model) renderProjectContextLines(width int) []string {
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

	if m.harnessExecutor != nil {
		lines = append(lines, "", "h: Harness defaults")
	}

	if m.projectDirty {
		lines = append(lines, "", shellMutedStyle.Render("Unsaved changes. Press s to save."))
	}

	return lines
}

func (m *Model) renderProjectFieldLine(field int, label, value string, width int) string {
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

func (m *Model) renderProjectFieldBody(field int, value string, width int) []string {
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

func (m *Model) renderProjectInput(width int) string {
	input := m.projectEdit.Input
	input.Width = width

	return input.View()
}

func (m *Model) renderProjectArtifactLines() []string {
	return []string{
		fmt.Sprintf("Board tickets     %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, core.WorkspaceNodeTicket)),
		fmt.Sprintf("Dumps             %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, core.WorkspaceNodeDump)),
		fmt.Sprintf("Quizzes           %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, core.WorkspaceNodeQuiz)),
		fmt.Sprintf("Specs             %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, core.WorkspaceNodeSpec)),
		fmt.Sprintf("Plans             %d", workspaceGraphProjectDocumentCount(m.workspaceGraph, core.WorkspaceNodePlan)),
		fmt.Sprintf("Diagnostics       %d", len(m.workspaceGraph.Diagnostics)),
	}
}

func (m *Model) renderProjectLocationLines() []string {
	return []string{
		"Path",
		"  " + emptyValue(m.projectDir),
		"",
		"File",
		"  " + emptyValue(projectMetadataRelativePath(m.projectDir, m.projectMetadata.SourcePath)),
	}
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

func (m *Model) projectDisplayName() string {
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
