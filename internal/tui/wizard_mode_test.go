package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Earlier and future tabs never send a current-stage mutation. Continue only
// navigates, and palette/manual source-edit paths cannot bypass the mode guard.
func TestWizardModeIsolation(t *testing.T) {
	model := guidedTerminalFixture()
	calls := 0

	model.guidedExecutor = guidedTestExecutor(func(GuidedRequest) (GuidedView, error) { calls++; return model.guided.View, nil })
	for _, view := range []View{ViewDumps, ViewSpecs, ViewPlans, ViewBoard} {
		model.selectView(view)
		checkWizardReadKeys(t, model, view)

		_, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		if command != nil || model.SelectedView() != ViewQuizzes {
			t.Fatal("Continue submitted instead of navigating")
		}
	}

	for _, action := range []CommandAction{CommandActionPlanDump, CommandActionPlanQuiz, CommandActionPlanSpec, CommandActionPlanPlan, CommandActionPlanTicket, CommandActionPlanArtifact, CommandActionPlanWork, CommandActionOpenSelectedSource} {
		if command := model.executeCommand(Command{Action: action}); command != nil {
			t.Fatal("manual palette bypassed wizard mode")
		}
	}

	if calls != 0 {
		t.Fatal("read-only navigation reached the owner")
	}
}

// A draft from another bound run is rejected before adapter IO, while explicit
// context clearing leaves manual editors reachable through their normal owner.
func TestWizardDraftScope(t *testing.T) {
	model := guidedTerminalFixture()
	request := model.guidedRequest("input")
	request.Command.RunID = "other-run"

	model.guidedExecutor = guidedTestExecutor(func(GuidedRequest) (GuidedView, error) {
		t.Fatal("foreign draft reached owner")

		return GuidedView{}, nil
	})
	if command := model.startGuidedRequest(request); command != nil {
		t.Fatal("foreign run accepted")
	}

	model.executeCommand(Command{Action: CommandActionClearCase})

	if model.wizardBound() || model.wizardManualCommand(Command{Action: CommandActionPlanDump}) {
		t.Fatal("explicit manual navigation retained wizard guard")
	}
}

func checkWizardReadKeys(t *testing.T, model *Model, view View) {
	t.Helper()

	for _, key := range []string{"e", "enter", "g", "s", "ctrl+s", "p", "x", "alt+e"} {
		message := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		if key == "enter" {
			message = tea.KeyMsg{Type: tea.KeyEnter}
		}

		if key == "ctrl+s" {
			message = tea.KeyMsg{Type: tea.KeyCtrlS}
		}

		if key == "alt+e" {
			message = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}, Alt: true}
		}

		_, command := model.Update(message)
		if command != nil || model.textForm.Active || model.dumpEdit.Active || !strings.Contains(model.statusMessage, "Read-only") {
			t.Fatal("non-current stage dispatched or opened a writer", view, key)
		}
	}
}
