package chctui

import (
	"sort"

	"forge.adrianpk.com/chc/internal/chc"
)

func casesFromWorkspaceGraph(graph chc.WorkspaceGraph, fallback []chc.CaseMetadata) []chc.CaseMetadata {
	if len(graph.Cases) == 0 {
		return append([]chc.CaseMetadata(nil), fallback...)
	}

	fallbackBySlug := make(map[string]chc.CaseMetadata, len(fallback))
	for _, metadata := range fallback {
		fallbackBySlug[metadata.Slug] = metadata
	}

	cases := make([]chc.CaseMetadata, 0, len(graph.Cases))
	for _, node := range graph.Cases {
		metadata := fallbackBySlug[node.Slug]
		metadata.SchemaVersion = chc.CaseMetadataSchemaVersion
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

func workspaceGraphProjectDocumentCount(graph chc.WorkspaceGraph, kind chc.WorkspaceNodeKind) int {
	count := 0

	for _, node := range graph.ProjectWide {
		if node.Kind == kind {
			count++
		}
	}

	return count
}

func workspaceGraphCaseDocumentCount(graph chc.WorkspaceGraph, slug string, kind chc.WorkspaceNodeKind) int {
	count := 0

	for _, node := range graph.CaseDocuments[slug] {
		if node.Kind == kind {
			count++
		}
	}

	return count
}

func workspaceGraphCaseReferenceCount(graph chc.WorkspaceGraph, slug string) int {
	for _, unresolved := range graph.UnresolvedCaseReferences {
		if unresolved.CaseSlug == slug {
			return len(unresolved.References)
		}
	}

	return 0
}

func workspaceGraphUnresolvedCaseReferences(graph chc.WorkspaceGraph) []chc.WorkspaceCaseReference {
	references := append([]chc.WorkspaceCaseReference(nil), graph.UnresolvedCaseReferences...)
	sort.Slice(references, func(i, j int) bool {
		return references[i].CaseSlug < references[j].CaseSlug
	})

	return references
}
