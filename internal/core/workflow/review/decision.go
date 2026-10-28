// Package review owns review policy's store-facing contracts.
package review

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Completion carries one owned review role's decision, failure and stop state.
type Completion struct {
	Decision    workflowstate.ReviewDecision
	Failure     error
	Interrupted bool
}

// Decisions accepts a decision against its bound claim and immutable input.
// The owner retains current policy checks, approval and receipt transactions.
type Decisions interface {
	Accept(context.Context, Completion) (workflowstate.ReviewBinding, error)
}
