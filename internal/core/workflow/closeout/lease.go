package closeout

import (
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// LeaseDuration preserves the finite closing lease, including the waiver bound.
func LeaseDuration(validation workflowstate.Validation) time.Duration {
	if !validation.Required {
		return time.Minute
	}

	return time.Duration(validation.TimeoutSeconds) * time.Second
}
