package core

import (
	"context"
	"path/filepath"
	"testing"

	daemoncore "forge.adrianpk.com/chr/internal/core/harness/daemon"
)

func TestNewDaemonStateRepositoryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := daemoncore.NewDaemonStateRepository(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewDaemonStateRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chr", "core.db"))
}
