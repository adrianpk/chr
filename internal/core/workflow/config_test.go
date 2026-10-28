package workflow

import (
	"os"
	"path/filepath"
	"testing"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func writeWorkspaceConfig(t *testing.T, projectDir, content string) {
	t.Helper()

	path := workspacecore.UserPath(projectDir, "config.yml")

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
}
