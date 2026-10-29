package app

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// Production terminal input creates complete forms, appends to the captured
// dump and round-trips unsaved external edits before an explicit canonical Save.
func TestTUIFormsTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)

	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\nprintf '\\nEXTERNAL_ROUND_TRIP\\n' >> \"$1\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", shellQuote(editor))

	terminal := startTUITerminal(t, operations, "TestTUIFormsProcess")
	terminal.see(t, "Alpha navigation")
	cancelTerminalCaseForm(t, terminal, operations)
	completeTerminalCaseForm(t, terminal)
	captureTerminalDump(t, terminal)
	terminal.keys(t, "a")
	terminal.see(t, "Append to Terminal dump")
	terminal.keys(t, "TERMINAL_APPEND_TEXT")
	terminal.keys(t, "\x13")
	terminal.see(t, "Appended dump")
	terminal.keys(t, "\r")
	terminal.see(t, "Title and scope are retained")
	terminal.keys(t, " UNSENT_BEFORE_EDITOR")
	terminal.keys(t, "\x1be")
	terminal.see(t, "External edits retained; Ctrl+S saves")

	before := findDumpViewByTitle(t, mustListDumps(t, operations), "Terminal dump")
	if strings.Contains(before.Content, "UNSENT_BEFORE_EDITOR") || strings.Contains(before.Content, "EXTERNAL_ROUND_TRIP") {
		t.Fatal("editor wrote canonical content before Save")
	}

	terminal.keys(t, "\x13")
	terminal.see(t, "Saved dump")
	terminal.stop(t)
	checkTerminalFormDocuments(t, operations)
}

func cancelTerminalCaseForm(t *testing.T, terminal *guidedTerminal, operations service) {
	t.Helper()

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal.keys(t, "n")
	terminal.see(t, "Create Case")
	terminal.keys(t, "CANCELLED_CASE")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Discard the unsubmitted form?")
	terminal.keys(t, "\r")
	terminal.see(t, "Form discarded.")

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("cancelled case form wrote canonical files")
	}
}

func completeTerminalCaseForm(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "n")
	terminal.see(t, "Create Case")
	terminal.keys(t, "\t")
	terminal.keys(t, "Description from terminal.")
	terminal.keys(t, "\t")
	terminal.keys(t, "Context from terminal.")
	resizeTUIParityTerminal(t, terminal, 80, 24)
	terminal.see(t, "Context from terminal.")
	terminal.keys(t, " AFTER_RESIZE")
	terminal.see(t, "Context from terminal. AFTER_RESIZE")
	resizeTUIParityTerminal(t, terminal, 120, 44)
	terminal.see(t, "Context from terminal. AFTER_RESIZE")
	terminal.keys(t, "\x13")
	terminal.see(t, "Operation: validation_failed")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Context from terminal.")
	terminal.keys(t, "\x1b[Z\x1b[Z")
	terminal.keys(t, "Form journey")
	terminal.keys(t, "\x13")
	terminal.see(t, "Case metadata saved.")
}

func captureTerminalDump(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "\x1bOQ")
	terminal.keys(t, "c")
	terminal.see(t, "Create Dump")
	terminal.see(t, "Scope: Form journey")
	terminal.keys(t, "\t")
	terminal.keys(t, "Terminal dump")
	terminal.keys(t, "\t")
	terminal.keys(t, "TERMINAL_DUMP_BODY")
	terminal.keys(t, "\x13")
	terminal.see(t, "TERMINAL_DUMP_BODY")
	terminal.see(t, "Captured dump")
}

func checkTerminalFormDocuments(t *testing.T, operations service) {
	t.Helper()

	metadata, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadPath(authoringcore.CaseMetadataPath(operations.projectContext.ProjectDir, "form-journey"))
	if err != nil || metadata.Description != "Description from terminal." || metadata.Context != "Context from terminal. AFTER_RESIZE" {
		t.Fatal(metadata, err)
	}

	dump := findDumpViewByTitle(t, mustListDumps(t, operations), "Terminal dump")
	for _, text := range []string{"TERMINAL_DUMP_BODY", "TERMINAL_APPEND_TEXT", "UNSENT_BEFORE_EDITOR", "EXTERNAL_ROUND_TRIP"} {
		if !strings.Contains(dump.Content, text) {
			t.Fatal("terminal input was lost", text, dump.Content)
		}
	}

	if dump.Scope.CaseSlug != "form-journey" {
		t.Fatal("terminal form lost its scope", dump)
	}
}

// A real concurrent writer must reject terminal Save. Cancelled reload retains
// input and canonical bytes; confirmed reload replaces it before the next Save.
func TestTUIFormsStaleTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)
	initial := findDumpViewByTitle(t, mustListDumps(t, operations), "Alpha navigation notes")
	terminal := startTUITerminal(t, operations, "TestTUIFormsProcess")
	terminal.see(t, "Alpha navigation")
	terminal.keys(t, "\x1bOQ")
	terminal.see(t, "Alpha navigation captured material")
	terminal.keys(t, "\r")
	terminal.see(t, "Title and scope are retained")
	terminal.keys(t, "UNSENT_STALE_TERMINAL")

	if _, err := operations.SaveDump(DumpSaveRequest{OperationID: "concurrent-terminal", ExpectedRevision: initial.Revision, RelativePath: initial.RelativePath, Content: "CONCURRENT_TERMINAL_BODY"}); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal.keys(t, "\x13")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "r")
	terminal.see(t, "Reload this dump and discard its unsaved text?")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "\x1b")
	terminal.see(t, "UNSENT_STALE_TERMINAL")

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("failed save or cancelled reload changed canonical bytes")
	}

	confirmTerminalDumpReload(t, terminal)
	terminal.keys(t, " FRESH_AFTER_RELOAD")
	terminal.keys(t, "\x13")
	terminal.see(t, "Saved dump")
	terminal.stop(t)

	final, err := operations.ReadDump(DumpReadRequest{RelativePath: initial.RelativePath})
	if err != nil || !strings.Contains(final.Content, "CONCURRENT_TERMINAL_BODY") || !strings.Contains(final.Content, "FRESH_AFTER_RELOAD") || strings.Contains(final.Content, "UNSENT_STALE_TERMINAL") {
		t.Fatal(final, err)
	}
}

func confirmTerminalDumpReload(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "\x1bo")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "r")
	terminal.see(t, "Reload this dump and discard its unsaved text?")
	terminal.keys(t, "\r")
	terminal.see(t, "CONCURRENT_TERMINAL_BODY")
}

// The child keeps the production root loader and host. Only the curation model
// boundary is deterministic; terminal, metadata, journal and filesystem IO are real.
func TestTUIFormsProcess(t *testing.T) {
	project := os.Getenv("CHR_GUIDED_TUI_PROJECT")
	if project == "" {
		t.Skip("launched by the owned form terminal fixture")
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

	operations := newService(t.Context(), env, projectContext, testAssetsFS())
	operations.dumpCurator = &countingDumpCurator{}
	operations.generationBinder = testGenerationBinder

	options.DumpExecutor = newTUIDumpActionExecutor(operations)
	if err = runManagedTUI(t.Context(), options, operations); err != nil {
		t.Fatal(err)
	}
}
