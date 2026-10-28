package review

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Execution is one bounded role output beside its immutable input and IO error.
// The host retains the claim, command process and cancellation ownership.
type Execution struct {
	Input   *workflowstate.ReviewInput
	Output  workflowstate.CommandResult
	Failure error
}

// Result retains guarded acceptance and the exact role-selected repair key.
// A failed acceptance may still return its original partial binding.
type Result struct {
	Binding       workflowstate.ReviewBinding
	RepairTaskKey string
}

// Complete checks observed role output and requests guarded decision acceptance.
// Cleanup is detached from execution cancellation for the original five seconds;
// operator interruption remains an explicit observation from the watched context.
func Complete(ctx, watched context.Context, execution Execution, decisions Decisions) (Result, error) {
	decision, parseErr := ParseDecision(execution.Output.Stdout, execution.Input)

	cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finish()

	binding, err := decisions.Accept(cleanup, Completion{Decision: decision, Failure: errors.Join(execution.Failure, parseErr), Interrupted: watched.Err() != nil})

	return Result{Binding: binding, RepairTaskKey: decision.RepairTaskKey}, err
}
