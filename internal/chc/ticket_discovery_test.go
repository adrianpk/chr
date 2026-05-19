package chc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTicketFileDiscovererDiscoversMarkdownTickets(t *testing.T) {
	projectDir := t.TempDir()

	writeDiscoveryTestFile(t, projectDir, ".chc/tickets/use-case/CHC-2-Second.md")
	writeDiscoveryTestFile(t, projectDir, ".chc/tickets/CHC-1-First.md")
	writeDiscoveryTestFile(t, projectDir, ".chc/tickets/notes.txt")

	files, err := NewTicketFileDiscoverer().DiscoverTicketFiles(
		context.Background(),
		DiscoverTicketFilesOptions{ProjectDir: projectDir},
	)
	if err != nil {
		t.Fatalf("DiscoverTicketFiles returned error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2", len(files))
	}

	assertDiscoveredTicketFile(t, files[0], projectDir, "CHC-1-First.md")
	assertDiscoveredTicketFile(t, files[1], projectDir, "use-case/CHC-2-Second.md")
}

func TestTicketFileDiscovererRejectsMissingTicketDirectory(t *testing.T) {
	_, err := NewTicketFileDiscoverer().DiscoverTicketFiles(
		context.Background(),
		DiscoverTicketFilesOptions{ProjectDir: t.TempDir()},
	)
	if err == nil {
		t.Fatal("DiscoverTicketFiles returned nil error")
	}
}

func writeDiscoveryTestFile(t *testing.T, projectDir, relativePath string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create test file dir: %v", err)
	}

	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func assertDiscoveredTicketFile(
	t *testing.T,
	file DiscoveredTicketFile,
	projectDir string,
	relativePath string,
) {
	t.Helper()

	if file.RelativePath != relativePath {
		t.Fatalf("RelativePath = %q, want %q", file.RelativePath, relativePath)
	}

	wantPath := filepath.Join(projectDir, ".chc", "tickets", filepath.FromSlash(relativePath))
	if file.Path != wantPath {
		t.Fatalf("Path = %q, want %q", file.Path, wantPath)
	}
}
