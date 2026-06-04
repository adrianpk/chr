package chctui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// DumpScope narrows a dump view to an optional use-case scope.
type DumpScope struct {
	CaseSlug string
}

// DumpView is the terminal projection of a dump document.
type DumpView struct {
	Index        int
	NodeID       string
	UseCaseTitle string
	Title        string
	Status       string
	SourcePath   string
	Scope        DumpScope
	RelativePath string
	UpdatedAt    time.Time
	Preview      string
}

func (m Model) renderDumpFocusList(limit int) []string {
	if len(m.dumps) == 0 {
		return []string{"No dumps found."}
	}

	lines := []string{shellMutedStyle.Render("Use Case")}

	for _, dump := range m.dumps {
		prefix := " "
		if dump.Index == m.selectedDump {
			prefix = ">"
		}

		lines = append(lines, prefix+" "+dumpUseCaseTitle(dump))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m Model) renderDumpDetailPanel(width, height int) (string, []string) {
	if m.dumpEdit.Active {
		return "Edit", m.renderSelectedDumpDetailLines(width, height)
	}

	return "Dump", m.renderSelectedDumpDetailLines(width, height)
}

func (m Model) renderSelectedDumpDetail() string {
	lines := append([]string{"Dump"}, m.renderSelectedDumpDetailLines(72, 16)...)

	return strings.Join(lines, "\n")
}

func (m Model) renderSelectedDumpDetailLines(width, height int) []string {
	dump, ok := m.selectedDumpView()
	if !ok {
		return []string{"No dump selected."}
	}

	lines := []string{
		dump.Title,
		shellMutedStyle.Render(dumpScopeLabel(dump.Scope)),
	}

	if m.dumpEdit.Active && m.dumpEdit.Path == dump.RelativePath {
		return m.renderSelectedDumpEditLines(lines, width, height)
	}

	lines = append(lines, "")
	if strings.TrimSpace(dump.SourcePath) != "" {
		lines = append(lines, shellMutedStyle.Render("Raw: "+dump.SourcePath), "")
	}

	lines = append(lines, renderContentReadLines(emptyValue(dump.Preview), maxInt(width, 1))...)
	lines = append(lines, "", shellMutedStyle.Render(m.dumpActionHint()))

	return lines
}

func renderContentReadLines(content string, width int) []string {
	lines := []string{}

	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "" {
			lines = append(lines, "")

			continue
		}

		lines = append(lines, wrapPlainText(line, width)...)
	}

	return lines
}

func (m Model) renderSelectedDumpEditLines(_ []string, width, height int) []string {
	lines := []string{shellMutedStyle.Render(m.dumpActionHint())}

	editorHeight := maxInt(height-len(lines)-1, 3)
	lines = append(lines, strings.Split(m.renderDumpEditInput(maxInt(width, 1), editorHeight), "\n")...)

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
		return "Dumps\n  No dumps found.\n\n  " + m.dumpActionHint()
	}

	var output strings.Builder

	output.WriteString("Dumps\n")

	for _, dump := range m.dumps {
		prefix := "  "
		if dump.Index == m.selectedDump {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(dumpUseCaseTitle(dump))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedDumpDetail())

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) dumpActionHint() string {
	if m.dumpEdit.Active {
		return "ctrl+s save  esc cancel  alt+e external"
	}

	if len(m.dumps) == 0 {
		return "c create  r refresh"
	}

	return "enter edit  alt+e external  c create  r refresh"
}

func (m Model) updateDumpsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		m.planDumpAction(DumpActionCreate)
	case "enter":
		m.planDumpAction(DumpActionEdit)

		return m, m.confirmDumpAction()
	case "r":
		m.planDumpAction(DumpActionRefresh)

		return m, m.confirmDumpAction()
	case "esc":
		m.selectPrimaryPane()
	default:
		return m.updateNavigationKey(msg)
	}

	return m, nil
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
		return "Project-wide"
	}

	return scope.CaseSlug
}

func dumpUseCaseTitle(dump DumpView) string {
	title := strings.TrimSpace(dump.UseCaseTitle)
	if title != "" {
		return title
	}

	return dumpScopeLabel(dump.Scope)
}
