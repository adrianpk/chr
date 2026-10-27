package core

import (
	"context"
	"path/filepath"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketIndexStatusCheckerReportsCleanIndex(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHR-1-Good.md", ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})
	rebuildProjectTicketsForStatusTest(t, index, projectDir)

	status := checkProjectTicketIndexForStatusTest(t, index, projectDir)

	assertTicketIndexStatusCounts(t, status, 0, 0, 0, 0)
}

func TestTicketIndexStatusCheckerReportsMissingTicket(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHR-1-Good.md", ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	status := checkProjectTicketIndexForStatusTest(t, index, projectDir)

	assertTicketIndexStatusCounts(t, status, 1, 0, 0, 0)

	if status.Missing[0].RelativePath != "CHR-1-Good.md" {
		t.Fatalf("Missing[0] = %#v, want CHR-1-Good.md", status.Missing[0])
	}
}

func TestTicketIndexStatusCheckerReportsStaleTicket(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHR-1-Good.md", ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})
	rebuildProjectTicketsForStatusTest(t, index, projectDir)
	writeRebuildTestTicket(t, projectDir, "CHR-1-Good.md", ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Changed",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	status := checkProjectTicketIndexForStatusTest(t, index, projectDir)

	assertTicketIndexStatusCounts(t, status, 0, 1, 0, 0)

	if status.Stale[0].Indexed.Title != "Good" || status.Stale[0].Parsed.Title != "Changed" {
		t.Fatalf("Stale[0] = %#v, want Good -> Changed", status.Stale[0])
	}
}

func TestTicketIndexStatusCheckerReportsOrphanedTicket(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)

	err := index.UpsertTicket(context.Background(), testIndexedTicket("CHR-1", nil))
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	writeRebuildTestTicket(t, projectDir, "CHR-2-Good.md", ticketcore.Ticket{
		Key:       "CHR-2",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	status := checkProjectTicketIndexForStatusTest(t, index, projectDir)

	assertTicketIndexStatusCounts(t, status, 1, 0, 1, 0)
}

func TestTicketIndexStatusCheckerReportsInvalidTicket(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeDiscoveryTestFile(t, projectDir, filepath.Join(".chr", "default", "tickets", "CHR-1-Bad.md"))

	status := checkProjectTicketIndexForStatusTest(t, index, projectDir)

	assertTicketIndexStatusCounts(t, status, 0, 0, 0, 1)

	if status.Invalid[0].RelativePath != "CHR-1-Bad.md" {
		t.Fatalf("Invalid[0] = %#v, want CHR-1-Bad.md", status.Invalid[0])
	}
}

func rebuildProjectTicketsForStatusTest(t *testing.T, index TicketIndex, projectDir string) {
	t.Helper()

	_, err := NewTicketIndexRebuilder(
		index,
		WithTicketIndexRebuilderClock(rebuildTestIndexedAt),
	).RebuildProjectTickets(context.Background(), RebuildProjectTicketsOptions{
		ProjectID:  "project-1",
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("RebuildProjectTickets returned error: %v", err)
	}
}

func checkProjectTicketIndexForStatusTest(
	t *testing.T,
	index TicketIndex,
	projectDir string,
) TicketIndexStatus {
	t.Helper()

	status, err := NewTicketIndexStatusChecker(
		index,
		WithTicketIndexStatusCheckerClock(rebuildTestIndexedAt),
	).CheckProjectTicketIndex(context.Background(), CheckProjectTicketIndexOptions{
		ProjectID:  "project-1",
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("CheckProjectTicketIndex returned error: %v", err)
	}

	return status
}

func assertTicketIndexStatusCounts(
	t *testing.T,
	status TicketIndexStatus,
	missing int,
	stale int,
	orphaned int,
	invalid int,
) {
	t.Helper()

	if len(status.Missing) != missing {
		t.Fatalf("len(Missing) = %d, want %d", len(status.Missing), missing)
	}

	if len(status.Stale) != stale {
		t.Fatalf("len(Stale) = %d, want %d", len(status.Stale), stale)
	}

	if len(status.Orphaned) != orphaned {
		t.Fatalf("len(Orphaned) = %d, want %d", len(status.Orphaned), orphaned)
	}

	if len(status.Invalid) != invalid {
		t.Fatalf("len(Invalid) = %d, want %d", len(status.Invalid), invalid)
	}
}
