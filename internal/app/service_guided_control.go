package app

import (
	"context"
	"errors"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ControlGuidedUseCase preserves interrupted stages and delegates native work
// controls with their displayed revision and a durable original request.
func (s service) ControlGuidedUseCase(ctx context.Context, command authoringcore.GuidedUseCaseCommand) (authoringcore.Session, error) {
	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return authoringcore.Session{}, err
	}

	command.GuidedUseCaseScope = scope

	return authoringcore.ApplyGuidedUseCaseControl(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *authoringcore.Session) error {
		bounded := s
		bounded.ctx, bounded.workflowNamespace = ctx, session.Namespace

		return bounded.controlGuidedWork(ctx, session)
	})
}

func (s service) controlGuidedWork(ctx context.Context, session *authoringcore.Session) error {
	if err := authoringcore.BindGuidedControlInvocation(ctx, s.projectContext.ProjectDir, session, workflowcore.NewGuidedWorkflowObserver(s.projectContext.ProjectDir)); err != nil {
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

func (s service) acknowledgeGuidedNativeControl(ctx context.Context, session *authoringcore.Session) error {
	control := session.Guided.Control

	run, err := workflowcore.ReadGuidedInvocation(ctx, s.projectContext.ProjectDir, *session)
	if err != nil {
		return err
	}

	if run.StopIntent == control.StopIntent && (run.Status == workflowstate.InvocationPaused || run.Status == workflowstate.InvocationCancelled) {
		control.AcknowledgedRevision = run.Revision
	}

	return nil
}

func (s service) applyGuidedNativeControl(ctx context.Context, store *workflowcore.Store, request workflowstate.ControlMutationRequest) (workflowstate.ControlResult, error) {
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

	return result, errors.Join(authoringcore.ErrGuidedControlRejected, err)
}

func guidedControlCommand(session authoringcore.Session) (authoringcore.GuidedUseCaseCommand, bool) {
	if session.Guided == nil {
		return authoringcore.GuidedUseCaseCommand{}, false
	}

	control := session.Guided.Control
	if control == nil || control.PendingOperation == nil {
		return authoringcore.GuidedUseCaseCommand{}, false
	}

	operation := control.PendingOperation

	return authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}, OperationID: operation.ID, Action: operation.Action, NativeRevision: operation.NativeRevision, ExpectedRevision: operation.ExpectedRevision, Content: operation.Content, QuestionID: operation.QuestionID, InputDigest: operation.InputDigest}, true
}
