package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// MaterializeWorkflowPlan explicitly opts approved input into ticket-owned
// workflow state. Ordinary legacy plan reads and defaults do not adopt it.
func (s service) MaterializeWorkflowPlan(ctx context.Context, plan generationcore.GeneratedWorkflowPlan) (core.WorkflowMaterializationResult, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	store := core.NewWorkflowStore(layout, s.projectContext.ProjectID)

	result, err := store.MaterializePlan(ctx, core.WorkflowMaterializationOptions{Plan: plan, ReserveTicketKey: func(ctx context.Context) (string, error) {
		operations := s
		operations.ctx = ctx
		number, err := operations.reserveNextTicketKeyNumber(config.TicketKeyPrefix)

		return fmt.Sprintf("%s-%d", config.TicketKeyPrefix, number), err
	}})
	if err != nil {
		return result, err
	}

	return result, s.refreshWorkflowProjection(ctx, store)
}

func (s service) refreshWorkflowProjection(ctx context.Context, store *core.WorkflowStore) (err error) {
	index, err := core.NewWorkflowTicketIndex(ctx, s.env, s.assetsFS)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := index.Close()
		if err == nil {
			err = closeErr
		}
	}()

	workflow, ok := index.(core.WorkflowIndex)
	if !ok {
		return fmt.Errorf("ticket index lacks workflow projection")
	}

	return workflow.RebuildWorkflowIndex(ctx, store)
}

// AdoptWorkflowPlan observes legacy runtime state and adopts only the explicit
// selections in the request. It never converts records while listing plans.
func (s service) AdoptWorkflowPlan(ctx context.Context, opts core.WorkflowPlanAdoptionOptions) (core.WorkflowMaterializationResult, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	store := core.NewWorkflowStore(layout, s.projectContext.ProjectID)
	opts.CheckQuiescent = s.checkWorkflowAdoptionExecutions

	result, err := store.AdoptPlanWorkflow(ctx, opts)
	if err != nil {
		return result, err
	}

	return result, s.refreshWorkflowProjection(ctx, store)
}

func (s service) checkWorkflowAdoptionExecutions(ctx context.Context, keys []string) (err error) {
	repository, err := executioncore.NewRepository(ctx, s.env, s.assetsFS)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, repository.Close()) }()

	for _, key := range keys {
		records, err := repository.ListTicketExecutions(ctx, s.projectContext.ProjectID, key)
		if err != nil {
			return err
		}

		if len(records) > workflowstate.RecordLimit {
			return fmt.Errorf("legacy execution observation bound exceeded")
		}

		for _, record := range records {
			if record.Status == executioncore.StatusRunning || record.Status == executioncore.StatusPending {
				return fmt.Errorf("legacy task %s has unfinished execution ownership", key)
			}
		}
	}

	return nil
}

func (s service) materializeApprovedWorkflow(ctx context.Context, request PlanTrackingRequest) (PlanTrackingResult, error) {
	scope := s.quizScope(request.Scope)
	if request.Workflow.UseCase != scope.CaseSlug {
		return PlanTrackingResult{}, fmt.Errorf("approved workflow use-case association differs from selected scope")
	}

	result, err := s.MaterializeWorkflowPlan(ctx, *request.Workflow)
	if err != nil {
		return PlanTrackingResult{}, err
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return PlanTrackingResult{}, err
	}

	relative := filepath.ToSlash(filepath.Join(layout.RelativeRoot(), result.Set.Set.PlanRef))

	return PlanTrackingResult{Message: fmt.Sprintf("Materialized %d task tickets in %d ordered slices.", len(result.Tasks), len(result.Set.Set.Slices)), Plan: s.planViewFromPath(scope, relative, filepath.Join(s.projectContext.ProjectDir, relative)), Workflow: &result}, nil
}
