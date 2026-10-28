// Package integrate owns candidate validation's store-facing contracts.
package integrate

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Completion carries one claimed validation's result and observed stop state.
type Completion struct {
	Candidate   *workflowstate.IntegrationCandidate
	Failure     error
	Interrupted bool
}

// Validation is bound to one reserved claim and its exact prepared input.
// Acceptance rechecks the claim and source before the owner's atomic transition.
type Validation interface {
	Validate(context.Context, *workflowstate.TaskResult) error
	Accept(context.Context, Completion) error
}

// RepairProof checks retained tree and replay repair evidence in its owning stage.
// It accepts immutable proof data and exposes no store or execution capability.
type RepairProof interface {
	Check(workflowstate.IntegrationProof) error
}
