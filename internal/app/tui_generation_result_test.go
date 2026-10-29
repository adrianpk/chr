package app

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/tui"
)

// Successful generation offers only retained results. Opening a document or
// its case-bound tickets uses the existing shell and performs no further write.
func TestTUIGenerationResult(t *testing.T) {
	for _, family := range []string{"spec", "plan"} {
		t.Run(family, func(t *testing.T) { checkTUIGenerationResult(t, family) })
	}
}

func checkTUIGenerationResult(t *testing.T, family string) {
	t.Helper()

	provider, _ := webPlanDeliveryFixture(t)
	operations := provider.service(t.Context())
	model := tuiGenerationModel(t, operations)

	view, key := tui.ViewPlans, tea.KeyF5
	if family == "spec" {
		view, key = tui.ViewSpecs, tea.KeyF4
	}

	model.Update(tea.KeyMsg{Type: key})
	runGuidedTUIKey(t, model, "g", true)

	result := model.View()
	if !strings.Contains(result, "o Open document") || strings.Contains(result, "t View tickets") != (family == "plan") {
		t.Fatal("result actions do not match retained outputs", result)
	}

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	runGuidedTUIKey(t, model, "o", true)

	if model.SelectedView() != view || !strings.Contains(model.View(), "Case: Named work") || strings.Contains(model.View(), "Operation: ok") {
		t.Fatal("document did not open in its case", model.View())
	}

	if family == "plan" {
		checkTUIGenerationTickets(t, model)
	}

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("result navigation changed canonical documents")
	}
}

func checkTUIGenerationTickets(t *testing.T, model *tui.Model) {
	t.Helper()
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o"), Alt: true})
	runGuidedTUIKey(t, model, "t", true)

	if model.SelectedView() != tui.ViewBoard || !strings.Contains(model.View(), "Case: Named work") || !strings.Contains(model.View(), "Implement foundation") {
		t.Fatal("prepared tickets did not open in their case", model.View())
	}
}

func tuiGenerationModel(t *testing.T, operations service) *tui.Model {
	t.Helper()

	options, err := (commandFactory{deps: testAppDependencies()}).loadTUIOptions(t.Context(), operations.env, operations.projectContext)
	if err != nil {
		t.Fatal(err)
	}

	options.SpecExecutor, options.PlanExecutor = newTUISpecActionExecutor(operations), newTUIPlanActionExecutor(operations)
	model := tui.NewModel(options)
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})

	return model
}
