package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTicketNewCreatesTicket(t *testing.T) {
	dir := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"new",
		"--dir",
		dir,
		"--priority",
		"high",
		"--tag",
		"uxui",
		"Do something",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantPath := filepath.Join(dir, ".chc", "tickets", "CHC-1-Do-something.md")
	if !strings.Contains(stdout.String(), "ticket: "+wantPath) {
		t.Fatalf("stdout = %q, want ticket path %q", stdout.String(), wantPath)
	}

	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`key: "CHC-1"`,
		`title: "Do something"`,
		`priority: "high"`,
		`tags: ["uxui"]`,
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestTicketNewUsesConfigKeyPrefix(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, ".chc", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	if err := os.WriteFile(configPath, []byte("tickets:\n  key_prefix: PRJ\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run(t.Context(), []string{"ticket", "new", "--dir", dir, "Do something"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(dir, ".chc", "tickets", "PRJ-1-Do-something.md"))
}

func TestTicketNewRejectsMissingTitle(t *testing.T) {
	err := run(t.Context(), []string{"ticket", "new"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat path %s: %v", path, err)
	}
}
