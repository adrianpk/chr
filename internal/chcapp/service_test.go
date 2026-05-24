package chcapp

import (
	"context"
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

func TestServiceListDumpsReturnsProjectAndCaseScopedDumps(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "project-context.md"), `---
title: "Project Context"
status: "raw"
---

# Project Context
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "blog-context.md"), `---
title: "Blog Context"
status: "curated"
case: "blog"
---

# Blog Context
`)

	operations := newTestService(t, projectDir)

	result, err := operations.ListDumps()
	if err != nil {
		t.Fatalf("ListDumps returned error: %v", err)
	}

	if result.Message != "Dumps loaded." {
		t.Fatalf("Message = %q, want Dumps loaded.", result.Message)
	}

	if len(result.Dumps) != 2 {
		t.Fatalf("Dumps = %#v, want two dumps", result.Dumps)
	}

	projectDump := findDumpViewByTitle(t, result.Dumps, "Project Context")
	if projectDump.Scope.CaseSlug != "" {
		t.Fatalf("project dump scope = %#v, want project-wide", projectDump.Scope)
	}

	caseDump := findDumpViewByTitle(t, result.Dumps, "Blog Context")
	if caseDump.Scope.CaseSlug != "blog" {
		t.Fatalf("case dump scope = %#v, want blog", caseDump.Scope)
	}

	if caseDump.RelativePath != ".chc/dumps/blog-context.md" {
		t.Fatalf("case dump path = %q, want .chc/dumps/blog-context.md", caseDump.RelativePath)
	}

	if operations.dumpDirectory() != filepath.Join(projectDir, ".chc", "dumps") {
		t.Fatalf("dumpDirectory = %q, want .chc/dumps", operations.dumpDirectory())
	}
}

func TestServiceCreateDumpWritesProjectAndCaseScopedDumps(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	operations := newTestService(t, projectDir)

	projectResult, err := operations.CreateDump(DumpCreateRequest{
		Title:   "Project Notes",
		Content: "Project-wide dump body.",
	})
	if err != nil {
		t.Fatalf("CreateDump project returned error: %v", err)
	}

	caseResult, err := operations.CreateDump(DumpCreateRequest{
		Title:   "Blog Notes",
		Content: "Case-scoped dump body.",
		Scope:   DumpScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("CreateDump case returned error: %v", err)
	}

	assertProjectDumpContent(t, projectDir)
	assertCaseDumpContent(t, projectDir)

	if len(projectResult.Dumps) != 1 {
		t.Fatalf("project result dumps = %#v, want one dump", projectResult.Dumps)
	}

	if len(caseResult.Dumps) != 2 {
		t.Fatalf("case result dumps = %#v, want two dumps", caseResult.Dumps)
	}

	caseDump := findDumpViewByTitle(t, caseResult.Dumps, "Blog Notes")
	if caseDump.Scope.CaseSlug != "blog" {
		t.Fatalf("case dump scope = %#v, want blog", caseDump.Scope)
	}
}

func TestServiceAppendAndOpenDump(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "project-notes.md"), `---
title: "Project Notes"
status: "raw"
---

# Project Notes

Original body.
`)

	operations := newTestService(t, projectDir)
	editor := &fakeDumpEditorLauncher{
		command: DumpEditorCommand{
			Program: "test-editor",
			Args:    []string{"--wait"},
			Dir:     projectDir,
		},
	}
	operations.dumpEditor = editor

	appendResult, err := operations.AppendDump(DumpAppendRequest{
		RelativePath: ".chc/dumps/project-notes.md",
		Content:      "Appended body.",
	})
	if err != nil {
		t.Fatalf("AppendDump returned error: %v", err)
	}

	content := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "project-notes.md"))
	for _, want := range []string{
		`title: "Project Notes"`,
		"Original body.",
		"Appended body.",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dump content missing %q:\n%s", want, content)
		}
	}

	if len(appendResult.Dumps) != 1 {
		t.Fatalf("append result dumps = %#v, want one dump", appendResult.Dumps)
	}

	readResult, err := operations.ReadDump(DumpReadRequest{
		RelativePath: ".chc/dumps/project-notes.md",
	})
	if err != nil {
		t.Fatalf("ReadDump returned error: %v", err)
	}

	if !strings.Contains(readResult.Content, "Appended body.") {
		t.Fatalf("read content missing appended body:\n%s", readResult.Content)
	}

	assertServiceOpenDumpEditorCommand(t, operations, editor, projectDir)
}

func assertServiceOpenDumpEditorCommand(
	t *testing.T,
	operations service,
	editor *fakeDumpEditorLauncher,
	projectDir string,
) {
	t.Helper()

	openResult, err := operations.OpenDump(DumpOpenRequest{
		RelativePath: ".chc/dumps/project-notes.md",
	})
	if err != nil {
		t.Fatalf("OpenDump returned error: %v", err)
	}

	if editor.path != filepath.Join(projectDir, ".chc", "dumps", "project-notes.md") {
		t.Fatalf("editor path = %q", editor.path)
	}

	if openResult.Message != "Editing dump .chc/dumps/project-notes.md." {
		t.Fatalf("open message = %q", openResult.Message)
	}

	if openResult.EditorCommand == nil {
		t.Fatal("EditorCommand = nil, want editor command")
	}

	if openResult.EditorCommand.Program != "test-editor" {
		t.Fatalf("editor program = %q", openResult.EditorCommand.Program)
	}
}

func TestDefaultDumpEditorLauncherUsesShellCommand(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("VISUAL", "code --wait")
	t.Setenv("EDITOR", "nvim")

	path := filepath.Join(t.TempDir(), "project context.md")

	command, err := defaultDumpEditorLauncher{}.EditorCommand(t.Context(), path)
	if err != nil {
		t.Fatalf("EditorCommand returned error: %v", err)
	}

	if command.Program != "/bin/zsh" {
		t.Fatalf("Program = %q, want /bin/zsh", command.Program)
	}

	if len(command.Args) != 2 || command.Args[0] != "-lc" {
		t.Fatalf("Args = %#v, want shell -lc script", command.Args)
	}

	for _, want := range []string{"cd ", "exec code --wait", "'project context.md'"} {
		if !strings.Contains(command.Args[1], want) {
			t.Fatalf("shell script missing %q: %q", want, command.Args[1])
		}
	}

	if command.Description != "code --wait project context.md" {
		t.Fatalf("Description = %q", command.Description)
	}
}

func TestServiceRejectsDumpPathsOutsideDumpDirectory(t *testing.T) {
	operations := newTestService(t, t.TempDir())

	_, err := operations.AppendDump(DumpAppendRequest{
		RelativePath: ".chc/project.md",
		Content:      "bad",
	})
	if err == nil {
		t.Fatal("AppendDump returned nil error, want invalid path error")
	}
}

func assertProjectDumpContent(t *testing.T, projectDir string) {
	t.Helper()

	projectContent := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "project-notes.md"))
	for _, want := range []string{
		`title: "Project Notes"`,
		`status: "raw"`,
		"# Project Notes",
		"Project-wide dump body.",
	} {
		if !strings.Contains(projectContent, want) {
			t.Fatalf("project dump missing %q:\n%s", want, projectContent)
		}
	}

	if strings.Contains(projectContent, "case:") {
		t.Fatalf("project dump has case field:\n%s", projectContent)
	}
}

func assertCaseDumpContent(t *testing.T, projectDir string) {
	t.Helper()

	caseContent := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "dumps", "blog-notes.md"))
	for _, want := range []string{
		`title: "Blog Notes"`,
		`case: "blog"`,
		"# Blog Notes",
		"Case-scoped dump body.",
	} {
		if !strings.Contains(caseContent, want) {
			t.Fatalf("case dump missing %q:\n%s", want, caseContent)
		}
	}
}

func findDumpViewByTitle(t *testing.T, dumps []DumpView, title string) DumpView {
	t.Helper()

	for _, dump := range dumps {
		if dump.Title == title {
			return dump
		}
	}

	t.Fatalf("dump %q not found in %#v", title, dumps)

	return DumpView{}
}

func readServiceWorkspaceFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	return string(content)
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

	if result.AuthoringSessions[0].Target != chc.AuthoringTargetSpec {
		t.Fatalf("session target = %q, want spec", result.AuthoringSessions[0].Target)
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

type fakeDumpEditorLauncher struct {
	path    string
	command DumpEditorCommand
	err     error
}

func (l *fakeDumpEditorLauncher) EditorCommand(_ context.Context, path string) (DumpEditorCommand, error) {
	l.path = path

	return l.command, l.err
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
