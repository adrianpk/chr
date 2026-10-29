package tui

import (
	"fmt"
	"strings"
)

func (m *Model) wizardBound() bool {
	session := m.guided.View.Session

	return m.guided.Active && session != nil && session.Guided != nil && m.caseContext.RunID == session.ID && m.caseContext.CaseSlug == session.Guided.CaseSlug
}

func (m *Model) wizardStageView() View {
	if m.guided.View.Session == nil {
		return ViewCases
	}

	return wizardDocumentView(m.guided.View.Session.Guided.Stage)
}

func wizardDocumentView(stage string) View {
	switch stage {
	case "name":
		return ViewCases
	case "dump":
		return ViewDumps
	case "quiz":
		return ViewQuizzes
	case "specification":
		return ViewSpecs
	case "preparation", "plan", "tracker":
		return ViewPlans
	default:
		return ViewBoard
	}
}

func (m *Model) wizardStageLabel() string {
	stage := m.guided.View.Session.Guided.Stage
	positions := map[string]int{"name": 1, "dump": 2, "quiz": 3, "specification": 4, "preparation": 5, "execution": 6, "results": 7}

	return fmt.Sprintf("Wizard: %d/7 %s", positions[stage], stage)
}

func (m *Model) wizardDocumentTab() bool {
	if !m.wizardBound() {
		return false
	}

	switch m.SelectedView() {
	case ViewDumps, ViewQuizzes, ViewSpecs, ViewPlans, ViewBoard:
		return true
	default:
		return false
	}
}

func (m *Model) wizardTabLines() []string {
	view := m.guided.View

	lines := []string{"Case: " + view.Session.Guided.CaseName, m.wizardStageLabel() + " · " + guidedCatalogStatus(view.Status)}
	if m.SelectedView() == m.wizardStageView() {
		lines = append(lines, view.Prompt, m.guidedControlHint(), "e Input · r Refresh · F1-F7 existing tabs")
	} else {
		lines = append(lines, "Read-only stage. c Continue to the current stage.")
	}

	lines = append(lines, view.Progress...)
	for _, problem := range view.Problems {
		lines = append(lines, "Prerequisite: "+problem)
	}

	if original := m.wizardTabOriginal(); original != nil {
		lines = append(lines, "Committed original (read-only)")
		if view.Document != nil {
			lines = append(lines, "Original commit: "+original.Commit)
		}

		lines = append(lines, original.Content)
	}

	if m.guided.Busy {
		lines = append(lines, "Reading or saving the retained step…")
	}

	if m.guided.Problem != "" {
		lines = append(lines, m.guided.Problem+" · y retries the original request; r reads current state.")
	}

	return strings.Split(strings.Join(lines, "\n"), "\n")
}

func (m *Model) completeWizardDraft(message guidedFinishedMsg) {
	if message.Request.Action == "start" {
		m.textForm = textFormState{}

		return
	}

	if draft := m.textForm.GuidedRequest; draft != nil && message.Request.Action == "input" && message.Request.Command.OperationID == draft.Command.OperationID {
		m.textForm = textFormState{}
	}
}

// Catalog state alone does not verify native delivery receipts.
func guidedCatalogStatus(status string) string {
	if status == "completed" {
		return "inspect results"
	}

	return status
}

func (m *Model) wizardTabOriginal() *GuidedDocument {
	if original := m.guided.View.Document; original != nil && wizardDocumentView(original.Kind) == m.SelectedView() {
		return original
	}

	for _, original := range m.guided.View.Originals {
		if wizardDocumentView(original.Kind) == m.SelectedView() {
			return &original
		}
	}

	return nil
}
