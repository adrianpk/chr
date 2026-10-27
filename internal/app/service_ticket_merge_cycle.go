package app

import (
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

type agentWorkMergeCycleOptions struct {
	Ticket       core.IndexedTicket
	Execution    core.ExecutionRecord
	WorkConfig   core.WorkConfig
	ReviewConfig core.ReviewWorkflowConfig
}

func (s service) runAgentWorkMergeCycle(opts agentWorkMergeCycleOptions) (core.TicketMergeCommandResult, error) {
	mergedAt := s.now().UTC()

	tickets, err := s.agentWorkMergeTickets(opts.Execution, mergedAt)
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	result, err := core.RunTicketMergeCommand(s.ctx, core.TicketMergeCommandOptions{
		Ticket:        opts.Ticket,
		Tickets:       tickets,
		Remote:        core.NewRemoteReviewAdapter(opts.ReviewConfig.Providers, reviewRemoteCommandRunner),
		ExpectedBase:  opts.WorkConfig.BranchPolicy.IntegrationBranch,
		MergeStrategy: core.TicketMergeStrategyRebase,
		ProjectDir:    s.projectContext.ProjectDir,
		BranchPolicy:  opts.WorkConfig.BranchPolicy,
		Workspace:     opts.WorkConfig.Workspace,
		MergedAt:      mergedAt,
	})
	if err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	if err := s.refreshAgenticReviewTicketIndex(opts.Execution, opts.Ticket.Path, mergedAt); err != nil {
		return core.TicketMergeCommandResult{}, err
	}

	return result, nil
}

func (s service) agentWorkMergeTickets(execution core.ExecutionRecord, indexedAt time.Time) ([]core.IndexedTicket, error) {
	if execution.Workspace.Mode == workspacecore.WorkWorkspaceModeWorktree {
		return parseProjectTicketsFromDir(s.ctx, s.projectContext.ProjectID, agenticReviewProjectDir(s.projectContext, execution), indexedAt)
	}

	return s.listProjectTickets()
}
