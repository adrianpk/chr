package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// A queued command uses the observed case/revision. Its completion must not
// replace a different selection or a newer observation of that same case.
func TestPlanBinding(t *testing.T) {
	for _, change := range []string{"selection", "revision"} {
		t.Run(change, func(t *testing.T) {
			executor := &fakePlanActionExecutor{result: PlanActionResult{Plans: []PlanView{{Title: "Old completion"}}}}
			model := NewModel(Options{PlanExecutor: executor, Plans: []PlanView{
				{Title: "Alpha", Scope: QuizScope{CaseSlug: "alpha"}, Revision: "alpha-v1"},
				{Title: "Beta", Scope: QuizScope{CaseSlug: "beta"}, Revision: "beta-v1"},
			}})
			model.selectView(ViewPlans)

			_, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
			if command == nil {
				t.Fatal("no captured command")
			}

			if change == "selection" {
				model.selectedPlan = 1
			} else {
				model.plans[0].Revision = "alpha-v2"
			}

			model.Update(command())

			request := executor.request
			if request.Scope.CaseSlug != "alpha" || request.Revision != "alpha-v1" || request.OperationID == "" || model.planBusy || len(model.plans) != 2 {
				t.Fatal("queued action or late result crossed its binding", request, model.plans)
			}
		})
	}
}
