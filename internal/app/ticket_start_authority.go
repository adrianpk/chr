package app

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (factory commandFactory) runTicketStartAuthority(ctx context.Context, cmd *cli.Command) (bool, error) {
	operations, err := factory.projectService(ctx, cmd)
	project := operations.projectContext

	if err != nil {
		return false, err
	}

	config, err := agentWorkLoopWorkspaceConfig(project.ProjectDir)
	if err != nil {
		return false, err
	}

	if workspaceconfig.WorkAuthority(config.Work) == workspaceconfig.WorkStateAuthorityTrackerCompat {
		if err := operations.checkLegacyAgentWorkKey(cmd.Args().First()); err != nil {
			return false, err
		}

		return false, printAgentWorkLines(factory.io.Stdout, []string{"authority: tracker_compat", "scope: unadopted_legacy"})
	}

	if cmd.Bool("dry-run") {
		return true, printNativeTicketStartPreview(operations, config, cmd.Args().First(), factory.io.Stdout)
	}

	result, err := operations.StartAgentWorkTicket(cmd.Args().First())
	if err != nil {
		return true, err
	}

	return true, printAgentWorkTicket(factory.io.Stdout, result)
}

func printNativeTicketStartPreview(operations service, config workspaceconfig.Config, key string, stdout io.Writer) error {
	result, err := operations.previewNativeAgentWorkSelection(config, workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: key}})
	if err != nil {
		return err
	}

	if err := printAgentWorkLines(stdout, []string{"dry_run: true"}); err != nil {
		return err
	}

	return printAgentWorkTicket(stdout, result)
}
