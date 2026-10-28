package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ContinuationRequest supplies host actions for the retained extent.
// It has no selector, filters, target or membership override. A zero StepLimit
// uses the invocation's frozen limit, whose compatibility default is 128.
type ContinuationRequest struct {
	AuthorizationCheck AdmissionCheck `json:"-"`
	ExpectedRevision   uint64
	Worker             []string
	PullRequest        workflowstate.PullRequestContent
	StepLimit          int
}

// ContinuationPorts supplies the same trusted publication/frontier IO.
type ContinuationPorts struct {
	Publication PublicationPorts
	Target      TargetGit
	Source      SourceGit
}

// ContinuationProgress reports accepted steps and the retained barrier.
type ContinuationProgress struct {
	Invocation workflowstate.Invocation
	Steps      int
	Reason     string
}

// ContinueWorkflow executes bounded transitions on one frozen invocation.
// Accepted task, PR and delivery checkpoints are reused after restart. Waiting
// authority, failed gates and set closure return without executing successors.
func ContinueWorkflow(ctx context.Context, store *Store, id string, request ContinuationRequest, ports ContinuationPorts) (ContinuationProgress, error) {
	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return ContinuationProgress{}, err
	}

	limit, err := workflowContinuationLimit(run, request.StepLimit)
	if err != nil {
		return ContinuationProgress{Invocation: run}, err
	}

	if request.ExpectedRevision != 0 {
		if _, err := store.acceptWorkflowContinuation(ctx, id, request.ExpectedRevision, request.AuthorizationCheck); err != nil {
			return ContinuationProgress{Invocation: run}, err
		}
	} else if request.AuthorizationCheck != nil {
		return ContinuationProgress{Invocation: run}, workflowstate.ErrConflict
	}

	step, err := (workflowContinuationHost{store: store}).run(ctx, id, request, ports, limit)

	return finishWorkflowContinuation(ctx, store, id, step, err)
}

func workflowContinuationLimit(run workflowstate.Invocation, requested int) (int, error) {
	if run.Continuation != nil && run.Continuation.Sequence >= workflowstate.RecordLimit {
		return 0, fmt.Errorf("continuation checkpoint history exhausted")
	}

	limit := run.Policy.CheckpointStepLimit
	if limit == 0 {
		limit = workflowstate.PageLimit
	}

	if requested == 0 {
		requested = limit
	}

	if requested < 1 || requested > limit || requested > workflowstate.PageLimit {
		return 0, fmt.Errorf("continuation step limit exceeds frozen bounds")
	}

	return requested, nil
}

func finishWorkflowContinuation(ctx context.Context, store *Store, id string, step workflowContinuationStep, failure error) (ContinuationProgress, error) {
	steps, reason := step.Steps, step.Reason

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	run, err := store.ReadWorkflow(cleanup, id)
	if err != nil {
		return ContinuationProgress{Steps: steps, Reason: reason}, errors.Join(failure, err)
	}

	if barrier := workflowContinuationBarrier(run); barrier != "" {
		reason = barrier
	}

	if !workflowstate.InvocationTerminal(run.Status) {
		run, err = store.retainWorkflowContinuation(cleanup, run, steps, reason)
	}

	return ContinuationProgress{Invocation: run, Steps: steps, Reason: reason}, errors.Join(failure, err)
}
