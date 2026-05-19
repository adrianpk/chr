package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestTicketParserRejectsInvalidMetadata(t *testing.T) {
	createdAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 30, 0, 0, time.UTC)
	content := renderParserTestTicket(t, Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		Priority:  "high",
		Severity:  "critical",
		Type:      "feature",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing key",
			content: strings.Replace(content, `key: "CHC-1"`, `key: ""`, 1),
			wantErr: "ticket key is required",
		},
		{
			name:    "missing title",
			content: strings.Replace(content, `title: "Do something"`, `title: ""`, 1),
			wantErr: "ticket title is required",
		},
		{
			name:    "invalid status",
			content: strings.Replace(content, `status: "open"`, `status: "unknown"`, 1),
			wantErr: "invalid ticket status: unknown",
		},
		{
			name:    "invalid priority",
			content: strings.Replace(content, `priority: "high"`, `priority: "urgent"`, 1),
			wantErr: "invalid ticket priority: urgent",
		},
		{
			name:    "invalid severity",
			content: strings.Replace(content, `severity: "critical"`, `severity: "maximum"`, 1),
			wantErr: "invalid ticket severity: maximum",
		},
		{
			name:    "invalid type",
			content: strings.Replace(content, `type: "feature"`, `type: "chore"`, 1),
			wantErr: "invalid ticket type: chore",
		},
		{
			name: "missing created_at",
			content: strings.Replace(
				content,
				`created_at: "2026-05-18T12:30:00Z"`,
				`created_at: ""`,
				1,
			),
			wantErr: "ticket created_at is required",
		},
		{
			name: "missing updated_at",
			content: strings.Replace(
				content,
				`updated_at: "2026-05-18T13:30:00Z"`,
				`updated_at: ""`,
				1,
			),
			wantErr: "ticket updated_at is required",
		},
		{
			name: "updated before created",
			content: strings.Replace(
				content,
				`updated_at: "2026-05-18T13:30:00Z"`,
				`updated_at: "2026-05-18T11:30:00Z"`,
				1,
			),
			wantErr: "ticket updated_at cannot be before created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ticket.md")
			writeTicketFile(t, path, tt.content)

			_, err := NewTicketParser().ParseTicket(context.Background(), ParseTicketOptions{
				Path:         path,
				RelativeBase: filepath.Dir(path),
			})
			if err == nil {
				t.Fatal("ParseTicket returned nil error")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ParseTicket error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func renderParserTestTicket(t *testing.T, ticket Ticket) string {
	t.Helper()

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	return content
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
