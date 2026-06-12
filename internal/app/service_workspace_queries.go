package app

import "forge.adrianpk.com/chc/internal/workspace"

// WorkspacePathResolutionResult reports a workspace index path resolution.
type WorkspacePathResolutionResult struct {
	Node    workspace.WorkspaceNode
	Found   bool
	Message string
}

// WorkspaceRelatedDocumentsResult contains documents related to a source path.
type WorkspaceRelatedDocumentsResult struct {
	Node     workspace.WorkspaceNode
	Related  []workspace.WorkspaceNode
	Resolved bool
	Message  string
}

// WorkspaceIndexSearchResult contains workspace index search matches.
type WorkspaceIndexSearchResult struct {
	Results []workspace.WorkspaceIndexSearchResult
	Message string
}

func (s service) ListWorkspaceCases() ([]workspace.WorkspaceCaseSuggestion, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return nil, err
	}

	return index.ListCases(), nil
}

func (s service) SuggestWorkspaceCases(query string) ([]workspace.WorkspaceCaseSuggestion, error) {
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

	cache, err := workspace.ReadWorkspaceSemanticDigestCache(
		workspace.WorkspaceSemanticDigestCachePath(s.projectContext.ProjectDir),
	)
	if err != nil {
		return WorkspaceIndexSearchResult{}, err
	}

	return WorkspaceIndexSearchResult{
		Results: index.Search(query, cache),
		Message: "Workspace index searched.",
	}, nil
}

func (s service) workspaceIndex() (workspace.WorkspaceIndex, error) {
	result, err := workspace.NewWorkspaceGraphBuilder(s.projectContext.ProjectDir).BuildCachedIndex(s.ctx)
	if err != nil {
		return workspace.WorkspaceIndex{}, err
	}

	return result.Index, nil
}
