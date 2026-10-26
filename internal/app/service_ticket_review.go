package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type ticketReviewInput struct {
	Key      string
	Result   string
	Findings []string
	DryRun   bool
}

func (s service) reviewTicket(ctx context.Context, request ticketReviewInput) (core.TicketReviewCommandResult, error) {
	index, err := core.NewTicketIndex(ctx, s.env, s.assetsFS)
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}
	defer func() {
		//nolint:errcheck // The operation result already owns its failure; the index is a rebuildable projection.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, s.projectContext); err != nil {
		return core.TicketReviewCommandResult{}, err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, s.projectContext.ProjectID, request.Key)
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}

	adapter := core.NewRemoteReviewAdapter(config.Review.Providers, reviewRemoteCommandRunner)

	release, err := acquireTicketReviewLock(ctx, s.projectContext, request.DryRun)
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}
	defer release()

	result, err := core.RunTicketReviewCommand(ctx, core.TicketReviewCommandOptions{
		Ticket: ticket, Config: config, Result: request.Result, Findings: request.Findings,
		Adapter: adapter, DryRun: request.DryRun, ReviewedAt: time.Now().UTC(),
	})
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}

	if err := refreshTicketIndexAfterReview(ctx, index, s.projectContext, request.DryRun); err != nil {
		return core.TicketReviewCommandResult{}, err
	}

	return result, nil
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
