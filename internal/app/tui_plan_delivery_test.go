package app

import (
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

// The production terminal adapter must produce the same native plan and ticket
// records as Web, without admitting execution or invoking a second UI action.
func TestTUIPlanDelivery(t *testing.T) {
	provider, generator := webPlanDeliveryFixture(t)
	operations := provider.service(t.Context())
	executor := newTUIPlanActionExecutor(operations)

	result, err := executor.ExecutePlanAction(tuiPlanDeliveryRequest(t, operations, "one-terminal-plan"))
	if err != nil {
		t.Fatal(err)
	}

	checkWebPlanDelivery(t, provider)

	if len(result.Tickets) != 1 || generator.calls != 1 || !strings.Contains(result.Message, "Work has not started") {
		t.Fatal("terminal action did not complete native preparation", result, generator.calls)
	}
}

func tuiPlanDeliveryRequest(t *testing.T, operations service, id string) tui.PlanActionRequest {
	t.Helper()

	revision, err := authoringcore.PlanTargetRevision(t.Context(), operations.projectContext.ProjectDir, operations.quizScope(quizcore.Scope{CaseSlug: "named-work"}))
	if err != nil {
		t.Fatal(err)
	}

	return tui.PlanActionRequest{Action: tui.PlanActionGenerate, Scope: tui.QuizScope{CaseSlug: "named-work"}, OperationID: id, Revision: revision}
}
