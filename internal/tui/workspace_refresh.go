package tui

import (
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkspaceRefreshExecutor rebuilds workspace graph and index state for the
// TUI.
type WorkspaceRefreshExecutor interface {
	RefreshWorkspace() (WorkspaceRefreshResult, error)
}

// WorkspaceRefreshResult contains refreshed workspace state for the TUI.
type WorkspaceRefreshResult struct {
	Message              string
	Graph                core.WorkspaceGraph
	Index                core.WorkspaceIndex
	IndexHealth          core.WorkspaceIndexHealth
	IndexChanges         core.WorkspaceIndexChanges
	IndexCache           core.WorkspaceIndexCacheState
	SemanticDigestStates []core.WorkspaceSemanticDigestState
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
	m.semanticDigestStates = append([]core.WorkspaceSemanticDigestState(nil), result.SemanticDigestStates...)
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
	m.semanticDigestStates = append([]core.WorkspaceSemanticDigestState(nil), result.SemanticDigestStates...)
	m.projectMetadata = projectMetadataFromWorkspaceGraph(m.projectMetadata, result.Graph)
	m.cases = casesFromWorkspaceIndex(result.Index, result.Graph, m.cases)
	m.clampSelectedCase()

	return true
}

func workspaceIndexFromRefreshResult(result WorkspaceRefreshResult) core.WorkspaceIndex {
	if len(result.Index.NodesByID) > 0 {
		return result.Index
	}

	return core.NewWorkspaceIndexFromGraph(result.Graph)
}

func projectMetadataFromWorkspaceGraph(
	current core.ProjectMetadata,
	graph core.WorkspaceGraph,
) core.ProjectMetadata {
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
