package app

import (
	"maps"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Browsing an earlier stage reads the canonical committed image, even after a
// working-copy replacement. Current-stage input is never submitted by navigation.
func TestTUIWizardReadOnly(t *testing.T) {
	operations, model := guidedTUIFixture(t)
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	runGuidedTUIKey(t, model, "w", false)
	runGuidedTUIKey(t, model, "Read only workflow", false)
	runGuidedTUIKey(t, model, "ctrl+s", true)
	enterGuidedTUIInput(t, model, "Preserve this committed capture.")

	session := onlyGuidedTUIRun(t, operations)

	original, err := operations.ReadGuidedUseCaseOriginal(t.Context(), session.ID, session.Namespace, "dump")
	if err != nil {
		t.Fatal(err)
	}

	replaceGuidedWebDocuments(t, operations.projectContext.ProjectDir, session)

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	checkTUIWizardTabOriginal(t, operations, session, original)
	model.Update(tea.KeyMsg{Type: tea.KeyF2})

	if !strings.Contains(model.View(), "Read-only stage") || !strings.Contains(model.View(), "Preserve this committed capture.") {
		t.Fatal(model.View())
	}

	for _, key := range []string{"e", "f", "s", "ctrl+s", "x", "c"} {
		runGuidedTUIKey(t, model, key, true)
	}

	after := onlyGuidedTUIRun(t, operations)
	if model.SelectedView() != tui.ViewQuizzes || after.Revision != session.Revision || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("read-only navigation wrote or resubmitted input")
	}
}

func checkTUIWizardTabOriginal(t *testing.T, operations service, session authoringcore.Session, original GuidedUseCaseArtifact) {
	t.Helper()

	executor := tuiGuidedExecutor{operations: operations}
	request := tui.GuidedRequest{Action: "inspect", Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}}}

	view, err := executor.ExecuteGuided(request)
	if err != nil || len(view.Originals) == 0 || view.Originals[0].Content != original.Content || view.Originals[0].Reference != original.Reference || view.Originals[0].Commit != original.Commit {
		t.Fatal("tab projection changed the original bytes or provenance", err)
	}
}
