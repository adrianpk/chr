package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowClosingRunGuard struct {
	store *Store
	run   workflowInvocationFile
}

func (guard workflowClosingRunGuard) Attempt(attempt workflowstate.Attempt) error {
	return workflowAttemptQuiescent(attempt)
}

func (guard workflowClosingRunGuard) Inputs(ctx context.Context) error {
	return guard.store.checkWorkflowInputs(ctx, guard.run)
}
