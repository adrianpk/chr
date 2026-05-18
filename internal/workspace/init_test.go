package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesProjectWorkspace(t *testing.T) {
	dir := t.TempDir()

	result, err := Init(context.Background(), InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if result.Root != filepath.Join(dir, DirectoryName) {
		t.Fatalf("Root = %q, want %q", result.Root, filepath.Join(dir, DirectoryName))
	}

	assertDir(t, dir, ".chc/tickets/open")
	assertDir(t, dir, ".chc/context")
	assertDir(t, dir, ".chc/runtime/traces")
	assertFile(t, dir, ".chc/manifest.md")
	assertFile(t, dir, ".chc/config.yml")
}

func TestInitDoesNotOverwriteExistingFiles(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".chc", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create .chc: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("custom: true\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Init(context.Background(), InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(content) != "custom: true\n" {
		t.Fatalf("config content = %q, want custom content unchanged", string(content))
	}
}

func TestInitRejectsFileProjectDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Init(context.Background(), InitOptions{ProjectDir: path})
	if err == nil {
		t.Fatal("Init returned nil error for file project dir")
	}
}

func assertDir(t *testing.T, root string, rel string) {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("stat dir %s: %v", rel, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", rel)
	}
}

func assertFile(t *testing.T, root string, rel string) {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("stat file %s: %v", rel, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory", rel)
	}
}
