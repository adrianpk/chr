package app

import (
	"maps"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Specs and Plans reuse the manual document renderer over exact committed
// originals. Working-copy replacement and tab navigation cannot change inputs.
func TestTUIWizardDocuments(t *testing.T) {
	operations, session, _ := guidedPreparationFixture(t, "default")

	prepared, err := operations.PrepareGuidedWork(t.Context(), guidedQuizCommand(session, "prepare", "prepare", ""))
	if err != nil {
		t.Fatal(err)
	}

	replaceGuidedWebDocuments(t, operations.projectContext.ProjectDir, prepared)

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	_, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)

	for _, kind := range []string{"specification", "plan", "tracker"} {
		assertTUIWizardDocument(t, operations, model, prepared, kind)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("document tabs changed canonical inputs")
	}

	if prepared.Guided.SpecApproval == nil || len(prepared.Guided.Checkpoints) != 4 {
		t.Fatal("owner approval or preparation changed")
	}

	assertOneGuidedMaterialization(t, operations)
}

func assertTUIWizardDocument(t *testing.T, operations service, model *tui.Model, prepared authoringcore.Session, kind string) {
	t.Helper()

	executor := tuiGuidedExecutor{operations: operations}

	original, readErr := operations.ReadGuidedUseCaseOriginal(t.Context(), prepared.ID, prepared.Namespace, kind)
	if readErr != nil {
		t.Fatal(readErr)
	}

	request := tui.GuidedRequest{Action: "document", Document: kind, Command: guidedQuizCommand(prepared, "", "", "")}

	view, readErr := executor.ExecuteGuided(request)
	if readErr != nil || view.Document == nil || view.Document.Content != original.Content || view.Document.Commit != original.Commit || view.Document.Reference != original.Reference {
		t.Fatal("original bytes or binding changed", kind, readErr)
	}

	key := map[string]string{"specification": "3", "plan": "4", "tracker": "5"}[kind]

	runGuidedTUIKey(t, model, "f5", false)
	runGuidedTUIKey(t, model, key, true)

	if !strings.Contains(model.View(), "Committed original (read-only)") || strings.Contains(model.View(), "Later replacement") {
		t.Fatal("shared document component missing", kind, model.View())
	}

	if !strings.Contains(model.View(), original.Reference) {
		t.Fatal("document tab selected another original", kind, model.View())
	}
}
