package app

import (
	"context"
	"io/fs"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
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

type ticketStartSelection struct {
	ProjectContext commandProjectContext
	Ticket         core.IndexedTicket
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

func (factory commandFactory) buildTicketStartPreflight(ctx context.Context, cmd *cli.Command) (ticketStartPreflight, error) {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	selection, err := selectTicketStartTicket(ctx, cmd, env, factory.deps.assetsFS)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	return buildTicketStartPreflightForTicketWithOptions(ctx, env, selection.ProjectContext, selection.Ticket, factory.deps.assetsFS, ticketStartPreflightOptions{
		PrepareWorktree: !cmd.Bool("dry-run"),
	})
}

func buildTicketStartPreflightForTicket(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	ticket core.IndexedTicket,
	assetsFS fs.FS,
) (ticketStartPreflight, error) {
	return buildTicketStartPreflightForTicketWithOptions(ctx, env, projectContext, ticket, assetsFS, ticketStartPreflightOptions{
		PrepareWorktree: true,
	})
}

func buildTicketStartPreflightForTicketWithOptions(ctx context.Context, env core.UserEnvironment, projectContext commandProjectContext, ticket core.IndexedTicket, assetsFS fs.FS, opts ticketStartPreflightOptions) (ticketStartPreflight, error) {
	if err := core.CheckLegacyWorkTicket(ticket.Path); err != nil {
		return ticketStartPreflight{}, err
	}

	return buildLegacyTicketStartPreflight(ctx, env, projectContext, ticket, assetsFS, opts)
}

func buildLegacyTicketStartPreflight(ctx context.Context, env core.UserEnvironment, projectContext commandProjectContext, ticket core.IndexedTicket, assetsFS fs.FS, opts ticketStartPreflightOptions) (ticketStartPreflight, error) {
	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	workspace, err := prepareTicketStartWorkspace(ctx, projectContext, ticket, config, opts)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	record := newPendingExecutionRecord(projectContext.ProjectID, workspace.Ticket)
	if workspace.TicketPath != "" {
		record.TicketPath = workspace.TicketPath
	}

	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: projectContext.ProjectDir})
	if err != nil {
		return ticketStartPreflight{}, err
	}

	artifact, err := core.CreateOpsArtifact(core.CreateOpsArtifactOptions{
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
		ProjectDir:            projectContext.ProjectDir,
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
		AssetsFS:       assetsFS,
		Env:            env,
		ProjectContext: projectContext,
		Ticket:         workspace.Ticket,
		Record:         record,
		Runtime:        result,
	}, nil
}

func selectTicketStartTicket(ctx context.Context, cmd *cli.Command, env core.UserEnvironment, assetsFS fs.FS) (ticketStartSelection, error) {
	projectContext, err := newCommandProjectContextResolver(env, assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return ticketStartSelection{}, err
	}

	index, err := core.NewTicketIndex(ctx, env, assetsFS)
	if err != nil {
		return ticketStartSelection{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return ticketStartSelection{}, err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return ticketStartSelection{}, err
	}

	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return ticketStartSelection{}, err
	}

	err = enforceWorkSelectionPreflight(ctx, projectContext, tickets, ticket.Key)
	if err != nil {
		return ticketStartSelection{}, err
	}

	return ticketStartSelection{
		ProjectContext: projectContext,
		Ticket:         ticket,
	}, nil
}
