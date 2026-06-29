package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketResultBlockedReasonWriterWritesBlocker(t *testing.T) {
	path := writeResultBlockedReasonTicket(t, "")
	application := testResultBlockedReasonApplication(path)

	err := NewTicketResultBlockedReasonWriter().WriteBlockedReason(context.Background(), WriteBlockedReasonOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 17, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteBlockedReason returned error: %v", err)
	}

	content := readResultBlockedReasonTicket(t, path)
	for _, want := range []string{
		`updated_at: "2026-05-19T17:00:00Z"`,
		"# Blocker\n\n## 2026-05-19T17:00:00Z exec-1",
		"- kind: harness",
		"- reason: harness: docker unavailable",
		"- details: see runtime artifacts",
		"# Validation",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketResultBlockedReasonWriterPreservesBlockerHistory(t *testing.T) {
	path := writeResultBlockedReasonTicket(t, "## 2026-05-19T16:00:00Z exec-0\n\n- kind: runtime\n")
	application := testResultBlockedReasonApplication(path)

	err := NewTicketResultBlockedReasonWriter().WriteBlockedReason(context.Background(), WriteBlockedReasonOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 17, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteBlockedReason returned error: %v", err)
	}

	content := readResultBlockedReasonTicket(t, path)
	for _, want := range []string{
		"## 2026-05-19T17:00:00Z exec-1",
		"## 2026-05-19T16:00:00Z exec-0",
		"- kind: runtime",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketResultBlockedReasonWriterSkipsReviewingApplication(t *testing.T) {
	path := writeResultBlockedReasonTicket(t, "")
	before := readResultBlockedReasonTicket(t, path)
	application := testResultBlockedReasonApplication(path)
	application.TicketStatus = TicketResultStatusReviewing

	err := NewTicketResultBlockedReasonWriter().WriteBlockedReason(context.Background(), WriteBlockedReasonOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 17, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteBlockedReason returned error: %v", err)
	}

	after := readResultBlockedReasonTicket(t, path)
	if after != before {
		t.Fatalf("reviewing application mutated ticket:\n%s", after)
	}
}

func TestTicketResultBlockedReasonWriterRejectsMissingTicketPath(t *testing.T) {
	application := testResultBlockedReasonApplication("")

	err := NewTicketResultBlockedReasonWriter().WriteBlockedReason(context.Background(), WriteBlockedReasonOptions{
		Application: application,
	})
	if err == nil {
		t.Fatal("WriteBlockedReason returned nil error")
	}
}

func writeResultBlockedReasonTicket(t *testing.T, blockerHistory string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ticket.md")
	ticket := Ticket{
		Key:       "CHR-1",
		Title:     "Do something",
		Status:    "blocked",
		CreatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
	}

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	if blockerHistory != "" {
		content = strings.Replace(content, "# Validation", blockerHistory+"\n# Validation", 1)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return path
}

func readResultBlockedReasonTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	return string(content)
}

func testResultBlockedReasonApplication(path string) ResultApplication {
	return ResultApplication{
		ExecutionID:  "exec-1",
		TicketPath:   path,
		TicketStatus: TicketResultStatusBlocked,
		BlockedReason: ResultApplicationBlockedReason{
			Kind:           ExecutionFailureHarness,
			Reason:         "harness: docker unavailable",
			ResultPath:     "/tmp/result.json",
			LogPath:        "/tmp/runtime.log",
			ValidationPath: "/tmp/validation.log",
		},
	}
}
