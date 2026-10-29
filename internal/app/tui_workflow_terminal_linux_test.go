package app

import (
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

// Production terminal keys browse retained state and confirm real stop controls.
// Resizing switches the Board presentation without submitting or moving work.
func TestTUIWorkflowTerminal(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)
	terminal := startTUITerminal(t, operations, "TestTUINavigationProcess")
	terminal.see(t, "Cases")

	if err := unix.IoctlSetWinsize(int(terminal.master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 44, Col: 160}); err != nil {
		t.Fatal(err)
	}

	terminal.keys(t, "\x1b[17~")
	terminal.see(t, "Board")
	terminal.see(t, "Attention")
	terminal.keys(t, "\x1b[18~")
	terminal.see(t, "Eligible:")
	terminal.keys(t, "w")
	terminal.see(t, "Retained workflows")
	terminal.see(t, "Invocation:")
	terminal.keys(t, "p")
	terminal.see(t, "Confirm pause")
	terminal.keys(t, "\r")
	terminal.see(t, "Workflow "+id+": paused.")
	terminal.keys(t, "u")
	terminal.see(t, "Confirm resume")
	terminal.keys(t, "\r")
	terminal.see(t, "Workflow "+id+": pending.")
	terminal.keys(t, "x")
	terminal.see(t, "Confirm cancel")
	terminal.keys(t, "\r")
	terminal.see(t, "Workflow "+id+": cancelled.")
	terminal.keys(t, "p")
	terminal.see(t, "terminal; inspect")
	terminal.stop(t)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Status != "cancelled" || status.TotalAttempts != 0 || status.IntegratedUnits != 0 || status.ConfirmedTickets != 0 {
		t.Fatal("terminal controls started or invented delivery", err, status.Status)
	}

	if !strings.Contains(strings.Join(tuiWorkflowView(status).Lines, "\n"), "confirmed tickets: 0") {
		t.Fatal("retained delivery counters missing")
	}
}
