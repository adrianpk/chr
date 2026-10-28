// Package merge owns merger policy's store-facing contracts.
package merge

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Evidence verifies the exact retained merger input and contained source images.
// It grants neither an editable store nor authority to accept or publish a proposal.
type Evidence interface {
	CheckWorkflowMergerEvidence(context.Context, workflowstate.MergerInput, time.Time) error
}
