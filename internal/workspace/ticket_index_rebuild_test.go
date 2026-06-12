package workspace

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestTicketIndexRebuilderRebuildsProjectTickets(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "nested/CHC-2-New.md", Ticket{
		Key:       "CHC-2",
		Title:     "New",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})

	err := index.UpsertTicket(context.Background(), testIndexedTicket("CHC-1", nil))
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	result, err := NewTicketIndexRebuilder(
		index,
		WithTicketIndexRebuilderClock(rebuildTestIndexedAt),
	).RebuildProjectTickets(context.Background(), RebuildProjectTicketsOptions{
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

	tickets, err := index.ListProjectTickets(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	if len(tickets) != 1 {
		t.Fatalf("len(tickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHC-2" || tickets[0].RelativePath != "nested/CHC-2-New.md" {
		t.Fatalf("rebuilt ticket = %#v, want nested CHC-2", tickets[0])
	}
}

func TestTicketIndexRebuilderPreservesParseErrors(t *testing.T) {
	projectDir := t.TempDir()
	index := newTestTicketIndex(t)
	writeRebuildTestTicket(t, projectDir, "CHC-1-Good.md", Ticket{
		Key:       "CHC-1",
		Title:     "Good",
		Status:    "open",
		CreatedAt: rebuildTestCreatedAt(),
		UpdatedAt: rebuildTestUpdatedAt(),
	})
	writeDiscoveryTestFile(t, projectDir, ".chc/user/tickets/CHC-2-Bad.md")

	result, err := NewTicketIndexRebuilder(
		index,
		WithTicketIndexRebuilderClock(rebuildTestIndexedAt),
	).RebuildProjectTickets(context.Background(), RebuildProjectTicketsOptions{
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

	if result.ParseErrors[0].RelativePath != "CHC-2-Bad.md" {
		t.Fatalf("ParseErrors[0] = %#v, want CHC-2-Bad.md", result.ParseErrors[0])
	}
}

func writeRebuildTestTicket(t *testing.T, projectDir, relativePath string, ticket Ticket) {
	t.Helper()

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	writeDiscoveryTestFile(t, projectDir, filepath.Join(".chc", "user", "tickets", relativePath))

	path := filepath.Join(projectDir, ".chc", "user", "tickets", filepath.FromSlash(relativePath))
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
