package app

import (
	"maps"
	"testing"
)

// The production loader, SQLite records and real PTY expose retained results
// and counters without starting a worker or changing canonical operations.
func TestTUIDashTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "work:\n  state_authority: tracker_compat\n  mode: no_git_enforcement\n")
	seedObservedWorkResults(t, operations)

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	terminal := startTUITerminal(t, operations, "TestTUINavigationProcess")
	terminal.see(t, "Alpha navigation")
	terminal.keys(t, "\x1b[18~")
	terminal.see(t, "Eligible: 0 / 0 queued")
	terminal.keys(t, "\t\t")
	terminal.keys(t, "\x1bm")
	terminal.see(t, "Harness succeeded")
	terminal.see(t, "1 processed / 0 blocked / 1 failed")
	terminal.see(t, "No recent activity.")
	terminal.stop(t)

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("passive Dash changed canonical files")
	}
}
