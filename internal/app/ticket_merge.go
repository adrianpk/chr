package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) ticketMergeCommand() *cli.Command {
	return &cli.Command{
		Name:   "merge",
		Usage:  "Merge a reviewed ticket pull request",
		Flags:  ticketMergeFlags(),
		Action: factory.runTicketMerge,
	}
}

func ticketMergeFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "strategy",
			Usage: "merge strategy: merge, rebase, squash",
			Value: core.TicketMergeStrategyRebase,
		},
		&cli.StringFlag{
			Name:  "base",
			Usage: "expected pull request base branch; defaults to the configured integration branch",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print merge preflight without merging or writing ticket state",
		},
	)
}

func (factory commandFactory) runTicketMerge(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket merge requires exactly one ticket key argument")
	}

	operations, err := factory.projectService(ctx, cmd)
	project := operations.projectContext

	if err != nil {
		return err
	}

	result, err := operations.mergeTicket(ctx, ticketMergeInput{Key: cmd.Args().First(), Base: optionalStringFlag(cmd, "base"), Strategy: cmd.String("strategy"), DryRun: cmd.Bool("dry-run")})
	if err != nil {
		return err
	}

	return printTicketMergeResult(factory.io.Stdout, project, result)
}

func printTicketMergeResult(stdout io.Writer, projectContext commandProjectContext, result core.TicketMergeCommandResult) error {
	lines := []string{
		"ticket_merge: result",
		"project_id: " + projectContext.ProjectID,
		"ticket: " + result.TicketKey,
		fmt.Sprintf("allowed: %t", result.Preflight.Allowed),
		"status: " + result.Status,
		"merge_state: " + result.MergeState,
		"merge_result: " + result.MergeResult,
		"merge_strategy: " + result.MergeStrategy,
		"merge_commit: " + result.MergeCommit,
		"review_ref: " + result.Preflight.ReviewRef.URL,
		"provider: " + string(result.Preflight.Remote.Reference.Provider),
		fmt.Sprintf("dry_run: %t", result.DryRun),
	}

	for _, blocker := range result.Preflight.Blockers {
		lines = append(lines, "blocker: "+blocker.Code+" kind="+blocker.Kind+" detail="+blocker.Detail)
	}

	if result.WorktreeCleanup.Status != "" {
		lines = append(lines,
			"worktree_cleanup: "+result.WorktreeCleanup.Status,
			"worktree_cleanup_detail: "+result.WorktreeCleanup.Detail,
		)
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket merge result: %w", err)
		}
	}

	return nil
}
