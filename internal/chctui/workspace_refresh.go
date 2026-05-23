package chctui

import (
	"fmt"

	"forge.adrianpk.com/chc/internal/chc"
)

type WorkspaceRefreshExecutor interface {
	RefreshWorkspace() (WorkspaceRefreshResult, error)
}

type WorkspaceRefreshResult struct {
	Message string
	Graph   chc.WorkspaceGraph
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

	m.workspaceGraph = result.Graph
	m.statusMessage = workspaceRefreshMessage(result)
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
