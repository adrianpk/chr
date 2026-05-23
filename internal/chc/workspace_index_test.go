package chc

import (
	"context"
	"path/filepath"
	"testing"
	"time"
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

func TestWorkspaceIndexChangesSinceDetectsAddedChangedAndRemovedFiles(t *testing.T) {
	previousTime := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	currentTime := previousTime.Add(time.Minute)
	previous := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"spec:changed.md": {
				Kind:       WorkspaceNodeSpec,
				SourceHash: "old",
				ModifiedAt: previousTime,
				Size:       10,
			},
			"spec:removed.md": {
				Kind:       WorkspaceNodeSpec,
				SourceHash: "removed",
				ModifiedAt: previousTime,
				Size:       12,
			},
		},
	}
	current := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"spec:added.md": {
				Kind:       WorkspaceNodeSpec,
				SourceHash: "added",
				ModifiedAt: currentTime,
				Size:       8,
			},
			"spec:changed.md": {
				Kind:       WorkspaceNodeSpec,
				SourceHash: "new",
				ModifiedAt: currentTime,
				Size:       11,
			},
		},
	}

	changes := current.ChangesSince(previous)

	assertStringSlice(t, changes.Added, []string{"spec:added.md"})
	assertStringSlice(t, changes.Changed, []string{"spec:changed.md"})
	assertStringSlice(t, changes.Removed, []string{"spec:removed.md"})
}

func TestWorkspaceIndexHealthReportsDiagnosticsAndUnresolvedReferences(t *testing.T) {
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"case:known":   {},
			"dump:missing": {},
		},
		Edges: []WorkspaceEdge{
			{From: "case:known", To: "dump:missing", Kind: WorkspaceEdgeContains},
		},
		ByCaseSlug: map[string][]string{
			"missing": {"dump:missing"},
		},
		CasesBySlug: map[string]string{
			"known": "case:known",
		},
		Diagnostics: []WorkspaceGraphDiagnostic{
			{Path: ".chc/cases/broken.md", Kind: WorkspaceNodeCase, Message: "invalid"},
		},
		BuiltAt: time.Date(2026, 5, 23, 11, 0, 0, 0, time.UTC),
	}

	health := index.Health()

	if health.NodeCount != 2 {
		t.Fatalf("NodeCount = %d, want 2", health.NodeCount)
	}

	if health.EdgeCount != 1 {
		t.Fatalf("EdgeCount = %d, want 1", health.EdgeCount)
	}

	if health.DiagnosticCount != 1 {
		t.Fatalf("DiagnosticCount = %d, want 1", health.DiagnosticCount)
	}

	if health.UnresolvedReferenceCount != 1 {
		t.Fatalf("UnresolvedReferenceCount = %d, want 1", health.UnresolvedReferenceCount)
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

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice = %#v, want %#v", got, want)
	}

	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("slice = %#v, want %#v", got, want)
		}
	}
}
