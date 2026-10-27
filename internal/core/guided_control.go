package core

import (
	"context"
	"errors"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrGuidedControlRejected marks a native rejection whose missing receipt was
// verified after journal recovery. Other errors retain uncertain stop intent.
var ErrGuidedControlRejected = errors.New("native guided control was rejected without an accepted receipt")

// ApplyGuidedUseCaseControl serializes one revision-bound stop/resume operation
// independently of the pending stage. Its callback composes the native owner.
func ApplyGuidedUseCaseControl(ctx context.Context, root string, command authoringcore.GuidedUseCaseCommand, apply GuidedUseCaseStep) (session authoringcore.Session, err error) {
	operation, err := guidedOperation(command)
	if err != nil || apply == nil {
		return session, errors.Join(workflowstate.ErrConflict, err)
	}

	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, command.RunID)
		if err != nil {
			return err
		}

		if err := authoringcore.ValidateGuidedScope(session, command.GuidedUseCaseScope); err != nil {
			return err
		}

		replay, err := guardGuidedControl(session, operation)
		if err != nil || replay {
			return err
		}

		if err := reserveGuidedControl(ctx, root, &session, operation); err != nil {
			return err
		}

		if err := apply(ctx, &session); err != nil {
			return r.retainGuidedControlFailure(ctx, &session, operation, err)
		}

		completeGuidedControl(&session, operation, false)

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})

	return session, err
}

func reserveGuidedControl(ctx context.Context, root string, session *authoringcore.Session, operation authoringcore.GuidedUseCaseOperation) error {
	state := session.Guided
	if state.Control != nil && state.Control.PendingOperation != nil {
		return nil
	}

	if state.Control == nil {
		state.Control = &authoringcore.GuidedControlState{}
	}

	control := state.Control

	control.PreviousStatus, control.PreviousIntent, control.PreviousBlocker = state.Status, control.StopIntent, state.Blocker
	if operation.Action == "pause" && control.StopIntent == "" {
		control.ResumeStatus = state.Status
	}

	if operation.Action != "resume" {
		control.StopIntent = operation.Action
	}

	control.PendingOperation, control.NativeReceipt, control.AcknowledgedRevision = &operation, nil, 0
	control.NativeRequest = guidedNativeControlRequest(*session, operation)
	state.Status, state.Blocker = "blocked", "Control request retained; reconcile its original native result."

	return RetainGuidedUseCaseProgress(ctx, root, session)
}

func guidedNativeControlRequest(session authoringcore.Session, operation authoringcore.GuidedUseCaseOperation) *workflowstate.ControlMutationRequest {
	work := session.Guided.Work
	if work == nil || work.Invocation == nil {
		return nil
	}

	return &workflowstate.ControlMutationRequest{Guided: &workflowstate.GuidedControlProof{RunID: session.ID, OperationDigest: operation.Digest}, Binding: workflowstate.MutationBinding{RequestID: "guided-control-" + operation.Digest, Operator: session.Guided.Actor}, Action: operation.Action, ID: work.Invocation.ID, ExpectedRevision: operation.NativeRevision}
}

func (r *fileAuthoringSessionRepository) retainGuidedControlFailure(ctx context.Context, session *authoringcore.Session, operation authoringcore.GuidedUseCaseOperation, cause error) error {
	if errors.Is(cause, ErrGuidedControlRejected) {
		completeGuidedControl(session, operation, true)
	} else {
		session.Guided.Blocker = "Control result needs recovery. Retry the original retained control before continuing."
	}

	return errors.Join(cause, RetainGuidedUseCaseProgress(ctx, r.projectDir, session))
}

func completeGuidedControl(session *authoringcore.Session, operation authoringcore.GuidedUseCaseOperation, rejected bool) {
	state := session.Guided
	control := state.Control
	control.CompletedOperations = append(control.CompletedOperations, authoringcore.GuidedControlReceipt{GuidedUseCaseOperation: operation, Native: control.NativeReceipt, Rejected: rejected})

	control.PendingOperation, control.NativeRequest = nil, nil
	if rejected {
		state.Status, state.Blocker, control.StopIntent = control.PreviousStatus, control.PreviousBlocker, control.PreviousIntent

		return
	}

	if operation.Action == "resume" {
		control.StopIntent, state.Status, state.Blocker = "", control.ResumeStatus, ""

		return
	}

	state.Status, state.Blocker = "blocked", "Stop requested; native work is stopping."
	if control.NativeReceipt == nil || control.AcknowledgedRevision != 0 {
		state.Status, state.Blocker = workflowStopStatus(operation.Action), ""
	}
}
