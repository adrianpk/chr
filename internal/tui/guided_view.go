package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) guidedLines() []string {
	view := m.guided.View

	lines := []string{"Project: " + view.Project + " · " + view.Namespace}
	if view.Session == nil {
		lines = append(lines, "Develop a new use case: n. This proceeds through specification approval and exact development under project policy.")

		lines = append(lines, view.Policy...)
		for index, run := range view.Runs {
			prefix := "  "
			if index == m.guided.Selected {
				prefix = "> "
			}

			lines = append(lines, prefix+run.Topic+" · "+run.Guided.Stage+" · "+run.Guided.Status)
		}
	} else {
		lines = append(lines, "Use case: "+view.Session.Topic, "Stage: "+view.Session.Guided.Stage+" · "+view.Status, view.Prompt)
		lines = append(lines, view.Progress...)
	}

	for _, problem := range view.Problems {
		lines = append(lines, "Prerequisite: "+problem)
	}

	if document := view.Document; document != nil {
		lines = append(lines, "Inspect "+document.Kind)
		if document.Commit != "" {
			lines = append(lines, "Original commit: "+document.Commit, "Source: "+document.Reference)
		}

		lines = append(lines, document.Content)
	}

	return lines
}

func (m Model) renderGuidedWorkspace(width, height int) string {
	var lines []string
	for _, line := range m.guidedLines() {
		lines = append(lines, strings.Split(lipgloss.NewStyle().Width(max(width-4, 1)).Render(line), "\n")...)
	}

	visible := max(height-5, 1)
	if m.guided.Editing {
		visible = max(visible-7, 1)
	}

	offset := min(m.guided.Scroll, max(len(lines)-visible, 0))
	if m.guided.View.Session == nil {
		for index, line := range lines {
			if strings.HasPrefix(line, "> ") {
				offset = max(0, index-visible+1)
			}
		}
	}

	lines = lines[offset:min(offset+visible, len(lines))]
	if m.guided.Editing {
		lines = append(lines, strings.Split(m.guided.Input.View(), "\n")...)
	}

	lines = append(lines, fmt.Sprintf("View lines %d+ · %s", offset+1, m.guidedHint()))
	if m.guided.Busy {
		lines = append(lines, "Reading or saving the retained step…")
	}

	if m.guided.Problem != "" {
		lines = append(lines, wrapPlainText(m.guided.Problem+" · y retries the original request; r reads current state.", max(width-4, 1))...)
	}

	return renderPanel("Guided use cases", lines, width, height, true)
}

func (m Model) guidedHint() string {
	if m.guided.Editing {
		return "Enter saves · Ctrl+J adds a line · Esc leaves input · Ctrl+C quits"
	}

	if m.guided.View.Session == nil {
		return "n Develop a new use case · Enter opens selected run · up/down selects · Esc returns"
	}

	return m.guidedControlHint() + "e input · i later idea · v results · 1-5 originals · scroll up/down · r refresh · d discard draft · b runs · Esc returns"
}
