package app

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlanDeliveryResult completes authoring with tickets and their internal order
// map. It neither admits execution nor starts a worker.
type PlanDeliveryResult struct {
	Plan     PlanView
	Workflow core.WorkflowMaterializationResult
}

// GeneratePlanDelivery creates a plan and its tickets in one user operation.
// A saved plan survives a later failure; repeating the operation resumes its
// retained generation and native materialization instead of allocating twice.
func (s service) GeneratePlanDelivery(ctx context.Context, request PlanGenerateRequest) (PlanDeliveryResult, error) {
	s.workflowNamespace = request.Scope.Namespace
	if err := s.validatePlanDeliveryScope(ctx, request.Scope); err != nil {
		return PlanDeliveryResult{}, err
	}

	request.TicketDelivery = true

	generated, err := s.GeneratePlan(ctx, request)
	if err != nil {
		return PlanDeliveryResult{}, err
	}

	result := PlanDeliveryResult{Plan: generated.Plan}

	err = s.withGenerationRuntime(ctx, "plan_delivery", s.quizScope(request.Scope), func(ctx context.Context) error {
		var materializeErr error

		result.Workflow, materializeErr = s.materializeGeneratedPlan(ctx, request, generated)

		return materializeErr
	})
	if err != nil {
		return result, fmt.Errorf("plan saved; ticket creation is incomplete. Retry Generate Plan to finish: %w", err)
	}

	return result, nil
}

func (s service) validatePlanDeliveryScope(ctx context.Context, scope quizcore.Scope) error {
	if scope.CaseSlug == "" {
		return fmt.Errorf("select an existing use case with a saved spec")
	}

	if _, err := authoringcore.NewCaseMetadataService(s.projectContext.ProjectDir).ReadBounded(ctx, scope.CaseSlug); err != nil {
		return err
	}

	documents, err := s.ListSpecDocuments()
	if err != nil {
		return err
	}

	for _, document := range documents {
		if document.CaseSlug == scope.CaseSlug {
			return nil
		}
	}

	return fmt.Errorf("save a spec for this use case before generating its plan")
}

func (s service) materializeGeneratedPlan(ctx context.Context, request PlanGenerateRequest, generated PlanOperationResult) (core.WorkflowMaterializationResult, error) {
	observed, err := s.planDeliveryInput(ctx, request, generated)
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	input, layout := observed.Input, observed.Layout

	store := core.NewWorkflowStore(layout, s.projectContext.ProjectID)

	if retained, err := store.ResumePlanDelivery(ctx, input); retained.Found || err != nil {
		if err == nil {
			err = s.refreshWorkflowProjection(ctx, store)
		}

		return retained.Result, err
	}

	retained := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), input.PlanRef))
	planView := s.planViewFromPath(s.quizScope(request.Scope), retained, filepath.Join(s.projectContext.ProjectDir, filepath.FromSlash(retained)))

	plan, err := s.generatePlanTicketContracts(ctx, input, planView)
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	return s.MaterializeWorkflowPlan(ctx, plan)
}

type planDeliveryObservation struct {
	Input  generationcore.GeneratedWorkflowPlan
	Layout opscore.Layout
}

func (s service) planDeliveryInput(ctx context.Context, request PlanGenerateRequest, generated PlanOperationResult) (planDeliveryObservation, error) {
	scope := s.quizScope(request.Scope)

	input, err := authoringcore.ObservePlanWorkflow(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return planDeliveryObservation{Input: input}, err
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: scope.Namespace})
	if err != nil {
		return planDeliveryObservation{Input: input, Layout: layout}, err
	}

	canonical := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), input.PlanRevision.Reference))
	retained := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), input.PlanRef))

	if generated.Plan.RelativePath != canonical && generated.Plan.RelativePath != retained {
		return planDeliveryObservation{Input: input, Layout: layout}, workflowstate.ErrConflict
	}

	if generated.Existing {
		if err := s.checkDeliveryTarget(ctx, scope, request.ExpectedRevision); err != nil {
			return planDeliveryObservation{Input: input, Layout: layout}, err
		}
	}

	input.UseCase = scope.CaseSlug

	return planDeliveryObservation{Input: input, Layout: layout}, nil
}

func (s service) checkDeliveryTarget(ctx context.Context, scope quizcore.Scope, expected string) error {
	current, err := authoringcore.PlanTargetRevision(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return err
	}

	if expected != "" && expected != current {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s service) generatePlanTicketContracts(ctx context.Context, input generationcore.GeneratedWorkflowPlan, plan PlanView) (generationcore.GeneratedWorkflowPlan, error) {
	bound, ctx, err := s.bindGeneration(ctx, input.MaterializationID, string(quizmodel.ActionMaterializePlanTracking))
	if err != nil {
		return input, err
	}

	state, err := bound.specGenerationRuntimeState(plan.Scope)
	if err != nil {
		return input, err
	}

	request, err := bound.newPlanTrackingModelRequest(state, plan)
	if err != nil {
		return input, err
	}

	encoded, err := json.Marshal(input)
	if err != nil {
		return input, err
	}

	request.Strategy = quizmodel.StrategyPlanDelivery
	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{Name: "planned_workflow", Layer: "plan", Authority: input.PlanRef, Content: string(encoded)})

	response, err := quizmodel.Invoke(ctx, bound.quizModel, request)
	if err != nil {
		return input, err
	}

	if response.PlanTracking.Workflow == nil {
		return input, fmt.Errorf("the provider did not return native ticket contracts")
	}

	return *response.PlanTracking.Workflow, nil
}
