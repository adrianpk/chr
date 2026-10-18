package app

import "forge.adrianpk.com/chr/internal/core"

// ReadTicketBoard uses the selected authority without rebuilding the legacy index.
func (s service) ReadTicketBoard(offset int) (core.Board, error) {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.Board{}, err
	}

	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return core.Board{}, err
	}

	return core.ReadTicketBoard(s.ctx, core.TicketBoardReadOptions{
		Layout: layout, ProjectID: s.projectContext.ProjectID,
		NativeWorkflow: core.WorkAuthority(config.Work) == core.WorkStateAuthorityTickets,
		Offset:         offset,
	})
}
