package chctui

import (
	"fmt"

	"forge.adrianpk.com/chc/internal/chc"
)

type WorkspaceRefreshExecutor interface {
	RefreshWorkspace() (WorkspaceRefreshResult, error)
}

type WorkspaceRefreshResult struct {
	Message              string
	Graph                chc.WorkspaceGraph
	Index                chc.WorkspaceIndex
	IndexHealth          chc.WorkspaceIndexHealth
	IndexChanges         chc.WorkspaceIndexChanges
	IndexCache           chc.WorkspaceIndexCacheState
	SemanticDigestStates []chc.WorkspaceSemanticDigestState
}

func (m *Model) refreshWorkspace() {
	if m.workspaceRefreshBlocked() {
		return
	}

	if m.workspaceExecutor == nil {
		m.statusMessage = "Workspace refresh executor is unavailable."

		return
	}

	result, err := m.workspaceExecutor.RefreshWorkspace()
	if err != nil {
		m.statusMessage = "Workspace refresh failed: " + err.Error()

		return
	}

	result.Index = workspaceIndexFromRefreshResult(result)
	result.IndexChanges = result.Index.ChangesSince(m.workspaceIndex)
	m.workspaceGraph = result.Graph
	m.workspaceIndex = result.Index
	m.semanticDigestStates = append([]chc.WorkspaceSemanticDigestState(nil), result.SemanticDigestStates...)
	m.projectMetadata = projectMetadataFromWorkspaceGraph(m.projectMetadata, result.Graph)
	m.cases = casesFromWorkspaceIndex(result.Index, result.Graph, m.cases)
	m.clampSelectedCase()

	m.statusMessage = workspaceRefreshMessage(result)
}

func (m *Model) refreshWorkspaceGraphAfterWrite() bool {
	if m.workspaceExecutor == nil {
		return true
	}

	result, err := m.workspaceExecutor.RefreshWorkspace()
	if err != nil {
		m.statusMessage = "Workspace refresh failed: " + err.Error()

		return false
	}

	result.Index = workspaceIndexFromRefreshResult(result)
	result.IndexChanges = result.Index.ChangesSince(m.workspaceIndex)
	m.workspaceGraph = result.Graph
	m.workspaceIndex = result.Index
	m.semanticDigestStates = append([]chc.WorkspaceSemanticDigestState(nil), result.SemanticDigestStates...)
	m.projectMetadata = projectMetadataFromWorkspaceGraph(m.projectMetadata, result.Graph)
	m.cases = casesFromWorkspaceIndex(result.Index, result.Graph, m.cases)
	m.clampSelectedCase()

	return true
}

func workspaceIndexFromRefreshResult(result WorkspaceRefreshResult) chc.WorkspaceIndex {
	if len(result.Index.NodesByID) > 0 {
		return result.Index
	}

	return chc.NewWorkspaceIndexFromGraph(result.Graph)
}

func projectMetadataFromWorkspaceGraph(
	current chc.ProjectMetadata,
	graph chc.WorkspaceGraph,
) chc.ProjectMetadata {
	if graph.Project == nil {
		return current
	}

	current.Name = graph.Project.Name
	current.Description = graph.Project.Description
	current.SourcePath = graph.Project.Path
	current.SourceHash = graph.Project.SourceHash
	current.Exists = true

	return current
}

func (m *Model) workspaceRefreshBlocked() bool {
	switch {
	case m.projectEdit.Active:
		m.statusMessage = "Finish or cancel the active project edit before refreshing."

		return true
	case m.caseEdit.Active:
		m.statusMessage = "Finish or cancel the active case edit before refreshing."

		return true
	case m.projectDirty:
		m.statusMessage = "Save or discard project changes before refreshing."

		return true
	case m.caseDirty:
		m.statusMessage = "Save or discard case changes before refreshing."

		return true
	default:
		return false
	}
}

func workspaceRefreshMessage(result WorkspaceRefreshResult) string {
	if len(result.Graph.Diagnostics) > 0 {
		return fmt.Sprintf("Workspace refreshed with %d diagnostics.", len(result.Graph.Diagnostics))
	}

	if len(result.Graph.UnresolvedCaseReferences) > 0 {
		return fmt.Sprintf(
			"Workspace refreshed with %d unresolved case references.",
			len(result.Graph.UnresolvedCaseReferences),
		)
	}

	if result.Message != "" {
		return result.Message
	}

	return "Workspace refreshed."
}
