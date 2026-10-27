package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketResultCommentWriterWritesExecutionComment(t *testing.T) {
	path := writeResultCommentTicket(t)
	application := testResultCommentApplication(path)

	err := NewTicketResultCommentWriter().WriteResultComment(context.Background(), WriteResultCommentOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteResultComment returned error: %v", err)
	}

	content := readResultCommentTicket(t, path)
	for _, want := range []string{
		`updated_at: "2026-05-19T14:00:00Z"`,
		"## 2026-05-19T14:00:00Z chr",
		"Execution `exec-1` finished with status `succeeded`.",
		"Summary: ready for review",
		"Existing comment.",
		"# Blocker",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketResultCommentWriterWritesFailureReason(t *testing.T) {
	path := writeResultCommentTicket(t)
	application := testResultCommentApplication(path)
	application.Execution.Status = ExecutionStatusFailed
	application.Execution.FailureReason = "harness: docker unavailable"
	application.Comment.Body = RenderResultApplicationCommentBody(application)

	err := NewTicketResultCommentWriter().WriteResultComment(context.Background(), WriteResultCommentOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteResultComment returned error: %v", err)
	}

	content := readResultCommentTicket(t, path)
	if strings.Contains(content, "harness: docker unavailable") {
		t.Fatalf("content exposes runtime failure reason:\n%s", content)
	}

	if !strings.Contains(content, "Execution details are available in the runtime artifacts.") {
		t.Fatalf("content does not contain runtime artifact pointer:\n%s", content)
	}
}

func TestTicketResultCommentWriterRejectsMissingTicketPath(t *testing.T) {
	application := testResultCommentApplication("")

	err := NewTicketResultCommentWriter().WriteResultComment(context.Background(), WriteResultCommentOptions{
		Application: application,
	})
	if err == nil {
		t.Fatal("WriteResultComment returned nil error")
	}
}

func writeResultCommentTicket(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ticket.md")
	ticket := ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Do something",
		Status:    "reviewing",
		CreatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
	}

	content, err := ticketcore.NewRenderer().Render(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	content = strings.Replace(content, "# Blocker", "## 2026-05-19T12:30:00Z user\n\n\"Existing comment.\"\n\n# Blocker", 1)

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return path
}

func readResultCommentTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	return string(content)
}

func testResultCommentApplication(path string) ResultApplication {
	application := ResultApplication{
		ExecutionID: "exec-1",
		TicketPath:  path,
		Execution: ResultApplicationExecution{
			Status: ExecutionStatusSucceeded,
		},
		Summary: "ready for review",
		Evidence: ResultApplicationEvidence{
			ResultPath:     "/tmp/result.json",
			LogPath:        "/tmp/runtime.log",
			ValidationPath: "/tmp/validation.log",
		},
	}
	application.Comment = ResultApplicationComment{
		Author: "chr",
		Body:   RenderResultApplicationCommentBody(application),
	}

	return application
}
