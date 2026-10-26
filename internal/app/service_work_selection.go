package app

import (
	"context"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type agentWorkLoopSelection struct {
	ExcludedAdopted int
	Tickets         []core.IndexedTicket
	EligibleCount   int
	ReviewConfig    core.ReviewWorkflowConfig
	WorkConfig      core.WorkConfig
}

func parseProjectTicketsFromDir(ctx context.Context, projectID, projectDir string, indexedAt time.Time) ([]core.IndexedTicket, error) {
	files, err := core.NewTicketFileDiscoverer().DiscoverTicketFiles(ctx, core.DiscoverTicketFilesOptions{ProjectDir: projectDir})
	if err != nil {
		return nil, err
	}

	parser := core.NewTicketParser()

	tickets := make([]core.IndexedTicket, 0, len(files))
	for _, file := range files {
		ticket, err := parser.ParseTicket(ctx, core.ParseTicketOptions{
			ProjectID:  projectID,
			ProjectDir: projectDir,
			Path:       file.Path,
			IndexedAt:  indexedAt,
		})
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (s service) agentWorkLoopSelection(opts AgentWorkSelectionOptions) (agentWorkLoopSelection, error) {
	index, err := core.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	defer func() {
		//nolint:errcheck // Query result already comes from the main workflow.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	config, err := agentWorkLoopWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	tickets, err := core.NewTicketLoopQuery(index).EligibleTickets(s.ctx, core.TicketLoopQueryOptions{
		ProjectID: s.projectContext.ProjectID,
		Statuses:  agentWorkStatuses(opts.Statuses),
		Filters:   opts.Filters,
		Order:     agentWorkOrder(opts.Order),
	})
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	tickets, excluded, err := core.FilterLegacyWorkTickets(s.ctx, tickets)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	tickets, err = s.applyAgentWorkSelectionPreflight(index, tickets, config.Work)
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	eligibleTickets := deduplicateTicketLoopTickets(tickets)
	tickets = ticketsForTicketLoopRun(eligibleTickets, ticketLoopRunOptions{MaxIterations: 1})

	return agentWorkLoopSelection{
		Tickets:         tickets,
		EligibleCount:   len(eligibleTickets),
		ExcludedAdopted: excluded,
		ReviewConfig:    config.Review,
		WorkConfig:      config.Work,
	}, nil
}

func (s service) applyAgentWorkSelectionPreflight(index core.TicketIndex, tickets []core.IndexedTicket, workConfig core.WorkConfig) ([]core.IndexedTicket, error) {
	allTickets, err := index.ListProjectTickets(s.ctx, s.projectContext.ProjectID)
	if err != nil {
		return nil, err
	}

	allTickets, _, err = core.FilterLegacyWorkTickets(s.ctx, allTickets)
	if err != nil {
		return nil, err
	}

	preflight, err := core.RunWorkSelectionPreflight(s.ctx, core.WorkSelectionPreflightOptions{
		ProjectDir:   s.projectContext.ProjectDir,
		Tickets:      allTickets,
		BranchPolicy: workConfig.BranchPolicy,
		Workspace:    workConfig.Workspace,
	})
	if err != nil {
		return nil, err
	}

	if !preflight.Allowed {
		return nil, formatWorkSelectionPreflightError(preflight)
	}

	if len(preflight.SelectedTickets) > 0 && hasReviewRepairTicket(preflight.SelectedTickets) {
		return preflight.SelectedTickets, nil
	}

	return tickets, nil
}

func agentWorkLoopWorkspaceConfig(projectDir string) (core.WorkspaceConfig, error) {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return core.WorkspaceConfig{}, err
	}

	if err := core.ValidateWorkConfig(config.Work); err != nil {
		return core.WorkspaceConfig{}, err
	}

	if err := core.ValidateReviewWorkflowConfig(config.Review); err != nil {
		return core.WorkspaceConfig{}, err
	}

	return config, nil
}

func validateAgentWorkSelectionOptions(opts AgentWorkSelectionOptions) error {
	if opts.MaxIterations < 0 {
		return fmt.Errorf("max-iterations must be greater than or equal to zero")
	}

	return nil
}

func defaultAgentWorkSelectionOptions(maxIterations int) AgentWorkSelectionOptions {
	return AgentWorkSelectionOptions{
		Statuses:      []string{core.TicketStatusReadyForDev},
		Order:         core.TicketLoopOrderCreatedAt,
		MaxIterations: maxIterations,
	}
}

func agentWorkStatuses(statuses []string) []string {
	if len(statuses) == 0 {
		return []string{core.TicketStatusReadyForDev}
	}

	return statuses
}

func agentWorkOrder(order core.TicketLoopOrder) core.TicketLoopOrder {
	if order == "" {
		return core.TicketLoopOrderCreatedAt
	}

	return order
}
