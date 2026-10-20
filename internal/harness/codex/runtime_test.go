package codex

import (
	"os"
	"path/filepath"
	"testing"
)

// Resolve only the selected files and preserve the legacy read-only targets.
func TestHarnessCodexCredentialMounts(t *testing.T) {
	homeDir := t.TempDir()
	writeRuntimePlanTestFile(t, filepath.Join(homeDir, ".codex", "auth.json"), "{}\n")
	writeRuntimePlanTestFile(t, filepath.Join(homeDir, ".codex", "config.toml"), "model = \"gpt-5.5\"\n")

	mounts := CredentialMounts(homeDir)
	if len(mounts) != 2 {
		t.Fatalf("len(mounts) = %d, want 2", len(mounts))
	}

	wantTargets := []string{
		"/root/.codex/auth.json",
		"/root/.codex/config.toml",
	}

	for index, wantTarget := range wantTargets {
		if mounts[index].Target != wantTarget {
			t.Fatalf("mount target %d = %q, want %q", index, mounts[index].Target, wantTarget)
		}

		if !mounts[index].ReadOnly {
			t.Fatalf("mount %d is writable, want read-only", index)
		}
	}
}

func writeRuntimePlanTestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
