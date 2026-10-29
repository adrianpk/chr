package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

// A reply for an earlier catalog read cannot replace form inputs. Retry uses
// the exact captured request; changed fields invalidate the older recovery.
func TestSpecFormOwnership(t *testing.T) {
	executor := &specFormExecutor{fakeSpecActionExecutor: fakeSpecActionExecutor{err: errors.New("temporary failure"), result: SpecActionResult{Outcome: operation.Result{Status: operation.Blocked, Retry: true}}}}
	model := NewModel(Options{SpecExecutor: executor})
	model.selectView(ViewSpecs)

	binding := model.beginOperation("spec")
	model.openSpecCreateForm()
	model.textForm.Fields[1].Input.SetValue("Retained title")
	model.textForm.Fields[2].Input.SetValue("UNSENT_SPEC")
	model.completeSpecAction(specActionFinishedMsg{Binding: binding, Result: SpecActionResult{Specs: []SpecView{{Title: "Older catalog"}}}})

	if !model.textForm.Active || model.textForm.Fields[2].value() != "UNSENT_SPEC" || len(model.specs) != 0 {
		t.Fatal("old response replaced the form")
	}

	model.submitSpecForm()

	request := executor.request

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if executor.calls != 2 || executor.request != request {
		t.Fatal("retry changed the original identity")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" changed")})
	model.Update(tea.KeyMsg{Type: tea.KeyF2})
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if model.SelectedView() != ViewSpecs || model.textForm.Fields[1].value() != "Retained title changed" || model.feedbackCurrent() {
		t.Fatal("navigation or resize lost ownership")
	}
}

type specFormExecutor struct {
	fakeSpecActionExecutor
	calls int
}

func (e *specFormExecutor) ExecuteSpecAction(request SpecActionRequest) (SpecActionResult, error) {
	e.calls++

	return e.fakeSpecActionExecutor.ExecuteSpecAction(request)
}

// Confirming a reload still cannot rebind the captured case identity silently.
func TestSpecReloadScope(t *testing.T) {
	executor := &fakeSpecActionExecutor{result: SpecActionResult{Document: &SpecView{Title: "Other", Content: "Other case", Scope: QuizScope{CaseSlug: "other"}}}}
	model := NewModel(Options{SpecExecutor: executor})
	request := SpecActionRequest{Action: SpecActionSave, Path: "captured", Scope: QuizScope{CaseSlug: "original"}}
	model.openSpecForm(request, SpecView{Title: "Retained", Content: "UNSENT_SPEC", Scope: request.Scope})
	model.reloadSpecForm(request)

	if model.textForm.Fields[2].value() != "UNSENT_SPEC" || model.textForm.SpecRequest.Scope != request.Scope {
		t.Fatal("reload silently rebound its case")
	}
}
