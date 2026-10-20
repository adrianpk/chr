package app

import (
	"sort"

	"forge.adrianpk.com/chr/internal/core"
)

// ListSpecDocuments keeps concrete-document reads separate from latest-in-scope
// inventory queries used by generation and existing CLI/TUI adapters.
func (s service) ListSpecDocuments() ([]core.SpecDocument, error) {
	return core.ListSpecDocuments(s.ctx, s.projectContext.ProjectDir)
}

func (s service) specGenerationSources() (core.ProjectDumpObservation, error) {
	observed, err := core.ObserveProjectDump(s.ctx, s.projectContext.ProjectDir)
	if err != nil {
		return observed, err
	}

	if observed.ReconciliationReason != "" {
		return observed, core.ErrProjectDumpReconciliation
	}

	return observed, nil
}

func specContextNodeIDs(index core.WorkspaceIndex, scope core.QuizScope, dumps core.ProjectDumpObservation) []string {
	allowedDumps := map[string]bool{dumps.Destination: true}
	for _, artifact := range dumps.Scoped {
		if scope.CaseSlug != "" && artifact.CaseSlug == scope.CaseSlug {
			allowedDumps[artifact.Reference] = true
		}
	}

	var ids []string

	for _, id := range sortedSpecContextNodeIDs(index) {
		node := index.NodesByID[id]
		if node.Kind == core.WorkspaceNodeDump && allowedDumps[node.RelativePath] || node.Kind == core.WorkspaceNodeQuiz && node.CaseSlug == scope.CaseSlug {
			ids = append(ids, id)
		}
	}

	sort.Strings(ids)

	return ids
}
