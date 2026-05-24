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
		return []string{"No dumps found."}
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

func (m Model) renderSelectedDumpDetail() string {
	dump, ok := m.selectedDumpView()
	if !ok {
		return "Selected Dump\n  No dump selected."
	}

	var output strings.Builder

	output.WriteString("Selected Dump\n")
	output.WriteString(fmt.Sprintf("  Scope: %s\n", dumpScopeLabel(dump.Scope)))
	output.WriteString(fmt.Sprintf("  Status: %s\n", dump.Status))
	output.WriteString(fmt.Sprintf("  Title: %s\n", dump.Title))
	output.WriteString(fmt.Sprintf("  Updated: %s\n", renderTime(dump.UpdatedAt)))
	output.WriteString(fmt.Sprintf("  Path: %s\n", dump.RelativePath))
	output.WriteString("  Preview:\n")
	output.WriteString(indentText(dump.Preview, "    "))

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) renderDumps() string {
	if len(m.dumps) == 0 {
		return "Dumps\n  No dumps found."
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
