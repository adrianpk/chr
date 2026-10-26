package tui

import (
	"strings"
	"time"
)

// SpecView is the terminal projection of a spec document.
type SpecView struct {
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        QuizScope
	RelativePath string
	UpdatedAt    time.Time
	Content      string
}

func (m *Model) renderSpecs() string { return m.documentCatalog(ViewSpecs).render() }

func (m *Model) specActionHint() string {
	return "g generate spec  f generate feature  r refresh"
}

func (m *Model) selectedSpecView() (SpecView, bool) {
	for _, spec := range m.specs {
		if spec.Index == m.selectedSpec {
			return spec, true
		}
	}

	if len(m.specs) > 0 {
		return m.specs[0], true
	}

	return SpecView{}, false
}

func (m *Model) selectPreviousSpec() {
	if len(m.specs) == 0 {
		return
	}

	position := m.selectedSpecPosition()
	if position == 0 {
		m.selectedSpec = m.specs[len(m.specs)-1].Index

		return
	}

	m.selectedSpec = m.specs[position-1].Index
}

func (m *Model) selectNextSpec() {
	if len(m.specs) == 0 {
		return
	}

	position := m.selectedSpecPosition()
	if position >= len(m.specs)-1 {
		m.selectedSpec = m.specs[0].Index

		return
	}

	m.selectedSpec = m.specs[position+1].Index
}

func (m *Model) selectedSpecPosition() int {
	for index, spec := range m.specs {
		if spec.Index == m.selectedSpec {
			return index
		}
	}

	return 0
}

func (m *Model) selectSpecByScope(scope QuizScope) {
	for _, spec := range m.specs {
		if spec.Scope.CaseSlug == scope.CaseSlug {
			m.selectedSpec = spec.Index

			return
		}
	}

	if len(m.specs) > 0 {
		m.selectedSpec = m.specs[0].Index

		return
	}

	m.selectedSpec = 0
}

func (m *Model) selectedSpecScope() QuizScope {
	spec, ok := m.selectedSpecView()
	if !ok {
		return QuizScope{}
	}

	return spec.Scope
}

func normalizeSpecViews(specs []SpecView) []SpecView {
	normalized := append([]SpecView(nil), specs...)
	for index := range normalized {
		normalized[index].Index = index
		if strings.TrimSpace(normalized[index].Status) == "" {
			normalized[index].Status = "missing"
		}
	}

	return normalized
}

func specUseCaseTitle(spec SpecView) string {
	title := strings.TrimSpace(spec.UseCaseTitle)
	if title != "" {
		return title
	}

	title = strings.TrimSpace(spec.Title)
	if title != "" {
		return title
	}

	return quizUseCaseLabel(spec.Scope)
}

func specScopeStatus(spec SpecView) string {
	status := strings.TrimSpace(spec.Status)
	if status == "" {
		status = "missing"
	}

	scope := quizUseCaseLabel(spec.Scope)
	if strings.TrimSpace(spec.RelativePath) == "" {
		return scope + "  " + status
	}

	return scope + "  " + status + "  " + spec.RelativePath
}
