package app

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// Production PTY input creates a manual spec, retains every field on conflict,
// cancels reload without writes, then explicitly reloads before a fresh Save.
func TestTUISpecTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)
	terminal := startTUITerminal(t, operations, "TestTUIFormsProcess")
	terminal.see(t, "Alpha navigation")
	terminal.keys(t, "\x1bOS")
	terminal.see(t, "c create  e edit")
	cancelTerminalSpecForm(t, terminal, operations)
	createTerminalSpecForm(t, terminal)

	documents, err := operations.ListSpecDocuments()
	checkSpecDocumentCount(t, documents, err, 1)

	initial := documents[0]
	if initial.Title != "Terminal contract" || initial.Body != "SPEC_TERMINAL_BODY" || initial.Status != "approved" || initial.CaseSlug != "alpha-navigation" {
		t.Fatal(initial)
	}

	terminal.keys(t, "e")
	terminal.see(t, "Scope (retained): Alpha navigation")
	terminal.keys(t, "\t")
	terminal.keys(t, "\x1b[F revised")
	terminal.keys(t, "\t")
	terminal.keys(t, "UNSENT_SPEC_TERMINAL")

	concurrent := SaveSpecRequest{OperationID: "concurrent-spec-terminal", Reference: initial.Reference, Revision: initial.Revision, Title: "Concurrent title", Body: "CONCURRENT_SPEC_BODY", Status: "draft", CaseSlug: initial.CaseSlug}
	if _, err := operations.SaveSpec(concurrent); err != nil {
		t.Fatal(err)
	}

	staleTerminalSpecSave(t, terminal, operations)
	reloadTerminalSpec(t, terminal)
	terminal.keys(t, "\t\t")
	terminal.keys(t, "\x1b[F FRESH_SPEC_BODY")
	terminal.keys(t, "\x13")
	terminal.see(t, "Spec saved.")
	terminal.stop(t)
	checkTerminalSpecSave(t, operations, initial.Reference)
}

func cancelTerminalSpecForm(t *testing.T, terminal *guidedTerminal, operations service) {
	t.Helper()

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal.keys(t, "c")
	terminal.see(t, "Create Spec")
	terminal.keys(t, "\tCANCELLED_SPEC")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Discard the unsubmitted form?")
	terminal.keys(t, "\r")
	terminal.see(t, "Form discarded.")

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("cancelled creation wrote a spec")
	}
}

func createTerminalSpecForm(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "c")
	terminal.see(t, "Create Spec")
	terminal.see(t, "Scope: Alpha navigation")
	terminal.keys(t, "\tTerminal contract\tSPEC_TERMINAL_BODY\t\x1b[C")
	terminal.keys(t, "\x13")
	terminal.see(t, "Spec saved.")
}

func staleTerminalSpecSave(t *testing.T, terminal *guidedTerminal, operations service) {
	t.Helper()

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal.keys(t, "\x13")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "r")
	terminal.see(t, "Reload this spec and discard its unsaved fields?")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "\x1b")
	terminal.see(t, "UNSENT_SPEC_TERMINAL")

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("stale Save or cancelled reload wrote canonical files")
	}
}

func reloadTerminalSpec(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "\x1bo")
	terminal.see(t, "Operation: stale")
	terminal.keys(t, "r")
	terminal.see(t, "Reload this spec and discard its unsaved fields?")
	terminal.keys(t, "\r")
	terminal.see(t, "Title: Concurrent title")
	terminal.see(t, "Body: CONCURRENT_SPEC_BODY")
}

func checkTerminalSpecSave(t *testing.T, operations service, reference string) {
	t.Helper()

	after, err := authoringcore.ReadSpecDocument(t.Context(), operations.projectContext.ProjectDir, reference)
	if err != nil || after.Title != "Concurrent title" || after.Status != "draft" || after.CaseSlug != "alpha-navigation" {
		t.Fatal(after, err)
	}

	if !strings.Contains(after.Body, "CONCURRENT_SPEC_BODY") || !strings.Contains(after.Body, "FRESH_SPEC_BODY") || strings.Contains(after.Body, "UNSENT_SPEC_TERMINAL") {
		t.Fatal("reload/save lost or silently retained older text", after.Body)
	}
}
