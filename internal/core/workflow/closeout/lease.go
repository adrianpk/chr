package closeout

import (
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowClosingLeaseSeconds preserves the existing closing contract.
func WorkflowClosingLeaseSeconds(validation workflowstate.Validation) time.Duration {
	if !validation.Required {
		return time.Minute
	}

	return time.Duration(validation.TimeoutSeconds) * time.Second
}
