package core

import (
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestSelectTicketBranchUsesExistingMetadata(t *testing.T) {
	selection, err := SelectTicketBranch(ticketcore.Indexed{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-1"}, IndexedSummary: ticketcore.IndexedSummary{Title: "Build CLI game"}, IndexedSource: ticketcore.IndexedSource{Branch: "feat/slice-1-cli-game"}})
	if err != nil {
		t.Fatalf("SelectTicketBranch returned error: %v", err)
	}

	if selection.Branch != "feat/slice-1-cli-game" {
		t.Fatalf("Branch = %q, want existing branch", selection.Branch)
	}

	if selection.Derived {
		t.Fatal("Derived = true, want false for existing branch")
	}
}

func TestSelectTicketBranchDerivesStableBranch(t *testing.T) {
	selection, err := SelectTicketBranch(ticketcore.Indexed{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-42"}, IndexedSummary: ticketcore.IndexedSummary{Title: "Add CLI tic-tac-toe!"}})
	if err != nil {
		t.Fatalf("SelectTicketBranch returned error: %v", err)
	}

	if selection.Branch != "chr/chr-42-add-cli-tic-tac-toe" {
		t.Fatalf("Branch = %q, want derived branch", selection.Branch)
	}

	if !selection.Derived {
		t.Fatal("Derived = false, want true")
	}
}

func TestSelectTicketBranchRejectsMissingKey(t *testing.T) {
	_, err := SelectTicketBranch(ticketcore.Indexed{IndexedSummary: ticketcore.IndexedSummary{Title: "Missing key"}})
	if err == nil {
		t.Fatal("SelectTicketBranch returned nil error")
	}
}
