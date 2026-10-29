package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Preparing a plan is one operation for missing and retained documents. The
// terminal must not advertise or confirm the former second tracking step.
func TestPlanPreparationControls(t *testing.T) {
	for _, status := range []string{"missing", "planned"} {
		t.Run(status, func(t *testing.T) {
			executor := &fakePlanActionExecutor{}
			model := NewModel(Options{PlanExecutor: executor, Plans: []PlanView{{Title: "Alpha", Status: status, Scope: QuizScope{CaseSlug: "alpha"}}}})
			model.selectView(ViewPlans)

			model.commandPalette.Filter = "tracking"
			if len(model.filteredCommands()) != 0 || strings.Contains(model.planActionHint(), "tracking") || !strings.Contains(model.planActionHint(), "plan + tickets") {
				t.Fatal("separate tracking step is advertised", model.planActionHint())
			}

			_, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t"), Alt: true})
			if command != nil || model.planAction.Pending || executor.request.Action != "" {
				t.Fatal("obsolete shortcut requested a second action")
			}
		})
	}
}
