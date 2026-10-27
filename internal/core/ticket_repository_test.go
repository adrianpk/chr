package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketRepositoryDerivesNextKeyFromExistingTickets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".chr", "default", "tickets", "CHR-3-Old.md"), "# Old\n")
	writeFile(t, filepath.Join(dir, ".chr", "default", "tickets", "nested", "ticket.md"), "---\nkey: \"CHR-9\"\n---\n")

	repository := newTestTicketRepository(dir, ticketcore.WithRepositoryClock(fixedTicketTime))

	result, err := repository.Create(context.Background(), ticketcore.CreateOptions{
		Title:    "Next",
		Synopsis: "Create the next ticket.",
	})
	if err != nil {
		t.Fatalf("CreateTicket returned error: %v", err)
	}

	if result.Key != "CHR-10" {
		t.Fatalf("Key = %q, want CHR-10", result.Key)
	}
}

func TestTicketRepositoryUsesCustomKeyPrefix(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".chr", "default", "tickets", "PRJ-4-Old.md"), "# Old\n")

	repository := newTestTicketRepository(
		dir,
		ticketcore.WithKeyPrefix("PRJ"),
		ticketcore.WithRepositoryClock(fixedTicketTime),
	)

	result, err := repository.Create(context.Background(), ticketcore.CreateOptions{
		Title:    "Next",
		Synopsis: "Create the next ticket.",
	})
	if err != nil {
		t.Fatalf("CreateTicket returned error: %v", err)
	}

	if result.Key != "PRJ-5" {
		t.Fatalf("Key = %q, want PRJ-5", result.Key)
	}
}

func fixedTicketTime() time.Time {
	return time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create parent dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func newTestTicketRepository(projectDir string, opts ...ticketcore.RepositoryOption) ticketcore.Repository {
	return ticketcore.NewRepository(projectDir, func(ctx context.Context, prefix string) (int, error) {
		return NextTicketKeyNumber(ctx, projectDir, prefix)
	}, opts...)
}
