package model

import (
	"encoding/json"
	"fmt"
	"reflect"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// StrategyPlanDelivery completes native ticket contracts for a retained
// plan. The application supplies identities; the model only supplies contracts.
const StrategyPlanDelivery = "plan_delivery"

func planDeliveryResponse(request Request, output []byte) (Response, error) {
	var contracts generationcore.GeneratedWorkflowPlan

	if len(output) > workflowstate.FileLimit {
		return Response{}, modelcore.ErrOutput
	}

	if err := authoringstate.Decode(modelcore.ExtractJSONObject(output), &contracts); err != nil {
		return Response{}, err
	}

	plan, err := planDeliveryRequestInput(request)
	if err != nil {
		return Response{}, err
	}

	if !reflect.DeepEqual(generationcore.OrderedPlanSlices(generationcore.PlannedWorkflowOrder(plan)), generationcore.OrderedPlanSlices(generationcore.PlannedWorkflowOrder(contracts))) {
		return Response{}, fmt.Errorf("ticket contracts changed the plan's declared slices or tasks")
	}

	plan, contracts = generationcore.OrderedWorkflowPlan(plan), generationcore.OrderedWorkflowPlan(contracts)
	for i := range contracts.Slices {
		contracts.Slices[i].Label = plan.Slices[i].Label
	}

	plan.Slices, plan.ClosingValidation = contracts.Slices, contracts.ClosingValidation
	if err := generationcore.ValidateGeneratedWorkflowPlan(plan); err != nil {
		return Response{}, err
	}

	return Response{ID: responseID(request), RequestID: request.ID, Provider: ProviderOpenAI, Status: ResponseReady, PlanTracking: generationcore.GeneratedPlanTracking{Workflow: &plan}, RawText: string(output), CreatedAt: request.CreatedAt.UTC()}, nil
}

func planDeliveryRequestInput(request Request) (generationcore.GeneratedWorkflowPlan, error) {
	for _, section := range request.ContextSections {
		if section.Name == "planned_workflow" {
			var plan generationcore.GeneratedWorkflowPlan

			err := json.Unmarshal([]byte(section.Content), &plan)

			return plan, err
		}
	}

	return generationcore.GeneratedWorkflowPlan{}, fmt.Errorf("retained workflow input is required")
}
