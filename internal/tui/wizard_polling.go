package tui

import authoringcore "forge.adrianpk.com/chr/internal/core/authoring"

func (m *Model) wizardWriteBusy() bool { return m.guided.Busy && !m.guided.Polling }

// At most one passive read can overlap an explicit request. The request takes
// presentation ownership; sequence checks ignore the read when it returns.
func (m *Model) guidedRequestBlocked(request GuidedRequest) bool {
	if !m.guided.Busy {
		return false
	}

	return !m.guided.Polling || request.Passive
}

func (m *Model) applyGuidedCases(cases []authoringcore.CaseMetadata) {
	selected := ""
	if metadata, ok := m.selectedCaseMetadata(); ok {
		selected = metadata.Slug
	}

	m.cases = append([]authoringcore.CaseMetadata(nil), cases...)
	if caseSliceContainsSlug(m.cases, selected) {
		m.selectedCase = m.caseIndexForSlug(selected)

		return
	}

	m.clampSelectedCase()
}

func (m *Model) wizardInputNotice() string {
	if m.wizardUnsentInput() {
		return "Unsent input remains available. Refresh does not submit or replace it."
	}

	return ""
}

func (m *Model) wizardUnsentInput() bool {
	if m.quizAnswer.Active && m.quizAnswer.Guided != nil && m.quizAnswer.Input.Value() != "" {
		return true
	}

	if m.dumpEdit.Active && m.dumpEdit.Guided != nil && m.dumpEdit.Input.Value() != m.dumpEdit.Original {
		return true
	}

	return m.textForm.Active && m.textForm.GuidedRequest != nil && m.textFormHasInput()
}
