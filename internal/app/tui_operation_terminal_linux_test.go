package app

import (
	"maps"
	"path/filepath"
	"testing"
)

// Actual terminal Escape must ask before dropping edited text. Cancelling and
// then explicitly confirming that discard leave all canonical sources unchanged.
func TestTUIOperationTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)
	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal := startTUITerminal(t, operations, "TestTUINavigationProcess")
	terminal.see(t, "Alpha navigation")
	terminal.keys(t, "\x1bOQ")
	terminal.see(t, "Alpha navigation captured material")
	terminal.keys(t, "\r")
	terminal.see(t, "Editing ")
	terminal.keys(t, "TERMINAL_UNSENT_SENTINEL")
	terminal.see(t, "TERMINAL_UNSENT_SENTINEL")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Discard unsaved dump text?")
	terminal.keys(t, "\x1b")
	terminal.see(t, "TERMINAL_UNSENT_SENTINEL")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Discard unsaved dump text?")
	terminal.keys(t, "\r")
	terminal.see(t, "Draft discarded.")
	terminal.stop(t)

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("terminal draft cancellation changed canonical files")
	}
}
