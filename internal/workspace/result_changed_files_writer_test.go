package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketResultChangedFilesWriterWritesSummaryComment(t *testing.T) {
	path := writeResultChangedFilesTicket(t)
	application := testResultChangedFilesApplication(path)

	err := NewTicketResultChangedFilesWriter().WriteChangedFiles(context.Background(), WriteChangedFilesOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 16, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteChangedFiles returned error: %v", err)
	}

	content := readResultChangedFilesTicket(t, path)
	for _, want := range []string{
		`updated_at: "2026-05-19T16:00:00Z"`,
		"## 2026-05-19T16:00:00Z chc",
		"Execution exec-1 changed files:",
		"- internal/workspace/result_application.go: Added changed-file inputs.",
		"- docs/reference/ticket-result-application.md",
		"Existing comment.",
		"# Blocker",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketResultChangedFilesWriterSkipsEmptyChangedFiles(t *testing.T) {
	path := writeResultChangedFilesTicket(t)
	before := readResultChangedFilesTicket(t, path)
	application := testResultChangedFilesApplication(path)
	application.ChangedFiles = nil

	err := NewTicketResultChangedFilesWriter().WriteChangedFiles(context.Background(), WriteChangedFilesOptions{
		Application: application,
		CreatedAt:   time.Date(2026, 5, 19, 16, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("WriteChangedFiles returned error: %v", err)
	}

	after := readResultChangedFilesTicket(t, path)
	if after != before {
		t.Fatalf("empty changed files mutated ticket:\n%s", after)
	}
}

func TestTicketResultChangedFilesWriterRejectsMissingTicketPath(t *testing.T) {
	application := testResultChangedFilesApplication("")

	err := NewTicketResultChangedFilesWriter().WriteChangedFiles(context.Background(), WriteChangedFilesOptions{
		Application: application,
	})
	if err == nil {
		t.Fatal("WriteChangedFiles returned nil error")
	}
}

func writeResultChangedFilesTicket(t *testing.T) string {
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

	content = strings.Replace(content, "# Blocker", "## 2026-05-19T12:30:00Z user\n\n\"Existing comment.\"\n\n# Blocker", 1)

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return path
}

func readResultChangedFilesTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	return string(content)
}

func testResultChangedFilesApplication(path string) ResultApplication {
	return ResultApplication{
		ExecutionID: "exec-1",
		TicketPath:  path,
		Comment: ResultApplicationComment{
			Author: "chc",
		},
		ChangedFiles: []ResultApplicationChangedFile{
			{
				Path:    "internal/workspace/result_application.go",
				Summary: "Added changed-file inputs.",
			},
			{
				Path: "docs/reference/ticket-result-application.md",
			},
		},
	}
}
