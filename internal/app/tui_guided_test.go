package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/tui"
)

func guidedTUIFixture(t *testing.T) (service, tui.Model) {
	t.Helper()

	provider, _ := guidedWebFixture(t)
	operations := provider.service(t.Context())
	guidedTelegramModel(t, &operations)

	return operations, tui.NewModel(tui.Options{ProjectDir: operations.projectContext.ProjectDir, ProjectID: operations.projectContext.ProjectID, GuidedExecutor: tuiGuidedExecutor{operations: operations}})
}

// TestGuidedUseCaseTUIJourney sends actual terminal model keys through the shared
// service, SQLite/filesystem and Git owners to clarification and default approval.
func TestGuidedUseCaseTUIJourney(t *testing.T) {
	t.Parallel()

	operations, model := guidedTUIFixture(t)
	openGuidedTUI(t, &model)
	runGuidedTUIKey(t, &model, "n", true)

	if !strings.Contains(model.View(), "use-case name") {
		t.Fatal("new journey did not request its name")
	}

	enterGuidedTUIInput(t, &model, "Terminal audit")
	enterGuidedTUIInput(t, &model, "Retain complete evidence and finite limits.")
	runGuidedTUIKey(t, &model, "f", true)

	if !strings.Contains(model.View(), "What should the workflow preserve") {
		t.Fatal("current clarification missing", model.View())
	}

	enterGuidedTUIInput(t, &model, "Preserve the original agreed scope.")

	session := onlyGuidedTUIRun(t, operations)
	if session.Guided.SpecApproval == nil || session.Guided.Stage != "preparation" || len(session.Guided.Checkpoints) != 3 {
		t.Fatal("terminal input did not approve the exact generated specification")
	}

	assertGuidedTUIOriginals(t, operations, session)

	_, reopened := guidedTUIModel(operations)
	openGuidedTUI(t, &reopened)

	if !strings.Contains(reopened.View(), "Terminal audit") || strings.Contains(reopened.View(), session.ID) {
		t.Fatal("retained navigation lost its visible name or exposed a transport ID")
	}

	runGuidedTUIKey(t, &reopened, "enter", true)

	after := onlyGuidedTUIRun(t, operations)
	if after.ID != session.ID || after.Revision != session.Revision {
		t.Fatal("navigation advanced or replaced the retained run")
	}
}

func guidedTUIModel(operations service) (tuiGuidedExecutor, tui.Model) {
	executor := tuiGuidedExecutor{operations: operations}

	return executor, tui.NewModel(tui.Options{ProjectID: operations.projectContext.ProjectID, GuidedExecutor: executor})
}

func onlyGuidedTUIRun(t *testing.T, operations service) core.AuthoringSession {
	t.Helper()

	view, err := operations.InspectGuidedUseCase(t.Context(), "", "default")
	if err != nil || len(view.Runs) != 1 {
		t.Fatal("missing exact guided run", err)
	}

	return view.Runs[0]
}

func openGuidedTUI(t *testing.T, model *tui.Model) {
	t.Helper()

	runGuidedTUIKey(t, model, "f2", false)
	runGuidedTUIKey(t, model, "g", true)
}

func enterGuidedTUIInput(t *testing.T, model *tui.Model, text string) {
	t.Helper()

	runGuidedTUIKey(t, model, "e", false)
	runGuidedTUIKey(t, model, text, false)
	runGuidedTUIKey(t, model, "enter", true)
}

func runGuidedTUIKey(t *testing.T, model *tui.Model, key string, execute bool) {
	t.Helper()

	message := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	if kind, ok := map[string]tea.KeyType{"f2": tea.KeyF2, "enter": tea.KeyEnter, "esc": tea.KeyEsc, "pgdown": tea.KeyPgDown}[key]; ok {
		message = tea.KeyMsg{Type: kind}
	}

	updated, command := model.Update(message)

	*model = guidedTUIUpdatedModel(t, updated)
	if execute && command != nil {
		updated, _ = model.Update(command())
		*model = guidedTUIUpdatedModel(t, updated)
	}
}

func guidedTUIUpdatedModel(t *testing.T, model tea.Model) tui.Model {
	t.Helper()

	value, ok := model.(tui.Model)
	if !ok {
		t.Fatal("unexpected terminal model")
	}

	return value
}

func assertGuidedTUIOriginals(t *testing.T, operations service, session core.AuthoringSession) {
	t.Helper()

	replaceGuidedWebDocuments(t, operations.projectContext.ProjectDir, session)

	before := session.Revision

	for _, kind := range []string{"dump", "quiz", "specification"} {
		executor := tuiGuidedExecutor{operations: operations}

		view, err := executor.ExecuteGuided(tui.GuidedRequest{Action: "document", Document: kind, Command: core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}}})
		if err != nil || view.Document == nil {
			t.Fatal("committed original unavailable", kind, err)
		}

		original, err := operations.ReadGuidedUseCaseOriginal(t.Context(), session.ID, session.Namespace, kind)
		if err != nil || view.Document.Content != original.Content || strings.Contains(view.Document.Content, "Later replacement") || view.Session.Revision != before {
			t.Fatal("terminal original changed bytes or advanced the owner", kind, err)
		}
	}
}
