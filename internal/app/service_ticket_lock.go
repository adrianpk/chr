package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

func acquireAgentOperationLock(ctx context.Context, projectContext commandProjectContext, operation string) (*workflowcore.AgentOperationLock, error) {
	return workflowcore.AcquireAgentOperationLock(ctx, workflowcore.AgentOperationLockOptions{
		ProjectDir: projectContext.ProjectDir,
		Operation:  operation,
		Owner:      "chr",
	})
}

func releaseAgentOperationLock(lock *workflowcore.AgentOperationLock) func() {
	return func() {
		//nolint:errcheck // Operation result already comes from the main workflow.
		_ = lock.Release()
	}
}

func noopRelease() {}
