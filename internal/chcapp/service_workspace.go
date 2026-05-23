package chcapp

import "forge.adrianpk.com/chc/internal/chc"

type WorkspaceRefreshResult struct {
	Message      string
	Graph        chc.WorkspaceGraph
	Index        chc.WorkspaceIndex
	IndexHealth  chc.WorkspaceIndexHealth
	IndexChanges chc.WorkspaceIndexChanges
}

func (s service) RefreshWorkspace() (WorkspaceRefreshResult, error) {
	graph, err := chc.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).Build(s.ctx)
	if err != nil {
		return WorkspaceRefreshResult{}, err
	}

	index := chc.NewWorkspaceIndexFromGraph(graph)

	return WorkspaceRefreshResult{
		Message:     "Workspace refreshed.",
		Graph:       graph,
		Index:       index,
		IndexHealth: index.Health(),
	}, nil
}
