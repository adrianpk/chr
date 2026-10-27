package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "accepted"
case: "blog"
---

# Blog Spec
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "plans", "project-plan.md"), `---
title: "Project Plan"
status: "ready"
---

# Project Plan
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "tickets", "CHR-1-Write.md"), ticketFixture(t, Ticket{
		Key:         "CHR-1",
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
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "dumps", "missing-case.md"), `---
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
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "dumps", "blog-notes.md"), `---
title: "Blog Notes"
status: "curated"
type: "dump"
source: ".chr/default/dumps/blog-notes-raw.md"
case: "blog"
---

# Blog Notes
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "dumps", "blog-notes-raw.md"), `---
title: "Blog Notes Raw Capture"
status: "raw"
type: "dump_raw"
curated: ".chr/default/dumps/blog-notes.md"
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
	if curated.DocumentType != "dump" || curated.SourcePath != ".chr/default/dumps/blog-notes-raw.md" {
		t.Fatalf("curated dump metadata = %#v", curated)
	}

	raw := findWorkspaceGraphNodeByTitle(t, caseNodes, "Blog Notes Raw Capture")
	if raw.DocumentType != "dump_raw" || raw.CuratedPath != ".chr/default/dumps/blog-notes.md" {
		t.Fatalf("raw dump metadata = %#v", raw)
	}

	if !IsRawDumpNode(raw) {
		t.Fatalf("IsRawDumpNode returned false for %#v", raw)
	}

	if !IsCuratedDumpNode(curated) {
		t.Fatalf("IsCuratedDumpNode returned false for %#v", curated)
	}
}

func TestWorkspaceGraphBuilderClassifiesQuizPairMetadata(t *testing.T) {
	projectDir := t.TempDir()

	writeWorkspaceGraphFile(t, CaseMetadataPath(projectDir, "blog"), RenderCaseMetadata(CaseMetadata{
		Name: "Blog Publishing",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "quizzes", "blog-quiz.md"), `---
title: "Blog Quiz"
status: "curated"
raw_path: ".chr/default/quizzes/blog-quiz-raw.md"
case: "blog"
---

# Blog Quiz
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, UserDirectoryName, "quizzes", "blog-quiz-raw.md"), `---
title: "Blog Quiz Raw Capture"
status: "raw"
curated_path: ".chr/default/quizzes/blog-quiz.md"
case: "blog"
---

# Blog Quiz Raw Capture
`)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	caseNodes, ok := graph.CaseView("blog")
	if !ok {
		t.Fatal("CaseView returned ok=false, want true")
	}

	curated := findWorkspaceGraphNodeByTitle(t, caseNodes, "Blog Quiz")
	if curated.SourcePath != ".chr/default/quizzes/blog-quiz-raw.md" {
		t.Fatalf("curated quiz metadata = %#v", curated)
	}

	raw := findWorkspaceGraphNodeByTitle(t, caseNodes, "Blog Quiz Raw Capture")
	if raw.CuratedPath != ".chr/default/quizzes/blog-quiz.md" {
		t.Fatalf("raw quiz metadata = %#v", raw)
	}

	if !IsRawQuizNode(raw) {
		t.Fatalf("IsRawQuizNode returned false for %#v", raw)
	}

	if !IsCuratedQuizNode(curated) {
		t.Fatalf("IsCuratedQuizNode returned false for %#v", curated)
	}
}

func TestIsRawQuizNodeClassifiesLegacyRawSuffix(t *testing.T) {
	node := WorkspaceNode{
		Kind: WorkspaceNodeQuiz,
		Slug: "project-quiz-raw",
	}

	if !IsRawQuizNode(node) {
		t.Fatalf("IsRawQuizNode returned false for legacy raw quiz %#v", node)
	}
}

func TestWorkspaceGraphBuilderToleratesMissingChrDirectory(t *testing.T) {
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

	path = canonicalWorkspaceGraphFixturePath(path)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}

func canonicalWorkspaceGraphFixturePath(path string) string {
	marker := filepath.Join(workspacecore.DirectoryName, UserDirectoryName)

	parts := strings.Split(filepath.ToSlash(path), filepath.ToSlash(marker)+"/")
	if len(parts) != 2 {
		return path
	}

	artifactTypes := map[string]string{
		"dumps":    "dump",
		"quizzes":  "quiz",
		"specs":    "spec",
		"plans":    "plan",
		"tickets":  "ticket",
		"trackers": "tracker",
	}

	legacyParts := strings.Split(parts[1], "/")
	if len(legacyParts) > 1 && legacyParts[1] == "archive" {
		return path
	}

	artifactType, ok := artifactTypes[legacyParts[0]]
	if !ok || len(legacyParts) == 1 {
		return path
	}

	projectDir := strings.TrimSuffix(parts[0], "/")
	slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	artifactID := "00000-" + opscore.OpsArtifactSlugOrDefault(slug)

	return filepath.Join(projectDir, opscore.OpsDirectoryName, opsnamespace.Default, artifactType, artifactID, opscore.OpsArtifactBodyFile())
}
