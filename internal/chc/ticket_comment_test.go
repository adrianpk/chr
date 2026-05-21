package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketCommenterAddsComment(t *testing.T) {
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

	err = NewTicketCommenter().AddComment(context.Background(), AddTicketCommentOptions{
		Path:      path,
		Author:    "adrian",
		Body:      "Looks ready.",
		CreatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("AddComment returned error: %v", err)
	}

	updatedContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`updated_at: "2026-05-18T13:00:00Z"`,
		"# Comments\n\n## 2026-05-18T13:00:00Z adrian\n\n\"Looks ready.\"",
		"# Blocker",
	} {
		if !strings.Contains(string(updatedContent), want) {
			t.Fatalf("updated content does not contain %q:\n%s", want, string(updatedContent))
		}
	}
}

func TestTicketCommenterRejectsMissingCommentBody(t *testing.T) {
	err := NewTicketCommenter().AddComment(context.Background(), AddTicketCommentOptions{
		Path: filepath.Join(t.TempDir(), "ticket.md"),
	})
	if err == nil {
		t.Fatal("AddComment returned nil error")
	}
}
