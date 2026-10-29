package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Real production terminal input creates and edits a ticket, enters distinct
// comments and a review reference, then opens the actual canonical source.
func TestTUITicketTerminal(t *testing.T) {
	operations := navigationTerminalFixture(t)

	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\nsed -i 's/TERMINAL_TICKET_BODY/TERMINAL_TICKET_BODY EXTERNAL_TICKET_EDIT/' \"$1\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", shellQuote(editor))

	terminal := startTUITerminal(t, operations, "TestTUINavigationProcess")
	terminal.see(t, "Alpha navigation")
	terminal.keys(t, "\x1b[17~")
	terminal.see(t, "Tickets")
	terminal.keys(t, "c")
	terminal.see(t, "Create Ticket")
	terminal.keys(t, "Terminal ticket\tTERMINAL_TICKET_BODY")
	terminal.keys(t, "\x13")
	terminal.see(t, "Ticket saved.")
	terminal.keys(t, "e")
	terminal.see(t, "Edit Ticket")
	terminal.keys(t, "\x1b[F revised")
	terminal.keys(t, "\x13")
	terminal.see(t, "Ticket saved.")
	terminalTicketComments(t, terminal)
	terminal.keys(t, "v")
	terminal.see(t, "Attach Review Reference")
	terminal.keys(t, "\tforgejo\thttps://forge.example.org/team/project/pulls/88\t88\topen\tfeat/terminal")
	terminal.keys(t, "\x13")
	terminal.see(t, "Ticket saved.")
	terminal.keys(t, "\x1be")
	terminal.see(t, "Workspace refreshed")
	terminal.stop(t)
	checkTerminalTicket(t, operations)
}

func terminalTicketComments(t *testing.T, terminal *guidedTerminal) {
	t.Helper()

	for _, comment := range []string{"FIRST_TERMINAL_COMMENT", "SECOND_TERMINAL_COMMENT"} {
		terminal.keys(t, "a")
		terminal.see(t, "Ticket Discussion")
		terminal.keys(t, comment)
		terminal.keys(t, "\x13")
		terminal.see(t, "Ticket saved.")
	}

	terminal.keys(t, "a")
	terminal.see(t, "Ticket Discussion")
	terminal.keys(t, "\x1b[Z")
	terminal.see(t, "FIRST_TERMINAL_COMMENT")
	terminal.keys(t, strings.Repeat("\x1b[B", 10))
	terminal.see(t, "SECOND_TERMINAL_COMMENT")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Form discarded.")
}

func checkTerminalTicket(t *testing.T, operations service) {
	t.Helper()

	board, err := operations.ReadTicketBoard(0)
	if err != nil || len(board.Tickets) != 1 {
		t.Fatal(board, err)
	}

	document, err := operations.ReadTicketDocument(board.Tickets[0].Key)
	if err != nil || document.Fields["title"] != "Terminal ticket revised" || document.Fields["use_case"] != "alpha-navigation" || len(document.Ticket.ReviewRefs) != 1 {
		t.Fatal(document.Fields, err)
	}

	if !strings.Contains(document.Description, "TERMINAL_TICKET_BODY") || !strings.Contains(document.Description, "EXTERNAL_TICKET_EDIT") {
		t.Fatal("canonical source was not opened", document.Description)
	}

	checkTerminalTicketComments(t, document.Discussion)
}

func checkTerminalTicketComments(t *testing.T, discussion string) {
	t.Helper()

	for _, comment := range []string{"FIRST_TERMINAL_COMMENT", "SECOND_TERMINAL_COMMENT"} {
		if strings.Count(discussion, comment) != 1 {
			t.Fatal("terminal comment lost or duplicated", discussion)
		}
	}
}
