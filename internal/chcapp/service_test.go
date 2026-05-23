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

func testTime() time.Time {
	return time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
}
