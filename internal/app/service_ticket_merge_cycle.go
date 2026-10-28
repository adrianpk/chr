package app

import (
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type agentWorkMergeCycleOptions struct {
	Ticket       ticketcore.Indexed
	Execution    executioncore.Record
	WorkConfig   workspaceconfig.WorkConfig
	ReviewConfig workspaceconfig.ReviewWorkflowConfig
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
		Remote:        workflowcore.NewRemoteReviewAdapter(opts.ReviewConfig.Providers, reviewRemoteCommandRunner),
		ExpectedBase:  opts.WorkConfig.BranchPolicy.IntegrationBranch,
		MergeStrategy: workflowcore.TicketMergeStrategyRebase,
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

func (s service) agentWorkMergeTickets(execution executioncore.Record, indexedAt time.Time) ([]ticketcore.Indexed, error) {
	if execution.Workspace.Mode == workspacecore.ExecutionWorktree {
		return parseProjectTicketsFromDir(s.ctx, s.projectContext.ProjectID, agenticReviewProjectDir(s.projectContext, execution), indexedAt)
	}

	return s.listProjectTickets()
}
