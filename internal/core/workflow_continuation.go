package core

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// WorkflowContinuationRequest supplies host actions for the retained extent.
// It has no selector, filters, target or membership override. A zero StepLimit
// uses the invocation's frozen limit, whose compatibility default is 128.
type WorkflowContinuationRequest struct {
	Worker      []string
	PullRequest WorkflowSlicePullRequestRequest
	StepLimit   int
}

// WorkflowContinuationPorts supplies the same trusted publication/frontier IO.
type WorkflowContinuationPorts struct {
	Publication WorkflowPublicationPorts
	Target      WorkflowTargetGit
	Source      WorkflowSourceGit
}

// WorkflowContinuationProgress reports accepted steps and the retained barrier.
type WorkflowContinuationProgress struct {
	Invocation WorkflowInvocation
	Steps      int
	Reason     string
}

// ContinueWorkflow executes bounded transitions on one frozen invocation.
// Accepted task, PR and delivery checkpoints are reused after restart. Waiting
// authority, failed gates and set closure return without executing successors.
func (s *WorkflowStore) ContinueWorkflow(ctx context.Context, id string, request WorkflowContinuationRequest, ports WorkflowContinuationPorts) (WorkflowContinuationProgress, error) {
	run, err := s.ReadWorkflow(ctx, id)
	if err != nil {
		return WorkflowContinuationProgress{}, err
	}

	limit, err := workflowContinuationLimit(run, request.StepLimit)
	if err != nil {
		return WorkflowContinuationProgress{Invocation: run}, err
	}

	steps, reason, err := s.continueWorkflowSteps(ctx, id, request, ports, limit)

	return s.finishWorkflowContinuation(ctx, id, steps, reason, err)
}

func workflowContinuationLimit(run WorkflowInvocation, requested int) (int, error) {
	if run.Continuation != nil && run.Continuation.Sequence >= WorkflowRecordLimit {
		return 0, fmt.Errorf("continuation checkpoint history exhausted")
	}

	limit := run.Policy.CheckpointStepLimit
	if limit == 0 {
		limit = WorkflowPageLimit
	}

	if requested == 0 {
		requested = limit
	}

	if requested < 1 || requested > limit || requested > WorkflowPageLimit {
		return 0, fmt.Errorf("continuation step limit exceeds frozen bounds")
	}

	return requested, nil
}

func (s *WorkflowStore) finishWorkflowContinuation(ctx context.Context, id string, steps int, reason string, failure error) (WorkflowContinuationProgress, error) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	run, err := s.ReadWorkflow(cleanup, id)
	if err != nil {
		return WorkflowContinuationProgress{Steps: steps, Reason: reason}, errors.Join(failure, err)
	}

	if barrier := workflowContinuationBarrier(run); barrier != "" {
		reason = barrier
	}

	if !workflowInvocationTerminal(run.Status) {
		run, err = s.retainWorkflowContinuation(cleanup, run, steps, reason)
	}

	return WorkflowContinuationProgress{Invocation: run, Steps: steps, Reason: reason}, errors.Join(failure, err)
}
