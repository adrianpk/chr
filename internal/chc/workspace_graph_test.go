package chc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceGraphBuilderBuildsCaseRootedViews(t *testing.T) {
	projectDir := t.TempDir()

	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assertCaseRootedWorkspaceGraph(t, graph)
}

func writeCaseRootedWorkspaceGraphFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeWorkspaceGraphFile(t, ProjectMetadataPath(projectDir), RenderProjectMetadata(ProjectMetadata{
		Name:        "Example",
		Description: "Example project.",
	}))
	writeWorkspaceGraphFile(t, CaseMetadataPath(projectDir, "blog"), RenderCaseMetadata(CaseMetadata{
		Name:        "Blog Publishing",
		Description: "Publish articles.",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "accepted"
case: "blog"
---

# Blog Spec
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "plans", "project-plan.md"), `---
title: "Project Plan"
status: "ready"
---

# Project Plan
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "tickets", "CHC-1-Write.md"), ticketFixture(t, Ticket{
		Key:         "CHC-1",
		Title:       "Write article",
		Status:      "ready",
		Priority:    "normal",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "blog",
		ProjectArea: "content",
		CreatedAt:   time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC),
	}))
}

func assertCaseRootedWorkspaceGraph(t *testing.T, graph WorkspaceGraph) {
	t.Helper()

	if graph.Project == nil || graph.Project.Name != "Example" {
		t.Fatalf("Project = %#v, want Example", graph.Project)
	}

	if len(graph.Cases) != 1 || graph.Cases[0].Slug != "blog" {
		t.Fatalf("Cases = %#v, want blog case", graph.Cases)
	}

	caseNodes, ok := graph.CaseView("blog")
	if !ok {
		t.Fatal("CaseView returned ok=false, want true")
	}

	if len(caseNodes) != 2 {
		t.Fatalf("len(caseNodes) = %d, want 2: %#v", len(caseNodes), caseNodes)
	}

	assertWorkspaceGraphNode(t, caseNodes, WorkspaceNodeSpec, "Blog Spec", "accepted", "blog")
	assertWorkspaceGraphNode(t, caseNodes, WorkspaceNodeTicket, "Write article", "ready", "blog")
	assertWorkspaceGraphNode(t, graph.ProjectWide, WorkspaceNodePlan, "Project Plan", "ready", "")

	if len(graph.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %#v, want none", graph.Diagnostics)
	}

	if len(graph.UnresolvedCaseReferences) != 0 {
		t.Fatalf("UnresolvedCaseReferences = %#v, want none", graph.UnresolvedCaseReferences)
	}
}

func TestWorkspaceGraphBuilderTracksInvalidFilesAndUnresolvedCases(t *testing.T) {
	projectDir := t.TempDir()

	writeWorkspaceGraphFile(t, CaseMetadataPath(projectDir, "valid"), RenderCaseMetadata(CaseMetadata{
		Name: "Valid",
	}))
	writeWorkspaceGraphFile(t, CaseMetadataPath(projectDir, "broken"), "# Missing frontmatter\n")
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

	if len(graph.Cases) != 1 || graph.Cases[0].Slug != "valid" {
		t.Fatalf("Cases = %#v, want only valid case", graph.Cases)
	}

	if len(graph.Diagnostics) != 1 {
		t.Fatalf("Diagnostics = %#v, want one invalid case diagnostic", graph.Diagnostics)
	}

	if graph.Diagnostics[0].Kind != WorkspaceNodeCase {
		t.Fatalf("Diagnostic kind = %q, want case", graph.Diagnostics[0].Kind)
	}

	if len(graph.UnresolvedCaseReferences) != 1 {
		t.Fatalf("UnresolvedCaseReferences = %#v, want one unresolved case", graph.UnresolvedCaseReferences)
	}

	if graph.UnresolvedCaseReferences[0].CaseSlug != "missing" {
		t.Fatalf("CaseSlug = %q, want missing", graph.UnresolvedCaseReferences[0].CaseSlug)
	}

	if len(graph.ProjectWide) != 0 {
		t.Fatalf("ProjectWide = %#v, want none", graph.ProjectWide)
	}
}

func TestWorkspaceGraphBuilderClassifiesDumpPairMetadata(t *testing.T) {
	projectDir := t.TempDir()

	writeWorkspaceGraphFile(t, CaseMetadataPath(projectDir, "blog"), RenderCaseMetadata(CaseMetadata{
		Name: "Blog Publishing",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "dumps", "blog-notes.md"), `---
title: "Blog Notes"
status: "curated"
type: "dump"
source: ".chc/dumps/blog-notes-raw.md"
case: "blog"
---

# Blog Notes
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "dumps", "blog-notes-raw.md"), `---
title: "Blog Notes Raw Capture"
status: "raw"
type: "dump_raw"
curated: ".chc/dumps/blog-notes.md"
case: "blog"
---

# Blog Notes Raw Capture
`)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	caseNodes, ok := graph.CaseView("blog")
	if !ok {
		t.Fatal("CaseView returned ok=false, want true")
	}

	curated := findWorkspaceGraphNodeByTitle(t, caseNodes, "Blog Notes")
	if curated.DocumentType != "dump" || curated.SourcePath != ".chc/dumps/blog-notes-raw.md" {
		t.Fatalf("curated dump metadata = %#v", curated)
	}

	raw := findWorkspaceGraphNodeByTitle(t, caseNodes, "Blog Notes Raw Capture")
	if raw.DocumentType != "dump_raw" || raw.CuratedPath != ".chc/dumps/blog-notes.md" {
		t.Fatalf("raw dump metadata = %#v", raw)
	}

	if !IsRawDumpNode(raw) {
		t.Fatalf("IsRawDumpNode returned false for %#v", raw)
	}

	if !IsCuratedDumpNode(curated) {
		t.Fatalf("IsCuratedDumpNode returned false for %#v", curated)
	}
}

func TestWorkspaceGraphBuilderToleratesMissingChcDirectory(t *testing.T) {
	graph, err := NewWorkspaceGraphBuilder(t.TempDir()).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if graph.Project != nil {
		t.Fatalf("Project = %#v, want nil", graph.Project)
	}

	if len(graph.Cases) != 0 || len(graph.ProjectWide) != 0 || len(graph.Diagnostics) != 0 {
		t.Fatalf("graph = %#v, want empty graph", graph)
	}
}

func findWorkspaceGraphNodeByTitle(t *testing.T, nodes []WorkspaceNode, title string) WorkspaceNode {
	t.Helper()

	for _, node := range nodes {
		if node.Title == title {
			return node
		}
	}

	t.Fatalf("node %q not found in %#v", title, nodes)

	return WorkspaceNode{}
}

func assertWorkspaceGraphNode(
	t *testing.T,
	nodes []WorkspaceNode,
	kind WorkspaceNodeKind,
	title string,
	status string,
	caseSlug string,
) {
	t.Helper()

	for _, node := range nodes {
		if node.Kind == kind && node.Title == title && node.Status == status && node.CaseSlug == caseSlug {
			return
		}
	}

	t.Fatalf("nodes = %#v, missing %s %q", nodes, kind, title)
}

func ticketFixture(t *testing.T, ticket Ticket) string {
	t.Helper()

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	return content
}

func writeWorkspaceGraphFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}
