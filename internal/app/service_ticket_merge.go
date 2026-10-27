package app

import (
	"context"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

type ticketMergeInput struct {
	Key      string
	Base     *string
	Strategy string
	DryRun   bool
}

func (s service) mergeTicket(ctx context.Context, request ticketMergeInput) (core.TicketMergeCommandResult, error) {
	index, err := core.NewWorkflowTicketIndex(ctx, s.env, s.assetsFS)
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}
	defer func() {
		//nolint:errcheck // The operation result already owns its failure; the index is a rebuildable projection.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, s.projectContext); err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	mergeSelection, err := ticketMergeCommandTickets(ctx, index, s.projectContext, request.Key)
	ticket, tickets := mergeSelection.Ticket, mergeSelection.Tickets

	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	if err := core.ValidateWorkConfig(config.Work); err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	base := ticketMergeBase(config.Work.BranchPolicy, request.Base)

	release, err := acquireTicketMergeLock(ctx, s.projectContext, request.DryRun)
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}
	defer release()

	result, err := core.RunTicketMergeCommand(ctx, core.TicketMergeCommandOptions{
		Ticket: ticket, Tickets: tickets, Remote: core.NewRemoteReviewAdapter(config.Review.Providers, reviewRemoteCommandRunner),
		ExpectedBase: base, MergeStrategy: request.Strategy, ProjectDir: s.projectContext.ProjectDir,
		BranchPolicy: config.Work.BranchPolicy, Workspace: config.Work.Workspace, DryRun: request.DryRun, MergedAt: time.Now().UTC(),
	})
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	if err := refreshTicketIndexAfterMerge(ctx, index, s.projectContext, request.DryRun); err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	return result, nil
}

func ticketMergeCommandTickets(ctx context.Context, index ticketcore.Index, projectContext commandProjectContext, key string) (ticketMergeSelection, error) {
	tickets, err := index.ListProject(ctx, projectContext.ProjectID)
	if err != nil {
		return ticketMergeSelection{Ticket: ticketcore.Indexed{}, Tickets: nil}, err
	}

	ticket, err := index.GetProjectByKey(ctx, projectContext.ProjectID, key)
	if err != nil {
		return ticketMergeSelection{Ticket: ticketcore.Indexed{}, Tickets: nil}, err
	}

	return ticketMergeSelection{Ticket: ticket, Tickets: tickets}, nil
}

func refreshTicketIndexAfterMerge(ctx context.Context, index ticketcore.Index, projectContext commandProjectContext, dryRun bool) error {
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

func ticketMergeBase(policy core.WorkBranchPolicy, supplied *string) string {
	if supplied != nil {
		if base := strings.TrimSpace(*supplied); base != "" {
			return base
		}
	}

	return policy.IntegrationBranch
}

type ticketMergeSelection struct {
	Ticket  ticketcore.Indexed
	Tickets []ticketcore.Indexed
}
