package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestInitCreatesProjectWorkspace(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	result, err := initializer.Init(context.Background(), InitOptions{ProjectDir: dir})
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
	assertFile(t, dir, ".chc/agents/CONVENTIONS.md")
	assertFile(t, dir, "AGENTS.md")
	assertNotExists(t, dir, ".chc/root/AGENTS.md")
}

func TestInitDoesNotOverwriteExistingFiles(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	configPath := filepath.Join(dir, ".chc", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create .chc: %v", err)
	}

	if err := os.WriteFile(configPath, []byte("custom: true\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := initializer.Init(context.Background(), InitOptions{ProjectDir: dir})
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

func TestInitDoesNotOverwriteExistingRootAgents(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("custom agents\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	_, err := initializer.Init(context.Background(), InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}

	if string(content) != "custom agents\n" {
		t.Fatalf("AGENTS.md content = %q, want custom content unchanged", string(content))
	}
}

func TestInitDryRunDoesNotCreateWorkspace(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	result, err := initializer.Init(context.Background(), InitOptions{ProjectDir: dir, DryRun: true})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if len(result.Created) == 0 {
		t.Fatal("dry run reported no planned creates")
	}

	if _, err := os.Stat(filepath.Join(dir, DirectoryName)); !os.IsNotExist(err) {
		t.Fatalf("dry run created workspace or returned unexpected stat error: %v", err)
	}
}

func TestInitDryRunReportsExistingPaths(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	if err := os.MkdirAll(filepath.Join(dir, ".chc", "agents"), 0o755); err != nil {
		t.Fatalf("create existing dir: %v", err)
	}

	result, err := initializer.Init(context.Background(), InitOptions{ProjectDir: dir, DryRun: true})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if !contains(result.Skipped, ".chc/agents") {
		t.Fatalf("Skipped = %#v, want .chc/agents", result.Skipped)
	}
}

func TestInitRejectsFileProjectDir(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	path := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := initializer.Init(context.Background(), InitOptions{ProjectDir: path})
	if err == nil {
		t.Fatal("Init returned nil error for file project dir")
	}
}

func TestInitUsesRequestedProfile(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	_, err := initializer.Init(context.Background(), InitOptions{ProjectDir: dir, Profile: "minimal"})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".chc", "manifest.md"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	if string(content) != "# Minimal Workspace\n" {
		t.Fatalf("manifest content = %q, want minimal profile content", string(content))
	}
}

func TestInitRejectsUnknownProfile(t *testing.T) {
	dir := t.TempDir()
	initializer := newTestInitializer()

	_, err := initializer.Init(context.Background(), InitOptions{ProjectDir: dir, Profile: "unknown"})
	if err == nil {
		t.Fatal("Init returned nil error for unknown profile")
	}
}

func newTestInitializer() Initializer {
	return NewInitializer(fstest.MapFS{
		"assets/workspace/profiles/standard/config.yml": {
			Data: []byte("version: 1\nworkspace_root: .chc\n"),
		},
		"assets/workspace/profiles/standard/manifest.md": {
			Data: []byte("# ChrysoCoord Workspace\n"),
		},
		"assets/workspace/profiles/standard/agents/CONVENTIONS.md": {
			Data: []byte("# Conventions\n"),
		},
		"assets/workspace/profiles/standard/root/AGENTS.md": {
			Data: []byte("# Agent Operating Contract\n"),
		},
		"assets/workspace/profiles/minimal/config.yml": {
			Data: []byte("version: 1\nworkspace_root: .chc\nprofile: minimal\n"),
		},
		"assets/workspace/profiles/minimal/manifest.md": {
			Data: []byte("# Minimal Workspace\n"),
		},
		"assets/workspace/profiles/minimal/root/AGENTS.md": {
			Data: []byte("# Minimal Agents\n"),
		},
		"assets/workspace/profiles/personal/config.yml": {
			Data: []byte("version: 1\nworkspace_root: .chc\nprofile: personal\n"),
		},
		"assets/workspace/profiles/personal/manifest.md": {
			Data: []byte("# Personal Workspace\n"),
		},
		"assets/workspace/profiles/personal/root/AGENTS.md": {
			Data: []byte("# Personal Agents\n"),
		},
	})
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}

func assertDir(t *testing.T, root, rel string) {
	t.Helper()

	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("stat dir %s: %v", rel, err)
	}

	if !info.IsDir() {
		t.Fatalf("%s is not a directory", rel)
	}
}

func assertFile(t *testing.T, root, rel string) {
	t.Helper()

	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("stat file %s: %v", rel, err)
	}

	if info.IsDir() {
		t.Fatalf("%s is a directory", rel)
	}
}

func assertNotExists(t *testing.T, root, rel string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Fatalf("%s exists or returned unexpected stat error: %v", rel, err)
	}
}
