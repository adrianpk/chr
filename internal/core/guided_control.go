package core

import (
	"context"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrGuidedControlRejected marks a native rejection whose missing receipt was
// verified after journal recovery. Other errors retain uncertain stop intent.
var ErrGuidedControlRejected = errors.New("native guided control was rejected without an accepted receipt")

// GuidedControlState keeps stop intent separate from an interrupted stage.
// Native receipts prove acceptance; native quiescence is acknowledged separately.
type GuidedControlState struct {
	StopIntent           string                          `json:"stop_intent,omitempty"`
	ResumeStatus         string                          `json:"resume_status,omitempty"`
	PreviousStatus       string                          `json:"previous_status,omitempty"`
	PreviousIntent       string                          `json:"previous_intent,omitempty"`
	PreviousBlocker      string                          `json:"previous_blocker,omitempty"`
	AcknowledgedRevision uint64                          `json:"acknowledged_revision,omitempty"`
	PendingOperation     *GuidedUseCaseOperation         `json:"pending_operation,omitempty"`
	NativeRequest        *WorkflowControlMutationRequest `json:"native_request,omitempty"`
	NativeReceipt        *WorkflowControlResult          `json:"native_receipt,omitempty"`
	CompletedOperations  []GuidedControlReceipt          `json:"completed_operations,omitempty"`
}

// GuidedControlReceipt retains the original control outcome across later changes.
type GuidedControlReceipt struct {
	GuidedUseCaseOperation
	Native   *WorkflowControlResult `json:"native,omitempty"`
	Rejected bool                   `json:"rejected,omitempty"`
}

func boundedGuidedControls(control *GuidedControlState) bool {
	return control == nil || len(control.CompletedOperations) <= 128
}

// ApplyGuidedUseCaseControl serializes one revision-bound stop/resume operation
// independently of the pending stage. Its callback composes the native owner.
func ApplyGuidedUseCaseControl(ctx context.Context, root string, command GuidedUseCaseCommand, apply GuidedUseCaseStep) (session AuthoringSession, err error) {
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

		if err := guardGuidedScope(session, command.GuidedUseCaseScope); err != nil {
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

func reserveGuidedControl(ctx context.Context, root string, session *AuthoringSession, operation GuidedUseCaseOperation) error {
	state := session.Guided
	if state.Control != nil && state.Control.PendingOperation != nil {
		return nil
	}

	if state.Control == nil {
		state.Control = &GuidedControlState{}
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

func guidedNativeControlRequest(session AuthoringSession, operation GuidedUseCaseOperation) *WorkflowControlMutationRequest {
	work := session.Guided.Work
	if work == nil || work.Invocation == nil {
		return nil
	}

	return &WorkflowControlMutationRequest{Guided: &GuidedWorkflowControlProof{RunID: session.ID, OperationDigest: operation.Digest}, Binding: WorkflowMutationBinding{RequestID: "guided-control-" + operation.Digest, Operator: session.Guided.Actor}, Action: operation.Action, ID: work.Invocation.ID, ExpectedRevision: operation.NativeRevision}
}

func (r *fileAuthoringSessionRepository) retainGuidedControlFailure(ctx context.Context, session *AuthoringSession, operation GuidedUseCaseOperation, cause error) error {
	if errors.Is(cause, ErrGuidedControlRejected) {
		completeGuidedControl(session, operation, true)
	} else {
		session.Guided.Blocker = "Control result needs recovery. Retry the original retained control before continuing."
	}

	return errors.Join(cause, RetainGuidedUseCaseProgress(ctx, r.projectDir, session))
}

func completeGuidedControl(session *AuthoringSession, operation GuidedUseCaseOperation, rejected bool) {
	state := session.Guided
	control := state.Control
	control.CompletedOperations = append(control.CompletedOperations, GuidedControlReceipt{GuidedUseCaseOperation: operation, Native: control.NativeReceipt, Rejected: rejected})

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
