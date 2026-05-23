package chcapp

import "forge.adrianpk.com/chc/internal/chc"

type WorkspacePathResolutionResult struct {
	Node    chc.WorkspaceNode
	Found   bool
	Message string
}

type WorkspaceRelatedDocumentsResult struct {
	Node     chc.WorkspaceNode
	Related  []chc.WorkspaceNode
	Resolved bool
	Message  string
}

type WorkspaceIndexSearchResult struct {
	Results []chc.WorkspaceIndexSearchResult
	Message string
}

func (s service) ListWorkspaceCases() ([]chc.WorkspaceCaseSuggestion, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return nil, err
	}

	return index.ListCases(), nil
}

func (s service) SuggestWorkspaceCases(query string) ([]chc.WorkspaceCaseSuggestion, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return nil, err
	}

	return index.SuggestCases(query), nil
}

func (s service) ResolveWorkspacePath(path string) (WorkspacePathResolutionResult, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return WorkspacePathResolutionResult{}, err
	}

	node, ok := index.ResolvePath(path)
	if !ok {
		return WorkspacePathResolutionResult{
			Message: "Workspace path not found.",
		}, nil
	}

	return WorkspacePathResolutionResult{
		Node:    node,
		Found:   true,
		Message: "Workspace path resolved.",
	}, nil
}

func (s service) RelatedWorkspaceDocuments(nodeID string) (WorkspaceRelatedDocumentsResult, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return WorkspaceRelatedDocumentsResult{}, err
	}

	result := index.RelatedDocuments(nodeID)
	if !result.Resolved {
		return WorkspaceRelatedDocumentsResult{
			Message: "Workspace node not found.",
		}, nil
	}

	return WorkspaceRelatedDocumentsResult{
		Node:     result.Node,
		Related:  result.Related,
		Resolved: true,
		Message:  "Workspace related documents resolved.",
	}, nil
}

func (s service) SearchWorkspaceIndex(query string) (WorkspaceIndexSearchResult, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return WorkspaceIndexSearchResult{}, err
	}

	cache, err := chc.ReadWorkspaceSemanticDigestCache(
		chc.WorkspaceSemanticDigestCachePath(s.projectContext.ProjectDir),
	)
	if err != nil {
		return WorkspaceIndexSearchResult{}, err
	}

	return WorkspaceIndexSearchResult{
		Results: index.Search(query, cache),
		Message: "Workspace index searched.",
	}, nil
}

func (s service) workspaceIndex() (chc.WorkspaceIndex, error) {
	result, err := chc.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).BuildCachedIndex(s.ctx)
	if err != nil {
		return chc.WorkspaceIndex{}, err
	}

	return result.Index, nil
}
