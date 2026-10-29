package app

import (
	"maps"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// One visible Name/Create calls the real named owner. Replaying the same start
// and opening its case retain the same run without a second name operation.
func TestTUIWizardEntry(t *testing.T) {
	operations, model := guidedTUIFixture(t)
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	runGuidedTUIKey(t, model, "w", false)

	if !strings.Contains(model.View(), "Wizard") || !strings.Contains(model.View(), "Name") {
		t.Fatal(model.View())
	}

	runGuidedTUIKey(t, model, "Named terminal workflow", false)

	_, command := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if command == nil {
		t.Fatal("Create did not dispatch")
	}

	message := command()
	model.Update(message)

	session := onlyGuidedTUIRun(t, operations)
	if session.Guided.Stage != "dump" || session.Guided.CaseName != "Named terminal workflow" || session.Guided.CaseSlug == "" {
		t.Fatal("one Create did not bind named dump", session.Guided)
	}

	checkTUIWizardReentry(t, operations, session)
	model.Update(message)

	if after := onlyGuidedTUIRun(t, operations); after.ID != session.ID || after.Revision != session.Revision {
		t.Fatal("late result duplicated a run")
	}
}

// Validation runs before start and an unrelated case cannot open this run.
func TestTUIWizardEntryGuards(t *testing.T) {
	operations, _ := guidedTUIFixture(t)
	executor := tuiGuidedExecutor{operations: operations}

	request := tui.GuidedRequest{Action: "start", Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{ProjectID: operations.projectContext.ProjectID, RunID: "named-entry"}}}
	if _, err := executor.ExecuteGuided(request); err == nil {
		t.Fatal("empty name admitted a run")
	}

	request.Command.Content = "Guarded workflow"

	first, err := executor.ExecuteGuided(request)
	if err != nil {
		t.Fatal(err)
	}

	replay, err := executor.ExecuteGuided(request)
	if err != nil || replay.Session.ID != first.Session.ID || replay.Session.Revision != first.Session.Revision {
		t.Fatal("named start replay changed the run", err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	request.Action, request.CaseSlug = "open", "another-case"
	if _, err := executor.ExecuteGuided(request); err == nil || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("foreign case opened or changed a run", err)
	}
}

func checkTUIWizardReentry(t *testing.T, operations service, session authoringcore.Session) {
	t.Helper()

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	executor := tuiGuidedExecutor{operations: operations}

	view, err := executor.ExecuteGuided(tui.GuidedRequest{Action: "open", CaseSlug: session.Guided.CaseSlug, Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{ProjectID: session.ProjectID, RunID: session.ID}}})
	if err != nil || view.Session == nil || view.Session.ID != session.ID || view.Session.Revision != session.Revision || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("case re-entry changed retained state", err)
	}
}
