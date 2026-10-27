package core

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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

	ticketID := index.TicketsByKey["CHR-1"]
	if ticketID == "" {
		t.Fatal("TicketsByKey[CHR-1] is empty")
	}
}

func assertWorkspaceIndexEdges(t *testing.T, index WorkspaceIndex, graph WorkspaceGraph) {
	t.Helper()

	projectID := WorkspaceNodeID(*graph.Project)
	caseID := WorkspaceNodeID(graph.Cases[0])
	ticketID := index.TicketsByKey["CHR-1"]

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
	writeWorkspaceGraphFile(t, authoringcore.CaseMetadataPath(projectDir, "valid"), authoringcore.RenderCaseMetadata(authoringcore.CaseMetadata{
		Name: "Valid",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "plans", "project-plan.md"), `---
title: "Project Plan"
status: "ready"
---

# Project Plan
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "dumps", "missing-case.md"), `---
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

func TestWorkspaceIndexExcludesArchivedGeneratedDocuments(t *testing.T) {
	projectDir := t.TempDir()
	writeWorkspaceGraphFile(t, authoringcore.CaseMetadataPath(projectDir, "blog"), authoringcore.RenderCaseMetadata(authoringcore.CaseMetadata{
		Name: "Blog Publishing",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case: "blog"
---

# Blog Plan
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "plans", "archive", "blog-plan-20260525.md"), `---
title: "Archived Blog Plan"
status: "planned"
case: "blog"
---

# Archived Blog Plan
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "specs", "archive", "blog-spec-20260525.md"), `---
title: "Archived Blog Spec"
status: "draft"
case: "blog"
---

# Archived Blog Spec
`)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	index := NewWorkspaceIndexFromGraph(graph)

	if len(index.ByKind[WorkspaceNodePlan]) != 1 {
		t.Fatalf("ByKind[plan] = %#v, want only canonical plan", index.ByKind[WorkspaceNodePlan])
	}

	if len(index.ByKind[WorkspaceNodeSpec]) != 1 {
		t.Fatalf("ByKind[spec] = %#v, want only canonical spec", index.ByKind[WorkspaceNodeSpec])
	}

	for path := range index.NodesByPath {
		if strings.Contains(path, "/archive/") {
			t.Fatalf("archived path indexed: %s", path)
		}
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
			{Path: ".chr/default/cases/broken.md", Kind: WorkspaceNodeCase, Message: "invalid"},
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

func TestWorkspaceIndexSuggestsCases(t *testing.T) {
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"case:blog": {
				Kind:        WorkspaceNodeCase,
				Slug:        "blog",
				Name:        "Blog Publishing",
				Description: "Publish articles.",
			},
			"case:runtime": {
				Kind:        WorkspaceNodeCase,
				Slug:        "runtime",
				Name:        "Runtime Operations",
				Description: "Observe runtime work.",
			},
		},
		CasesBySlug: map[string]string{
			"blog":    "case:blog",
			"runtime": "case:runtime",
		},
	}

	suggestions := index.SuggestCases("blog")

	if len(suggestions) != 1 {
		t.Fatalf("suggestions = %#v, want one match", suggestions)
	}

	if suggestions[0].Slug != "blog" {
		t.Fatalf("Slug = %q, want blog", suggestions[0].Slug)
	}
}

func TestWorkspaceIndexResolvesPathsAndRelatedDocuments(t *testing.T) {
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"case:blog": {
				Kind: WorkspaceNodeCase,
				Slug: "blog",
			},
			"spec:ops/default/spec/00000-blog/artifact.md": {
				Kind:         WorkspaceNodeSpec,
				RelativePath: "ops/default/spec/00000-blog/artifact.md",
				Title:        "Blog Spec",
				CaseSlug:     "blog",
			},
			"ticket:ops/default/ticket/00000-chr-1/artifact.md": {
				Kind:         WorkspaceNodeTicket,
				RelativePath: "ops/default/ticket/00000-chr-1/artifact.md",
				Title:        "Blog Ticket",
				CaseSlug:     "blog",
			},
		},
		NodesByPath: map[string]string{
			"ops/default/spec/00000-blog/artifact.md": "spec:ops/default/spec/00000-blog/artifact.md",
		},
		ByCaseSlug: map[string][]string{
			"blog": {
				"spec:ops/default/spec/00000-blog/artifact.md",
				"ticket:ops/default/ticket/00000-chr-1/artifact.md",
			},
		},
	}

	node, ok := index.ResolvePath("ops/default/spec/00000-blog/artifact.md")
	if !ok {
		t.Fatal("ResolvePath returned ok=false, want true")
	}

	if node.Title != "Blog Spec" {
		t.Fatalf("node.Title = %q, want Blog Spec", node.Title)
	}

	related := index.RelatedDocuments("spec:ops/default/spec/00000-blog/artifact.md")
	if !related.Resolved {
		t.Fatal("RelatedDocuments returned Resolved=false, want true")
	}

	if len(related.Related) != 1 || related.Related[0].Title != "Blog Ticket" {
		t.Fatalf("Related = %#v, want Blog Ticket", related.Related)
	}
}

func TestWorkspaceIndexSearchUsesStructuralMetadataAndSemanticDigests(t *testing.T) {
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"spec:blog": {
				Kind:       WorkspaceNodeSpec,
				Title:      "Blog Spec",
				SourceHash: "source-blog",
			},
			"plan:ops": {
				Kind:       WorkspaceNodePlan,
				Title:      "Operations",
				SourceHash: "source-ops",
			},
		},
	}
	cache := NewWorkspaceSemanticDigestCache()
	cache.Entries["plan:ops"] = WorkspaceSemanticDigest{
		NodeID: "plan:ops",
		Digest: "publishing workflow and editorial operations",
	}

	results := index.Search("publishing", cache)

	if len(results) == 0 {
		t.Fatal("Search returned no results, want semantic digest match")
	}

	if results[0].NodeID != "plan:ops" || !results[0].MatchedSemanticDigest {
		t.Fatalf("first result = %#v, want semantic plan match", results[0])
	}
}

func TestWorkspaceIndexSearchReportsStaleSemanticDigestStatus(t *testing.T) {
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"plan:ops": {
				Kind:       WorkspaceNodePlan,
				Title:      "Operations",
				SourceHash: "source-new",
			},
		},
	}
	cache := NewWorkspaceSemanticDigestCache()
	cache.Entries["plan:ops"] = WorkspaceSemanticDigest{
		NodeID:     "plan:ops",
		Digest:     "release coordination",
		SourceHash: "source-old",
	}
	states := []WorkspaceSemanticDigestState{
		{NodeID: "plan:ops", Status: WorkspaceSemanticDigestStale},
	}

	results := index.SearchWithSemanticDigestStates("release", cache, states)

	if len(results) != 1 {
		t.Fatalf("results = %#v, want one result", results)
	}

	if results[0].SemanticDigestStatus != WorkspaceSemanticDigestStale {
		t.Fatalf("SemanticDigestStatus = %q, want stale", results[0].SemanticDigestStatus)
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
