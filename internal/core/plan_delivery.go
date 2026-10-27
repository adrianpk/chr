package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// QuizModelStrategyPlanDelivery completes native ticket contracts for a retained
// plan. The application supplies identities; the model only supplies contracts.
const QuizModelStrategyPlanDelivery = "plan_delivery"

// ScopePlanDeliveryOrder gives each generated revision namespace-unique native
// identities while retaining model-supplied order, titles and human labels.
func ScopePlanDeliveryOrder(slices []PlanOrderSlice, operation PlanPublicationOperation) []PlanOrderSlice {
	if !operation.TicketDelivery {
		return slices
	}

	identity := operation.ID + "\x00" + operation.Scope.ProjectID + "\x00" + operation.Scope.CaseSlug + "\x00" + operation.ExpectedRevision
	prefix := "p" + contenthash.Sum([]byte(identity))[:16] + "-"
	result := orderedPlanSlices(slices)

	for i := range result {
		result[i].ID = prefix + fmt.Sprintf("slice-%d", result[i].Order)
		for j := range result[i].Tasks {
			result[i].Tasks[j].ID = prefix + fmt.Sprintf("task-%d-%d", result[i].Order, result[i].Tasks[j].Order)
		}
	}

	return result
}

// ObservePlanWorkflow reads the immutable planned input even after tickets have
// been bound, so a retry uses the same materialization identity and contracts.
func ObservePlanWorkflow(ctx context.Context, projectDir string, scope QuizScope) (GeneratedWorkflowPlan, error) {
	binding, err := ObservePlanRevision(ctx, projectDir, scope)
	if err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	if binding.Reference == "" {
		return GeneratedWorkflowPlan{}, fmt.Errorf("generate a revision-bound plan before creating its tickets")
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir, Namespace: scope.Namespace})
	if err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	r := &fileAuthoringSessionRepository{projectDir: projectDir}
	root := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), filepath.Dir(binding.OrderRef)))

	data, err := r.authoringRead(root + "/planned-binding.json")
	if err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	var planned PlanRevisionBinding
	if err := decodeAuthoringState(data, &planned); err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	if !samePlannedWorkflowBinding(planned, binding) || !validPlanBindingPaths(layout.Namespace(), planned) {
		return GeneratedWorkflowPlan{}, workflowstate.ErrConflict
	}

	return r.plannedWorkflowInput(ctx, layout, planned)
}

func samePlannedWorkflowBinding(planned, current PlanRevisionBinding) bool {
	return planned.Reference == current.Reference && planned.Revision == current.Revision && planned.BodyHash == current.BodyHash && planned.PlannedOrderHash == current.PlannedOrderHash
}

func (r *fileAuthoringSessionRepository) plannedWorkflowInput(ctx context.Context, layout OpsLayout, binding PlanRevisionBinding) (GeneratedWorkflowPlan, error) {
	reference := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), binding.Reference))

	document, err := ReadPlanDocument(ctx, r.projectDir, reference)
	if err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	data, err := r.authoringRead(filepath.ToSlash(filepath.Join(layout.RelativeRoot(), filepath.Dir(binding.OrderRef), "planned-order.json")))
	if err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	if contenthash.Sum(data) != binding.PlannedOrderHash {
		return GeneratedWorkflowPlan{}, ErrWorkflowProjectionStale
	}

	var order PlanOrderDocument
	if err := decodeAuthoringState(data, &order); err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	if err := ValidatePlannedOrder(order.Slices); err != nil {
		return GeneratedWorkflowPlan{}, err
	}

	plan := GeneratedWorkflowPlan{MaterializationID: "delivery-" + binding.PlannedOrderHash, Title: document.Title, PlanRef: binding.BodyRef, PlanHash: binding.BodyHash, PlanRevision: &binding}
	for _, slice := range orderedPlanSlices(order.Slices) {
		item := GeneratedWorkflowSlice{ID: slice.ID, Order: slice.Order, Title: slice.Title, Label: fmt.Sprintf("Slice %d", slice.Order)}
		for _, task := range slice.Tasks {
			item.Tasks = append(item.Tasks, GeneratedWorkflowTask{ID: task.ID, Order: task.Order, Title: task.Title, Label: fmt.Sprintf("T%d.%d", slice.Order, task.Order)})
		}

		plan.Slices = append(plan.Slices, item)
	}

	return plan, nil
}

func quizModelPlanDeliveryResponse(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	var contracts GeneratedWorkflowPlan

	if len(output) > WorkflowFileLimit {
		return QuizModelResponse{}, modelcore.ErrOutput
	}

	if err := decodeAuthoringState(extractJSONObject(output), &contracts); err != nil {
		return QuizModelResponse{}, err
	}

	plan, err := planDeliveryRequestInput(request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	if !reflect.DeepEqual(orderedPlanSlices(plannedWorkflowOrder(plan)), orderedPlanSlices(plannedWorkflowOrder(contracts))) {
		return QuizModelResponse{}, fmt.Errorf("ticket contracts changed the plan's declared slices or tasks")
	}

	plan, contracts = orderedWorkflowPlan(plan), orderedWorkflowPlan(contracts)
	for i := range contracts.Slices {
		contracts.Slices[i].Label = plan.Slices[i].Label
	}

	plan.Slices, plan.ClosingValidation = contracts.Slices, contracts.ClosingValidation
	if err := ValidateGeneratedWorkflowPlan(plan); err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{ID: quizModelResponseID(request), RequestID: request.ID, Provider: QuizModelProviderOpenAI, Status: QuizModelResponseReady, PlanTracking: GeneratedPlanTracking{Workflow: &plan}, RawText: string(output), CreatedAt: request.CreatedAt.UTC()}, nil
}

func planDeliveryRequestInput(request QuizModelRequest) (GeneratedWorkflowPlan, error) {
	for _, section := range request.ContextSections {
		if section.Name == "planned_workflow" {
			var plan GeneratedWorkflowPlan

			err := json.Unmarshal([]byte(section.Content), &plan)

			return plan, err
		}
	}

	return GeneratedWorkflowPlan{}, fmt.Errorf("retained workflow input is required")
}
