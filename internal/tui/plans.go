package tui

import (
	"strings"
	"time"
)

// PlanView is the terminal projection of a generated plan document.
type PlanView struct {
	Revision     string
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        QuizScope
	RelativePath string
	UpdatedAt    time.Time
	Content      string
}

func (m *Model) renderPlans() string { return m.documentCatalog(ViewPlans).render() }

func (m *Model) planActionHint() string {
	if m.selectedPlanCanMaterializeTracking() {
		return "g generate missing  G regenerate  alt+t tracking  r refresh"
	}

	return "g generate missing  G regenerate  r refresh"
}

func (m *Model) selectedPlanCanMaterializeTracking() bool {
	plan, ok := m.selectedPlanView()
	if !ok {
		return false
	}

	return strings.TrimSpace(plan.RelativePath) != "" &&
		strings.TrimSpace(plan.Content) != "" &&
		plan.Status != "missing"
}

func (m *Model) selectedPlanView() (PlanView, bool) {
	for _, plan := range m.visiblePlans() {
		if plan.Index == m.selectedPlan {
			return plan, true
		}
	}

	if len(m.visiblePlans()) > 0 {
		return m.visiblePlans()[0], true
	}

	return PlanView{}, false
}

func (m *Model) selectPreviousPlan() {
	if len(m.visiblePlans()) == 0 {
		return
	}

	position := m.selectedPlanPosition()
	if position == 0 {
		m.selectedPlan = m.visiblePlans()[len(m.visiblePlans())-1].Index

		return
	}

	m.selectedPlan = m.visiblePlans()[position-1].Index
}

func (m *Model) selectNextPlan() {
	if len(m.visiblePlans()) == 0 {
		return
	}

	position := m.selectedPlanPosition()
	if position >= len(m.visiblePlans())-1 {
		m.selectedPlan = m.visiblePlans()[0].Index

		return
	}

	m.selectedPlan = m.visiblePlans()[position+1].Index
}

func (m *Model) selectedPlanPosition() int {
	for index, plan := range m.visiblePlans() {
		if plan.Index == m.selectedPlan {
			return index
		}
	}

	return 0
}

func (m *Model) selectPlanByScope(scope QuizScope) {
	for _, plan := range m.visiblePlans() {
		if plan.Scope.CaseSlug == scope.CaseSlug {
			m.selectedPlan = plan.Index

			return
		}
	}

	if len(m.visiblePlans()) > 0 {
		m.selectedPlan = m.visiblePlans()[0].Index

		return
	}

	m.selectedPlan = 0
}

func (m *Model) selectedPlanScope() QuizScope {
	plan, ok := m.selectedPlanView()
	if !ok {
		return QuizScope{CaseSlug: m.caseContext.CaseSlug}
	}

	return plan.Scope
}

func normalizePlanViews(plans []PlanView) []PlanView {
	normalized := append([]PlanView(nil), plans...)
	for index := range normalized {
		normalized[index].Index = index
		if strings.TrimSpace(normalized[index].Status) == "" {
			normalized[index].Status = "missing"
		}
	}

	return normalized
}

func planUseCaseTitle(plan PlanView) string {
	title := strings.TrimSpace(plan.UseCaseTitle)
	if title != "" {
		return title
	}

	title = strings.TrimSpace(plan.Title)
	if title != "" {
		return title
	}

	return quizUseCaseLabel(plan.Scope)
}

func planScopeStatus(plan PlanView) string {
	status := strings.TrimSpace(plan.Status)
	if status == "" {
		status = "missing"
	}

	scope := quizUseCaseLabel(plan.Scope)
	if strings.TrimSpace(plan.RelativePath) == "" {
		return scope + "  " + status
	}

	return scope + "  " + status + "  " + plan.RelativePath
}
