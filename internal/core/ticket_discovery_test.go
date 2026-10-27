package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketFileDiscovererDiscoversMarkdownTickets(t *testing.T) {
	projectDir := t.TempDir()

	writeDiscoveryTestFile(t, projectDir, ".chr/default/tickets/use-case/CHR-2-Second.md")
	writeDiscoveryTestFile(t, projectDir, ".chr/default/tickets/CHR-1-First.md")
	writeDiscoveryTestFile(t, projectDir, ".chr/default/tickets/notes.txt")

	files, err := ticketcore.NewTicketFileDiscoverer().DiscoverTicketFiles(
		context.Background(),
		ticketcore.DiscoverTicketFilesOptions{ProjectDir: projectDir},
	)
	if err != nil {
		t.Fatalf("DiscoverTicketFiles returned error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2", len(files))
	}

	assertDiscoveredTicketFile(t, files[0], projectDir, "CHR-1-First.md")
	assertDiscoveredTicketFile(t, files[1], projectDir, "use-case/CHR-2-Second.md")
}

func TestTicketFileDiscovererRejectsMissingTicketDirectory(t *testing.T) {
	_, err := ticketcore.NewTicketFileDiscoverer().DiscoverTicketFiles(
		context.Background(),
		ticketcore.DiscoverTicketFilesOptions{ProjectDir: t.TempDir()},
	)
	if err == nil {
		t.Fatal("DiscoverTicketFiles returned nil error")
	}
}

func writeDiscoveryTestFile(t *testing.T, projectDir, relativePath string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create test file dir: %v", err)
	}

	err = os.WriteFile(path, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func assertDiscoveredTicketFile(
	t *testing.T,
	file ticketcore.DiscoveredTicketFile,
	projectDir string,
	relativePath string,
) {
	t.Helper()

	if file.RelativePath != relativePath {
		t.Fatalf("RelativePath = %q, want %q", file.RelativePath, relativePath)
	}

	wantPath := filepath.Join(projectDir, ".chr", "default", "tickets", filepath.FromSlash(relativePath))
	if file.Path != wantPath {
		t.Fatalf("Path = %q, want %q", file.Path, wantPath)
	}
}
