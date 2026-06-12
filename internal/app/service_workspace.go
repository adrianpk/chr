package app

import "forge.adrianpk.com/chc/internal/core"

// WorkspaceRefreshResult reports the outcome of rebuilding workspace indexes.
type WorkspaceRefreshResult struct {
	Message              string
	Graph                core.WorkspaceGraph
	Index                core.WorkspaceIndex
	IndexHealth          core.WorkspaceIndexHealth
	IndexChanges         core.WorkspaceIndexChanges
	IndexCache           core.WorkspaceIndexCacheState
	SemanticDigestStates []core.WorkspaceSemanticDigestState
}

func (s service) RefreshWorkspace() (WorkspaceRefreshResult, error) {
	result, err := core.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).BuildCachedIndex(s.ctx)
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
	index core.WorkspaceIndex,
) ([]core.WorkspaceSemanticDigestState, error) {
	cache, err := core.ReadWorkspaceSemanticDigestCache(
		core.WorkspaceSemanticDigestCachePath(s.projectContext.ProjectDir),
	)
	if err != nil {
		return nil, err
	}

	return core.WorkspaceSemanticDigestStates(index, cache, "", ""), nil
}
