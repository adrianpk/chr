package tui

import (
	"sort"

	"forge.adrianpk.com/chc/internal/workspace"
)

func casesFromWorkspaceGraph(graph workspace.WorkspaceGraph, fallback []workspace.CaseMetadata) []workspace.CaseMetadata {
	if len(graph.Cases) == 0 {
		return append([]workspace.CaseMetadata(nil), fallback...)
	}

	fallbackBySlug := make(map[string]workspace.CaseMetadata, len(fallback))
	for _, metadata := range fallback {
		fallbackBySlug[metadata.Slug] = metadata
	}

	cases := make([]workspace.CaseMetadata, 0, len(graph.Cases))
	for _, node := range graph.Cases {
		metadata := fallbackBySlug[node.Slug]
		metadata.SchemaVersion = workspace.CaseMetadataSchemaVersion
		metadata.Slug = node.Slug
		metadata.Name = node.Name
		metadata.Description = node.Description
		metadata.SourcePath = node.Path
		metadata.SourceHash = node.SourceHash
		metadata.Exists = true
		cases = append(cases, metadata)
	}

	return cases
}

func casesFromWorkspaceIndex(
	index workspace.WorkspaceIndex,
	graph workspace.WorkspaceGraph,
	fallback []workspace.CaseMetadata,
) []workspace.CaseMetadata {
	suggestions := index.ListCases()
	if len(suggestions) == 0 {
		return casesFromWorkspaceGraph(graph, fallback)
	}

	fallbackBySlug := make(map[string]workspace.CaseMetadata, len(fallback))
	for _, metadata := range fallback {
		fallbackBySlug[metadata.Slug] = metadata
	}

	cases := make([]workspace.CaseMetadata, 0, len(suggestions))
	for _, suggestion := range suggestions {
		metadata := fallbackBySlug[suggestion.Slug]
		metadata.SchemaVersion = workspace.CaseMetadataSchemaVersion
		metadata.Slug = suggestion.Slug
		metadata.Name = suggestion.Name
		metadata.Description = suggestion.Description
		metadata.SourcePath = suggestion.Path
		metadata.Exists = true
		cases = append(cases, metadata)
	}

	return cases
}

func workspaceIndexCaseDocumentCount(index workspace.WorkspaceIndex, slug string, kind workspace.WorkspaceNodeKind) int {
	count := 0

	for _, nodeID := range index.ByCaseSlug[slug] {
		node := index.NodesByID[nodeID]
		if kind == workspace.WorkspaceNodeDump && workspace.IsRawDumpNode(node) {
			continue
		}

		if node.Kind == kind {
			count++
		}
	}

	return count
}

func workspaceGraphProjectDocumentCount(graph workspace.WorkspaceGraph, kind workspace.WorkspaceNodeKind) int {
	count := 0

	for _, node := range graph.ProjectWide {
		if kind == workspace.WorkspaceNodeDump && workspace.IsRawDumpNode(node) {
			continue
		}

		if node.Kind == kind {
			count++
		}
	}

	return count
}

func workspaceGraphCaseReferenceCount(graph workspace.WorkspaceGraph, slug string) int {
	for _, unresolved := range graph.UnresolvedCaseReferences {
		if unresolved.CaseSlug == slug {
			return len(unresolved.References)
		}
	}

	return 0
}

func workspaceGraphUnresolvedCaseReferences(graph workspace.WorkspaceGraph) []workspace.WorkspaceCaseReference {
	references := append([]workspace.WorkspaceCaseReference(nil), graph.UnresolvedCaseReferences...)
	sort.Slice(references, func(i, j int) bool {
		return references[i].CaseSlug < references[j].CaseSlug
	})

	return references
}
