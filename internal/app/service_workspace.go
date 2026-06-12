package app

import "forge.adrianpk.com/chc/internal/workspace"

// WorkspaceRefreshResult reports the outcome of rebuilding workspace indexes.
type WorkspaceRefreshResult struct {
	Message              string
	Graph                workspace.WorkspaceGraph
	Index                workspace.WorkspaceIndex
	IndexHealth          workspace.WorkspaceIndexHealth
	IndexChanges         workspace.WorkspaceIndexChanges
	IndexCache           workspace.WorkspaceIndexCacheState
	SemanticDigestStates []workspace.WorkspaceSemanticDigestState
}

func (s service) RefreshWorkspace() (WorkspaceRefreshResult, error) {
	result, err := workspace.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).BuildCachedIndex(s.ctx)
	if err != nil {
		return WorkspaceRefreshResult{}, err
	}

	digestStates, err := s.workspaceSemanticDigestStates(result.Index)
	if err != nil {
		return WorkspaceRefreshResult{}, err
	}

	return WorkspaceRefreshResult{
		Message:              "Workspace refreshed.",
		Graph:                result.Graph,
		Index:                result.Index,
		IndexHealth:          result.Index.HealthWithSemanticDigestStates(digestStates),
		IndexCache:           result.Cache,
		SemanticDigestStates: digestStates,
	}, nil
}

func (s service) workspaceSemanticDigestStates(
	index workspace.WorkspaceIndex,
) ([]workspace.WorkspaceSemanticDigestState, error) {
	cache, err := workspace.ReadWorkspaceSemanticDigestCache(
		workspace.WorkspaceSemanticDigestCachePath(s.projectContext.ProjectDir),
	)
	if err != nil {
		return nil, err
	}

	return workspace.WorkspaceSemanticDigestStates(index, cache, "", ""), nil
}
