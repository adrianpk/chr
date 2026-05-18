package chc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTicketParserParsesRenderedTicket(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".chc", "tickets")
	path := filepath.Join(ticketsDir, "onboarding", "CHC-1-Do-something.md")
	createdAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 30, 0, 0, time.UTC)
	indexedAt := time.Date(2026, 5, 18, 14, 30, 0, 0, time.UTC)

	content, err := NewTicketRenderer().RenderTicket(Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "critical",
		Type:        "feature",
		Tags:        []string{"uxui", "onboarding"},
		UseCase:     "onboarding",
		ProjectArea: "frontend",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	writeTicketFile(t, path, content)

	ticket, err := NewTicketParser().ParseTicket(context.Background(), ParseTicketOptions{
		ProjectID:  "project-1",
		ProjectDir: dir,
		TicketsDir: ticketsDir,
		Path:       path,
		IndexedAt:  indexedAt,
	})
	if err != nil {
		t.Fatalf("ParseTicket returned error: %v", err)
	}

	assertParsedTicketMetadata(t, ticket)
	assertParsedTicketTimes(t, ticket, createdAt, updatedAt)
	assertParsedTicketFileState(t, ticket)
}

func TestTicketParserRejectsMissingFrontmatter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ticket.md")
	writeTicketFile(t, path, "# Description\n")

	_, err := NewTicketParser().ParseTicket(context.Background(), ParseTicketOptions{Path: path})
	if err == nil {
		t.Fatal("ParseTicket returned nil error")
	}
}

func writeTicketFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket file: %v", err)
	}
}

func assertParsedTicketMetadata(t *testing.T, ticket IndexedTicket) {
	t.Helper()

	if ticket.ProjectID != "project-1" {
		t.Fatalf("ProjectID = %q, want project-1", ticket.ProjectID)
	}

	if ticket.RelativePath != "onboarding/CHC-1-Do-something.md" {
		t.Fatalf("RelativePath = %q, want onboarding/CHC-1-Do-something.md", ticket.RelativePath)
	}

	if ticket.Key != "CHC-1" || ticket.Title != "Do something" || ticket.Status != "open" {
		t.Fatalf("parsed metadata = %#v", ticket)
	}

	if len(ticket.Tags) != 2 || ticket.Tags[0] != "uxui" || ticket.Tags[1] != "onboarding" {
		t.Fatalf("Tags = %#v, want uxui and onboarding", ticket.Tags)
	}
}

func assertParsedTicketTimes(t *testing.T, ticket IndexedTicket, createdAt, updatedAt time.Time) {
	t.Helper()

	if ticket.CreatedAt == nil || !ticket.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %v, want %s", ticket.CreatedAt, createdAt)
	}

	if ticket.UpdatedAt == nil || !ticket.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("UpdatedAt = %v, want %s", ticket.UpdatedAt, updatedAt)
	}
}

func assertParsedTicketFileState(t *testing.T, ticket IndexedTicket) {
	t.Helper()

	if ticket.ContentHash == "" {
		t.Fatal("ContentHash is empty")
	}

	if ticket.FileSize == 0 {
		t.Fatal("FileSize is zero")
	}
}
