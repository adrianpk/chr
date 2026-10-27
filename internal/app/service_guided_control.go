package app

import (
	"context"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ControlGuidedUseCase preserves interrupted stages and delegates native work
// controls with their displayed revision and a durable original request.
func (s service) ControlGuidedUseCase(ctx context.Context, command core.GuidedUseCaseCommand) (core.AuthoringSession, error) {
	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	command.GuidedUseCaseScope = scope

	return core.ApplyGuidedUseCaseControl(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *core.AuthoringSession) error {
		bounded := s
		bounded.ctx, bounded.workflowNamespace = ctx, session.Namespace

		return bounded.controlGuidedWork(ctx, session)
	})
}

func (s service) controlGuidedWork(ctx context.Context, session *core.AuthoringSession) error {
	if err := core.BindGuidedControlInvocation(ctx, s.projectContext.ProjectDir, session); err != nil {
		return err
	}

	control := session.Guided.Control
	if control.NativeRequest == nil {
		return nil
	}

	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	request := *control.NativeRequest

	result, err := s.applyGuidedNativeControl(ctx, store, request)
	if err != nil {
		return err
	}

	if s.afterGuidedControl != nil {
		if err := s.afterGuidedControl(); err != nil {
			return err
		}
	}

	control.NativeReceipt = &result

	return s.acknowledgeGuidedNativeControl(ctx, session)
}

func (s service) acknowledgeGuidedNativeControl(ctx context.Context, session *core.AuthoringSession) error {
	control := session.Guided.Control

	run, err := core.ReadGuidedInvocation(ctx, s.projectContext.ProjectDir, *session)
	if err != nil {
		return err
	}

	if run.StopIntent == control.StopIntent && (run.Status == "paused" || run.Status == "cancelled") {
		control.AcknowledgedRevision = run.Revision
	}

	return nil
}

func (s service) applyGuidedNativeControl(ctx context.Context, store *core.WorkflowStore, request workflowstate.WorkflowControlMutationRequest) (workflowstate.WorkflowControlResult, error) {
	result, found, err := store.ReadWorkflowControlReceipt(ctx, request)
	if err != nil || found {
		return result, err
	}

	ports, err := s.workflowLifecycleControlPorts(ctx, store, WorkflowControlRequest{Action: request.Action, ID: request.ID, ExpectedRevision: request.ExpectedRevision})
	if err != nil {
		return result, err
	}

	result, err = store.ControlWorkflowLifecycle(ctx, request, ports, nil)
	if !errors.Is(err, workflowstate.ErrConflict) {
		return result, err
	}
	// A failed write can still have an accepted journal entry. Only a verified
	// missing original receipt permits rolling back the guided stop reservation.
	prior, found, readErr := store.ReadWorkflowControlReceipt(ctx, request)
	if readErr != nil {
		return result, errors.Join(err, readErr)
	}

	if found {
		return prior, nil
	}

	return result, errors.Join(core.ErrGuidedControlRejected, err)
}

func guidedControlCommand(session core.AuthoringSession) (core.GuidedUseCaseCommand, bool) {
	if session.Guided == nil {
		return core.GuidedUseCaseCommand{}, false
	}

	control := session.Guided.Control
	if control == nil || control.PendingOperation == nil {
		return core.GuidedUseCaseCommand{}, false
	}

	operation := control.PendingOperation

	return core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}, OperationID: operation.ID, Action: operation.Action, NativeRevision: operation.NativeRevision, ExpectedRevision: operation.ExpectedRevision, Content: operation.Content, QuestionID: operation.QuestionID, InputDigest: operation.InputDigest}, true
}
