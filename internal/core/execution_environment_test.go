package core

import (
	"context"
	"path/filepath"
	"testing"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func TestNewExecutionRepositoryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := executioncore.NewRepository(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chr", "core.db"))
}

func TestNewLoopRunRepositoryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := executioncore.NewLoopRunRepository(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewLoopRunRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chr", "core.db"))
}
