// Package closeout owns closing gate policy and bounded execution.
package closeout

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Workspace prepares and checks the isolated checkout for one exact closing input.
// Its host owns filesystem access and repository observations.
type Workspace interface {
	Prepare(context.Context) (string, error)
	Check(context.Context, string) error
}

// Commands runs a command in the prepared checkout. The host retains process
// supervision, output bounds, capacity and guarded process journals.
type Commands interface {
	Run(context.Context, string, []string) (workflowstate.CommandResult, error)
}

// Evidence reads retained closing images and checks immutable prerequisite hashes.
// The host retains safe path resolution, file bounds and storage error semantics.
type Evidence interface {
	Read(string) ([]byte, error)
	Check(map[string]string) error
}

// InputSource supplies bounded observations for one frozen closing input.
// Configuration, Git IO, current-member guards and storage checks stay host-owned.
type InputSource interface {
	ValidationTimeout(context.Context, string, int) (int, error)
	Target(context.Context, Request) (workflowstate.PreviewTarget, error)
	Tree(context.Context, string) (string, error)
	Snapshot(context.Context) (workflowstate.Snapshot, error)
	Slice(context.Context, workflowstate.ClosingInput, workflowstate.WorkflowSlice) (workflowstate.ClosingSlice, error)
	Check(map[string]string) error
}
