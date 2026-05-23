package chcapp

import "forge.adrianpk.com/chc/internal/chc"

type WorkspaceRefreshResult struct {
	Message      string
	Graph        chc.WorkspaceGraph
	Index        chc.WorkspaceIndex
	IndexHealth  chc.WorkspaceIndexHealth
	IndexChanges chc.WorkspaceIndexChanges
	IndexCache   chc.WorkspaceIndexCacheState
}

func (s service) RefreshWorkspace() (WorkspaceRefreshResult, error) {
	result, err := chc.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).BuildCachedIndex(s.ctx)
	if err != nil {
		return WorkspaceRefreshResult{}, err
	}

	return WorkspaceRefreshResult{
		Message:     "Workspace refreshed.",
		Graph:       result.Graph,
		Index:       result.Index,
		IndexHealth: result.Index.Health(),
		IndexCache:  result.Cache,
	}, nil
}
