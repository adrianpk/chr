package closeout

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Readiness binds the public checkpoint to its closing set and host owner scope.
type Readiness struct {
	Run       workflowstate.Invocation
	SetID     string
	OwnerRoot string
	Root      string
}

// RunGuard keeps live process and current-input fences with the host.
type RunGuard interface {
	Attempt(workflowstate.Attempt) error
	Inputs(context.Context) error
}

// Projection supplies the accepted attempt outcome for one guarded host update.
type Projection struct {
	State  Attempt
	SetID  string
	Target *workflowstate.PreviewTarget
}
