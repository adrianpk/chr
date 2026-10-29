package app

import (
	"maps"
	"path/filepath"
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

// Real terminal keys inspect a native preview or retained invocation. Cancelling
// preview and polling status leave ticket bytes, attempts and delivery unchanged.
func TestTUIDashNativeTerminal(t *testing.T) {
	for _, mode := range []string{"preview", "workflow"} {
		t.Run(mode, func(t *testing.T) {
			operations, hierarchy, _ := workflowCLIFixture(t)
			if mode == "workflow" {
				workflowCLIManualStart(t, operations, hierarchy)
			}

			before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
			terminal := startTUITerminal(t, operations, "TestTUINavigationProcess")
			terminal.see(t, "Cases")
			terminal.keys(t, "\x1b[18~")
			terminal.see(t, "Eligible:")

			if mode == "preview" {
				terminal.keys(t, "v")
				terminal.see(t, "Confirm native workflow extent")
				terminal.see(t, "Policy:")
				terminal.keys(t, "\x1b")
				terminal.see(t, "Work action cancelled.")
			} else {
				terminal.keys(t, "w")
				terminal.see(t, "Invocation:")
				terminal.keys(t, "\x1bm")
				terminal.see(t, "confirmed tickets: 0")
			}

			terminal.stop(t)

			if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
				t.Fatal("terminal inspection changed native work")
			}
		})
	}
}
