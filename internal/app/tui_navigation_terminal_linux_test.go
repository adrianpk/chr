package app

import (
	"maps"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// The production loader and terminal loop must preserve the chosen case through
// document tabs, contextual Help and legacy views without changing source files.
func TestTUINavigationTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)
	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal := startTUITerminal(t, operations, "TestTUINavigationProcess")
	terminal.see(t, "Alpha navigation")
	terminal.keys(t, "\x1b[B")
	terminal.keys(t, "\r")
	terminal.see(t, "Case: Beta navigation")
	terminal.keys(t, "\x1bOQ")
	terminal.see(t, "Beta navigation captured material")
	terminal.keys(t, "?")
	terminal.see(t, "Help: Dumps")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Beta navigation captured material")
	terminal.keys(t, "\x1bp")
	terminal.see(t, "Harness defaults")
	terminal.keys(t, ":")
	terminal.see(t, "Command Palette")
	terminal.keys(t, "view:runtime")
	terminal.see(t, "Go to Runtime")
	terminal.keys(t, "\r")
	terminal.see(t, "Workspace Index")
	terminal.keys(t, "\x1bOP")
	terminal.see(t, "Case: Beta navigation")
	terminal.stop(t)

	if after := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops")); !maps.Equal(before, after) {
		t.Fatal("navigation changed canonical case or dump documents")
	}
}

func navigationTerminalFixture(t *testing.T) service {
	t.Helper()

	operations := newTestService(t, newTestWorkspace(t))
	if err := os.MkdirAll(ticketcore.Directory(operations.projectContext.ProjectDir), 0o755); err != nil {
		t.Fatal(err)
	}

	for _, item := range []struct{ slug, title string }{{"alpha-navigation", "Alpha navigation"}, {"beta-navigation", "Beta navigation"}} {
		if _, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).Save(authoringcore.CaseMetadata{Slug: item.slug, Name: item.title}); err != nil {
			t.Fatal(err)
		}

		if _, err := operations.CreateDump(DumpCreateRequest{OperationID: "navigation-" + item.slug, Title: item.title + " notes", Content: item.title + " captured material", Scope: DumpScope{CaseSlug: item.slug}}); err != nil {
			t.Fatal(err)
		}
	}

	return operations
}

// This child uses the same loader as chr's root command, not a hand-built
// projection; all providers and workers remain idle throughout navigation.
func TestTUINavigationProcess(t *testing.T) {
	project := os.Getenv("CHR_GUIDED_TUI_PROJECT")
	if project == "" {
		t.Skip("launched by the owned navigation terminal fixture")
	}

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: os.Getenv("CHR_GUIDED_TUI_ROOT")})
	if err != nil {
		t.Fatal(err)
	}

	projectContext := commandProjectContext{ProjectDir: project, ProjectID: os.Getenv("CHR_GUIDED_TUI_ID")}
	factory := commandFactory{deps: testAppDependencies(), io: IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}}

	options, err := factory.loadTUIOptions(t.Context(), env, projectContext)
	if err != nil {
		t.Fatal(err)
	}

	if err := runManagedTUI(t.Context(), options, newService(t.Context(), env, projectContext, testAssetsFS())); err != nil {
		t.Fatal(err)
	}
}
