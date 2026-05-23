package chcapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestServiceRefreshWorkspaceBuildsGraph(t *testing.T) {
	projectDir := t.TempDir()

	err := os.MkdirAll(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create .chc: %v", err)
	}

	err = os.WriteFile(
		chc.ProjectMetadataPath(projectDir),
		[]byte(chc.RenderProjectMetadata(chc.ProjectMetadata{
			Name:        "Service Project",
			Description: "Loaded by service.",
		})),
		0o644,
	)
	if err != nil {
		t.Fatalf("write project metadata: %v", err)
	}

	env, err := chcTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chcTestUserEnvironment returned error: %v", err)
	}

	operations := newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
	})

	result, err := operations.RefreshWorkspace()
	if err != nil {
		t.Fatalf("RefreshWorkspace returned error: %v", err)
	}

	if result.Message != "Workspace refreshed." {
		t.Fatalf("Message = %q, want Workspace refreshed.", result.Message)
	}

	if result.Graph.Project == nil || result.Graph.Project.Name != "Service Project" {
		t.Fatalf("Project = %#v, want Service Project", result.Graph.Project)
	}

	if result.IndexHealth.NodeCount != 1 {
		t.Fatalf("IndexHealth.NodeCount = %d, want 1", result.IndexHealth.NodeCount)
	}

	if len(result.Index.NodesByID) != 1 {
		t.Fatalf("Index.NodesByID = %#v, want one project node", result.Index.NodesByID)
	}

	assertWorkspaceIndexCacheMiss(t, result.IndexCache)
}

func TestServiceSaveProjectMetadata(t *testing.T) {
	projectDir := t.TempDir()
	operations := newTestService(t, projectDir)

	result, err := operations.SaveProjectMetadata(chc.ProjectMetadata{
		SchemaVersion: chc.ProjectMetadataSchemaVersion,
		Name:          "Saved Project",
		Description:   "Saved through service.",
	})
	if err != nil {
		t.Fatalf("SaveProjectMetadata returned error: %v", err)
	}

	if result.Message != "Project metadata saved." {
		t.Fatalf("Message = %q, want Project metadata saved.", result.Message)
	}

	if result.Metadata.Name != "Saved Project" {
		t.Fatalf("Metadata.Name = %q, want Saved Project", result.Metadata.Name)
	}

	loaded, err := chc.NewProjectMetadataService(projectDir).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Name != "Saved Project" {
		t.Fatalf("loaded.Name = %q, want Saved Project", loaded.Name)
	}
}

func TestServiceSaveCaseMetadataReloadsCases(t *testing.T) {
	projectDir := t.TempDir()
	operations := newTestService(t, projectDir)

	result, err := operations.SaveCaseMetadata(chc.CaseMetadata{
		SchemaVersion: chc.CaseMetadataSchemaVersion,
		Name:          "Saved Case",
		Description:   "Saved through service.",
	})
	if err != nil {
		t.Fatalf("SaveCaseMetadata returned error: %v", err)
	}

	if result.Message != "Case metadata saved." {
		t.Fatalf("Message = %q, want Case metadata saved.", result.Message)
	}

	if result.Case.Slug != "saved-case" {
		t.Fatalf("Case.Slug = %q, want saved-case", result.Case.Slug)
	}

	if len(result.Cases) != 1 || result.Cases[0].Slug != "saved-case" {
		t.Fatalf("Cases = %#v, want saved-case", result.Cases)
	}
}

func TestServiceWorkspaceIndexQueries(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	operations := newTestService(t, projectDir)

	assertServiceSuggestsWorkspaceCases(t, operations)
	assertServiceResolvesWorkspaceRelatedDocuments(t, operations)
}

func assertServiceSuggestsWorkspaceCases(t *testing.T, operations service) {
	t.Helper()

	suggestions, err := operations.SuggestWorkspaceCases("blog")
	if err != nil {
		t.Fatalf("SuggestWorkspaceCases returned error: %v", err)
	}

	if len(suggestions) != 1 || suggestions[0].Slug != "blog" {
		t.Fatalf("suggestions = %#v, want blog", suggestions)
	}
}

func assertServiceResolvesWorkspaceRelatedDocuments(t *testing.T, operations service) {
	t.Helper()

	resolved, err := operations.ResolveWorkspacePath(filepath.ToSlash(filepath.Join(chc.DirectoryName, "specs", "blog-spec.md")))
	if err != nil {
		t.Fatalf("ResolveWorkspacePath returned error: %v", err)
	}

	if !resolved.Found || resolved.Node.Title != "Blog Spec" {
		t.Fatalf("resolved = %#v, want Blog Spec", resolved)
	}

	related, err := operations.RelatedWorkspaceDocuments(chc.WorkspaceNodeID(resolved.Node))
	if err != nil {
		t.Fatalf("RelatedWorkspaceDocuments returned error: %v", err)
	}

	if !related.Resolved || len(related.Related) != 1 || related.Related[0].Kind != chc.WorkspaceNodeTicket {
		t.Fatalf("related = %#v, want related ticket", related)
	}
}

func TestServiceSearchWorkspaceIndexUsesSemanticDigestCache(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	graph, err := chc.NewWorkspaceGraphBuilder(projectDir).Build(t.Context())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	index := chc.NewWorkspaceIndexFromGraph(graph)
	cache := chc.NewWorkspaceSemanticDigestCache()

	for nodeID, node := range index.NodesByID {
		if node.Kind == chc.WorkspaceNodePlan {
			cache.Entries[nodeID] = chc.WorkspaceSemanticDigest{
				NodeID: nodeID,
				Digest: "editorial roadmap publishing workflow",
			}
		}
	}

	err = chc.WriteWorkspaceSemanticDigestCache(chc.WorkspaceSemanticDigestCachePath(projectDir), cache)
	if err != nil {
		t.Fatalf("WriteWorkspaceSemanticDigestCache returned error: %v", err)
	}

	operations := newTestService(t, projectDir)

	result, err := operations.SearchWorkspaceIndex("editorial")
	if err != nil {
		t.Fatalf("SearchWorkspaceIndex returned error: %v", err)
	}

	if len(result.Results) == 0 {
		t.Fatal("SearchWorkspaceIndex returned no results, want semantic match")
	}

	if !result.Results[0].MatchedSemanticDigest {
		t.Fatalf("first result = %#v, want semantic digest match", result.Results[0])
	}
}

func TestServiceCreateDefaultTicketReloadsTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	operations := newTestService(t, projectDir)

	result, err := operations.CreateDefaultTicket()
	if err != nil {
		t.Fatalf("CreateDefaultTicket returned error: %v", err)
	}

	if len(result.Tickets) != 1 {
		t.Fatalf("Tickets = %#v, want one ticket", result.Tickets)
	}

	if result.Tickets[0].Title != "TUI created ticket" {
		t.Fatalf("ticket title = %q, want TUI created ticket", result.Tickets[0].Title)
	}

	if result.AuthoringSessions == nil {
		t.Fatal("AuthoringSessions = nil, want reloaded empty slice")
	}
}

func TestServiceCreateTUIAuthoringSessionReloadsSessions(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "tickets"), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	operations := newTestService(t, projectDir)

	result, err := operations.CreateTUIAuthoringSession()
	if err != nil {
		t.Fatalf("CreateTUIAuthoringSession returned error: %v", err)
	}

	if len(result.AuthoringSessions) != 1 {
		t.Fatalf("AuthoringSessions = %#v, want one session", result.AuthoringSessions)
	}

	if result.AuthoringSessions[0].Mode != chc.AuthoringSessionModeTUI {
		t.Fatalf("session mode = %q, want tui", result.AuthoringSessions[0].Mode)
	}

	if result.Tickets == nil {
		t.Fatal("Tickets = nil, want reloaded empty slice")
	}
}

func TestServiceMutatesSelectedTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	})
	operations := newTestService(t, projectDir)

	commentResult, err := operations.AddSelectedTicketComment("CHC-1")
	if err != nil {
		t.Fatalf("AddSelectedTicketComment returned error: %v", err)
	}

	if commentResult.Message != "Added comment to ticket CHC-1." {
		t.Fatalf("comment message = %q", commentResult.Message)
	}

	statusResult, err := operations.MoveSelectedTicketStatus("CHC-1", "ready")
	if err != nil {
		t.Fatalf("MoveSelectedTicketStatus returned error: %v", err)
	}

	if statusResult.Message != "Moved ticket CHC-1 to ready." {
		t.Fatalf("status message = %q", statusResult.Message)
	}

	updatedContent, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	content := string(updatedContent)
	for _, want := range []string{
		"status: \"ready\"",
		"# Comments",
		"TUI action recorded.",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content missing %q:\n%s", want, content)
		}
	}
}

func TestServiceRequestDaemonStopReturnsStoppedMessage(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "tickets"), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	operations := newTestService(t, projectDir)

	result, err := operations.RequestDaemonStop()
	if err != nil {
		t.Fatalf("RequestDaemonStop returned error: %v", err)
	}

	if result.Message != "Daemon is already stopped." {
		t.Fatalf("Message = %q, want daemon stopped message", result.Message)
	}

	if result.Tickets == nil {
		t.Fatal("Tickets = nil, want reloaded empty slice")
	}
}

func newTestService(t *testing.T, projectDir string) service {
	t.Helper()

	assetsFS = testAssetsFS()

	env, err := chcTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chcTestUserEnvironment returned error: %v", err)
	}

	return newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
	})
}

func assertWorkspaceIndexCacheMiss(t *testing.T, cache chc.WorkspaceIndexCacheState) {
	t.Helper()

	if cache.Status != chc.WorkspaceIndexCacheMiss {
		t.Fatalf("IndexCache.Status = %q, want %q", cache.Status, chc.WorkspaceIndexCacheMiss)
	}
}

func writeServiceWorkspaceQueryFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeServiceWorkspaceFile(t, chc.ProjectMetadataPath(projectDir), chc.RenderProjectMetadata(chc.ProjectMetadata{
		Name:        "Example",
		Description: "Example project.",
	}))
	writeServiceWorkspaceFile(t, chc.CaseMetadataPath(projectDir, "blog"), chc.RenderCaseMetadata(chc.CaseMetadata{
		Name:        "Blog Publishing",
		Description: "Publish articles.",
	}))
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, chc.DirectoryName, "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "accepted"
case: "blog"
---

# Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, chc.DirectoryName, "plans", "project-plan.md"), `---
title: "Project Plan"
status: "ready"
---

# Project Plan
`)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Write.md", chc.Ticket{
		Key:         "CHC-1",
		Title:       "Write article",
		Status:      "ready",
		Priority:    "normal",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "blog",
		ProjectArea: "content",
		CreatedAt:   testTime(),
		UpdatedAt:   testTime(),
	})
}

func writeServiceWorkspaceFile(t *testing.T, path, content string) {
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

func testTime() time.Time {
	return time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
}
