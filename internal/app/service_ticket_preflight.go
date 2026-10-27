package app

import (
	"context"
	"io/fs"

	"forge.adrianpk.com/chr/internal/core"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

type ticketStartPreflight struct {
	AssetsFS       fs.FS
	Env            core.UserEnvironment
	ProjectContext commandProjectContext
	Ticket         core.IndexedTicket
	Record         core.ExecutionRecord
	Runtime        core.RuntimeValidationResult
}

type ticketStartPreflightOptions struct {
	PrepareWorktree bool
}

type ticketStartWorkspace struct {
	Ticket                core.IndexedTicket
	TicketPath            string
	TicketBranch          string
	ExecutionWorkspaceDir string
	WorkspaceBaseRef      string
	WorkspaceBaseSHA      string
	PreparedWorktree      bool
}

func (s service) prepareTicketStart(ctx context.Context, ticket core.IndexedTicket, opts ticketStartPreflightOptions) (ticketStartPreflight, error) {
	if err := core.CheckLegacyWorkTicket(ticket.Path); err != nil {
		return ticketStartPreflight{}, err
	}

	return s.buildLegacyTicketStartPreflight(ctx, ticket, opts)
}

func (s service) buildLegacyTicketStartPreflight(ctx context.Context, ticket core.IndexedTicket, opts ticketStartPreflightOptions) (ticketStartPreflight, error) {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	workspace, err := prepareTicketStartWorkspace(ctx, s.projectContext, ticket, config, opts)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record := newPendingExecutionRecord(s.projectContext.ProjectID, workspace.Ticket)
	if workspace.TicketPath != "" {
		record.TicketPath = workspace.TicketPath
	}

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return ticketStartPreflight{}, err
	}

	artifact, err := opscore.CreateOpsArtifact(opscore.CreateOpsArtifactOptions{
		Layout:      layout,
		Type:        "run",
		Title:       ticketStartArtifactTitle(workspace.Ticket),
		Producer:    "chr",
		ExternalIDs: []string{record.ID},
	})
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record.Artifacts.ArtifactID = artifact.ID

	result, err := validateCodexRuntime(core.RuntimeValidationOptions{
		ProjectDir:            s.projectContext.ProjectDir,
		Config:                config.Runtime,
		Workspace:             config.Work.Workspace,
		ExecutionWorkspaceDir: workspace.ExecutionWorkspaceDir,
		TicketKey:             workspace.Ticket.Key,
		TicketBranch:          workspace.TicketBranch,
		WorkspaceBaseRef:      workspace.WorkspaceBaseRef,
		WorkspaceBaseSHA:      workspace.WorkspaceBaseSHA,
		RunID:                 record.ID,
		OpsLayout:             &layout,
		ArtifactID:            artifact.ID,
	})
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record = core.ExecutionRecordFromRuntimePlan(record, result.Plan)

	return ticketStartPreflight{
		AssetsFS:       s.assetsFS,
		Env:            s.env,
		ProjectContext: s.projectContext,
		Ticket:         workspace.Ticket,
		Record:         record,
		Runtime:        result,
	}, nil
}

func (s service) selectTicketStartTicket(ctx context.Context, key string) (core.IndexedTicket, error) {
	index, err := core.NewTicketIndex(ctx, s.env, s.assetsFS)
	if err != nil {
		return core.IndexedTicket{}, err
	}
	defer func() {
		//nolint:errcheck // The operation result already owns its failure; the index is a rebuildable projection.
		_ = index.Close()
	}()

	if err := refreshProjectTicketIndex(ctx, index, s.projectContext); err != nil {
		return core.IndexedTicket{}, err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, s.projectContext.ProjectID, key)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	tickets, err := index.ListProjectTickets(ctx, s.projectContext.ProjectID)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	if err := enforceWorkSelectionPreflight(ctx, s.projectContext, tickets, ticket.Key); err != nil {
		return core.IndexedTicket{}, err
	}

	return ticket, nil
}
