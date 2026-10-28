package workflow

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

func TestAgentOperationLockBlocksConcurrentOperation(t *testing.T) {
	projectDir := t.TempDir()

	lock, err := AcquireAgentOperationLock(context.Background(), AgentOperationLockOptions{
		ProjectDir: projectDir,
		Operation:  AgentOperationDevelopment,
		Owner:      "test",
	})
	if err != nil {
		t.Fatalf("AcquireAgentOperationLock returned error: %v", err)
	}

	if want := filepath.Join(projectDir, opscore.DirectoryName, opsnamespace.Default, "lock", "agent-operation.lock"); lock.path != want {
		t.Fatalf("lock path = %q, want %q", lock.path, want)
	}

	defer func() {
		if err := lock.Release(); err != nil {
			t.Fatalf("Release returned error: %v", err)
		}
	}()

	_, err = AcquireAgentOperationLock(context.Background(), AgentOperationLockOptions{
		ProjectDir: projectDir,
		Operation:  AgentOperationReview,
		Owner:      "test",
	})
	if err == nil || !strings.Contains(err.Error(), "agent operation lock is already held") {
		t.Fatalf("second AcquireAgentOperationLock error = %v, want held lock", err)
	}
}
