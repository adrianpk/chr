package chc

import (
	"context"
	"path/filepath"
	"testing"
)

func TestTicketReconcilerRefreshesCleanProject(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHC-1-Good.md", Ticket{
		Key:       "CHC-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	result, err := NewTicketReconciler(index).ReconcileProjectTickets(
		context.Background(),
		ReconcileProjectTicketsOptions{
			ProjectID:    "project-1",
			ProjectDir:   projectDir,
			ReconciledAt: rebuildTestIndexedAt(),
		},
	)
	if err != nil {
		t.Fatalf("ReconcileProjectTickets returned error: %v", err)
	}

	if result.IndexedTickets != 1 {
		t.Fatalf("IndexedTickets = %d, want 1", result.IndexedTickets)
	}

	if result.HasBlockers() {
		t.Fatalf("Blockers = %#v, want none", result.Blockers)
	}

	ticket, err := index.GetProjectTicketByKey(context.Background(), "project-1", "CHC-1")
	if err != nil {
		t.Fatalf("GetProjectTicketByKey returned error: %v", err)
	}

	if ticket.RelativePath != "CHC-1-Good.md" {
		t.Fatalf("RelativePath = %q, want CHC-1-Good.md", ticket.RelativePath)
	}
}

func TestTicketReconcilerReturnsParseErrorBlockers(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeDiscoveryTestFile(t, projectDir, filepath.Join(".chc", "tickets", "CHC-1-Bad.md"))

	result, err := NewTicketReconciler(index).ReconcileProjectTickets(
		context.Background(),
		ReconcileProjectTicketsOptions{
			ProjectID:    "project-1",
			ProjectDir:   projectDir,
			ReconciledAt: rebuildTestIndexedAt(),
		},
	)
	if err != nil {
		t.Fatalf("ReconcileProjectTickets returned error: %v", err)
	}

	if result.IndexedTickets != 0 {
		t.Fatalf("IndexedTickets = %d, want 0", result.IndexedTickets)
	}

	if !result.HasBlockers() {
		t.Fatal("HasBlockers returned false")
	}

	blocker, found := result.FirstBlocker()
	if !found {
		t.Fatal("FirstBlocker returned false")
	}

	if blocker.Kind != TicketReconciliationBlockerParseError {
		t.Fatalf("Kind = %q, want %q", blocker.Kind, TicketReconciliationBlockerParseError)
	}

	if blocker.RelativePath != "CHC-1-Bad.md" {
		t.Fatalf("RelativePath = %q, want CHC-1-Bad.md", blocker.RelativePath)
	}

	if blocker.Err == nil {
		t.Fatal("Err is nil")
	}
}
