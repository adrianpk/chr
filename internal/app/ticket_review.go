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

func (factory commandFactory) ticketReviewRefCommand() *cli.Command {
	return &cli.Command{
		Name:   "review-ref",
		Usage:  "Attach branch and pull request review metadata to one ticket",
		Flags:  ticketReviewRefFlags(),
		Action: factory.runTicketReviewRef,
	}
}

func (factory commandFactory) ticketReviewCommand() *cli.Command {
	return &cli.Command{
		Name:   "review",
		Usage:  "Apply an inspection-only ticket review result",
		Flags:  ticketReviewFlags(),
		Action: factory.runTicketReview,
	}
}

func ticketReviewRefFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "branch",
			Usage: "ticket implementation branch",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "review reference type",
			Value: "pull_request",
		},
		&cli.StringFlag{
			Name:  "provider",
			Usage: "review provider",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "pull request URL",
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "provider pull request identifier",
		},
		&cli.StringFlag{
			Name:  "ref-status",
			Usage: "provider pull request status",
			Value: "open",
		},
	)
}

func ticketReviewFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "result",
			Usage: "review result: pass, needs_changes, insufficient_evidence, needs_human_decision",
			Value: core.TicketReviewCommandResultPass,
		},
		&cli.StringSliceFlag{
			Name:  "finding",
			Usage: "blocking review finding to write remotely",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print review plan without writing ticket or remote comments",
		},
	)
}

func (factory commandFactory) runTicketReviewRef(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket review-ref requires exactly one ticket key argument")
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

	request, err := ticketReviewReferenceRequestFromCommand(cmd)
	if err != nil {
		return err
	}

	result, err := newService(ctx, env, projectContext, factory.deps.assetsFS).AttachTicketReviewReference(request)
	if err != nil {
		return err
	}

	updated, err := ticketOperationResultTicket(result, request.Key)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"ticket: %s review_ref\n", updated.Key)
	if err != nil {
		return fmt.Errorf("write ticket review-ref result: %w", err)
	}

	return nil
}

func (factory commandFactory) runTicketReview(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket review requires exactly one ticket key argument")
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

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	adapter := core.NewRemoteReviewAdapter(config.Review.Providers, reviewRemoteCommandRunner)

	release, err := acquireTicketReviewLock(ctx, projectContext, cmd.Bool("dry-run"))
	if err != nil {
		return err
	}
	defer release()

	result, err := core.RunTicketReviewCommand(ctx, core.TicketReviewCommandOptions{
		Ticket:     ticket,
		Config:     config,
		Result:     cmd.String("result"),
		Findings:   cmd.StringSlice("finding"),
		Adapter:    adapter,
		DryRun:     cmd.Bool("dry-run"),
		ReviewedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	err = refreshTicketIndexAfterReview(ctx, index, projectContext, cmd.Bool("dry-run"))
	if err != nil {
		return err
	}

	return printTicketReviewResult(factory.io.Stdout, projectContext, result)
}

func refreshTicketIndexAfterReview(ctx context.Context, index core.TicketIndex, projectContext commandProjectContext, dryRun bool) error {
	if dryRun {
		return nil
	}

	return refreshProjectTicketIndex(ctx, index, projectContext)
}

func acquireTicketReviewLock(ctx context.Context, projectContext commandProjectContext, dryRun bool) (func(), error) {
	if dryRun {
		return noopRelease, nil
	}

	lock, err := acquireAgentOperationLock(ctx, projectContext, core.AgentOperationReview)
	if err != nil {
		return noopRelease, err
	}

	return releaseAgentOperationLock(lock), nil
}

func printTicketReviewResult(
	stdout io.Writer,
	projectContext commandProjectContext,
	result core.TicketReviewCommandResult,
) error {
	lines := []string{
		"ticket_review: result",
		"project_id: " + projectContext.ProjectID,
		"ticket: " + result.TicketKey,
		"result: " + result.ReviewResult,
		"review_state: " + result.ReviewState,
		fmt.Sprintf("review_round: %d", result.ReviewRound),
		"status: " + result.Status,
		"review_ref: " + result.ReviewRef.URL,
		"provider: " + string(result.RemoteSnapshot.Reference.Provider),
		fmt.Sprintf("remote_comments: %d", len(result.RemoteComments)),
		fmt.Sprintf("dry_run: %t", result.DryRun),
		"policy: " + core.TicketReviewCommandInspectionOnlyPolicySummary,
	}

	for _, finding := range result.BlockingFindings {
		lines = append(lines, "finding: "+finding)
	}

	for _, comment := range result.RemoteComments {
		lines = append(lines, "remote_comment: "+comment.Placement+" "+comment.Kind)
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket review result: %w", err)
		}
	}

	return nil
}

func ticketReviewReferenceRequestFromCommand(cmd *cli.Command) (TicketReviewReferenceRequest, error) {
	request := TicketReviewReferenceRequest{
		Key:    cmd.Args().First(),
		Branch: cmd.String("branch"),
	}

	if strings.TrimSpace(cmd.String("url")) == "" {
		return request, nil
	}

	if strings.TrimSpace(cmd.String("provider")) == "" {
		return TicketReviewReferenceRequest{}, fmt.Errorf("ticket review-ref provider is required when url is set")
	}

	request.Ref = &core.TicketReviewRef{
		Type:     cmd.String("type"),
		Provider: cmd.String("provider"),
		URL:      cmd.String("url"),
		ID:       cmd.String("id"),
		Status:   cmd.String("ref-status"),
	}

	return request, nil
}
