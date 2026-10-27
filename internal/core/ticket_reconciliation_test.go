package core

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketReconcilerRefreshesCleanProject(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHR-1-Good.md", ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	result, err := ticketcore.NewReconciler(index).ReconcileProject(
		context.Background(),
		ticketcore.ReconcileProjectOptions{
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

	ticket, err := index.GetProjectByKey(context.Background(), "project-1", "CHR-1")
	if err != nil {
		t.Fatalf("GetProjectTicketByKey returned error: %v", err)
	}

	if ticket.RelativePath != "CHR-1-Good.md" {
		t.Fatalf("RelativePath = %q, want CHR-1-Good.md", ticket.RelativePath)
	}
}

func TestTicketReconcilerReturnsParseErrorBlockers(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeDiscoveryTestFile(t, projectDir, filepath.Join(".chr", "default", "tickets", "CHR-1-Bad.md"))

	result, err := ticketcore.NewReconciler(index).ReconcileProject(
		context.Background(),
		ticketcore.ReconcileProjectOptions{
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

	if blocker.Kind != ticketcore.ReconcileBlockerParseError {
		t.Fatalf("Kind = %q, want %q", blocker.Kind, ticketcore.ReconcileBlockerParseError)
	}

	if blocker.RelativePath != "CHR-1-Bad.md" {
		t.Fatalf("RelativePath = %q, want CHR-1-Bad.md", blocker.RelativePath)
	}

	if blocker.Err == nil {
		t.Fatal("Err is nil")
	}
}

func TestTicketReconcilerRefreshesAffectedTicket(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHR-1-Good.md", ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	reconcileAllProjectTickets(t, index, projectDir)

	ticketPath := filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1-Good.md")

	err := NewTicketResultStateWriter().WriteResultState(context.Background(), WriteResultStateOptions{
		Application: ResultApplication{
			TicketPath:   ticketPath,
			TicketStatus: TicketResultStatusReviewing,
		},
		UpdatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteResultState returned error: %v", err)
	}

	result, err := ticketcore.NewReconciler(index).ReconcileTicket(
		context.Background(),
		ticketcore.ReconcileTicketOptions{
			ProjectID:    "project-1",
			ProjectDir:   projectDir,
			TicketPath:   ticketPath,
			ReconciledAt: time.Date(2026, 5, 19, 12, 5, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("ReconcileProjectTicket returned error: %v", err)
	}

	if result.IndexedTickets != 1 {
		t.Fatalf("IndexedTickets = %d, want 1", result.IndexedTickets)
	}

	if result.HasBlockers() {
		t.Fatalf("Blockers = %#v, want none", result.Blockers)
	}

	ticket, err := index.GetProjectByKey(context.Background(), "project-1", "CHR-1")
	if err != nil {
		t.Fatalf("GetProjectTicketByKey returned error: %v", err)
	}

	if ticket.Status != TicketResultStatusReviewing {
		t.Fatalf("Status = %q, want reviewing", ticket.Status)
	}

	if !ticket.UpdatedAt.Equal(time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("UpdatedAt = %s, want result write time", ticket.UpdatedAt.Format(time.RFC3339))
	}
}

func TestTicketReconcilerSurfacesAffectedTicketParseBlocker(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	ticketPath := filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1-Bad.md")
	writeFile(t, ticketPath, "missing frontmatter")

	result, err := ticketcore.NewReconciler(index).ReconcileTicket(
		context.Background(),
		ticketcore.ReconcileTicketOptions{
			ProjectID:    "project-1",
			ProjectDir:   projectDir,
			TicketPath:   ticketPath,
			ReconciledAt: rebuildTestIndexedAt(),
		},
	)
	if err != nil {
		t.Fatalf("ReconcileProjectTicket returned error: %v", err)
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

	if blocker.RelativePath != "CHR-1-Bad.md" {
		t.Fatalf("RelativePath = %q, want CHR-1-Bad.md", blocker.RelativePath)
	}
}

func reconcileAllProjectTickets(t *testing.T, index ticketcore.Index, projectDir string) {
	t.Helper()

	_, err := ticketcore.NewReconciler(index).ReconcileProject(
		context.Background(),
		ticketcore.ReconcileProjectOptions{
			ProjectID:    "project-1",
			ProjectDir:   projectDir,
			ReconciledAt: rebuildTestIndexedAt(),
		},
	)
	if err != nil {
		t.Fatalf("ReconcileProjectTickets returned error: %v", err)
	}
}
