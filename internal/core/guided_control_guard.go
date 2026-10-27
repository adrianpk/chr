package core

import (
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func guardGuidedControl(session authoringcore.Session, operation authoringcore.GuidedUseCaseOperation) (bool, error) {
	state := session.Guided

	if !guidedControlAction(operation.Action) {
		return false, workflowstate.ErrConflict
	}

	if state.Control != nil {
		for _, receipt := range state.Control.CompletedOperations {
			if receipt.ID == operation.ID {
				return replayGuidedControl(receipt, operation)
			}
		}

		if pending := state.Control.PendingOperation; pending != nil {
			return false, guidedDigestConflict(*pending, operation)
		}
	}

	if err := guardFreshGuidedControl(session, operation); err != nil {
		return false, err
	}

	return false, nil
}

func replayGuidedControl(receipt authoringcore.GuidedControlReceipt, operation authoringcore.GuidedUseCaseOperation) (bool, error) {
	if receipt.Rejected {
		return false, workflowstate.ErrConflict
	}

	return true, guidedDigestConflict(receipt.GuidedUseCaseOperation, operation)
}

func guidedControlAction(action string) bool {
	return action == "pause" || action == "resume" || action == "cancel"
}

func guidedControlOwnsID(control *authoringcore.GuidedControlState, id string) bool {
	if control == nil {
		return false
	}

	if control.PendingOperation != nil && control.PendingOperation.ID == id {
		return true
	}

	for _, receipt := range control.CompletedOperations {
		if receipt.ID == id {
			return true
		}
	}

	return false
}

func guardFreshGuidedControl(session authoringcore.Session, operation authoringcore.GuidedUseCaseOperation) error {
	state := session.Guided
	if session.Revision != operation.ExpectedRevision {
		return workflowstate.ErrConflict
	}

	if err := guardGuidedControlStatus(state, operation.Action); err != nil {
		return err
	}

	if err := guardGuidedControlIdentity(state, operation.ID); err != nil {
		return err
	}

	return guardGuidedControlNativeRevision(state.Work, operation.NativeRevision)
}

func guardGuidedControlIdentity(state *authoringcore.GuidedUseCaseState, id string) error {
	if guidedIdeaOwnsID(state, id) {
		return workflowstate.ErrConflict
	}

	if state.PendingOperation != nil && state.PendingOperation.ID == id {
		return workflowstate.ErrConflict
	}

	for _, completed := range state.CompletedOperations {
		if completed.ID == id {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func guardGuidedControlNativeRevision(work *authoringcore.GuidedWorkState, revision uint64) error {
	hasNative := work != nil && work.Invocation != nil
	if hasNative != (revision != 0) {
		return workflowstate.ErrConflict
	}

	return nil
}

func guardGuidedControlStatus(state *authoringcore.GuidedUseCaseState, action string) error {
	if state.Status == "cancelled" || state.Status == "completed" || state.Status == "failed" {
		return workflowstate.ErrConflict
	}

	if state.Control != nil && len(state.Control.CompletedOperations) >= 128 {
		return workflowstate.ErrConflict
	}

	if action == "resume" && (state.Control == nil || state.Control.StopIntent != "pause") {
		return workflowstate.ErrConflict
	}

	return nil
}
