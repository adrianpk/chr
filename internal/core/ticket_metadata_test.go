//nolint:wsl_v5 // The metadata round-trip test keeps setup and assertions adjacent.
package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketMetadataUpdaterUpdatesEditableFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHR-1.md")
	content, err := NewTicketRenderer().RenderTicket(Ticket{
		Key:       "CHR-1",
		Title:     "Old title",
		Synopsis:  "Old synopsis.",
		Status:    TicketStatusReadyForDev,
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		Body:      "# Comments\n",
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	err = NewTicketMetadataUpdater().UpdateMetadata(context.Background(), UpdateTicketMetadataOptions{
		Path:      path,
		Title:     "New title",
		Synopsis:  "New synopsis.",
		UpdatedAt: time.Date(2026, 6, 30, 8, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("UpdateMetadata returned error: %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`title: "New title"`,
		`synopsis: "New synopsis."`,
		`updated_at: "2026-06-30T08:00:00Z"`,
		"# Comments",
	} {
		if !strings.Contains(string(updated), want) {
			t.Fatalf("updated ticket does not contain %q:\n%s", want, updated)
		}
	}
}
