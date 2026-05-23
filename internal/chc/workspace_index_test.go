package chc

import (
	"context"
	"path/filepath"
	"testing"
)

func TestWorkspaceIndexBuildsLookupsAndEdgesFromGraph(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	index := NewWorkspaceIndexFromGraph(graph)

	assertWorkspaceIndexLookups(t, index, graph)
	assertWorkspaceIndexEdges(t, index, graph)
}

func assertWorkspaceIndexLookups(t *testing.T, index WorkspaceIndex, graph WorkspaceGraph) {
	t.Helper()

	projectID := WorkspaceNodeID(*graph.Project)
	caseID := WorkspaceNodeID(graph.Cases[0])

	if index.NodesByID[projectID].Name != "Example" {
		t.Fatalf("project node = %#v, want Example", index.NodesByID[projectID])
	}

	if index.CasesBySlug["blog"] != caseID {
		t.Fatalf("CasesBySlug[blog] = %q, want %q", index.CasesBySlug["blog"], caseID)
	}

	if index.NodesByPath[graph.Cases[0].RelativePath] != caseID {
		t.Fatalf("NodesByPath case = %q, want %q", index.NodesByPath[graph.Cases[0].RelativePath], caseID)
	}

	if len(index.ByKind[WorkspaceNodeSpec]) != 1 {
		t.Fatalf("ByKind[spec] = %#v, want one spec", index.ByKind[WorkspaceNodeSpec])
	}

	if len(index.ByCaseSlug["blog"]) != 2 {
		t.Fatalf("ByCaseSlug[blog] = %#v, want spec and ticket", index.ByCaseSlug["blog"])
	}

	ticketID := index.TicketsByKey["CHC-1"]
	if ticketID == "" {
		t.Fatal("TicketsByKey[CHC-1] is empty")
	}
}

func assertWorkspaceIndexEdges(t *testing.T, index WorkspaceIndex, graph WorkspaceGraph) {
	t.Helper()

	projectID := WorkspaceNodeID(*graph.Project)
	caseID := WorkspaceNodeID(graph.Cases[0])
	ticketID := index.TicketsByKey["CHC-1"]

	if !workspaceIndexHasEdge(index, projectID, caseID, WorkspaceEdgeContains) {
		t.Fatalf("missing project contains case edge in %#v", index.Edges)
	}

	if !workspaceIndexHasEdge(index, caseID, ticketID, WorkspaceEdgeContains) {
		t.Fatalf("missing case contains ticket edge in %#v", index.Edges)
	}

	if !workspaceIndexHasEdge(index, ticketID, caseID, WorkspaceEdgeReferencesCase) {
		t.Fatalf("missing ticket references case edge in %#v", index.Edges)
	}
}

func TestWorkspaceIndexTracksProjectWideAndUnresolvedCaseDocuments(t *testing.T) {
	projectDir := t.TempDir()
	writeWorkspaceGraphFile(t, CaseMetadataPath(projectDir, "valid"), RenderCaseMetadata(CaseMetadata{
		Name: "Valid",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "plans", "project-plan.md"), `---
title: "Project Plan"
status: "ready"
---

# Project Plan
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "dumps", "missing-case.md"), `---
title: "Missing Case Dump"
case: "missing"
---

# Missing Case Dump
`)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	index := NewWorkspaceIndexFromGraph(graph)

	if len(index.ByKind[WorkspaceNodePlan]) != 1 {
		t.Fatalf("ByKind[plan] = %#v, want one project-wide plan", index.ByKind[WorkspaceNodePlan])
	}

	if len(index.ByCaseSlug["missing"]) != 1 {
		t.Fatalf("ByCaseSlug[missing] = %#v, want unresolved dump", index.ByCaseSlug["missing"])
	}

	unresolvedID := index.ByCaseSlug["missing"][0]
	if index.NodesByID[unresolvedID].Title != "Missing Case Dump" {
		t.Fatalf("unresolved node = %#v, want Missing Case Dump", index.NodesByID[unresolvedID])
	}
}

func TestWorkspaceGraphBuilderBuildIndex(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	index, err := NewWorkspaceGraphBuilder(projectDir).BuildIndex(context.Background())
	if err != nil {
		t.Fatalf("BuildIndex returned error: %v", err)
	}

	if len(index.CasesBySlug) != 1 || index.CasesBySlug["blog"] == "" {
		t.Fatalf("CasesBySlug = %#v, want blog", index.CasesBySlug)
	}
}

func workspaceIndexHasEdge(index WorkspaceIndex, from, to string, kind WorkspaceEdgeKind) bool {
	for _, edge := range index.Edges {
		if edge.From == from && edge.To == to && edge.Kind == kind {
			return true
		}
	}

	return false
}
