package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

type recordedTUIPlanExecutor struct {
	delegate tuiPlanActionExecutor
	requests []tui.PlanActionRequest
}

func (e *recordedTUIPlanExecutor) ExecutePlanAction(request tui.PlanActionRequest) (tui.PlanActionResult, error) {
	e.requests = append(e.requests, request)

	return e.delegate.ExecutePlanAction(request)
}

// A real saved plan survives failed ticket-contract generation. The visible
// retry must submit the original identity, then replay without another ticket.
func TestTUIPlanRetry(t *testing.T) {
	provider, generator := webPlanDeliveryFixture(t)
	operations := provider.service(t.Context())
	executor := &recordedTUIPlanExecutor{delegate: newTUIPlanActionExecutor(operations)}
	model := tuiPlanModel(t, operations, executor)
	generator.fail = true

	runGuidedTUIKey(t, model, "g", true)

	if !strings.Contains(model.View(), "ticket creation is incomplete") || !strings.Contains(model.View(), "y Retry the original operation") {
		t.Fatal("retained preparation has no original retry", model.View())
	}

	generator.fail = false

	runGuidedTUIKey(t, model, "y", true)

	if len(executor.requests) != 2 || executor.requests[0] != executor.requests[1] {
		t.Fatal("retry changed captured intent", executor.requests)
	}

	checkWebPlanDelivery(t, provider)

	replay, err := executor.ExecutePlanAction(executor.requests[0])
	if err != nil || len(replay.Tickets) != 1 || len(generator.base.requests) != 1 || generator.calls != 2 {
		t.Fatal("retry duplicated retained work", replay, err, generator.calls, len(generator.base.requests))
	}

	checkReopenedTUIPlan(t, provider, generator, executor)
}

func checkReopenedTUIPlan(t *testing.T, provider webViewProvider, generator *planDeliveryTestModel, executor *recordedTUIPlanExecutor) {
	t.Helper()

	operations := provider.service(t.Context())

	reopened := tuiPlanModel(t, operations, executor)
	runGuidedTUIKey(t, reopened, "g", true)

	if len(executor.requests) != 4 || executor.requests[3].OperationID == executor.requests[0].OperationID || generator.calls != 2 || len(generator.base.requests) != 1 {
		t.Fatal("reopened preparation failed to reconcile its retained plan", executor.requests, reopened.View())
	}

	checkWebPlanDelivery(t, provider)
}

func tuiPlanModel(t *testing.T, operations service, executor tui.PlanActionExecutor) *tui.Model {
	t.Helper()

	plans, err := loadTUIPlans(operations)
	if err != nil {
		t.Fatal(err)
	}

	cases, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	model := tui.NewModel(tui.Options{Plans: plans, Cases: cases, PlanExecutor: executor})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	model.Update(tea.KeyMsg{Type: tea.KeyF5})

	return model
}
