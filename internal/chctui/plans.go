package chctui

import (
	"strings"
	"time"
)

type PlanView struct {
	Index        int
	UseCaseTitle string
	Title        string
	Status       string
	Scope        QuizScope
	RelativePath string
	UpdatedAt    time.Time
	Content      string
}

func (m Model) renderPlanFocusList(limit int) []string {
	if len(m.plans) == 0 {
		return []string{"No use cases found."}
	}

	lines := []string{shellMutedStyle.Render("Use Case")}

	for _, plan := range m.plans {
		prefix := " "
		if plan.Index == m.selectedPlan {
			prefix = ">"
		}

		lines = append(lines, prefix+" "+planUseCaseTitle(plan))
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (m Model) renderSelectedPlanDetailLines(width int) []string {
	plan, ok := m.selectedPlanView()
	if !ok {
		return []string{"No use case selected."}
	}

	lines := []string{
		shellMutedStyle.Render(m.planActionHint()),
		"",
		plan.Title,
		shellMutedStyle.Render(planScopeStatus(plan)),
	}

	if strings.TrimSpace(plan.Content) == "" {
		lines = append(lines, "", "No generated plan found for this use case.")
		if m.planBusy {
			lines = append(lines, "", "Working...")
			lines = append(lines, "  "+m.statusMessage)
		}

		return lines
	}

	if m.planBusy {
		lines = append(lines, "", "Working...")
		lines = append(lines, "  "+m.statusMessage)
	}

	lines = append(lines, "")
	lines = append(lines, renderContentReadLines(plan.Content, maxInt(width, 1))...)

	return lines
}

func (m Model) renderPlans() string {
	if len(m.plans) == 0 {
		return "Plans\n  No use cases found.\n\n  " + m.planActionHint()
	}

	var output strings.Builder

	output.WriteString("Plans\n")

	for _, plan := range m.plans {
		prefix := "  "
		if plan.Index == m.selectedPlan {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(planUseCaseTitle(plan))
		output.WriteString("  ")
		output.WriteString(shellMutedStyle.Render(plan.Status))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(strings.Join(m.renderSelectedPlanDetailLines(72), "\n"))

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) planActionHint() string {
	return "g generate missing  G regenerate  r refresh"
}

func (m Model) selectedPlanView() (PlanView, bool) {
	for _, plan := range m.plans {
		if plan.Index == m.selectedPlan {
			return plan, true
		}
	}

	if len(m.plans) > 0 {
		return m.plans[0], true
	}

	return PlanView{}, false
}

func (m *Model) selectPreviousPlan() {
	if len(m.plans) == 0 {
		return
	}

	position := m.selectedPlanPosition()
	if position == 0 {
		m.selectedPlan = m.plans[len(m.plans)-1].Index

		return
	}

	m.selectedPlan = m.plans[position-1].Index
}

func (m *Model) selectNextPlan() {
	if len(m.plans) == 0 {
		return
	}

	position := m.selectedPlanPosition()
	if position >= len(m.plans)-1 {
		m.selectedPlan = m.plans[0].Index

		return
	}

	m.selectedPlan = m.plans[position+1].Index
}

func (m Model) selectedPlanPosition() int {
	for index, plan := range m.plans {
		if plan.Index == m.selectedPlan {
			return index
		}
	}

	return 0
}

func (m *Model) selectPlanByScope(scope QuizScope) {
	for _, plan := range m.plans {
		if plan.Scope.CaseSlug == scope.CaseSlug {
			m.selectedPlan = plan.Index

			return
		}
	}

	if len(m.plans) > 0 {
		m.selectedPlan = m.plans[0].Index

		return
	}

	m.selectedPlan = 0
}

func (m Model) selectedPlanScope() QuizScope {
	plan, ok := m.selectedPlanView()
	if !ok {
		return QuizScope{}
	}

	return plan.Scope
}

func (m Model) planExistsForScope(scope QuizScope) bool {
	for _, plan := range m.plans {
		if plan.Scope.CaseSlug != scope.CaseSlug {
			continue
		}

		return planHasCanonicalContent(plan)
	}

	return false
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

func planHasCanonicalContent(plan PlanView) bool {
	if strings.TrimSpace(plan.RelativePath) != "" {
		return true
	}

	if strings.TrimSpace(plan.Content) != "" {
		return true
	}

	status := strings.TrimSpace(plan.Status)

	return status != "" && status != "missing"
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
