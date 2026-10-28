package app

import (
	"context"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type agentWorkLoopSelection struct {
	ExcludedAdopted int
	Tickets         []ticketcore.Indexed
	EligibleCount   int
	ReviewConfig    workspaceconfig.ReviewWorkflowConfig
	WorkConfig      workspaceconfig.WorkConfig
}

func parseProjectTicketsFromDir(ctx context.Context, projectID, projectDir string, indexedAt time.Time) ([]ticketcore.Indexed, error) {
	files, err := ticketcore.NewDiscoverer().Discover(ctx, ticketcore.DiscoverOptions{ProjectDir: projectDir})
	if err != nil {
		return nil, err
	}

	parser := ticketcore.NewParser()

	tickets := make([]ticketcore.Indexed, 0, len(files))
	for _, file := range files {
		ticket, err := parser.Parse(ctx, ticketcore.ParseOptions{
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
	index, err := workflowcore.NewWorkflowTicketIndex(s.ctx, s.env, s.assetsFS)
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

	tickets, err := ticketcore.NewLoopQuery(index).EligibleTickets(s.ctx, ticketcore.LoopQueryOptions{
		ProjectID: s.projectContext.ProjectID,
		Statuses:  agentWorkStatuses(opts.Statuses),
		Filters:   opts.Filters,
		Order:     agentWorkOrder(opts.Order),
	})
	if err != nil {
		return agentWorkLoopSelection{}, err
	}

	legacySelection, legacySelectionErr := workflowcore.FilterLegacyWorkTickets(s.ctx, tickets)

	tickets, excluded, err := legacySelection.Tickets, legacySelection.Excluded, legacySelectionErr
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

func (s service) applyAgentWorkSelectionPreflight(index ticketcore.Index, tickets []ticketcore.Indexed, workConfig workspaceconfig.WorkConfig) ([]ticketcore.Indexed, error) {
	allTickets, err := index.ListProject(s.ctx, s.projectContext.ProjectID)
	if err != nil {
		return nil, err
	}

	legacySelection2, legacySelectionErr2 := workflowcore.FilterLegacyWorkTickets(s.ctx, allTickets)

	allTickets, _, err = legacySelection2.Tickets, legacySelection2.Excluded, legacySelectionErr2
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

func agentWorkLoopWorkspaceConfig(projectDir string) (workspaceconfig.Config, error) {
	config, err := workspaceconfig.Load(projectDir)
	if err != nil {
		return workspaceconfig.Config{}, err
	}

	if err := workspaceconfig.ValidateWorkConfig(config.Work); err != nil {
		return workspaceconfig.Config{}, err
	}

	if err := core.ValidateReviewWorkflowConfig(config.Review); err != nil {
		return workspaceconfig.Config{}, err
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
		Statuses:      []string{string(ticketcore.StatusReadyForDev)},
		Order:         workflowstate.TicketOrderCreatedAt,
		MaxIterations: maxIterations,
	}
}

func agentWorkStatuses(statuses []string) []string {
	if len(statuses) == 0 {
		return []string{string(ticketcore.StatusReadyForDev)}
	}

	return statuses
}

func agentWorkOrder(order workflowstate.TicketOrder) workflowstate.TicketOrder {
	if order == "" {
		return workflowstate.TicketOrderCreatedAt
	}

	return order
}
