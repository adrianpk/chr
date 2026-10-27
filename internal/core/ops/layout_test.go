package ops

import (
	"path/filepath"
	"testing"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

func TestResolveOpsLayoutDefaultsToDefaultNamespace(t *testing.T) {
	projectDir := t.TempDir()

	layout, err := ResolveLayout(LayoutOptions{
		ProjectDir: projectDir,
		LookupEnv:  func(string) string { return "" },
	})
	if err != nil {
		t.Fatalf("ResolveOpsLayout returned error: %v", err)
	}

	if layout.Namespace() != opsnamespace.Default {
		t.Fatalf("Namespace = %q, want %q", layout.Namespace(), opsnamespace.Default)
	}

	if layout.Root() != filepath.Join(projectDir, "ops", "default") {
		t.Fatalf("Root = %q", layout.Root())
	}
}

func TestResolveOpsLayoutUsesExplicitNamespaceBeforeEnvironment(t *testing.T) {
	layout, err := ResolveLayout(LayoutOptions{
		ProjectDir: t.TempDir(),
		Namespace:  "team-a",
		LookupEnv:  func(string) string { return "other-user" },
	})
	if err != nil {
		t.Fatalf("ResolveOpsLayout returned error: %v", err)
	}

	if layout.Namespace() != "team-a" {
		t.Fatalf("Namespace = %q, want team-a", layout.Namespace())
	}
}

func TestResolveOpsLayoutUsesOpsUser(t *testing.T) {
	layout, err := ResolveLayout(LayoutOptions{
		ProjectDir: t.TempDir(),
		LookupEnv: func(name string) string {
			if name == opsnamespace.UserEnv {
				return "adrian"
			}

			return ""
		},
	})
	if err != nil {
		t.Fatalf("ResolveOpsLayout returned error: %v", err)
	}

	if layout.Namespace() != "adrian" {
		t.Fatalf("Namespace = %q, want adrian", layout.Namespace())
	}
}

func TestResolveOpsLayoutRejectsInvalidNamespace(t *testing.T) {
	_, err := ResolveLayout(LayoutOptions{
		ProjectDir: t.TempDir(),
		Namespace:  "../outside",
	})
	if err == nil {
		t.Fatal("ResolveOpsLayout returned nil error")
	}
}

func TestOpsLayoutBuildsCanonicalArtifactAndWorktreePaths(t *testing.T) {
	projectDir := t.TempDir()

	layout, err := ResolveLayout(LayoutOptions{ProjectDir: projectDir, Namespace: "adrian"})
	if err != nil {
		t.Fatalf("ResolveOpsLayout returned error: %v", err)
	}

	runDir, err := layout.ArtifactDirectory("run", "00001-parser-validation")
	if err != nil {
		t.Fatalf("ArtifactDirectory returned error: %v", err)
	}

	if want := filepath.Join(projectDir, "ops", "adrian", "run", "00001-parser-validation"); runDir != want {
		t.Fatalf("ArtifactDirectory = %q, want %q", runDir, want)
	}

	worktreeDir, err := layout.WorktreeDirectory("00001-parser-validation")
	if err != nil {
		t.Fatalf("WorktreeDirectory returned error: %v", err)
	}

	if want := filepath.Join(projectDir, "ops", "adrian", "worktree", "00001-parser-validation"); worktreeDir != want {
		t.Fatalf("WorktreeDirectory = %q, want %q", worktreeDir, want)
	}
}

func TestOpsLayoutRejectsUnknownArtifactTypeAndUnsafeID(t *testing.T) {
	layout, err := ResolveLayout(LayoutOptions{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ResolveOpsLayout returned error: %v", err)
	}

	if _, err := layout.ArtifactDirectory("runs", "00001-parser-validation"); err == nil {
		t.Fatal("ArtifactDirectory accepted plural type")
	}

	if _, err := layout.ArtifactDirectory("run", "../outside"); err == nil {
		t.Fatal("ArtifactDirectory accepted unsafe ID")
	}
}
