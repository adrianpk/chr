package tui

import "strings"

// Guided originals use the same document detail component as the manual catalog.
// The owner supplies committed bytes; wizard controls cannot save a manual copy.
func (m *Model) wizardDocumentLines() []string {
	original := m.wizardTabOriginal()
	if original == nil {
		return nil
	}

	if m.SelectedView() != ViewSpecs && m.SelectedView() != ViewPlans {
		return []string{"Committed original (read-only)", original.Content}
	}

	kind, title := "spec", "Specification"
	if m.SelectedView() == ViewPlans {
		kind, title = "plan", "Plan and tickets"
	}

	catalog := documentCatalogView{kind: kind, hint: "Committed original (read-only)", items: []documentCatalogItem{{Title: title, ScopeStatus: original.Reference, Status: "committed", Content: original.Content}}}

	lines := catalog.detailLines(m.wizardContentWidth())
	if m.guided.View.Document != nil {
		lines = append([]string{"Original commit: " + original.Commit}, lines...)
	}

	return lines
}

func (m *Model) wizardContentWidth() int {
	if m.width == 0 {
		return 72
	}

	return max(1, m.width-6)
}

func (m *Model) wizardOriginalLines() []string {
	lines := m.wizardDocumentLines()

	original := m.wizardTabOriginal()
	if original != nil && m.SelectedView() != ViewSpecs && m.SelectedView() != ViewPlans && m.guided.View.Document != nil {
		lines = append([]string{"Original commit: " + original.Commit}, lines...)
	}

	return strings.Split(strings.Join(lines, "\n"), "\n")
}
