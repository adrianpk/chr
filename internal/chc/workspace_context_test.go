package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssembleWorkspaceContextLoadsCaseRootedDocuments(t *testing.T) {
	projectDir := t.TempDir()
	writeWorkspaceContextFixture(t, projectDir)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assembly, err := AssembleWorkspaceContext(NewCaseRootedWorkspaceContextRequest(projectDir, graph, "blog"))
	if err != nil {
		t.Fatalf("AssembleWorkspaceContext returned error: %v", err)
	}

	if assembly.Scope != WorkspaceContextScopeCase {
		t.Fatalf("Scope = %q, want case", assembly.Scope)
	}

	if !strings.Contains(assembly.Project.Content, "Project context body.") {
		t.Fatal("project document body was not loaded")
	}

	if !strings.Contains(assembly.Case.Content, "Case context body.") {
		t.Fatal("case document body was not loaded")
	}

	rendered := RenderWorkspaceContextAssembly(assembly)
	for _, want := range []string{
		"# Workspace Context",
		"Case Spec",
		"Case Plan",
		"Unresolved Case References",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered context missing %q:\n%s", want, rendered)
		}
	}

	for _, forbidden := range []string{"Project Plan", "Board Ticket"} {
		if strings.Contains(rendered, forbidden) {
			t.Fatalf("rendered case context contains non-case document %q:\n%s", forbidden, rendered)
		}
	}
}

func TestAssembleWorkspaceContextLoadsProjectWideDocuments(t *testing.T) {
	projectDir := t.TempDir()
	writeWorkspaceContextFixture(t, projectDir)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assembly, err := AssembleWorkspaceContext(NewProjectWideWorkspaceContextRequest(projectDir, graph))
	if err != nil {
		t.Fatalf("AssembleWorkspaceContext returned error: %v", err)
	}

	rendered := RenderWorkspaceContextAssembly(assembly)
	for _, want := range []string{
		"## Scope\n\nproject",
		"Project Plan",
		"missing",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered project context missing %q:\n%s", want, rendered)
		}
	}

	if strings.Contains(rendered, "Case Spec") {
		t.Fatalf("rendered project context contains case-scoped spec:\n%s", rendered)
	}
}

func TestAssembleWorkspaceContextReportsMissingFileAfterGraphBuild(t *testing.T) {
	projectDir := t.TempDir()
	writeWorkspaceContextFixture(t, projectDir)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	missingPath := filepath.Join(projectDir, ".chc", "plans", "blog-plan.md")

	err = os.Remove(missingPath)
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	assembly, err := AssembleWorkspaceContext(NewCaseRootedWorkspaceContextRequest(projectDir, graph, "blog"))
	if err != nil {
		t.Fatalf("AssembleWorkspaceContext returned error: %v", err)
	}

	var found bool

	for _, source := range assembly.Sources {
		if source.Path == ".chc/plans/blog-plan.md" {
			found = true

			if source.Included {
				t.Fatal("missing plan source was marked included")
			}

			if source.Reason != "file was not found" {
				t.Fatalf("missing plan reason = %q", source.Reason)
			}
		}
	}

	if !found {
		t.Fatal("missing plan source was not recorded")
	}
}

func TestAssembleWorkspaceContextReportsMissingSelectedCase(t *testing.T) {
	projectDir := t.TempDir()
	writeWorkspaceContextFixture(t, projectDir)

	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assembly, err := AssembleWorkspaceContext(NewCaseRootedWorkspaceContextRequest(projectDir, graph, "missing"))
	if err != nil {
		t.Fatalf("AssembleWorkspaceContext returned error: %v", err)
	}

	if len(assembly.Sources) == 0 {
		t.Fatal("missing case source was not recorded")
	}

	last := assembly.Sources[len(assembly.Sources)-1]
	if last.ID != "workspace:case:missing" || last.Included || last.Reason != "case was not found in workspace graph" {
		t.Fatalf("missing case source = %#v", last)
	}
}

func writeWorkspaceContextFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeTestFile(t, ProjectMetadataPath(projectDir), RenderProjectMetadata(ProjectMetadata{
		SchemaVersion: ProjectMetadataSchemaVersion,
		Name:          "Workspace Context",
		Description:   "Project description.",
		Context:       "Project context body.",
	}))
	writeTestFile(t, CaseMetadataPath(projectDir, "blog"), RenderCaseMetadata(CaseMetadata{
		SchemaVersion: CaseMetadataSchemaVersion,
		Slug:          "blog",
		Name:          "Blog",
		Description:   "Case description.",
		Context:       "Case context body.",
	}))
	writeTestFile(t, filepath.Join(projectDir, ".chc", "specs", "blog-spec.md"), `---
title: Case Spec
case: blog
---

# Case Spec

Case specification body.
`)
	writeTestFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: Case Plan
case: blog
---

# Case Plan

Case plan body.
`)
	writeTestFile(t, filepath.Join(projectDir, ".chc", "plans", "project-plan.md"), `---
title: Project Plan
---

# Project Plan

Project plan body.
`)
	writeTestFile(t, filepath.Join(projectDir, ".chc", "dumps", "missing-case.md"), `---
title: Missing Case Dump
case: missing
---

# Missing Case Dump
`)
	writeTestFile(t, filepath.Join(projectDir, ".chc", "tickets", "CHC-1-board.md"), renderProjectDocumentTestTicket(t, Ticket{
		Key:      "CHC-1",
		Title:    "Board Ticket",
		Status:   "ready",
		Priority: "normal",
		Severity: "normal",
		Type:     "feature",
		UseCase:  "blog",
	}, "\n# Description\n\nTicket body.\n"))
}
