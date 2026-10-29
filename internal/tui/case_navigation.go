package tui

import "fmt"

type caseNavigationContext struct {
	CaseSlug string
	RunID    string
}

func (m *Model) captureCaseContext() {
	if m.SelectedView() != ViewCases {
		return
	}

	metadata, ok := m.selectedCaseMetadata()
	if !ok || metadata.Slug == "" {
		return
	}

	runID := ""
	if metadata.Slug == m.caseContext.CaseSlug {
		runID = m.caseContext.RunID
	}

	m.bindCaseContext(metadata.Slug, runID)
}

func (m *Model) bindCaseContext(slug, runID string) {
	context := caseNavigationContext{CaseSlug: slug, RunID: runID}
	if context == m.caseContext {
		return
	}

	if session := m.guided.View.Session; session != nil && (runID != session.ID || slug != session.Guided.CaseSlug) {
		m.guided.Active = false
	}

	m.caseContext = context
	if caseSliceContainsSlug(m.cases, slug) {
		m.selectedCase = m.caseIndexForSlug(slug)
	}

	if dump, ok := m.selectedDumpView(); ok {
		m.selectedDump = dump.Index
	}

	if quiz, ok := m.selectedQuizView(); ok {
		m.selectedQuiz = quiz.Index
	}

	if spec, ok := m.selectedSpecView(); ok {
		m.selectedSpec = spec.Index
	}

	if plan, ok := m.selectedPlanView(); ok {
		m.selectedPlan = plan.Index
	}

	m.boardColumn = defaultBoardColumn(m.caseTickets())
	m.selectFirstFilteredTicket()
}

func (m *Model) caseContextAvailable() bool {
	if m.caseContext.CaseSlug == "" {
		return true
	}

	return caseSliceContainsSlug(m.cases, m.caseContext.CaseSlug)
}

func (m *Model) caseContextLabel() string {
	slug := m.caseContext.CaseSlug
	if slug == "" {
		return ""
	}

	for _, metadata := range m.cases {
		if metadata.Slug == slug {
			return "Case: " + metadata.Name
		}
	}

	return "Case: " + slug + " (unavailable)"
}

func (m *Model) explainCaseContext() {
	if !m.caseContextAvailable() {
		m.statusMessage = fmt.Sprintf("Case %s is unavailable. Refresh Cases or choose another case; the palette can clear case context.", m.caseContext.CaseSlug)
	}
}

func (m *Model) reconcileCaseTicketSelection() {
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		if _, ok := m.selectedTicketView(); !ok {
			m.selectFirstFilteredTicket()
		}
	}
}
