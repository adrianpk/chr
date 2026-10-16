package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

// ErrWorkflowStop reports a durable pause or cancellation request.
var ErrWorkflowStop = errors.New("workflow stop requested")

// StopWorkflow records intent. A running attempt releases ownership only after
// its process group is quiescent and its retained checkpoint is journaled.
func (s *WorkflowStore) StopWorkflow(ctx context.Context, id, intent string, expected ...uint64) (result WorkflowInvocation, err error) {
	return s.stopWorkflowChecked(ctx, id, intent, nil, expected...)
}

func (s *WorkflowStore) stopWorkflowChecked(ctx context.Context, id, intent string, check WorkflowAdmissionCheck, expected ...uint64) (result WorkflowInvocation, err error) {
	if !slices.Contains([]string{"pause", "cancel"}, intent) {
		return result, fmt.Errorf("workflow stop intent must be pause or cancel")
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if err := checkWorkflowRevision(run.WorkflowInvocation, expected); err != nil {
			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		readErr = owner.stopWorkflow(ctx, &run, intent, nil)

		result = run.WorkflowInvocation

		return readErr
	})

	return result, err
}

// ResumeWorkflow clears a completed pause without resetting attempts or evidence.
func (s *WorkflowStore) ResumeWorkflow(ctx context.Context, id string, ports ...WorkflowResumePorts) (result WorkflowInvocation, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if len(ports) > 1 {
		return result, ErrWorkflowConflict
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		result, readErr = owner.resumeGuardedWorkflow(ctx, run, ports, nil)

		return readErr
	})

	return result, err
}

func workflowStopStatus(intent string) string {
	if intent == "cancel" {
		return "cancelled"
	}

	return "paused"
}

func (s *WorkflowStore) watchWorkflowStop(ctx context.Context, claim *WorkflowClaim) (context.Context, context.CancelFunc) {
	watched, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-watched.Done():
				return
			case <-ticker.C:
				if err := s.checkWorkflowClaimStop(watched, claim); err != nil {
					cancel()

					return
				}
			}
		}
	}()

	return watched, cancel
}

func workflowRetainedStopIntent(run workflowInvocationFile, intent string) string {
	if run.StopIntent == "cancel" {
		return "cancel"
	}

	return intent
}

func (s *WorkflowStore) resumeGuardedWorkflow(ctx context.Context, run workflowInvocationFile, ports []WorkflowResumePorts, control *workflowControlCommit) (WorkflowInvocation, error) {
	if err := checkWorkflowResumeRevision(run.WorkflowInvocation, ports); err != nil {
		return WorkflowInvocation{}, err
	}

	if len(ports) == 1 {
		if err := workflowCheckAuthorization(ctx, ports[0].AuthorizationCheck); err != nil {
			return WorkflowInvocation{}, err
		}
	}

	if run.Status != "paused" || run.StopIntent == "cancel" {
		return WorkflowInvocation{}, ErrWorkflowOwned
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return WorkflowInvocation{}, err
	}

	changes, err := s.resumeWorkflowPrerequisite(ctx, &run, ports)
	if err != nil {
		return WorkflowInvocation{}, err
	}

	run.Status = "pending"
	if run.ResumeStatus != "" {
		run.Status = run.ResumeStatus
	}

	run.StopIntent, run.ResumeStatus = "", ""

	if err := s.commitWorkflowControlResume(ctx, &run, changes, control); err != nil {
		return WorkflowInvocation{}, err
	}

	return run.WorkflowInvocation, nil
}

func (s *WorkflowStore) stopWorkflow(ctx context.Context, run *workflowInvocationFile, intent string, control *workflowControlCommit) error {
	intent = workflowRetainedStopIntent(*run, intent)
	if handled, err := s.stopWorkflowNativeSubmission(ctx, run, intent, control); handled || err != nil {
		return err
	}

	if workflowInvocationTerminal(run.Status) || run.StopIntent == intent {
		return s.retainWorkflowControlNoop(ctx, *run, control)
	}

	run.StopIntent = intent
	if run.Status != "running" {
		if run.Status != "paused" {
			run.ResumeStatus = run.Status
		}

		run.Status = workflowStopStatus(intent)
	}

	return s.commitWorkflowControlState(ctx, run, control)
}
