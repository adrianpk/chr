package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) ticketReviewModeCommand() *cli.Command {
	return &cli.Command{
		Name:  "review-mode",
		Usage: "Inspect ticket review workflow mode",
		Commands: []*cli.Command{
			factory.ticketReviewModeDiagnosticsCommand(),
			factory.ticketReviewModeStatusCommand(),
		},
	}
}

func (factory commandFactory) ticketReviewModeDiagnosticsCommand() *cli.Command {
	return &cli.Command{
		Name:   "diagnostics",
		Usage:  "Verify ticket review provider access",
		Flags:  ticketReviewModeDiagnosticsFlags(),
		Action: factory.runTicketReviewModeDiagnostics,
	}
}

func (factory commandFactory) ticketReviewModeStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Print active ticket review workflow mode",
		Flags:  ticketReviewModeStatusFlags(),
		Action: factory.runTicketReviewModeStatus,
	}
}

func ticketReviewModeStatusFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print the active mode without performing review work",
		},
	)
}

func ticketReviewModeDiagnosticsFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "url",
			Usage: "pull request URL to inspect",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "skip remote pull request readability checks",
		},
	)
}

func (factory commandFactory) runTicketReviewModeStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket review-mode status does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	return printTicketReviewModeStatus(factory.io.Stdout, projectContext, config.Review, cmd.Bool("dry-run"))
}

func (factory commandFactory) runTicketReviewModeDiagnostics(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket review-mode diagnostics does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	diagnostic := core.RunReviewProviderDiagnostics(ctx, core.ReviewProviderDiagnosticOptions{
		Config: config.Review.Providers,
		URL:    cmd.String("url"),
		DryRun: cmd.Bool("dry-run"),
		Runner: reviewProviderDiagnosticRunner,
	})

	return printTicketReviewModeDiagnostics(factory.io.Stdout, projectContext, diagnostic, cmd.Bool("dry-run"))
}

func printTicketReviewModeDiagnostics(stdout io.Writer, projectContext commandProjectContext, diagnostic core.ReviewProviderDiagnostic, dryRun bool) error {
	reference := diagnostic.Reference
	lines := []string{
		"review_provider_diagnostics: status",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		"url: " + reference.URL,
		"provider: " + string(diagnostic.Provider),
		"provider_status: " + diagnostic.Status,
		"host: " + reference.Host,
		"owner: " + reference.Owner,
		"repo: " + reference.Repo,
		fmt.Sprintf("number: %d", reference.Number),
		fmt.Sprintf("dry_run: %t", dryRun),
	}

	for _, check := range diagnostic.Checks {
		line := fmt.Sprintf("check.%s: %s", check.Name, check.Status)
		if check.Reason != "" {
			line += " reason=" + check.Reason
		}

		if check.Detail != "" {
			line += " detail=" + check.Detail
		}

		lines = append(lines, line)
	}

	return printTicketReviewModeStatusLines(stdout, lines)
}

func printTicketReviewModeStatus(stdout io.Writer, projectContext commandProjectContext, config core.ReviewWorkflowConfig, dryRun bool) error {
	status := "ready"

	err := core.ValidateReviewWorkflowConfig(config)
	if err != nil {
		status = "invalid"
	}

	lines := []string{
		"review_mode: status",
		"project_id: " + projectContext.ProjectID,
		"project_dir: " + projectContext.ProjectDir,
		"mode: " + string(config.Mode),
		"mode_status: " + status,
		"description: " + core.ReviewWorkflowModeDescription(config.Mode),
		fmt.Sprintf("dry_run: %t", dryRun),
		fmt.Sprintf("requires_remote_review: %t", config.Mode == core.ReviewWorkflowModeAgenticReview),
		fmt.Sprintf("requires_human_review: %t", config.Mode == core.ReviewWorkflowModeHumanReview),
		fmt.Sprintf("trusted_auto_merge: %t", config.Mode == core.ReviewWorkflowModeTrustedAutoMerge),
		"clean_pass_remote_comment: false",
		fmt.Sprintf("non_blocking_comment_limit: %d", config.NonBlockingCommentLimit),
	}

	if err != nil {
		lines = append(lines, "error: "+err.Error())
	}

	return printTicketReviewModeStatusLines(stdout, lines)
}

func printTicketReviewModeStatusLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket review mode status: %w", err)
		}
	}

	return nil
}
