package app

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/tui"
)

type tuiWizardLostAck struct {
	owner    tuiGuidedExecutor
	requests []tui.GuidedRequest
	lost     bool
}

func (e *tuiWizardLostAck) ExecuteGuided(request tui.GuidedRequest) (tui.GuidedView, error) {
	view, err := e.owner.ExecuteGuided(request)
	if request.Action == "input" && request.Command.Action == "finish" {
		e.requests = append(e.requests, request)
		if err == nil && !e.lost {
			e.lost = true

			return view, errors.New("lost terminal acknowledgement")
		}
	}

	return view, err
}

// A real finish seals the edited shared Dump body once. Losing only its terminal
// acknowledgement keeps the draft; repeating Continue reconciles the same input.
func TestTUIWizardDumpReplay(t *testing.T) {
	operations, _ := guidedTUIFixture(t)
	executor := &tuiWizardLostAck{owner: tuiGuidedExecutor{operations: operations}}
	model := tui.NewModel(tui.Options{ProjectID: operations.projectContext.ProjectID, GuidedExecutor: executor})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	runGuidedTUIKey(t, model, "w", false)
	runGuidedTUIKey(t, model, "Exact capture workflow", false)
	runGuidedTUIKey(t, model, "ctrl+s", true)
	runGuidedTUIKey(t, model, "e", false)
	runGuidedTUIKey(t, model, "Initial body", false)
	runGuidedTUIKey(t, model, " with the final unsent addition.", false)

	const content = "Initial body with the final unsent addition."

	runGuidedTUIKey(t, model, "ctrl+s", true)

	if !strings.Contains(model.View(), content) || !strings.Contains(model.View(), "lost terminal acknowledgement") {
		t.Fatal("lost acknowledgement discarded draft or failure", model.View())
	}

	checkTUIWizardDumpRetry(t, operations, model, executor, content)
}

func checkTUIWizardDumpRetry(t *testing.T, operations service, model *tui.Model, executor *tuiWizardLostAck, content string) {
	t.Helper()

	before := onlyGuidedTUIRun(t, operations)
	if before.Guided.Capture != content || len(before.Guided.Checkpoints) != 1 {
		t.Fatal("finish lost final content or duplicated checkpoint")
	}

	runGuidedTUIKey(t, model, "ctrl+s", true)

	after := onlyGuidedTUIRun(t, operations)
	if len(executor.requests) != 2 || executor.requests[0] != executor.requests[1] || after.ID != before.ID || after.Revision != before.Revision || len(after.Guided.Checkpoints) != 1 {
		t.Fatal("retry replaced input or advanced the owner again")
	}

	original, err := operations.ReadGuidedUseCaseOriginal(t.Context(), after.ID, after.Namespace, "dump")
	if err != nil || strings.Count(original.Content, content) != 1 {
		t.Fatal("committed original did not retain exact content once", err)
	}
}
