package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketResultValidationWriterWritesEvidence(t *testing.T) {
	path := writeResultValidationTicket(t, "")
	application := testResultValidationApplication(path)

	err := NewTicketResultValidationWriter().WriteResultValidation(context.Background(), WriteResultValidationOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 15, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteResultValidation returned error: %v", err)
	}

	content := readResultValidationTicket(t, path)
	for _, want := range []string{
		`updated_at: "2026-05-19T15:00:00Z"`,
		"# Validation\n\n## 2026-05-19T15:00:00Z exec-1",
		"- command: make verify",
		"- result: /tmp/result.json",
		"- log: /tmp/runtime.log",
		"- validation: /tmp/validation.log",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketResultValidationWriterPreservesValidationHistory(t *testing.T) {
	path := writeResultValidationTicket(t, "## 2026-05-19T14:00:00Z exec-0\n\n- command: go test ./...\n")
	application := testResultValidationApplication(path)

	err := NewTicketResultValidationWriter().WriteResultValidation(context.Background(), WriteResultValidationOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 15, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteResultValidation returned error: %v", err)
	}

	content := readResultValidationTicket(t, path)
	for _, want := range []string{
		"## 2026-05-19T15:00:00Z exec-1",
		"## 2026-05-19T14:00:00Z exec-0",
		"- command: go test ./...",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketResultValidationWriterRejectsMissingTicketPath(t *testing.T) {
	application := testResultValidationApplication("")

	err := NewTicketResultValidationWriter().WriteResultValidation(context.Background(), WriteResultValidationOptions{
		Application: application,
	})
	if err == nil {
		t.Fatal("WriteResultValidation returned nil error")
	}
}

func writeResultValidationTicket(t *testing.T, validationHistory string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ticket.md")
	ticket := Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "reviewing",
		CreatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
	}

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	if validationHistory != "" {
		content += "\n" + validationHistory
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return path
}

func readResultValidationTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	return string(content)
}

func testResultValidationApplication(path string) ResultApplication {
	return ResultApplication{
		ExecutionID: "exec-1",
		TicketPath:  path,
		Evidence: ResultApplicationEvidence{
			ValidationCommand: "make verify",
			ResultPath:        "/tmp/result.json",
			LogPath:           "/tmp/runtime.log",
			ValidationPath:    "/tmp/validation.log",
		},
	}
}
