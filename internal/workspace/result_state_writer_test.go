package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketResultStateWriterWritesReviewingState(t *testing.T) {
	path := writeResultStateTicket(t, "open")
	application := testResultStateApplication(path, TicketResultStatusReviewing)

	err := NewTicketResultStateWriter().WriteResultState(context.Background(), WriteResultStateOptions{
		Application: application,
		UpdatedAt:   time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteResultState returned error: %v", err)
	}

	content := readResultStateTicket(t, path)
	for _, want := range []string{
		`status: "reviewing"`,
		`updated_at: "2026-05-19T13:00:00Z"`,
		`title: "Do something"`,
		"## 2026-05-19T12:30:00Z chc",
		`"Existing comment."`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketResultStateWriterWritesBlockedState(t *testing.T) {
	path := writeResultStateTicket(t, "in_progress")
	application := testResultStateApplication(path, TicketResultStatusBlocked)

	err := NewTicketResultStateWriter().WriteResultState(context.Background(), WriteResultStateOptions{
		Application: application,
		UpdatedAt:   time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteResultState returned error: %v", err)
	}

	content := readResultStateTicket(t, path)
	if !strings.Contains(content, `status: "blocked"`) {
		t.Fatalf("content does not contain blocked status:\n%s", content)
	}
}

func TestTicketResultStateWriterRejectsMissingTicketPath(t *testing.T) {
	application := testResultStateApplication("", TicketResultStatusReviewing)

	err := NewTicketResultStateWriter().WriteResultState(context.Background(), WriteResultStateOptions{
		Application: application,
	})
	if err == nil {
		t.Fatal("WriteResultState returned nil error")
	}
}

func writeResultStateTicket(t *testing.T, status string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ticket.md")
	ticket := Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    status,
		CreatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
	}

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	content += "\n## 2026-05-19T12:30:00Z chc\n\n\"Existing comment.\"\n"

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return path
}

func readResultStateTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	return string(content)
}

func testResultStateApplication(path, status string) ResultApplication {
	return ResultApplication{
		ExecutionID:  "exec-1",
		ProjectID:    "project-1",
		TicketKey:    "CHC-1",
		TicketPath:   path,
		TicketStatus: status,
	}
}
