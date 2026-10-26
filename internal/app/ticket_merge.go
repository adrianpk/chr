package app

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

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

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, index, err := factory.openTicketIndexForCommand(ctx, cmd, env)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, projectContext); err != nil {
		return err
	}

	ticket, tickets, err := ticketMergeCommandTickets(ctx, index, projectContext, cmd.Args().First())
	if err != nil {
		return err
	}

	config, expectedBase, err := ticketMergeCommandConfig(projectContext.ProjectDir, cmd)
	if err != nil {
		return err
	}

	release, err := acquireTicketMergeLock(ctx, projectContext, cmd.Bool("dry-run"))
	if err != nil {
		return err
	}
	defer release()

	result, err := core.RunTicketMergeCommand(ctx, core.TicketMergeCommandOptions{
		Ticket:        ticket,
		Tickets:       tickets,
		Remote:        core.NewRemoteReviewAdapter(config.Review.Providers, reviewRemoteCommandRunner),
		ExpectedBase:  expectedBase,
		MergeStrategy: cmd.String("strategy"),
		ProjectDir:    projectContext.ProjectDir,
		BranchPolicy:  config.Work.BranchPolicy,
		Workspace:     config.Work.Workspace,
		DryRun:        cmd.Bool("dry-run"),
		MergedAt:      time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if err := refreshTicketIndexAfterMerge(ctx, index, projectContext, cmd.Bool("dry-run")); err != nil {
		return err
	}

	return printTicketMergeResult(factory.io.Stdout, projectContext, result)
}

func ticketMergeCommandConfig(projectDir string, cmd *cli.Command) (core.WorkspaceConfig, string, error) {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return core.WorkspaceConfig{}, "", err
	}

	if err := core.ValidateWorkConfig(config.Work); err != nil {
		return core.WorkspaceConfig{}, "", err
	}

	return config, ticketMergeExpectedBase(cmd, config.Work.BranchPolicy), nil
}

func ticketMergeExpectedBase(cmd *cli.Command, policy core.WorkBranchPolicy) string {
	if base := strings.TrimSpace(cmd.String("base")); cmd.IsSet("base") && base != "" {
		return base
	}

	return policy.IntegrationBranch
}

func ticketMergeCommandTickets(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, key string) (core.IndexedTicket, []core.IndexedTicket, error) {
	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return core.IndexedTicket{}, nil, err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, key)
	if err != nil {
		return core.IndexedTicket{}, nil, err
	}

	return ticket, tickets, nil
}

func refreshTicketIndexAfterMerge(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, dryRun bool) error {
	if dryRun {
		return nil
	}

	return refreshProjectTicketIndex(ctx, index, projectContext)
}

func acquireTicketMergeLock(ctx context.Context, projectContext commandProjectContext, dryRun bool) (func(), error) {
	if dryRun {
		return noopRelease, nil
	}

	lock, err := acquireAgentOperationLock(ctx, projectContext, core.AgentOperationMerge)
	if err != nil {
		return noopRelease, err
	}

	return releaseAgentOperationLock(lock), nil
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
