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

// QuizModelStrategyPlanDelivery completes native ticket contracts for a retained
// plan. The application supplies identities; the model only supplies contracts.
const QuizModelStrategyPlanDelivery = "plan_delivery"

func quizModelPlanDeliveryResponse(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	var contracts generationcore.GeneratedWorkflowPlan

	if len(output) > workflowstate.FileLimit {
		return QuizModelResponse{}, modelcore.ErrOutput
	}

	if err := authoringstate.Decode(modelcore.ExtractJSONObject(output), &contracts); err != nil {
		return QuizModelResponse{}, err
	}

	plan, err := planDeliveryRequestInput(request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	if !reflect.DeepEqual(generationcore.OrderedPlanSlices(generationcore.PlannedWorkflowOrder(plan)), generationcore.OrderedPlanSlices(generationcore.PlannedWorkflowOrder(contracts))) {
		return QuizModelResponse{}, fmt.Errorf("ticket contracts changed the plan's declared slices or tasks")
	}

	plan, contracts = generationcore.OrderedWorkflowPlan(plan), generationcore.OrderedWorkflowPlan(contracts)
	for i := range contracts.Slices {
		contracts.Slices[i].Label = plan.Slices[i].Label
	}

	plan.Slices, plan.ClosingValidation = contracts.Slices, contracts.ClosingValidation
	if err := generationcore.ValidateGeneratedWorkflowPlan(plan); err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{ID: quizModelResponseID(request), RequestID: request.ID, Provider: QuizModelProviderOpenAI, Status: QuizModelResponseReady, PlanTracking: generationcore.GeneratedPlanTracking{Workflow: &plan}, RawText: string(output), CreatedAt: request.CreatedAt.UTC()}, nil
}

func planDeliveryRequestInput(request QuizModelRequest) (generationcore.GeneratedWorkflowPlan, error) {
	for _, section := range request.ContextSections {
		if section.Name == "planned_workflow" {
			var plan generationcore.GeneratedWorkflowPlan

			err := json.Unmarshal([]byte(section.Content), &plan)

			return plan, err
		}
	}

	return generationcore.GeneratedWorkflowPlan{}, fmt.Errorf("retained workflow input is required")
}
