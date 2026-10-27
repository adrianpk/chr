package app

import (
	"context"
	"errors"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// UpgradeWorkflowConfiguration owns legacy runtime observation for the explicit
// project upgrade. A supplied adoption map is checked by the same operation owner.
func (s service) UpgradeWorkflowConfiguration(ctx context.Context, opts core.WorkflowUpgradeOptions) (core.WorkflowUpgradeResult, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return core.WorkflowUpgradeResult{}, err
	}

	opts.CheckQuiescent = s.checkWorkflowUpgradeExecutions
	if opts.Adoption != nil {
		selection := *opts.Adoption
		selection.CheckQuiescent = func(ctx context.Context, _ []string) error { return s.checkWorkflowUpgradeExecutions(ctx) }
		opts.Adoption = &selection
	}

	store := core.NewWorkflowStore(layout, s.projectContext.ProjectID)

	result, err := store.UpgradeWorkflowConfiguration(ctx, opts)
	if err != nil || result.Adopted == nil {
		return result, err
	}

	if err := s.refreshWorkflowProjection(ctx, store); err != nil {
		return result, fmt.Errorf("workflow upgraded; derived ticket index is stale: %w", err)
	}

	return result, nil
}

func (s service) checkWorkflowUpgradeExecutions(ctx context.Context) (err error) {
	repository, err := core.NewExecutionRepository(ctx, s.env, s.assetsFS)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, repository.Close()) }()

	return repository.RequireProjectQuiescent(ctx, s.projectContext.ProjectID)
}
