package app

import (
	"maps"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Terminal results distinguish a validated worker and open PR from confirmed
// integration, then display exact native delivery, task and closing evidence.
func TestTUIWizardResults(t *testing.T) {
	operations, session, host, provider := guidedResultFixture(t)
	_, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)

	if strings.Contains(model.View(), "All admitted work has confirmed") || !strings.Contains(model.View(), "Delivery is unconfirmed") || !strings.Contains(model.View(), "Review PR:") {
		t.Fatal("open PR claimed delivery or hid review", model.View())
	}

	withIdea, ideaErr := operations.ChangeGuidedUseCase(t.Context(), guidedQuizCommand(session, "retained-result-idea", "idea", "Keep a separate result export."))
	if ideaErr != nil {
		t.Fatal(ideaErr)
	}

	completed := finishGuidedResultFixture(t, operations, withIdea, host, provider)
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	runGuidedTUIKey(t, model, "r", true)

	view, err := (tuiGuidedExecutor{operations: operations}).ExecuteGuided(tui.GuidedRequest{Action: "inspect", Command: guidedQuizCommand(completed, "", "", "")})
	if err != nil || view.Result == nil || !view.Result.Complete || len(view.Result.Deliveries) != 1 {
		t.Fatal("terminal result omitted verified delivery", err)
	}

	assertTUIWizardResultView(t, model, view)
	assertTUIWizardResultExport(t, operations, model, completed)

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("result inspection or export wrote owner state")
	}
}

func assertTUIWizardResultView(t *testing.T, model *tui.Model, view tui.GuidedView) {
	t.Helper()

	delivery := view.Result.Deliveries[0]
	for _, marker := range []string{"All admitted work has confirmed", "Confirmed delivery", "Branch: dev", delivery.Commit, delivery.Reference, delivery.Hash, "Set closing:", "Validation:", "exit 0"} {
		if !strings.Contains(model.View(), marker) {
			t.Fatal("terminal results omitted native evidence", marker, model.View())
		}
	}

	if delivery.Repository == "" || delivery.PullRequest == "" || len(view.Result.Validation) < 2 {
		t.Fatal("incomplete structured result")
	}
}

func assertTUIWizardResultExport(t *testing.T, operations service, model *tui.Model, completed authoringcore.Session) {
	t.Helper()
	runGuidedTUIKey(t, model, "v", true)

	if !strings.Contains(model.View(), `"Complete": true`) {
		t.Fatal("JSON export unavailable", model.View())
	}

	runGuidedTUIKey(t, model, "v", false)

	if !strings.Contains(model.View(), "Confirmed delivery") {
		t.Fatal("JSON replaced the shared Board")
	}

	assertGuidedTUIExport(t, tuiGuidedExecutor{operations: operations}, completed, true)
}
