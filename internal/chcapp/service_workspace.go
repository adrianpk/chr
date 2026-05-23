package chcapp

import "forge.adrianpk.com/chc/internal/chc"

type WorkspaceRefreshResult struct {
	Message string
	Graph   chc.WorkspaceGraph
}

func (s service) RefreshWorkspace() (WorkspaceRefreshResult, error) {
	graph, err := chc.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).Build(s.ctx)
	if err != nil {
		return WorkspaceRefreshResult{}, err
	}

	return WorkspaceRefreshResult{
		Message: "Workspace refreshed.",
		Graph:   graph,
	}, nil
}
