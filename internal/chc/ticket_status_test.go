package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketStatusUpdaterUpdatesStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ticket.md")
	createdAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)

	content, err := NewTicketRenderer().RenderTicket(Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	err = NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:      path,
		Status:    "ready",
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	updatedContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`status: "ready"`,
		`updated_at: "2026-05-18T13:00:00Z"`,
	} {
		if !strings.Contains(string(updatedContent), want) {
			t.Fatalf("updated content does not contain %q:\n%s", want, string(updatedContent))
		}
	}
}

func TestTicketStatusUpdaterRejectsInvalidStatus(t *testing.T) {
	err := NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:   filepath.Join(t.TempDir(), "ticket.md"),
		Status: "unknown",
	})
	if err == nil {
		t.Fatal("UpdateStatus returned nil error")
	}
}
