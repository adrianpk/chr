package core

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketIndexRebuilderRebuildsProjectTickets(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "nested/CHR-2-New.md", ticketcore.Ticket{
		Key:       "CHR-2",
		Title:     "New",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	err := index.Upsert(context.Background(), testIndexedTicket("CHR-1", nil))
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	result, err := ticketcore.NewRebuilder(
		index,
		ticketcore.WithRebuildClock(rebuildTestIndexedAt),
	).Rebuild(context.Background(), ticketcore.RebuildOptions{
		ProjectID:  "project-1",
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("RebuildProjectTickets returned error: %v", err)
	}

	if result.IndexedTickets != 1 {
		t.Fatalf("IndexedTickets = %d, want 1", result.IndexedTickets)
	}

	if len(result.ParseErrors) != 0 {
		t.Fatalf("ParseErrors = %#v, want none", result.ParseErrors)
	}

	tickets, err := index.ListProject(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	if len(tickets) != 1 {
		t.Fatalf("len(tickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHR-2" || tickets[0].RelativePath != "nested/CHR-2-New.md" {
		t.Fatalf("rebuilt ticket = %#v, want nested CHR-2", tickets[0])
	}
}

func TestTicketIndexRebuilderPreservesParseErrors(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHR-1-Good.md", ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})
	writeDiscoveryTestFile(t, projectDir, ".chr/default/tickets/CHR-2-Bad.md")

	result, err := ticketcore.NewRebuilder(
		index,
		ticketcore.WithRebuildClock(rebuildTestIndexedAt),
	).Rebuild(context.Background(), ticketcore.RebuildOptions{
		ProjectID:  "project-1",
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("RebuildProjectTickets returned error: %v", err)
	}

	if result.IndexedTickets != 1 {
		t.Fatalf("IndexedTickets = %d, want 1", result.IndexedTickets)
	}

	if len(result.ParseErrors) != 1 {
		t.Fatalf("len(ParseErrors) = %d, want 1", len(result.ParseErrors))
	}

	if result.ParseErrors[0].RelativePath != "CHR-2-Bad.md" {
		t.Fatalf("ParseErrors[0] = %#v, want CHR-2-Bad.md", result.ParseErrors[0])
	}
}

func writeRebuildTestTicket(t *testing.T, projectDir, relativePath string, ticket ticketcore.Ticket) {
	t.Helper()

	content, err := ticketcore.NewRenderer().Render(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	writeDiscoveryTestFile(t, projectDir, filepath.Join(".chr", "default", "tickets", relativePath))

	path := filepath.Join(projectDir, ".chr", "default", "tickets", filepath.FromSlash(relativePath))
	writeFile(t, path, content)
}

func rebuildTestCreatedAt() time.Time {
	return time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
}

func rebuildTestUpdatedAt() time.Time {
	return time.Date(2026, 5, 19, 10, 30, 0, 0, time.UTC)
}

func rebuildTestIndexedAt() time.Time {
	return time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC)
}
