package authoring

import (
	"encoding/json"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// Native contract generation uses both provider transports' shared decoder and
// refuses changes to the retained hierarchy before allocation can occur.
func TestPlanDeliveryModel(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "retained hierarchy", true: "changed hierarchy"}[changed], func(t *testing.T) {
			checkPlanDeliveryModel(t, changed)
		})
	}
}

func checkPlanDeliveryModel(t *testing.T, changed bool) {
	t.Helper()

	plan := workflowPlanFixture()

	input, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}

	request := validCodexQuizModelRequest(quizmodel.ActionMaterializePlanTracking)
	request.Strategy = quizmodel.StrategyPlanDelivery
	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{Name: "planned_workflow", Content: string(input)})

	if changed {
		plan.Slices[0].Tasks[0].Title = "Unplanned work"
	}

	output, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}

	response, err := quizmodel.ResponseFromOutput(request, output, "openai")
	if changed {
		if err == nil {
			t.Fatal("changed hierarchy was accepted")
		}

		return
	}

	if err != nil || response.PlanTracking.Workflow == nil || quizmodel.ValidateResponse(response) != nil {
		t.Fatal(response, err)
	}

	if quizmodel.OutputSchema(request).Name != "plan_delivery" {
		t.Fatal("native contracts used the legacy tracking schema")
	}
}

// Independent plan operations cannot collide on model-local IDs such as
// slice-1. Retries preserve scoped IDs without mutating the model response.
func TestPlanDeliveryOrderScope(t *testing.T) {
	plan := orderedPlanCandidate()
	first := ScopePlanDeliveryOrder(plan.Order, generationcore.PlanPublicationOperation{ID: "first", Scope: plan.Scope, TicketDelivery: true})
	retry := ScopePlanDeliveryOrder(plan.Order, generationcore.PlanPublicationOperation{ID: "first", Scope: plan.Scope, TicketDelivery: true})
	second := ScopePlanDeliveryOrder(plan.Order, generationcore.PlanPublicationOperation{ID: "second", Scope: plan.Scope, TicketDelivery: true})

	if first[0].ID != retry[0].ID || first[0].ID == second[0].ID || plan.Order[0].ID != "slice-1" || first[0].Tasks[0].ID == second[0].Tasks[0].ID {
		t.Fatal("unstable or colliding scoped identities", first, retry, second)
	}

	if err := generationcore.ValidatePlannedOrder(first); err != nil {
		t.Fatal(err)
	}
}
