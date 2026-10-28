package closeout

import (
	"fmt"
	"path/filepath"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidateWorkflowClosingBinding preserves the existing closing contract.
func ValidateWorkflowClosingBinding(set workflowstate.WorkflowSet) error {
	if set.Closing == nil {
		return nil
	}

	if len(set.Closing.Attempts) < 1 || len(set.Closing.Attempts) > workflowstate.RecordLimit {
		return fmt.Errorf("set closing history bound exceeded")
	}

	for _, binding := range set.Closing.Attempts {
		if !filepath.IsLocal(binding.Ref) || len(binding.Hash) != 64 || filepath.Base(binding.Ref) != "state.json" {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
