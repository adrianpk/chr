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

func TestTicketRepositoryCreatesProjectWiseTicket(t *testing.T) {
	dir := t.TempDir()
	repository := newTestTicketRepository(
		dir,
		ticketcore.WithTicketRepositoryClock(fixedTicketTime),
	)

	result, err := repository.CreateTicket(context.Background(), ticketcore.CreateTicketOptions{
		Title:    "Do something",
		Synopsis: "Do the implementation work.",
		Priority: "high",
		Tags:     []string{"uxui"},
	})
	if err != nil {
		t.Fatalf("CreateTicket returned error: %v", err)
	}

	wantPath := filepath.Join(dir, ".chr", "default", "tickets", "CHR-1-Do-something.md")
	if result.Path != wantPath {
		t.Fatalf("Path = %q, want %q", result.Path, wantPath)
	}

	content := readFile(t, wantPath)
	for _, want := range []string{
		`key: "CHR-1"`,
		`title: "Do something"`,
		`synopsis: "Do the implementation work."`,
		`status: "backlog"`,
		`priority: "high"`,
		`tags: ["uxui"]`,
		`created_at: "2026-05-18T12:30:00Z"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketRepositoryCreatesUseCaseTicket(t *testing.T) {
	dir := t.TempDir()
	repository := newTestTicketRepository(dir, ticketcore.WithTicketRepositoryClock(fixedTicketTime))

	result, err := repository.CreateTicket(context.Background(), ticketcore.CreateTicketOptions{
		Title:    "Add setup flow",
		Synopsis: "Add the onboarding setup flow.",
		UseCase:  "User Onboarding",
	})
	if err != nil {
		t.Fatalf("CreateTicket returned error: %v", err)
	}

	wantPath := filepath.Join(dir, ".chr", "default", "tickets", "User-Onboarding", "CHR-1-Add-setup-flow.md")
	if result.Path != wantPath {
		t.Fatalf("Path = %q, want %q", result.Path, wantPath)
	}
}

func TestTicketRepositoryDerivesNextKeyFromExistingTickets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".chr", "default", "tickets", "CHR-3-Old.md"), "# Old\n")
	writeFile(t, filepath.Join(dir, ".chr", "default", "tickets", "nested", "ticket.md"), "---\nkey: \"CHR-9\"\n---\n")

	repository := newTestTicketRepository(dir, ticketcore.WithTicketRepositoryClock(fixedTicketTime))

	result, err := repository.CreateTicket(context.Background(), ticketcore.CreateTicketOptions{
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
		ticketcore.WithTicketRepositoryKeyPrefix("PRJ"),
		ticketcore.WithTicketRepositoryClock(fixedTicketTime),
	)

	result, err := repository.CreateTicket(context.Background(), ticketcore.CreateTicketOptions{
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

func TestTicketRepositoryRejectsMissingSynopsis(t *testing.T) {
	dir := t.TempDir()
	repository := ticketcore.NewTicketRepository(dir, func(context.Context, string) (int, error) {
		t.Fatal("missing synopsis reached the key allocator")

		return 0, nil
	}, ticketcore.WithTicketRepositoryClock(fixedTicketTime))

	_, err := repository.CreateTicket(context.Background(), ticketcore.CreateTicketOptions{Title: "Next"})
	if err == nil {
		t.Fatal("CreateTicket returned nil error")
	}
}

func fixedTicketTime() time.Time {
	return time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}

	return string(content)
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

func newTestTicketRepository(projectDir string, opts ...ticketcore.TicketRepositoryOption) ticketcore.TicketRepository {
	return ticketcore.NewTicketRepository(projectDir, func(ctx context.Context, prefix string) (int, error) {
		return NextTicketKeyNumber(ctx, projectDir, prefix)
	}, opts...)
}
