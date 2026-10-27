package execution

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestChangedFileDiscovererDiscoversWorkspaceChanges(t *testing.T) {
	projectDir := newChangedFileDiscoveryGitRepo(t)
	writeChangedFileDiscoveryFile(t, projectDir, "internal/core/new.go", "package core\n")
	writeChangedFileDiscoveryFile(t, projectDir, "README.md", "changed\n")
	writeChangedFileDiscoveryFile(t, projectDir, ".chr/default/runtime/runs/CHR-1/exec-1/result.json", "{}")

	changedFiles, err := NewChangedFileDiscoverer().DiscoverChangedFiles(
		context.Background(),
		ChangedFilesOptions{ProjectDir: projectDir},
	)
	if err != nil {
		t.Fatalf("DiscoverChangedFiles returned error: %v", err)
	}

	assertChangedFilePaths(t, changedFiles, []string{
		"README.md",
		"internal/core/new.go",
	})
}

func TestChangedFileDiscovererHandlesCleanWorktree(t *testing.T) {
	projectDir := newChangedFileDiscoveryGitRepo(t)

	changedFiles, err := NewChangedFileDiscoverer().DiscoverChangedFiles(
		context.Background(),
		ChangedFilesOptions{ProjectDir: projectDir},
	)
	if err != nil {
		t.Fatalf("DiscoverChangedFiles returned error: %v", err)
	}

	if len(changedFiles) != 0 {
		t.Fatalf("len(changedFiles) = %d, want 0", len(changedFiles))
	}
}

func TestChangedFileDiscovererUsesExplicitWorkspaceDir(t *testing.T) {
	projectDir := newChangedFileDiscoveryGitRepo(t)
	workspaceDir := newChangedFileDiscoveryGitRepo(t)
	writeChangedFileDiscoveryFile(t, projectDir, "README.md", "dirty user checkout\n")
	writeChangedFileDiscoveryFile(t, workspaceDir, "internal/core/result.go", "package core\n")

	changedFiles, err := NewChangedFileDiscoverer().DiscoverChangedFiles(
		context.Background(),
		ChangedFilesOptions{
			ProjectDir:   projectDir,
			WorkspaceDir: workspaceDir,
		},
	)
	if err != nil {
		t.Fatalf("DiscoverChangedFiles returned error: %v", err)
	}

	assertChangedFilePaths(t, changedFiles, []string{
		"internal/core/result.go",
	})
}

func TestChangedFileDiscovererRejectsMissingProjectDir(t *testing.T) {
	_, err := NewChangedFileDiscoverer().DiscoverChangedFiles(
		context.Background(),
		ChangedFilesOptions{},
	)
	if err == nil {
		t.Fatal("DiscoverChangedFiles returned nil error")
	}
}

func TestParseGitStatusChangedFilesRejectsNonProjectRelativePaths(t *testing.T) {
	changedFiles := parseGitStatusChangedFiles(" M /tmp/worktree/main.go\n M ../outside.go\n M internal/core/result.go\n")

	assertChangedFilePaths(t, changedFiles, []string{
		"internal/core/result.go",
	})
}

func newChangedFileDiscoveryGitRepo(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()
	runChangedFileDiscoveryGit(t, projectDir, "init")

	return projectDir
}

func runChangedFileDiscoveryGit(t *testing.T, projectDir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", projectDir}, args...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func writeChangedFileDiscoveryFile(t *testing.T, projectDir, relativePath, content string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create test file dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func assertChangedFilePaths(t *testing.T, changedFiles []ResultChangedFile, want []string) {
	t.Helper()

	if len(changedFiles) != len(want) {
		t.Fatalf("len(changedFiles) = %d, want %d: %#v", len(changedFiles), len(want), changedFiles)
	}

	for index := range want {
		if changedFiles[index].Path != want[index] {
			t.Fatalf("changedFiles[%d].Path = %q, want %q", index, changedFiles[index].Path, want[index])
		}
	}
}
