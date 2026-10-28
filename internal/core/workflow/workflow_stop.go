package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowStopWatch struct {
	Claim  *Claim
	Cancel context.CancelFunc
}

// ErrWorkflowStop reports a durable pause or cancellation request.
var ErrWorkflowStop = errors.New("workflow stop requested")

// StopWorkflow records intent. A running attempt releases ownership only after
// its process group is quiescent and its retained checkpoint is journaled.
func (s *Store) StopWorkflow(ctx context.Context, id, intent string, expected ...uint64) (result Invocation, err error) {
	return s.stopWorkflowChecked(ctx, id, intent, nil, expected...)
}

func (s *Store) stopWorkflowChecked(ctx context.Context, id, intent string, check AdmissionCheck, expected ...uint64) (result Invocation, err error) {
	if !slices.Contains([]string{"pause", "cancel"}, intent) {
		return result, fmt.Errorf("workflow stop intent must be pause or cancel")
	}

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if err := checkWorkflowRevision(invocationLocation.Run.Invocation, expected); err != nil {
			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		readErr = invocationLocation.Owner.stopWorkflow(ctx, &invocationLocation.Run, intent, nil)

		result = invocationLocation.Run.Invocation

		return readErr
	})

	return result, err
}

// ResumeWorkflow clears a completed pause without resetting attempts or evidence.
func (s *Store) ResumeWorkflow(ctx context.Context, id string, ports ...ResumePorts) (result Invocation, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if len(ports) > 1 {
		return result, workflowstate.ErrConflict
	}

	err = s.withLock(ctx, func() error {
		invocationLocation2, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		result, readErr = invocationLocation2.Owner.resumeGuardedWorkflow(ctx, invocationLocation2.Run, ports, nil)

		return readErr
	})

	return result, err
}

func (s *Store) watchWorkflowStop(watched context.Context, watch workflowStopWatch) {
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-watched.Done():
				return
			case <-ticker.C:
				if err := s.checkWorkflowClaimStop(watched, watch.Claim); err != nil {
					watch.Cancel()

					return
				}
			}
		}
	}()
}

func workflowRetainedStopIntent(run workflowInvocationFile, intent string) string {
	if run.StopIntent == "cancel" {
		return "cancel"
	}

	return intent
}

func (s *Store) resumeGuardedWorkflow(ctx context.Context, run workflowInvocationFile, ports []ResumePorts, control *workflowControlCommit) (Invocation, error) {
	if err := checkWorkflowResumeRevision(run.Invocation, ports); err != nil {
		return Invocation{}, err
	}

	if len(ports) == 1 {
		if err := workflowCheckAuthorization(ctx, ports[0].AuthorizationCheck); err != nil {
			return Invocation{}, err
		}
	}

	if run.Status != workflowstate.InvocationPaused || run.StopIntent == "cancel" {
		return Invocation{}, ErrWorkflowOwned
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return Invocation{}, err
	}

	changes, err := s.resumeWorkflowPrerequisite(ctx, &run, ports)
	if err != nil {
		return Invocation{}, err
	}

	run.Status = workflowstate.InvocationPending
	if run.ResumeStatus != workflowstate.InvocationUnspecified {
		run.Status = run.ResumeStatus
	}

	run.StopIntent, run.ResumeStatus = "", workflowstate.InvocationUnspecified
	restoreWorkflowInterruptedContinuation(&run)

	if err := s.commitWorkflowControlResume(ctx, &run, changes, control); err != nil {
		return Invocation{}, err
	}

	return run.Invocation, nil
}

func restoreWorkflowInterruptedContinuation(run *workflowInvocationFile) {
	if run.Outcome != "interrupted" {
		return
	}
	// Interrupted set closure retains its own retry marker. A resumed task or
	// validation must reach its original owner rather than the closing branch.
	next := workflowNextClosingSet(run.Invocation)
	if next != "" && workflowClosingRunFrontier(*run, next) {
		return
	}

	run.Outcome = "resumed"
}

func (s *Store) stopWorkflow(ctx context.Context, run *workflowInvocationFile, intent string, control *workflowControlCommit) error {
	intent = workflowRetainedStopIntent(*run, intent)
	if handled, err := s.stopWorkflowNativeSubmission(ctx, run, intent, control); handled || err != nil {
		return err
	}

	if workflowInvocationTerminal(run.Status) || run.StopIntent == intent {
		return s.retainWorkflowControlNoop(ctx, *run, control)
	}

	run.StopIntent = intent
	if run.Status != workflowstate.InvocationRunning {
		if run.Status != workflowstate.InvocationPaused {
			run.ResumeStatus = run.Status
		}

		run.Status = workflowstate.StopStatus(intent)
	}

	return s.commitWorkflowControlState(ctx, run, control)
}
