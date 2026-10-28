// Package forge owns publication policy's store-facing contracts.
package forge

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Response retains an untrusted provider response and its original IO failure.
type Response struct {
	Reply   workflowstate.SubmissionResponse
	Failure error
}

// Responses is bound to one invocation and journaled submission identity.
// Retention cannot confirm delivery or dispatch an uncertain submission again.
type Responses interface {
	Retain(context.Context, Response) (workflowstate.Submission, error)
}
