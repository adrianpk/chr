package chctui

import (
	"fmt"
	"strings"
	"time"
)

type DumpScope struct {
	CaseSlug string
}

type DumpView struct {
	Index        int
	NodeID       string
	Title        string
	Status       string
	Scope        DumpScope
	RelativePath string
	UpdatedAt    time.Time
	Preview      string
}

func (m Model) renderDumpFocusList(limit int) []string {
	if len(m.dumps) == 0 {
		return []string{"No dumps found.", "Use the command palette to create one."}
	}

	lines := []string{shellMutedStyle.Render("Scope        Status      Title")}

	for _, dump := range m.dumps {
		prefix := " "
		if dump.Index == m.selectedDump {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %-12s %-11s %s", prefix, dumpScopeLabel(dump.Scope), dump.Status, dump.Title))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m Model) renderDumpDetailPanel(width, height int) (string, []string) {
	if m.dumpEdit.Active {
		return "Edit Dump", m.renderSelectedDumpDetailLines(width, height)
	}

	return "Preview", m.renderSelectedDumpDetailLines(width, height)
}

func (m Model) renderSelectedDumpDetail() string {
	lines := append([]string{"Selected Dump"}, m.renderSelectedDumpDetailLines(72, 16)...)

	return strings.Join(lines, "\n")
}

func (m Model) renderSelectedDumpDetailLines(width, height int) []string {
	dump, ok := m.selectedDumpView()
	if !ok {
		return []string{"No dump selected."}
	}

	lines := []string{
		fmt.Sprintf("Scope: %s", dumpScopeLabel(dump.Scope)),
		fmt.Sprintf("Status: %s", dump.Status),
		fmt.Sprintf("Title: %s", dump.Title),
		fmt.Sprintf("Updated: %s", renderTime(dump.UpdatedAt)),
		fmt.Sprintf("Path: %s", dump.RelativePath),
	}

	if m.dumpEdit.Active && m.dumpEdit.Path == dump.RelativePath {
		return m.renderSelectedDumpEditLines(lines, width, height)
	}

	lines = append(lines, "", "Preview:")
	lines = append(lines, indentLines(wrapPlainText(emptyValue(dump.Preview), maxInt(width-2, 1)), "  ")...)

	return lines
}

func (m Model) renderSelectedDumpEditLines(header []string, width, height int) []string {
	lines := append([]string(nil), header...)
	lines = append(lines,
		"",
		shellMutedStyle.Render("Editing Markdown. ctrl+s save | esc cancel | alt+e external editor"),
		"",
	)

	editorHeight := maxInt(height-len(lines), 3)
	lines = append(lines, strings.Split(m.renderDumpEditInput(maxInt(width-2, 1), editorHeight), "\n")...)

	return lines
}

func indentLines(lines []string, prefix string) []string {
	indented := make([]string, 0, len(lines))
	for _, line := range lines {
		indented = append(indented, prefix+line)
	}

	return indented
}

func (m Model) renderDumpEditInput(width, height int) string {
	input := m.dumpEdit.Input
	input.SetWidth(maxInt(width, 1))
	input.SetHeight(maxInt(height, 3))

	return input.View()
}

func (m Model) renderDumps() string {
	if len(m.dumps) == 0 {
		return "Dumps\n  No dumps found.\n  Use the command palette to create one."
	}

	var output strings.Builder

	output.WriteString("Dumps\n")
	output.WriteString("  Scope        Status      Title\n")

	for _, dump := range m.dumps {
		prefix := "  "
		if dump.Index == m.selectedDump {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(fmt.Sprintf("%-12s %-11s %s", dumpScopeLabel(dump.Scope), dump.Status, dump.Title))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedDumpDetail())

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) selectedDumpView() (DumpView, bool) {
	for _, dump := range m.dumps {
		if dump.Index == m.selectedDump {
			return dump, true
		}
	}

	if len(m.dumps) > 0 {
		return m.dumps[0], true
	}

	return DumpView{}, false
}

func (m *Model) selectPreviousDump() {
	if len(m.dumps) == 0 {
		return
	}

	position := m.selectedDumpPosition()
	if position == 0 {
		m.selectedDump = m.dumps[len(m.dumps)-1].Index

		return
	}

	m.selectedDump = m.dumps[position-1].Index
}

func (m *Model) selectNextDump() {
	if len(m.dumps) == 0 {
		return
	}

	position := m.selectedDumpPosition()
	if position >= len(m.dumps)-1 {
		m.selectedDump = m.dumps[0].Index

		return
	}

	m.selectedDump = m.dumps[position+1].Index
}

func (m Model) selectedDumpPosition() int {
	for index, dump := range m.dumps {
		if dump.Index == m.selectedDump {
			return index
		}
	}

	return 0
}

func (m *Model) selectDumpByPath(path string) {
	path = strings.TrimSpace(path)
	if path != "" {
		for _, dump := range m.dumps {
			if dump.RelativePath == path {
				m.selectedDump = dump.Index

				return
			}
		}
	}

	if len(m.dumps) > 0 {
		m.selectedDump = m.dumps[0].Index

		return
	}

	m.selectedDump = 0
}

func normalizeDumpViews(dumps []DumpView) []DumpView {
	normalized := append([]DumpView(nil), dumps...)
	for index := range normalized {
		normalized[index].Index = index
		if strings.TrimSpace(normalized[index].Status) == "" {
			normalized[index].Status = "unknown"
		}

		if strings.TrimSpace(normalized[index].Preview) == "" {
			normalized[index].Preview = "No preview available."
		}
	}

	return normalized
}

func dumpScopeLabel(scope DumpScope) string {
	if strings.TrimSpace(scope.CaseSlug) == "" {
		return "project"
	}

	return scope.CaseSlug
}
