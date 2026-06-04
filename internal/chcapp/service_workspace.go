package chcapp

import "forge.adrianpk.com/chc/internal/chc"

// WorkspaceRefreshResult reports the outcome of rebuilding workspace indexes.
type WorkspaceRefreshResult struct {
	Message              string
	Graph                chc.WorkspaceGraph
	Index                chc.WorkspaceIndex
	IndexHealth          chc.WorkspaceIndexHealth
	IndexChanges         chc.WorkspaceIndexChanges
	IndexCache           chc.WorkspaceIndexCacheState
	SemanticDigestStates []chc.WorkspaceSemanticDigestState
}

func (s service) RefreshWorkspace() (WorkspaceRefreshResult, error) {
	result, err := chc.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).BuildCachedIndex(s.ctx)
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
	index chc.WorkspaceIndex,
) ([]chc.WorkspaceSemanticDigestState, error) {
	cache, err := chc.ReadWorkspaceSemanticDigestCache(
		chc.WorkspaceSemanticDigestCachePath(s.projectContext.ProjectDir),
	)
	if err != nil {
		return nil, err
	}

	return chc.WorkspaceSemanticDigestStates(index, cache, "", ""), nil
}
