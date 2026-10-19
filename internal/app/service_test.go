package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func TestServiceRefreshWorkspaceBuildsGraph(t *testing.T) {
	projectDir := t.TempDir()

	err := os.MkdirAll(filepath.Dir(core.ProjectMetadataPath(projectDir)), 0o755)
	if err != nil {
		t.Fatalf("create .chr: %v", err)
	}

	err = os.WriteFile(
		core.ProjectMetadataPath(projectDir),
		[]byte(core.RenderProjectMetadata(core.ProjectMetadata{
			Name:        "Service Project",
			Description: "Loaded by service.",
		})),
		0o644,
	)
	if err != nil {
		t.Fatalf("write project metadata: %v", err)
	}

	env, err := chrTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chrTestUserEnvironment returned error: %v", err)
	}

	operations := newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
	}, testAssetsFS())

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

	result, err := operations.SaveProjectMetadata(core.ProjectMetadata{
		SchemaVersion: core.ProjectMetadataSchemaVersion,
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

	loaded, err := core.NewProjectMetadataService(projectDir).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Name != "Saved Project" {
		t.Fatalf("loaded.Name = %q, want Saved Project", loaded.Name)
	}
}

func TestNewQuizModelAdapterRoutesArtifactGenerationToCodexByDefault(t *testing.T) {
	adapter := newQuizModelAdapter(t.TempDir())

	routing, ok := adapter.(core.ActionRoutingQuizModelAdapter)
	if !ok {
		t.Fatalf("adapter = %T, want action routing adapter", adapter)
	}

	for _, action := range []core.QuizModelAction{
		core.QuizModelActionGenerateSpec,
		core.QuizModelActionGenerateAcceptanceFeature,
		core.QuizModelActionGeneratePlan,
		core.QuizModelActionMaterializePlanTracking,
	} {
		if _, ok := routing.ActionAdapters[action]; !ok {
			t.Fatalf("%s action is not routed: %#v", action, routing.ActionAdapters)
		}
	}
}

func TestNewQuizModelAdapterCanRouteArtifactGenerationToOpenAI(t *testing.T) {
	t.Setenv("CHR_SPEC_GENERATION_BACKEND", "openai")
	t.Setenv("CHR_ACCEPTANCE_FEATURE_GENERATION_BACKEND", "openai")
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")
	t.Setenv("CHR_PLAN_TRACKING_GENERATION_BACKEND", "openai")

	adapter := newQuizModelAdapter(t.TempDir())

	if _, ok := adapter.(core.OpenAIQuizModelAdapter); !ok {
		t.Fatalf("adapter = %T, want OpenAI adapter", adapter)
	}
}

func TestNewDumpCurationProviderUsesCodexByDefault(t *testing.T) {
	provider := newDumpCurationProvider(t.TempDir())

	if _, ok := provider.(harnessDumpCurationProvider); !ok {
		t.Fatalf("dump curation provider = %T, want codex provider", provider)
	}
}

func TestNewDumpCurationProviderCanUseOpenAI(t *testing.T) {
	t.Setenv("CHR_DUMP_GENERATION_BACKEND", "openai")

	provider := newDumpCurationProvider(t.TempDir())

	if _, ok := provider.(openAIDumpCurationProvider); !ok {
		t.Fatalf("dump curation provider = %T, want OpenAI provider", provider)
	}
}

func TestServiceSaveCaseMetadataReloadsCases(t *testing.T) {
	projectDir := t.TempDir()
	operations := newTestService(t, projectDir)

	result, err := operations.SaveCaseMetadata(core.CaseMetadata{
		SchemaVersion: core.CaseMetadataSchemaVersion,
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
	writeServiceDumpListFixture(t, projectDir)

	operations := newTestService(t, projectDir)

	result, err := operations.ListDumps()
	if err != nil {
		t.Fatalf("ListDumps returned error: %v", err)
	}

	assertServiceDumpListResult(t, operations, result)
}

func writeServiceDumpListFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "project-context.md"), `---
title: "Project Context"
status: "raw"
---

# Project Context
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "blog-context.md"), `---
title: "Blog Context"
status: "curated"
type: "dump"
source: "ops/default/dump/00000-blog-context-raw/artifact.md"
case: "blog"
---

# Blog Context
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "blog-context-raw.md"), `---
title: "Blog Context Raw Capture"
status: "raw"
type: "dump_raw"
curated: "ops/default/dump/00000-blog-context/artifact.md"
case: "blog"
---

# Blog Context Raw Capture
`)
}

func assertServiceDumpListResult(t *testing.T, operations service, result DumpListResult) {
	t.Helper()

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

	if !strings.Contains(projectDump.Content, "# Project Context") {
		t.Fatalf("project dump content = %q, want full dump content", projectDump.Content)
	}

	caseDump := findDumpViewByTitle(t, result.Dumps, "Blog Context")
	if caseDump.Scope.CaseSlug != "blog" {
		t.Fatalf("case dump scope = %#v, want blog", caseDump.Scope)
	}

	if caseDump.RelativePath != "ops/default/dump/00000-blog-context/artifact.md" {
		t.Fatalf("case dump path = %q, want ops/default/dump/00000-blog-context/artifact.md", caseDump.RelativePath)
	}

	if caseDump.SourcePath != "ops/default/dump/00000-blog-context-raw/artifact.md" {
		t.Fatalf("case dump source path = %q, want raw sibling path", caseDump.SourcePath)
	}

	if operations.dumpDirectory() != filepath.Join(operations.projectContext.ProjectDir, "ops", "default", "dump") {
		t.Fatalf("dumpDirectory = %q, want ops/default/dump", operations.dumpDirectory())
	}
}

func TestServiceCreateDumpWritesProjectAndCaseScopedDumps(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	operations := newTestService(t, projectDir)

	projectResult, err := operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(""),
		Title:       "Project Notes",
		Content:     "Project-wide dump body.",
	})
	if err != nil {
		t.Fatalf("CreateDump project returned error: %v", err)
	}

	caseResult, err := operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(""),
		Title:       "Blog Notes",
		Content:     "Case-scoped dump body.",
		Scope:       DumpScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("CreateDump case returned error: %v", err)
	}

	assertProjectDumpContent(t, projectDir)
	assertCaseDumpContent(t, projectDir)

	if len(projectResult.Dumps) != 1 {
		t.Fatalf("project result dumps = %#v, want curated dump only", projectResult.Dumps)
	}

	if len(caseResult.Dumps) != 2 {
		t.Fatalf("case result dumps = %#v, want curated dumps only for both captures", caseResult.Dumps)
	}

	caseDump := findDumpViewByTitle(t, caseResult.Dumps, "Blog Notes")
	if caseDump.Scope.CaseSlug != "blog" {
		t.Fatalf("case dump scope = %#v, want blog", caseDump.Scope)
	}
}

func TestServiceCreateDumpExtendsProjectDump(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	operations := newTestService(t, projectDir)

	first, err := operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(""),
		Title:       "Collision Notes",
		Content:     "First body.",
	})
	if err != nil {
		t.Fatalf("CreateDump first returned error: %v", err)
	}

	second, err := operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(""),
		Title:       "Collision Notes",
		Content:     "Second body.",
	})
	if err != nil {
		t.Fatalf("CreateDump second returned error: %v", err)
	}

	if first.RelativePath != second.RelativePath || len(second.Dumps) != 1 {
		t.Fatal("general capture created another logical dump", first, second)
	}

	observed, err := operations.ReadProjectDump()
	if err != nil || len(observed.Candidates) != 1 || len(observed.Components) != 2 {
		t.Fatal("raw captures lost ownership", observed, err)
	}

	content := second.Dumps[0].Content
	for _, text := range []string{"First body.", "Second body."} {
		if strings.Count(content, text) != 1 {
			t.Fatal("capture lost or duplicated material", content)
		}
	}
}

func TestServiceAppendAndOpenDump(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "project-notes.md"), `---
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

	appendResult, err := appendObservedDump(t, operations, DumpAppendRequest{
		RelativePath: "ops/default/dump/00000-project-notes/artifact.md",
		Content:      "Appended body.",
	})
	if err != nil {
		t.Fatalf("AppendDump returned error: %v", err)
	}

	content := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "project-notes.md"))
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
		RelativePath: "ops/default/dump/00000-project-notes/artifact.md",
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
		RelativePath: "ops/default/dump/00000-project-notes/artifact.md",
	})
	if err != nil {
		t.Fatalf("OpenDump returned error: %v", err)
	}

	if editor.path != filepath.Join(projectDir, filepath.FromSlash("ops/default/dump/00000-project-notes/artifact.md")) {
		t.Fatalf("editor path = %q", editor.path)
	}

	if openResult.Message != "Editing dump ops/default/dump/00000-project-notes/artifact.md." {
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
		OperationID: "invalid-path", ExpectedRevision: strings.Repeat("a", 64),
		RelativePath: ".chr/default/project.md",
		Content:      "bad",
	})
	if err == nil {
		t.Fatal("AppendDump returned nil error, want invalid path error")
	}
}

func assertProjectDumpContent(t *testing.T, projectDir string) {
	t.Helper()

	projectContent := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "project-notes.md"))
	for _, want := range []string{
		`title: "Project Notes"`,
		`status: "curated"`,
		`type: "dump"`,
		`source: "ops/default/dump/00000-project-notes-raw/artifact.md"`,
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

	rawContent := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "project-notes-raw.md"))
	for _, want := range []string{
		`status: "raw"`,
		`type: "dump_raw"`,
		`curated: "ops/default/dump/00000-project-notes/artifact.md"`,
		"## Highlights",
		"## Raw Capture",
		"Project-wide dump body.",
	} {
		if !strings.Contains(rawContent, want) {
			t.Fatalf("project raw dump missing %q:\n%s", want, rawContent)
		}
	}
}

func assertCaseDumpContent(t *testing.T, projectDir string) {
	t.Helper()

	caseContent := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "blog-notes.md"))
	for _, want := range []string{
		`title: "Blog Notes"`,
		`status: "curated"`,
		`type: "dump"`,
		`source: "ops/default/dump/00000-blog-notes-raw/artifact.md"`,
		`case: "blog"`,
		"# Blog Notes",
		"Case-scoped dump body.",
	} {
		if !strings.Contains(caseContent, want) {
			t.Fatalf("case dump missing %q:\n%s", want, caseContent)
		}
	}

	rawContent := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "dumps", "blog-notes-raw.md"))
	for _, want := range []string{
		`status: "raw"`,
		`type: "dump_raw"`,
		`curated: "ops/default/dump/00000-blog-notes/artifact.md"`,
		`case: "blog"`,
		"Case-scoped dump body.",
	} {
		if !strings.Contains(rawContent, want) {
			t.Fatalf("case raw dump missing %q:\n%s", want, rawContent)
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

	path = canonicalServiceWorkspaceFixturePath(path)

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	return string(content)
}

func assertServiceWorkspacePathMissing(t *testing.T, projectDir, relativePath string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))

	_, err := os.Stat(path)
	if err == nil {
		t.Fatalf("path exists, want missing: %s", path)
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat returned error %v, want not exists", err)
	}
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

	resolved, err := operations.ResolveWorkspacePath("ops/default/spec/00000-blog-spec/artifact.md")
	if err != nil {
		t.Fatalf("ResolveWorkspacePath returned error: %v", err)
	}

	if !resolved.Found || resolved.Node.Title != "Blog Spec" {
		t.Fatalf("resolved = %#v, want Blog Spec", resolved)
	}

	related, err := operations.RelatedWorkspaceDocuments(core.WorkspaceNodeID(resolved.Node))
	if err != nil {
		t.Fatalf("RelatedWorkspaceDocuments returned error: %v", err)
	}

	if !related.Resolved || len(related.Related) != 1 || related.Related[0].Kind != core.WorkspaceNodeTicket {
		t.Fatalf("related = %#v, want related ticket", related)
	}
}

func TestServiceSearchWorkspaceIndexUsesSemanticDigestCache(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	graph, err := core.NewWorkspaceGraphBuilder(projectDir).Build(t.Context())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	index := core.NewWorkspaceIndexFromGraph(graph)
	cache := core.NewWorkspaceSemanticDigestCache()

	for nodeID, node := range index.NodesByID {
		if node.Kind == core.WorkspaceNodePlan {
			cache.Entries[nodeID] = core.WorkspaceSemanticDigest{
				NodeID: nodeID,
				Digest: "editorial roadmap publishing workflow",
			}
		}
	}

	err = core.WriteWorkspaceSemanticDigestCache(core.WorkspaceSemanticDigestCachePath(projectDir), cache)
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

	if result.Tickets[0].Status != core.TicketStatusBacklog {
		t.Fatalf("ticket status = %q, want backlog", result.Tickets[0].Status)
	}

	if result.Tickets[0].Synopsis != "Ticket created from the TUI Board action." {
		t.Fatalf("ticket synopsis = %q, want TUI default synopsis", result.Tickets[0].Synopsis)
	}

	if result.AuthoringSessions == nil {
		t.Fatal("AuthoringSessions = nil, want reloaded empty slice")
	}
}

func TestServiceBoardReturnsSharedBoardReadModel(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Backlog.md", core.Ticket{
		Key:       "CHR-1",
		Title:     "Backlog work",
		Synopsis:  "Prepare backlog work.",
		Status:    core.TicketStatusBacklog,
		Priority:  "normal",
		Severity:  "normal",
		Type:      "task",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	operations := newTestService(t, projectDir)

	board, err := operations.Board()
	if err != nil {
		t.Fatalf("Board returned error: %v", err)
	}

	if board.TotalCount != 1 || board.NormalCount != 1 {
		t.Fatalf("board counts = %#v, want one normal ticket", board)
	}

	if len(board.Tickets) != 1 || board.Tickets[0].Key != "CHR-1" {
		t.Fatalf("board tickets = %#v, want CHR-1", board.Tickets)
	}
}

func TestServiceCreateTUIAuthoringSessionReloadsSessions(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := os.MkdirAll(filepath.Join(projectDir, ".chr", "default", "tickets"), 0o755)
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

	if result.AuthoringSessions[0].Mode != core.AuthoringSessionModeTUI {
		t.Fatalf("session mode = %q, want tui", result.AuthoringSessions[0].Mode)
	}

	if result.AuthoringSessions[0].Target != core.AuthoringTargetSpec {
		t.Fatalf("session target = %q, want spec", result.AuthoringSessions[0].Target)
	}

	if result.Tickets == nil {
		t.Fatal("Tickets = nil, want reloaded empty slice")
	}
}

func TestServiceMutatesSelectedTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHR-1-Do-something.md", core.Ticket{
		Key:       "CHR-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	})
	operations := newTestService(t, projectDir)

	commentResult, err := operations.AddSelectedTicketComment("CHR-1")
	if err != nil {
		t.Fatalf("AddSelectedTicketComment returned error: %v", err)
	}

	if commentResult.Message != "Added comment to ticket CHR-1." {
		t.Fatalf("comment message = %q", commentResult.Message)
	}

	statusResult, err := operations.MoveSelectedTicketStatus("CHR-1", "ready")
	if err != nil {
		t.Fatalf("MoveSelectedTicketStatus returned error: %v", err)
	}

	if statusResult.Message != "Moved ticket CHR-1 to ready." {
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

func TestServiceUpsertTicketEvidenceStrategyChecklistPreservesTicketContent(t *testing.T) {
	projectDir := newTestWorkspace(t)
	ticketPath := writeServiceEvidenceStrategyTicket(t, projectDir)
	operations := newTestService(t, projectDir)

	result, err := operations.UpsertTicketEvidenceStrategyChecklist(TicketEvidenceStrategyChecklistRequest{
		Key:       "CHR-1",
		Checklist: serviceTestEvidenceStrategyChecklist(),
	})
	if err != nil {
		t.Fatalf("UpsertTicketEvidenceStrategyChecklist returned error: %v", err)
	}

	if result.Message != "Updated evidence strategy checklist for ticket CHR-1." {
		t.Fatalf("Message = %q", result.Message)
	}

	if len(result.Tickets) != 1 || result.Tickets[0].Key != "CHR-1" {
		t.Fatalf("Tickets = %#v, want reloaded CHR-1", result.Tickets)
	}

	content := readServiceWorkspaceFile(t, ticketPath)
	assertServiceEvidenceStrategyTicketContent(t, content)

	assertServiceEvidenceStrategyValidationPending(t, content)
}

func TestServiceRequestDaemonStopReturnsStoppedMessage(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := os.MkdirAll(filepath.Join(projectDir, ".chr", "default", "tickets"), 0o755)
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

func serviceTestEvidenceStrategyChecklist() core.EvidenceStrategyChecklist {
	return core.EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]core.EvidenceStrategyChecklistItem{
			core.EvidenceStrategyChecklistLabelStrategy: {
				Label:  core.EvidenceStrategyChecklistLabelStrategy,
				Status: "planned",
				Detail: "change_kind=bugfix; risk=medium; test_timing=before; note=regression",
			},
			core.EvidenceStrategyChecklistLabelGherkin: {
				Label:  core.EvidenceStrategyChecklistLabelGherkin,
				Status: "required",
				Detail: "path=.chr/default/acceptance/bug.feature; reason=user-visible",
			},
			core.EvidenceStrategyChecklistLabelTests: {
				Label:  core.EvidenceStrategyChecklistLabelTests,
				Status: "mixed",
				Detail: "reason=unit and fuzz coverage",
			},
			core.EvidenceStrategyChecklistLabelValidation: {
				Label:  core.EvidenceStrategyChecklistLabelValidation,
				Status: "pending",
				Detail: "commands=make verify; evidence=none",
			},
			core.EvidenceStrategyChecklistLabelSummary: {
				Label:  core.EvidenceStrategyChecklistLabelSummary,
				Status: "pending",
			},
		},
	}
}

func writeServiceEvidenceStrategyTicket(t *testing.T, projectDir string) string {
	t.Helper()

	return writeTicketMarkdownOnly(t, projectDir, "CHR-1-Do-something.md", core.Ticket{
		Key:       "CHR-1",
		Title:     "Do something",
		Synopsis:  "Preserve ticket body while adding evidence strategy.",
		Status:    core.TicketStatusReadyForDev,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
		Body: `# Description

Keep this description.

# Comments

Keep comments.

# Blocker

Keep blockers.

# Validation

Keep validation notes.
`,
	})
}

func assertServiceEvidenceStrategyTicketContent(t *testing.T, content string) {
	t.Helper()

	assertServiceTextOrder(t, content, core.EvidenceStrategyChecklistStartAnchor, "# Description")

	for _, want := range []string{
		"Keep this description.",
		"# Comments\n\nKeep comments.",
		"# Blocker\n\nKeep blockers.",
		"# Validation\n\nKeep validation notes.",
		`updated_at: "`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content missing %q:\n%s", want, content)
		}
	}
}

func assertServiceEvidenceStrategyValidationPending(t *testing.T, content string) {
	t.Helper()

	checklist, err := core.ParseEvidenceStrategyChecklist(content)
	if err != nil {
		t.Fatalf("ParseEvidenceStrategyChecklist returned error: %v", err)
	}

	validation, ok := checklist.Item(core.EvidenceStrategyChecklistLabelValidation)
	if !ok {
		t.Fatal("validation checklist item missing")
	}

	if validation.Status != "pending" {
		t.Fatalf("validation status = %q, want pending", validation.Status)
	}
}

func assertServiceTextOrder(t *testing.T, content, before, after string) {
	t.Helper()

	beforeIndex := strings.Index(content, before)
	if beforeIndex < 0 {
		t.Fatalf("content missing %q:\n%s", before, content)
	}

	afterIndex := strings.Index(content, after)
	if afterIndex < 0 {
		t.Fatalf("content missing %q:\n%s", after, content)
	}

	if beforeIndex >= afterIndex {
		t.Fatalf("%q appears after %q in:\n%s", before, after, content)
	}
}

func TestServiceOpenSourceUsesProjectRelativeFile(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1-example.md"), "# Ticket\n")

	operations := newTestService(t, projectDir)
	editor := &fakeDumpEditorLauncher{
		command: DumpEditorCommand{
			Program: "test-editor",
			Args:    []string{"--wait"},
			Dir:     projectDir,
		},
	}
	operations.dumpEditor = editor

	result, err := operations.OpenSource(".chr/default/tickets/CHR-1-example.md")
	if err != nil {
		t.Fatalf("OpenSource returned error: %v", err)
	}

	if editor.path != filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1-example.md") {
		t.Fatalf("editor path = %q", editor.path)
	}

	if result.EditorCommand == nil {
		t.Fatal("EditorCommand = nil, want editor command")
	}
}

func newTestService(t *testing.T, projectDir string) service {
	t.Helper()

	env, err := chrTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chrTestUserEnvironment returned error: %v", err)
	}

	operations := newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
	}, testAssetsFS())
	operations.dumpCurator = fallbackDumpCurationProvider{Reason: "test fallback"}

	return operations
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

func assertWorkspaceIndexCacheMiss(t *testing.T, cache core.WorkspaceIndexCacheState) {
	t.Helper()

	if cache.Status != core.WorkspaceIndexCacheMiss {
		t.Fatalf("IndexCache.Status = %q, want %q", cache.Status, core.WorkspaceIndexCacheMiss)
	}
}

func writeServiceWorkspaceQueryFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeServiceWorkspaceFile(t, core.ProjectMetadataPath(projectDir), core.RenderProjectMetadata(core.ProjectMetadata{
		Name:        "Example",
		Description: "Example project.",
	}))
	writeServiceWorkspaceFile(t, core.CaseMetadataPath(projectDir, "blog"), core.RenderCaseMetadata(core.CaseMetadata{
		Name:        "Blog Publishing",
		Description: "Publish articles.",
	}))
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, core.DirectoryName, core.UserDirectoryName, "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "accepted"
case: "blog"
---

# Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, core.DirectoryName, core.UserDirectoryName, "plans", "project-plan.md"), `---
title: "Project Plan"
status: "ready"
---

# Project Plan
`)

	ticket := core.Ticket{
		Key:         "CHR-1",
		Title:       "Write article",
		Status:      "ready",
		Priority:    "normal",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "blog",
		ProjectArea: "content",
		CreatedAt:   testTime(),
		UpdatedAt:   testTime(),
	}
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Write.md", ticket)

	content, err := core.NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "ops", "default", "ticket", "00000-chr-1-write", "artifact.md"), content)
}

func writeServiceWorkspaceFile(t *testing.T, path, content string) {
	t.Helper()

	path = canonicalServiceWorkspaceFixturePath(path)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	root := path
	for range 5 {
		root = filepath.Dir(root)
	}

	if kind, ok := core.OpsArtifactTypeForBodyPath(root, path); ok && kind == "dump" {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}

		if err := core.EnsureOpsArtifactBodyMetadata(root, filepath.ToSlash(relative)); err != nil {
			t.Fatal(err)
		}
	}
}

func canonicalServiceWorkspaceFixturePath(path string) string {
	marker := filepath.Join(core.DirectoryName, core.UserDirectoryName)

	parts := strings.Split(filepath.ToSlash(path), filepath.ToSlash(marker)+"/")
	if len(parts) != 2 {
		return path
	}

	artifactTypes := map[string]string{
		"acceptance": "spec",
		"dumps":      "dump",
		"quizzes":    "quiz",
		"specs":      "spec",
		"plans":      "plan",
		"trackers":   "tracker",
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
	if legacyParts[0] == "acceptance" {
		slug = strings.TrimSuffix(slug, "-feature-round") + "-acceptance-feature-round"
		if filepath.Ext(path) == ".feature" {
			slug = strings.TrimSuffix(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), "-feature") + "-acceptance-feature"
		}
	}

	artifactID := "00000-" + opsArtifactSlugForServiceFixture(slug)

	return filepath.Join(projectDir, core.OpsDirectoryName, core.DefaultOpsNamespace, artifactType, artifactID, core.OpsArtifactBodyFile())
}

func opsArtifactSlugForServiceFixture(value string) string {
	slug := dumpSlug(value)
	if slug == "" {
		return "artifact"
	}

	return slug
}

func testTime() time.Time {
	return time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
}
