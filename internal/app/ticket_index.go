package app

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

func (factory commandFactory) openTicketIndexForCommand(ctx context.Context, cmd *cli.Command, env core.UserEnvironment) (commandTicketIndex, error) {
	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return commandTicketIndex{Project: commandProjectContext{}, Index: nil}, err
	}

	index, err := workflowcore.NewWorkflowTicketIndex(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return commandTicketIndex{Project: commandProjectContext{}, Index: nil}, err
	}

	return commandTicketIndex{Project: projectContext, Index: index}, nil
}

func refreshProjectTicketIndex(ctx context.Context, index ticketcore.Index, projectContext commandProjectContext) error {
	result, err := ticketcore.NewReconciler(index).ReconcileProject(ctx, ticketcore.ReconcileProjectOptions{
		ProjectID:    projectContext.ProjectID,
		ProjectDir:   projectContext.ProjectDir,
		ReconciledAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if result.HasBlockers() {
		firstBlocker, _ := result.FirstBlocker()

		return fmt.Errorf("refresh ticket index: %d blockers: %s %w", len(result.Blockers), firstBlocker.RelativePath, firstBlocker.Err)
	}

	return nil
}

type commandTicketIndex struct {
	Project commandProjectContext
	Index   ticketcore.Index
}
