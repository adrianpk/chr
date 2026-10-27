package app

import (
	"sort"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

//nolint:wsl_v5 // Tight loop keeps filtering and candidate assignment together.
func latestDocumentNodeForScope(scope quizcore.Scope, index core.WorkspaceIndex, kind core.WorkspaceNodeKind) (core.WorkspaceNode, bool) {
	nodes := []core.WorkspaceNode{}
	for _, nodeID := range index.ByKind[kind] {
		node, ok := index.NodesByID[nodeID]
		if !ok {
			continue
		}

		if node.CaseSlug != scope.CaseSlug {
			continue
		}

		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return core.WorkspaceNode{}, false
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].ModifiedAt.After(nodes[j].ModifiedAt)
	})

	return nodes[0], true
}
