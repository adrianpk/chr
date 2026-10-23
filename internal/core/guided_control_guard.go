package core

func guardGuidedControl(session AuthoringSession, operation GuidedUseCaseOperation) (bool, error) {
	state := session.Guided

	if !guidedControlAction(operation.Action) {
		return false, ErrWorkflowConflict
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

func replayGuidedControl(receipt GuidedControlReceipt, operation GuidedUseCaseOperation) (bool, error) {
	if receipt.Rejected {
		return false, ErrWorkflowConflict
	}

	return true, guidedDigestConflict(receipt.GuidedUseCaseOperation, operation)
}

func guidedControlAction(action string) bool {
	return action == "pause" || action == "resume" || action == "cancel"
}

func guidedControlOwnsID(control *GuidedControlState, id string) bool {
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

func guardFreshGuidedControl(session AuthoringSession, operation GuidedUseCaseOperation) error {
	state := session.Guided
	if session.Revision != operation.ExpectedRevision {
		return ErrWorkflowConflict
	}

	if err := guardGuidedControlStatus(state, operation.Action); err != nil {
		return err
	}

	if err := guardGuidedControlIdentity(state, operation.ID); err != nil {
		return err
	}

	return guardGuidedControlNativeRevision(state.Work, operation.NativeRevision)
}

func guardGuidedControlIdentity(state *GuidedUseCaseState, id string) error {
	if guidedIdeaOwnsID(state, id) {
		return ErrWorkflowConflict
	}

	if state.PendingOperation != nil && state.PendingOperation.ID == id {
		return ErrWorkflowConflict
	}

	for _, completed := range state.CompletedOperations {
		if completed.ID == id {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func guardGuidedControlNativeRevision(work *GuidedWorkState, revision uint64) error {
	hasNative := work != nil && work.Invocation != nil
	if hasNative != (revision != 0) {
		return ErrWorkflowConflict
	}

	return nil
}

func guardGuidedControlStatus(state *GuidedUseCaseState, action string) error {
	if state.Status == "cancelled" || state.Status == "completed" || state.Status == "failed" {
		return ErrWorkflowConflict
	}

	if state.Control != nil && len(state.Control.CompletedOperations) >= 128 {
		return ErrWorkflowConflict
	}

	if action == "resume" && (state.Control == nil || state.Control.StopIntent != "pause") {
		return ErrWorkflowConflict
	}

	return nil
}
