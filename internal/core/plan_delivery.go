package core

import (
	"context"

	"fmt"
	"path/filepath"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ScopePlanDeliveryOrder gives each generated revision namespace-unique native
// identities while retaining model-supplied order, titles and human labels.
func ScopePlanDeliveryOrder(slices []generationcore.PlanOrderSlice, operation generationcore.PlanPublicationOperation) []generationcore.PlanOrderSlice {
	if !operation.TicketDelivery {
		return slices
	}

	identity := operation.ID + "\x00" + operation.Scope.ProjectID + "\x00" + operation.Scope.CaseSlug + "\x00" + operation.ExpectedRevision
	prefix := "p" + contenthash.Sum([]byte(identity))[:16] + "-"
	result := generationcore.OrderedPlanSlices(slices)

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
func ObservePlanWorkflow(ctx context.Context, projectDir string, scope quizcore.Scope) (generationcore.GeneratedWorkflowPlan, error) {
	binding, err := ObservePlanRevision(ctx, projectDir, scope)
	if err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	if binding.Reference == "" {
		return generationcore.GeneratedWorkflowPlan{}, fmt.Errorf("generate a revision-bound plan before creating its tickets")
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir, Namespace: scope.Namespace})
	if err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	r := &fileAuthoringSessionRepository{projectDir: projectDir}
	root := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), filepath.Dir(binding.OrderRef)))

	data, err := r.authoringRead(root + "/planned-binding.json")
	if err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	var planned generationcore.PlanRevisionBinding
	if err := authoringstate.Decode(data, &planned); err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	if !samePlannedWorkflowBinding(planned, binding) || !validPlanBindingPaths(layout.Namespace(), planned) {
		return generationcore.GeneratedWorkflowPlan{}, workflowstate.ErrConflict
	}

	return r.plannedWorkflowInput(ctx, layout, planned)
}

func samePlannedWorkflowBinding(planned, current generationcore.PlanRevisionBinding) bool {
	return planned.Reference == current.Reference && planned.Revision == current.Revision && planned.BodyHash == current.BodyHash && planned.PlannedOrderHash == current.PlannedOrderHash
}

func (r *fileAuthoringSessionRepository) plannedWorkflowInput(ctx context.Context, layout OpsLayout, binding generationcore.PlanRevisionBinding) (generationcore.GeneratedWorkflowPlan, error) {
	reference := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), binding.Reference))

	document, err := ReadPlanDocument(ctx, r.projectDir, reference)
	if err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	data, err := r.authoringRead(filepath.ToSlash(filepath.Join(layout.RelativeRoot(), filepath.Dir(binding.OrderRef), "planned-order.json")))
	if err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	if contenthash.Sum(data) != binding.PlannedOrderHash {
		return generationcore.GeneratedWorkflowPlan{}, ErrWorkflowProjectionStale
	}

	var order generationcore.PlanOrderDocument
	if err := authoringstate.Decode(data, &order); err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	if err := generationcore.ValidatePlannedOrder(order.Slices); err != nil {
		return generationcore.GeneratedWorkflowPlan{}, err
	}

	plan := generationcore.GeneratedWorkflowPlan{MaterializationID: "delivery-" + binding.PlannedOrderHash, Title: document.Title, PlanRef: binding.BodyRef, PlanHash: binding.BodyHash, PlanRevision: &binding}
	for _, slice := range generationcore.OrderedPlanSlices(order.Slices) {
		item := generationcore.GeneratedWorkflowSlice{ID: slice.ID, Order: slice.Order, Title: slice.Title, Label: fmt.Sprintf("Slice %d", slice.Order)}
		for _, task := range slice.Tasks {
			item.Tasks = append(item.Tasks, generationcore.GeneratedWorkflowTask{ID: task.ID, Order: task.Order, Title: task.Title, Label: fmt.Sprintf("T%d.%d", slice.Order, task.Order)})
		}

		plan.Slices = append(plan.Slices, item)
	}

	return plan, nil
}
