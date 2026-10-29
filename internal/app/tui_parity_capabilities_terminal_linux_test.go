package app

import (
	"maps"
	"path/filepath"
	"testing"
)

// The production palette explains both unsupported writers before confirmation
// and leaves canonical files unchanged, including when no ticket is selected.
func TestTUIParityCapabilities(t *testing.T) {
	operations := navigationTerminalFixture(t)
	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal := startTUITerminal(t, operations, "TestTUINavigationProcess")
	terminal.see(t, "Alpha navigation")

	for _, id := range []string{"ticket-checklist-toggle", "ticket-checklist-add"} {
		terminal.keys(t, ":")
		terminal.see(t, "Command Palette")
		terminal.keys(t, id)
		terminal.see(t, "(unavailable)")
		terminal.keys(t, "\r")
		terminal.see(t, "Operation: blocked")
		terminal.see(t, "No safe checklist item writer is available.")
		terminal.keys(t, "\r\x1b")
		terminal.see(t, "Alpha navigation")
	}

	terminal.stop(t)

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("unavailable palette action changed canonical files")
	}
}
