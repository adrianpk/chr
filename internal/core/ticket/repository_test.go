package ticket_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketRepositoryCreatesProjectWiseTicket(t *testing.T) {
	dir := t.TempDir()
	repository := newTestTicketRepository(
		dir,
		ticketcore.WithRepositoryClock(fixedTicketTime),
	)

	result, err := repository.Create(context.Background(), ticketcore.CreateOptions{
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
	repository := newTestTicketRepository(dir, ticketcore.WithRepositoryClock(fixedTicketTime))

	result, err := repository.Create(context.Background(), ticketcore.CreateOptions{
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

func TestTicketRepositoryRejectsMissingSynopsis(t *testing.T) {
	dir := t.TempDir()
	repository := ticketcore.NewRepository(dir, func(context.Context, string) (int, error) {
		t.Fatal("missing synopsis reached the key allocator")

		return 0, nil
	}, ticketcore.WithRepositoryClock(fixedTicketTime))

	_, err := repository.Create(context.Background(), ticketcore.CreateOptions{Title: "Next"})
	if err == nil {
		t.Fatal("CreateTicket returned nil error")
	}
}
