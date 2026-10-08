package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// MaterializeWorkflowPlan explicitly opts approved input into ticket-owned
// workflow state. Ordinary legacy plan reads and defaults do not adopt it.
func (s service) MaterializeWorkflowPlan(ctx context.Context, plan core.GeneratedWorkflowPlan) (core.WorkflowMaterializationResult, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
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
	index, err := core.NewTicketIndex(ctx, s.env, s.assetsFS)
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
