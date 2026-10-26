package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

func acquireAgentOperationLock(ctx context.Context, projectContext commandProjectContext, operation string) (*core.AgentOperationLock, error) {
	return core.AcquireAgentOperationLock(ctx, core.AgentOperationLockOptions{
		ProjectDir: projectContext.ProjectDir,
		Operation:  operation,
		Owner:      "chr",
	})
}

func releaseAgentOperationLock(lock *core.AgentOperationLock) func() {
	return func() {
		//nolint:errcheck // Operation result already comes from the main workflow.
		_ = lock.Release()
	}
}

func noopRelease() {}
