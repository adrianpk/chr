package app

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func runTicketStartAuthority(ctx context.Context, cmd *cli.Command, stdout io.Writer) (bool, error) {
	deps, err := appDependenciesFromCommand(cmd)
	if err != nil {
		return false, err
	}

	operations, project, err := workServiceFromCommand(ctx, cmd, deps)
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

		return false, printAgentWorkLines(stdout, []string{"authority: tracker_compat", "scope: unadopted_legacy"})
	}

	if cmd.Bool("dry-run") {
		return true, printNativeTicketStartPreview(operations, config, cmd.Args().First(), stdout)
	}

	result, err := operations.StartAgentWorkTicket(cmd.Args().First())
	if err != nil {
		return true, err
	}

	return true, printAgentWorkTicket(stdout, result)
}

func printNativeTicketStartPreview(operations service, config core.WorkspaceConfig, key string, stdout io.Writer) error {
	result, err := operations.previewNativeAgentWorkSelection(config, core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: key}})
	if err != nil {
		return err
	}

	if err := printAgentWorkLines(stdout, []string{"dry_run: true"}); err != nil {
		return err
	}

	return printAgentWorkTicket(stdout, result)
}
