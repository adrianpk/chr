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
	PullRequest        SlicePullRequestRequest
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
	Invocation Invocation
	Steps      int
	Reason     string
}

// ContinueWorkflow executes bounded transitions on one frozen invocation.
// Accepted task, PR and delivery checkpoints are reused after restart. Waiting
// authority, failed gates and set closure return without executing successors.
func (s *Store) ContinueWorkflow(ctx context.Context, id string, request ContinuationRequest, ports ContinuationPorts) (ContinuationProgress, error) {
	run, err := s.ReadWorkflow(ctx, id)
	if err != nil {
		return ContinuationProgress{}, err
	}

	limit, err := workflowContinuationLimit(run, request.StepLimit)
	if err != nil {
		return ContinuationProgress{Invocation: run}, err
	}

	if request.ExpectedRevision != 0 {
		if _, err := s.acceptWorkflowContinuation(ctx, id, request.ExpectedRevision, request.AuthorizationCheck); err != nil {
			return ContinuationProgress{Invocation: run}, err
		}
	} else if request.AuthorizationCheck != nil {
		return ContinuationProgress{Invocation: run}, workflowstate.ErrConflict
	}

	steps, reason, err := s.continueWorkflowSteps(ctx, id, request, ports, limit)

	return s.finishWorkflowContinuation(ctx, id, steps, reason, err)
}

func workflowContinuationLimit(run Invocation, requested int) (int, error) {
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

func (s *Store) finishWorkflowContinuation(ctx context.Context, id string, steps int, reason string, failure error) (ContinuationProgress, error) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	run, err := s.ReadWorkflow(cleanup, id)
	if err != nil {
		return ContinuationProgress{Steps: steps, Reason: reason}, errors.Join(failure, err)
	}

	if barrier := workflowContinuationBarrier(run); barrier != "" {
		reason = barrier
	}

	if !workflowInvocationTerminal(run.Status) {
		run, err = s.retainWorkflowContinuation(cleanup, run, steps, reason)
	}

	return ContinuationProgress{Invocation: run, Steps: steps, Reason: reason}, errors.Join(failure, err)
}
