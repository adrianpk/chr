package core

import (
	"context"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowLifecycleControlPorts supplies existing native resume and continuation
// dependencies. Ports cannot replace the invocation's retained selection or policy.
type WorkflowLifecycleControlPorts struct {
	Resume       WorkflowResumePorts
	Continuation WorkflowContinuationRequest
	Execution    WorkflowContinuationPorts
}

type workflowLifecycleAcceptance struct {
	receipt WorkflowControlResult
	replay  bool
	limit   int
}

// ControlWorkflowLifecycle commits lifecycle acceptance and its request receipt
// together. Continuation runs outside exclusion after acceptance; a replay only
// returns that original receipt and never starts another checkpoint.
func (s *WorkflowStore) ControlWorkflowLifecycle(ctx context.Context, request WorkflowControlMutationRequest, ports WorkflowLifecycleControlPorts, check WorkflowAdmissionCheck) (WorkflowControlResult, error) {
	request, digest, err := freezeWorkflowControlMutation(request)
	if err != nil || !slices.Contains([]string{"pause", "cancel", "resume", "continue", "reconcile"}, request.Action) {
		return WorkflowControlResult{}, errorsOrConflict(err)
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
	steps, reason, err := s.continueWorkflowSteps(ctx, request.ID, ports.Continuation, ports.Execution, accepted.limit)
	_, err = s.finishWorkflowContinuation(ctx, request.ID, steps, reason, err)

	return accepted.receipt, err
}

func (s *WorkflowStore) acceptWorkflowLifecycleControl(ctx context.Context, request WorkflowControlMutationRequest, digest string, ports WorkflowLifecycleControlPorts, check WorkflowAdmissionCheck) (workflowLifecycleAcceptance, error) {
	run, owner, err := s.locateWorkflowInvocation(ctx, request.ID)
	if err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	prior, found, err := owner.readWorkflowControlReceipt(request, digest)
	if found || err != nil {
		return workflowLifecycleAcceptance{receipt: prior, replay: found}, err
	}

	if err := owner.checkGuidedWorkflowControl(ctx, run.WorkflowInvocation, request); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	if err := workflowCheckAuthorization(ctx, check); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	if err := owner.reserveWorkflowControlBinding(request, digest); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	if err := checkWorkflowRevision(run.WorkflowInvocation, []uint64{request.ExpectedRevision}); err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	commit := owner.newWorkflowControlCommit(request, digest, check)

	limit, err := owner.applyWorkflowLifecycleControl(ctx, run, request, ports, commit)
	if err != nil {
		return workflowLifecycleAcceptance{}, err
	}

	return workflowLifecycleAcceptance{receipt: commit.receipt.Result, limit: limit}, nil
}

func (s *WorkflowStore) applyWorkflowLifecycleControl(ctx context.Context, run workflowInvocationFile, request WorkflowControlMutationRequest, ports WorkflowLifecycleControlPorts, commit *workflowControlCommit) (int, error) {
	switch request.Action {
	case "pause", "cancel":
		return 0, s.stopWorkflow(ctx, &run, request.Action, commit)
	case "resume":
		ports.Resume.ExpectedRevision, ports.Resume.AuthorizationCheck = request.ExpectedRevision, commit.check
		_, err := s.resumeGuardedWorkflow(ctx, run, []WorkflowResumePorts{ports.Resume}, commit)

		return 0, err
	case "reconcile":
		return 0, s.reconcileWorkflowInvocation(ctx, run, commit)
	case "continue":
		return s.acceptWorkflowControlContinuation(ctx, &run, request.StepLimit, commit)
	default:
		return 0, workflowstate.ErrConflict
	}
}

func (s *WorkflowStore) acceptWorkflowControlContinuation(ctx context.Context, run *workflowInvocationFile, steps int, commit *workflowControlCommit) (int, error) {
	if workflowInvocationTerminal(run.Status) {
		return 0, s.retainWorkflowControlNoop(ctx, *run, commit)
	}

	limit, err := workflowContinuationLimit(run.WorkflowInvocation, steps)
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

func validateWorkflowLifecyclePayload(request WorkflowControlMutationRequest) error {
	if request.Integration != nil || request.Grant != nil || request.Answer != nil || request.ReviewHash != "" || request.DecisionRevision != 0 || request.DecisionHash != "" {
		return workflowstate.ErrConflict
	}

	if request.StepLimit < 0 || request.StepLimit > workflowstate.WorkflowPageLimit {
		return workflowstate.ErrConflict
	}

	return nil
}
