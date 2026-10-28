package workflow

import (
	"context"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// LifecycleControlPorts supplies existing native resume and continuation
// dependencies. Ports cannot replace the invocation's retained selection or policy.
type LifecycleControlPorts struct {
	Resume       ResumePorts
	Continuation ContinuationRequest
	Execution    ContinuationPorts
}

type workflowLifecycleAcceptance struct {
	receipt workflowstate.ControlResult
	replay  bool
	limit   int
}

// ControlWorkflowLifecycle commits lifecycle acceptance and its request receipt
// together. Continuation runs outside exclusion after acceptance; a replay only
// returns that original receipt and never starts another checkpoint.
func (s *Store) ControlWorkflowLifecycle(ctx context.Context, request workflowstate.ControlMutationRequest, ports LifecycleControlPorts, check AdmissionCheck) (workflowstate.ControlResult, error) {
	frozen, err := freezeWorkflowControlMutation(request)

	request, digest := frozen.Request, frozen.Digest

	if err != nil || !slices.Contains([]string{"pause", "cancel", "resume", "continue", "reconcile"}, request.Action) {
		return workflowstate.ControlResult{}, errorsOrConflict(err)
	}

	acceptCtx, cancel := context.WithTimeout(ctx, time.Minute)

	var accepted workflowLifecycleAcceptance

	err = s.withLock(acceptCtx, func() error {
		var err error

		accepted, err = s.acceptWorkflowLifecycleControl(acceptCtx, request, digest, ports, check)

		return err
	})

	cancel()

	if err != nil || accepted.replay || accepted.limit == 0 {
		return accepted.receipt, err
	}

	ports.Continuation.StepLimit = request.StepLimit
	ports.Continuation.ExpectedRevision = request.ExpectedRevision
	step, err := (workflowContinuationHost{store: s}).run(ctx, request.ID, ports.Continuation, ports.Execution, accepted.limit)
	_, err = finishWorkflowContinuation(ctx, s, request.ID, step, err)

	return accepted.receipt, err
}

func (s *Store) acceptWorkflowLifecycleControl(ctx context.Context, request workflowstate.ControlMutationRequest, digest string, ports LifecycleControlPorts, check AdmissionCheck) (workflowLifecycleAcceptance, error) {
	invocationLocation, err := s.locateWorkflowInvocation(ctx, request.ID)
	if err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	prior, err := invocationLocation.Owner.readWorkflowControlReceipt(request, digest)
	if prior.Found || err != nil {
		return workflowLifecycleAcceptance{receipt: prior.Result, replay: prior.Found}, err
	}

	if err := invocationLocation.Owner.checkGuidedWorkflowControl(ctx, invocationLocation.Run.Invocation, request); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	if err := workflowCheckAuthorization(ctx, check); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	if err := invocationLocation.Owner.reserveWorkflowControlBinding(request, digest); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	if err := checkWorkflowRevision(invocationLocation.Run.Invocation, []uint64{request.ExpectedRevision}); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	commit := invocationLocation.Owner.newWorkflowControlCommit(request, digest, check)

	limit, err := invocationLocation.Owner.applyWorkflowLifecycleControl(ctx, invocationLocation.Run, request, ports, commit)
	if err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	return workflowLifecycleAcceptance{receipt: commit.receipt.Result, limit: limit}, nil
}

func (s *Store) applyWorkflowLifecycleControl(ctx context.Context, run workflowInvocationFile, request workflowstate.ControlMutationRequest, ports LifecycleControlPorts, commit *workflowControlCommit) (int, error) {
	switch request.Action {
	case "pause", "cancel":
		return 0, s.stopWorkflow(ctx, &run, request.Action, commit)
	case "resume":
		ports.Resume.ExpectedRevision, ports.Resume.AuthorizationCheck = request.ExpectedRevision, commit.check
		_, err := s.resumeGuardedWorkflow(ctx, run, []ResumePorts{ports.Resume}, commit)

		return 0, err
	case "reconcile":
		return 0, s.reconcileWorkflowInvocation(ctx, run, commit)
	case "continue":
		return s.acceptWorkflowControlContinuation(ctx, &run, request.StepLimit, commit)
	default:
		return 0, workflowstate.ErrConflict
	}
}

func (s *Store) acceptWorkflowControlContinuation(ctx context.Context, run *workflowInvocationFile, steps int, commit *workflowControlCommit) (int, error) {
	if workflowstate.InvocationTerminal(run.Status) {
		return 0, s.retainWorkflowControlNoop(ctx, *run, commit)
	}

	limit, err := workflowContinuationLimit(run.Invocation, steps)
	if err != nil {
		return 0, err
	}

	if err := s.checkWorkflowInputs(ctx, *run); err != nil {
		return 0, err
	}

	files, err := commit.files(ctx, run.Revision+1)
	if err != nil {
		return 0, err
	}

	return limit, s.commitWorkflowContinuation(run, 0, "accepted", files...)
}

func validateWorkflowLifecyclePayload(request workflowstate.ControlMutationRequest) error {
	if request.Integration != nil || request.Grant != nil || request.Answer != nil || request.ReviewHash != "" || request.DecisionRevision != 0 || request.DecisionHash != "" {
		return workflowstate.ErrConflict
	}

	if request.StepLimit < 0 || request.StepLimit > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	return nil
}
