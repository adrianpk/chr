package app

import (
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// The same typed application failures keep validation, stale sources, ownership
// blocks and retained-operation recovery distinct in both adapter mappings.
func TestOperationOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name          string
		err           error
		status        string
		reload, retry bool
	}{
		{"validation", errors.New("content is required"), operation.ValidationFailed, false, false},
		{"stale", workflowstate.ErrConflict, operation.Stale, true, false},
		{"owned", workflowstate.ErrOwned, operation.Blocked, true, false},
		{"reconciliation", authoringcore.ErrProjectDumpReconciliation, operation.Blocked, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outcome := operationResult("", tc.err)
			if outcome.Status != tc.status || outcome.Reload != tc.reload || outcome.Retry != tc.retry || webDumpWriteFailure(tc.err, "bound.md").Status != tc.status {
				t.Fatal("adapters disagreed about operation disposition", outcome)
			}
		})
	}
}
