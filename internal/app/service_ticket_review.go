package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type ticketReviewInput struct {
	Key      string
	Result   string
	Findings []string
	DryRun   bool
}

func (s service) reviewTicket(ctx context.Context, request ticketReviewInput) (core.TicketReviewCommandResult, error) {
	index, err := workflowcore.NewWorkflowTicketIndex(ctx, s.env, s.assetsFS)
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

	ticket, err := index.GetProjectByKey(ctx, s.projectContext.ProjectID, request.Key)
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}

	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}

	adapter := workflowcore.NewRemoteReviewAdapter(config.Review.Providers, reviewRemoteCommandRunner)

	release, err := acquireTicketReviewLock(ctx, s.projectContext, request.DryRun)
	if err != nil {
		return core.TicketReviewCommandResult{}, err
	}
	defer release()

	result, err := core.RunTicketReviewCommand(ctx, core.TicketReviewCommandOptions{
		Ticket: ticket, Config: config, Result: ticketcore.ReviewResult(request.Result), Findings: request.Findings,
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

func refreshTicketIndexAfterReview(ctx context.Context, index ticketcore.Index, projectContext commandProjectContext, dryRun bool) error {
	if dryRun {
		return nil
	}

	return refreshProjectTicketIndex(ctx, index, projectContext)
}

func acquireTicketReviewLock(ctx context.Context, projectContext commandProjectContext, dryRun bool) (func(), error) {
	if dryRun {
		return noopRelease, nil
	}

	lock, err := acquireAgentOperationLock(ctx, projectContext, workflowcore.AgentOperationReview)
	if err != nil {
		return noopRelease, err
	}

	return releaseAgentOperationLock(lock), nil
}
