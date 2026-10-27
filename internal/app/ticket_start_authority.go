package app

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
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

	if core.WorkAuthority(config.Work) == core.WorkStateAuthorityTrackerCompat {
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

func printNativeTicketStartPreview(operations service, config core.WorkspaceConfig, key string, stdout io.Writer) error {
	result, err := operations.previewNativeAgentWorkSelection(config, workflowstate.WorkflowPreviewRequest{Selector: workflowstate.WorkflowSelector{Extent: "ticket", Identity: key}})
	if err != nil {
		return err
	}

	if err := printAgentWorkLines(stdout, []string{"dry_run: true"}); err != nil {
		return err
	}

	return printAgentWorkTicket(stdout, result)
}
