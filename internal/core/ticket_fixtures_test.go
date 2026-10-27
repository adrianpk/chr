package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func writeTicketFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket file: %v", err)
	}
}

func validIndexedTicket(now time.Time) ticketcore.Indexed {
	return ticketcore.Indexed{
		IndexedIdentity: ticketcore.IndexedIdentity{
			ProjectID:    "project-1",
			Path:         "/tmp/project/.chr/default/tickets/CHR-1-Test.md",
			RelativePath: "CHR-1-Test.md",
			Key:          "CHR-1",
		}, IndexedSummary: ticketcore.IndexedSummary{Title: "Test ticket"}, IndexedClassification: ticketcore.IndexedClassification{
			Status:   ticketcore.StatusBacklog,
			Priority: "normal",
			Severity: "normal",
			Type:     "task",
		}, IndexedTimes: ticketcore.IndexedTimes{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}
}
